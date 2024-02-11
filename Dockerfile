FROM golang:1.22

WORKDIR /app/

COPY go.mod go.sum ./
RUN go mod download 

COPY ./ /app/
RUN go build

EXPOSE 8080
ENTRYPOINT ["/app/graphcache-go"]
