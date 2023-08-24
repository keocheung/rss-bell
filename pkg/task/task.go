package task

import (
	"fmt"

	"rss-bell/pkg/config"
	"rss-bell/util/http"
	"rss-bell/util/logger"

	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron/v3"
)

type Task interface {
	cron.Job
	UpdateConfig(config config.Task)
}

type taskImpl struct {
	config     config.Task
	httpClient http.HTTPClient
	lastGUID   string
}

func NewTask(config config.Task) (Task, error) {
	t := &taskImpl{
		config:     config,
		httpClient: http.NewHTTPClient(),
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
	return t, nil
}

func (t *taskImpl) Run() {
	logger.Infof("check %s", t.config.Name)
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
	for _, item := range feed.Items {
		if item.GUID == t.lastGUID {
			return
		}
		sender, err := shoutrrr.CreateSender(t.config.NotificationURL)
		if err != nil {
			logger.Errorf("create sender for %s error: %v", t.config.NotificationURL, err)
			return
		}
		params := types.Params(map[string]string{
			"title": feed.Title,
		})
		if err := sender.Send(item.Title, &params); err != nil {
			logger.Errorf("send notification for %s error: %v", t.config.NotificationURL, err)
			return
		}
		logger.Infof("sent notification for %s", t.config.Name)
	}
	t.lastGUID = feed.Items[0].GUID
}

func (t *taskImpl) UpdateConfig(config config.Task) {
	t.config = config
}
