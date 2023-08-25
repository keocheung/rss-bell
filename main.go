package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"rss-bell/pkg/config"
	"rss-bell/pkg/task"
	"rss-bell/util/logger"

	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

const (
	configPathEnvKey  = "CONFIG_PATH"
	defaultConfigPath = "./config.yaml"
)

func main() {
	c := cron.New()
	c.Start()
	conf, err := loadConfigFromFile()
	if err != nil {
		log.Fatal(err)
	}
	tasks, entries := registerTasks(conf, c)

	// Watch config file for changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	configPath := getConfigPath()
	err = watcher.Add(configPath)
	if err != nil {
		log.Fatal(err)
	}
	logger.Infof("watching config file: %s", configPath)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					logger.Errorf("get config file watcher events error")
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
					logger.Errorf("get config file watcher errors error")
					return
				}
				logger.Warnf("config file watcher error: %s", err)
			}
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)
	<-stop
}

func registerTasks(conf config.Config, c *cron.Cron) (map[string]task.Task, map[string]cron.EntryID) {
	tasks := make(map[string]task.Task)
	entries := make(map[string]cron.EntryID)
	wg := sync.WaitGroup{}
	wg.Add(len(conf.Tasks))
	for tID, tConf := range conf.Tasks {
		go func(tID string, tConf config.Task) {
			t, err := task.NewTask(tID, tConf)
			if err != nil {
				logger.Errorf("NewTask %s error: %v", tID, err)
				return
			}
			tasks[tID] = t
			entryID, err := c.AddJob(tConf.Cron, t)
			if err != nil {
				logger.Errorf("addJob %s error: %v", tID, err)
				return
			}
			entries[tID] = entryID
			logger.Infof("task added, ID: %s, Name: %s", tID, tConf.Name)
			wg.Done()
		}(tID, tConf)
	}
	wg.Wait()
	return tasks, entries
}

func updateTasks(conf config.Config, tasks map[string]task.Task, entries map[string]cron.EntryID,
	c *cron.Cron) (map[string]task.Task, map[string]cron.EntryID) {

	// Remove tasks that doesn't exist anymore
	for tID := range tasks {
		if _, ok := conf.Tasks[tID]; !ok {
			c.Remove(entries[tID])
			delete(tasks, tID)
			delete(entries, tID)
			logger.Infof("task removed: %s", tID)
		}
	}

	for tID, tConf := range conf.Tasks {
		// Update task configs
		if _, ok := tasks[tID]; ok {
			// TODO: handle cron expression changes
			tasks[tID].UpdateConfig(tConf)
			logger.Infof("task updated, ID: %s", tID)
			continue
		}

		// Add new tasks
		logger.Infof("task added, ID: %s, Name: %s", tID, tConf.Name)
		t, err := task.NewTask(tID, tConf)
		if err != nil {
			logger.Errorf("NewTask %s error: %v", tID, err)
			continue
		}
		tasks[tID] = t
		entryID, err := c.AddJob(tConf.Cron, t)
		if err != nil {
			logger.Errorf("AddJob %s error: %v", tID, err)
			continue
		}
		entries[tID] = entryID
	}
	logger.Infof("config reloaded")
	sendAppNotification(conf.AppNotificationURL, "config reloaded")

	return tasks, entries
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
