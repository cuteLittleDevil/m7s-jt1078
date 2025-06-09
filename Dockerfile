FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN cd ./example/jt1078 &&  go build -o jt1078

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates && update-ca-certificates
# github action中curl下载的
COPY --from=builder /app/example/jt1078/ /tmp/
RUN if [ -f /tmp/admin.zip ]; then cp /tmp/admin.zip .; fi

COPY --from=builder /app/example/testdata/data.txt .
COPY --from=builder /app/example/testdata/audio_data.txt .
COPY --from=builder /app/example/jt1078/jt1078 .
COPY --from=builder /app/example/jt1078/docker_config.yaml ./config.yaml
EXPOSE 12079 12081 12051 12052
CMD ["./jt1078"]