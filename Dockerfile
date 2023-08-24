FROM golang:1.20.2-alpine3.17 as builder
ADD . /go/src/rss-bell
WORKDIR /go/src/rss-bell
RUN go install -ldflags="-s -w"

FROM alpine:3.17
COPY --from=builder /go/bin/rss-bell /app/rss-bell
ENTRYPOINT ["/app/rss-bell"]