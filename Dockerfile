FROM golang:1.20.2-alpine3.17 as builder
RUN apk add git
ADD . /go/src/rss-bell
WORKDIR /go/src/rss-bell
RUN go install -ldflags="-s -w -X rss-bell/internal/meta.Version=$(git describe --tags)"

FROM alpine:3.17
COPY --from=builder /go/bin/rss-bell /app/rss-bell
ENTRYPOINT ["/app/rss-bell"]