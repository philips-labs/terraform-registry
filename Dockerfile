FROM golang:1.24.3 as builder
WORKDIR /build
COPY . /build/

RUN go mod download
RUN CGO_ENABLED=0 go build -o app

FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /build/app /app
EXPOSE 8080
CMD ["/app/app"]
