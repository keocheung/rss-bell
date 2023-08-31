package task

import (
	"fmt"
	"sort"
	"time"

	"rss-bell/pkg/config"
	"rss-bell/pkg/util/http"
	"rss-bell/pkg/util/logger"

	"github.com/keocheung/shoutrrr"
	"github.com/keocheung/shoutrrr/pkg/types"
	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron/v3"
)

// Task is a task that checks RSS feed and sends notifications if the feed has updates
type Task interface {
	// GetConfig returns the task config
	GetConfig() config.Task
	// UpdateConfig updates the task config
	UpdateConfig(config config.Task)
	cron.Job
}

type taskImpl struct {
	id            string
	Config        config.Task
	httpClient    http.Client
	lastGUID      string
	lastPublished time.Time
}

// NewTask creates a new task with task ID and task config
func NewTask(id string, config config.Task) (Task, error) {
	var client http.Client
	if config.Proxy != "" {
		client = http.NewClientWithProxy(config.Proxy)
	} else {
		client = http.NewClient()
	}
	t := &taskImpl{
		id:         id,
		Config:     config,
		httpClient: client,
	}
	feed, err := t.getFeed()
	if err != nil {
		return nil, err
	}
	t.lastGUID = feed.Items[0].GUID
	t.lastPublished = *feed.Items[0].PublishedParsed
	return t, nil
}

func (t *taskImpl) Run() {
	logger.Infof("checking %s", t.id)
	feed, err := t.getFeed()
	if err != nil {
		logger.Errorf(err.Error())
	}
	if len(feed.Items) == 0 {
		logger.Infof("no feed items for %s", t.id)
		return
	}
	var items []*gofeed.Item
	for _, item := range feed.Items {
		if t.itemIsOld(item) {
			break
		}
		items = append(items, item)
	}
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		sender, err := shoutrrr.CreateSender(t.Config.NotificationURL)
		if err != nil {
			logger.Errorf("create sender for %s error: %v", t.Config.NotificationURL, err)
			continue
		}
		title := t.Config.Name
		if title == "" {
			title = feed.Title
		}
		params := types.Params(map[string]string{
			"title": title,
			"url":   item.Link,
		})
		errs := sender.Send(item.Title, &params)
		logger.Infof("sent notification for %s, errs: %+v", t.id, errs)
	}
	t.lastGUID = feed.Items[0].GUID
	t.lastPublished = *feed.Items[0].PublishedParsed
}

func (t *taskImpl) GetConfig() config.Task {
	return t.Config
}

func (t *taskImpl) UpdateConfig(config config.Task) {
	t.Config = config
}

func (t *taskImpl) getFeed() (*gofeed.Feed, error) {
	data, err := t.httpClient.Get(t.Config.FeedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("get %s error: %v", t.Config.FeedURL, err)
	}
	feed, err := gofeed.NewParser().ParseString(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse %s error: %v", t.Config.FeedURL, err)
	}
	// Disable sorting if any feed item does not have a publish time
	canSort := true
	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			canSort = false
			break
		}
	}
	if canSort {
		sort.Sort(sort.Reverse(feed))
	}
	return feed, nil
}

func (t *taskImpl) itemIsOld(item *gofeed.Item) bool {
	if item.PublishedParsed != nil {
		return item.PublishedParsed.Add(-1).Before(t.lastPublished)
	}
	return item.GUID == t.lastGUID
}
