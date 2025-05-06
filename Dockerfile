FROM golang:alpine as builder
RUN apk add git
ADD . /go/src/rss-bell
WORKDIR /go/src/rss-bell
RUN go build -ldflags="-s -w -X rss-bell/internal/meta.Version=$(git describe --tags)"

FROM alpine
COPY --from=builder /go/src/rss-bell/rss-bell /app/rss-bell
ENTRYPOINT ["/app/rss-bell"]