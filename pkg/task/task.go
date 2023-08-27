package task

import (
	"fmt"
	"time"

	"rss-bell/pkg/config"
	"rss-bell/pkg/util/http"
	"rss-bell/pkg/util/logger"

	"github.com/keocheung/shoutrrr"
	"github.com/keocheung/shoutrrr/pkg/types"
	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron/v3"
)

type Task interface {
	cron.Job
	UpdateConfig(config config.Task)
}

type taskImpl struct {
	id            string
	config        config.Task
	httpClient    http.Client
	lastGUID      string
	lastPublished time.Time
}

func NewTask(id string, config config.Task) (Task, error) {
	var client http.Client
	if config.Proxy != "" {
		client = http.NewClientWithProxy(config.Proxy)
	} else {
		client = http.NewClient()
	}
	t := &taskImpl{
		id:         id,
		config:     config,
		httpClient: client,
	}
	data, err := t.httpClient.Get(t.config.FeedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("get %s error: %v", t.config.FeedURL, err)
	}
	feed, err := gofeed.NewParser().ParseString(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse %s error: %v", t.config.FeedURL, err)
	}
	t.lastGUID = feed.Items[0].GUID
	t.lastPublished = *feed.Items[0].PublishedParsed
	return t, nil
}

func (t *taskImpl) Run() {
	logger.Infof("checking %s", t.id)
	data, err := t.httpClient.Get(t.config.FeedURL, nil)
	if err != nil {
		logger.Errorf("get %s error: %v", t.config.FeedURL, err)
		return
	}
	feed, err := gofeed.NewParser().ParseString(string(data))
	if err != nil {
		logger.Errorf("parse %s error: %v", t.config.FeedURL, err)
		return
	}
	if len(feed.Items) == 0 {
		logger.Infof("no feed items for %s", t.id)
		return
	}
	for _, item := range feed.Items {
		if t.itemIsOld(item) {
			break
		}
		sender, err := shoutrrr.CreateSender(t.config.NotificationURL)
		if err != nil {
			logger.Errorf("create sender for %s error: %v", t.config.NotificationURL, err)
			continue
		}
		title := t.config.Name
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

func (t *taskImpl) UpdateConfig(config config.Task) {
	t.config = config
}

func (t *taskImpl) itemIsOld(item *gofeed.Item) bool {
	if item.PublishedParsed != nil {
		return item.PublishedParsed.Add(-1).Before(t.lastPublished)
	}
	return item.GUID == t.lastGUID
}
