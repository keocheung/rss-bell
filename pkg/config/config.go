// Package config contains app configs
package config

type Config struct {
	AppNotificationURL string          `yaml:"app_notification_url"`
	Tasks              map[string]Task `yaml:"tasks"`
}

type Task struct {
	Name             string `yaml:"name"`
	FeedURL          string `yaml:"feed_url"`
	Cron             string `yaml:"cron"`
	Proxy            string `yaml:"proxy"`
	NotificationURL  string `yaml:"notification_url"`
	MaxDelayInSecond int32  `yaml:"max_delay_in_second"`
}
