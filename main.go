package main

import (
	"os"
	"os/signal"
	"syscall"

	"rss-bell/cmd"
	"rss-bell/internal/meta"
	"rss-bell/pkg/util/logger"
)

func main() {
	logger.Infof("rss-bell %s", meta.Version)

	cmd.StartApp()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)
	<-stop
}
