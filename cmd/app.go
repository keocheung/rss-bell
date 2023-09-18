package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"rss-bell/pkg/config"
	"rss-bell/pkg/schedule"
	"rss-bell/pkg/task"
	"rss-bell/pkg/util/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/keocheung/shoutrrr"
	"github.com/keocheung/shoutrrr/pkg/types"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

const (
	configPathEnvKey  = "CONFIG_PATH"
	defaultConfigPath = "./config.yaml"
)

// StartApp starts the main application
func StartApp() error {
	c := cron.New()
	c.Start()
	conf, err := loadConfigFromFile()
	if err != nil {
		return err
	}
	tasks, entries, errIDs := registerTasks(conf, c)
	sendAppNotification(conf.AppNotificationURL, fmt.Sprintf("Tasks added, errs: %+v", errIDs))

	// Watch config file for changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
					logger.Infof("config file changed: %s", event.Name)
					newConf, err := loadConfigFromFile()
					if err != nil {
						logger.Warnf("config file watcher error: %s", err)
						continue
					}
					tasks, entries = updateTasks(newConf, tasks, entries, c)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Warnf("config file watcher error: %s", err)
			}
		}
	}()
	configPath := getConfigPath()
	err = watcher.Add(configPath)
	if err != nil {
		return err
	}
	logger.Infof("watching config file: %s", configPath)
	return nil
}

func registerTasks(conf config.Config, c *cron.Cron) (map[string]task.Task, map[string]cron.EntryID, []string) {
	tasks := make(map[string]task.Task)
	entries := make(map[string]cron.EntryID)
	wg := sync.WaitGroup{}
	wg.Add(len(conf.Tasks))
	var errIDs []string
	for tID, tConf := range conf.Tasks {
		go func(tID string, tConf config.Task) {
			defer wg.Done()
			t, err := task.NewTask(tID, tConf)
			if err != nil {
				logger.Errorf("NewTask %s error: %v", tID, err)
				errIDs = append(errIDs, tID)
				return
			}
			tasks[tID] = t
			schedule, err := schedule.NewRandomDelaySchedule(tConf.Cron, tConf.MaxDelayInSecond)
			if err != nil {
				logger.Errorf("addJob %s error: %v", tID, err)
				errIDs = append(errIDs, tID)
				return
			}
			entryID := c.Schedule(schedule, t)
			entries[tID] = entryID
			logger.Infof("task added, ID: %s, Name: %s", tID, tConf.Name)
		}(tID, tConf)
	}
	wg.Wait()
	return tasks, entries, errIDs
}

func updateTasks(conf config.Config, tasks map[string]task.Task, entries map[string]cron.EntryID,
	c *cron.Cron) (map[string]task.Task, map[string]cron.EntryID) {

	// Remove tasks that doesn't exist anymore
	for tID := range tasks {
		if _, ok := conf.Tasks[tID]; !ok {
			removeTask(tID, tasks, entries, c)
		}
	}

	for tID, tConf := range conf.Tasks {
		// Update task configs
		if _, ok := tasks[tID]; ok {
			if tConf.Cron != tasks[tID].GetConfig().Cron || tConf.MaxDelayInSecond != tasks[tID].GetConfig().MaxDelayInSecond {
				removeTask(tID, tasks, entries, c)
				addTask(tID, tConf, tasks, entries, c)
			} else if tConf != tasks[tID].GetConfig() {
				tasks[tID].UpdateConfig(tConf)
				logger.Infof("task updated: %s", tID)
			}
			continue
		}

		addTask(tID, tConf, tasks, entries, c)
	}
	logger.Infof("config reloaded")
	sendAppNotification(conf.AppNotificationURL, "Config reloaded")

	return tasks, entries
}

func removeTask(tID string, tasks map[string]task.Task, entries map[string]cron.EntryID, c *cron.Cron) {
	c.Remove(entries[tID])
	delete(tasks, tID)
	delete(entries, tID)
	logger.Infof("task removed: %s", tID)
}

func addTask(tID string, tConf config.Task, tasks map[string]task.Task, entries map[string]cron.EntryID, c *cron.Cron) {
	t, err := task.NewTask(tID, tConf)
	if err != nil {
		logger.Errorf("NewTask %s error: %v", tID, err)
		return
	}
	tasks[tID] = t
	entryID, err := c.AddJob(tConf.Cron, t)
	if err != nil {
		logger.Errorf("AddJob %s error: %v", tID, err)
		return
	}
	entries[tID] = entryID
	logger.Infof("task added: %s", tID)
}

func loadConfigFromFile() (config.Config, error) {
	configPath := getConfigPath()
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return config.Config{}, fmt.Errorf("Read from config file error: %v", err)
	}
	var c config.Config
	if strings.HasSuffix(configPath, ".yaml") || strings.HasSuffix(configPath, ".yml") {
		if err := yaml.Unmarshal(configFile, &c); err != nil {
			return config.Config{}, fmt.Errorf("Unmarshal config file error: %v", err)
		}
	} else {
		return c, fmt.Errorf("config file name should end with yaml or yml")
	}
	return c, nil
}

func getConfigPath() string {
	configPath := os.Getenv(configPathEnvKey)
	if configPath == "" {
		configPath = defaultConfigPath
	}
	return configPath
}

func sendAppNotification(url, message string) {
	sender, err := shoutrrr.CreateSender(url)
	if err != nil {
		logger.Errorf("create sender for %s error: %v", url, err)
		return
	}
	params := types.Params(map[string]string{
		"title": "RSS Bell",
	})
	errs := sender.Send(message, &params)
	logger.Infof("sent app notification, errs: %+v", errs)
}
