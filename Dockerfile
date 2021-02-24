FROM golang:1.13-alpine

RUN apk add --no-cache git && \
	go get -v github.com/debdutdeb/upptime-email-notify

EXPOSE 8080

ENTRYPOINT ["upptime-email-notify"]
