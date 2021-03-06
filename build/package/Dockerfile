# build stage
FROM golang:1.14.4-alpine3.11 AS builder
RUN apk update \
    && apk add --no-cache \
        git \
        ca-certificates \
        tzdata \
        protobuf-dev \
    && update-ca-certificates
RUN adduser -D -u 10001 appuser
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
RUN go get github.com/gogo/protobuf/protoc-gen-gogofast
COPY . .
RUN go generate ./... && \
    export BUILD_TIME=`cat .build_time` && \
    cd cmd/ws-test && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s -X main.buildTime=${BUILD_TIME}" -o ws-test

# final stage
FROM alpine:3.11.6
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
USER appuser
WORKDIR /app
ENTRYPOINT ["./ws-test"]
CMD ["--help"]
COPY --from=builder /workspace/cmd/ws-test/ws-test ./ws-test