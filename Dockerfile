FROM golang:1.22

WORKDIR /app/

COPY go.mod go.sum ./
RUN go mod download

COPY ./ /app/
RUN go build -tags=go_json
RUN go build ./cmd/startupCacher

EXPOSE 8080
ENTRYPOINT ["/app/graphcache-go"]
