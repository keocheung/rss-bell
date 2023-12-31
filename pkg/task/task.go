package task

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"rss-bell/pkg/config"
	"rss-bell/pkg/dlwebhook"
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
	t.initLastPublished()
	return t, nil
}

func (t *taskImpl) Run() {
	logger.Infof("checking %s", t.id)
	feed, err := t.getFeed()
	if err != nil {
		logger.Errorf(err.Error())
		return
	}
	if feed == nil || len(feed.Items) == 0 {
		logger.Infof("no feed item for %s", t.id)
		return
	}
	var items []*gofeed.Item
	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			continue
		}
		if t.itemIsOld(item) {
			break
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return
	}
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		if t.Config.NotificationURL != "" {
			t.sendNotification(feed, item)
		}
		if t.Config.DownloadWebhook.APIURL != "" {
			t.triggerDownloadWebhook(item)
		}
	}
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

func (t *taskImpl) initLastPublished() {
	feed, err := t.getFeed()
	if err != nil {
		logger.Errorf(err.Error())
		t.lastPublished = time.Now()
		return
	}
	if feed == nil || len(feed.Items) == 0 || feed.Items[0].PublishedParsed == nil {
		t.lastPublished = time.Now()
		return
	}
	t.lastPublished = *feed.Items[0].PublishedParsed
}

func (t *taskImpl) itemIsOld(item *gofeed.Item) bool {
	return item.PublishedParsed.Add(-1).Before(t.lastPublished)
}

func (t *taskImpl) sendNotification(feed *gofeed.Feed, item *gofeed.Item) {
	sender, err := shoutrrr.CreateSender(t.Config.NotificationURL)
	if err != nil {
		logger.Errorf("create sender for %s error: %v", t.Config.NotificationURL, err)
		return
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

func (t *taskImpl) triggerDownloadWebhook(item *gofeed.Item) {
	client := http.NewClient()
	data := dlwebhook.Data{
		URL:          item.Link,
		Secret:       t.Config.DownloadWebhook.Secret,
		Engine:       t.Config.DownloadWebhook.Engine,
		Path:         t.Config.DownloadWebhook.Path,
		Name:         t.Config.DownloadWebhook.Name,
		ExtraOptions: t.Config.DownloadWebhook.ExtraOptions,
	}
	b, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("triggerDownloadWebhook marshal error: %v", err)
		return
	}
	rsp, err := client.Post(t.Config.DownloadWebhook.APIURL, b, map[string]string{
		"Content-Type": "application/json",
	})
	if err != nil {
		logger.Errorf("triggerDownloadWebhook post error: %v", err)
		return
	}
	logger.Infof("triggered download webhook for %s, rsp: %s", t.id, rsp)
}
