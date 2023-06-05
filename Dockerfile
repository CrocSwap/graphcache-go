FROM golang:1.20

WORKDIR /app/
COPY ./ /app/

RUN go build

EXPOSE 8080

ENTRYPOINT ["/app/graphcache-go"]
