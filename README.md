# RSS-Bell
Send notifications when RSS feeds have new items

## Install
### Docker Compose
```yaml
version: '3'
services:
  rss-bell:
    image: ghcr.io/keocheung/rss-bell
    container_name: rss-bell
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - ./config.yaml:/config/config.yaml    # YAML config file location
    network_mode: bridge
    environment:
      - CONFIG_PATH=/config/config.yaml      # Optional. Default is /config/config.yaml
      - HTTP_PROXY=http://127.0.0.1:9090     # Optional. Default is empty
      - HTTPS_PROXY=http://127.0.0.1:9090    # Optional. Default is empty
      - NO_PROXY=example.com,192.168.0.0/16  # Optional. Default is empty
    restart: unless-stopped
```

## Usage
**RSS-Bell** uses [Shoutrrr](https://github.com/containrrr/shoutrrr) as notification library. Please refer [Shoutrrr Docs](https://containrrr.dev/shoutrrr/v0.8/) for more details.
### Config file example
```yaml
app_notification_url: bark://:key@api.day.app # Shoutrrr URL for rss-bell itself. Please refer https://containrrr.dev/shoutrrr/v0.8/
tasks:
  "RSSHub New Routes":
    name: RSSHub New Routes
    feed_url: https://rsshub.app/rsshub/routes
    cron: '*/30 * * * *' # For more supported expression, please refer https://pkg.go.dev/github.com/robfig/cron
    notification_url: bark://:key@api.day.app # Shoutrrr URL for feed items. Please refer https://containrrr.dev/shoutrrr/v0.8/
```