# rss-bell
### Send notification when RSS feeds have new items

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