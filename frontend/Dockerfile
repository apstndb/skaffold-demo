FROM golang:1.9.4-alpine3.7

WORKDIR /go/src/github.com/GoogleCloudPlatform/skaffold/examples/frontend
CMD ["./app"]
COPY main.go .
COPY vendor ./vendor
RUN go build -o app main.go
