FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
# 不用数据库也可以跑 只是默认页面会报错 https://github.com/cuteLittleDevil/m7s-jt1078/issues/4
RUN cd ./example/jt1078 &&   go build -tags sqlite -o jt1078

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates && update-ca-certificates
# github action中curl下载的
COPY --from=builder /app/example/jt1078/admin.zip /tmp/
RUN if [ -f /tmp/admin.zip ]; then cp /tmp/admin.zip . && rm -rf /tmp/*; fi
COPY --from=builder /app/example/testdata/data.txt .
COPY --from=builder /app/example/testdata/audio_data.txt .
COPY --from=builder /app/example/jt1078/jt1078 .
COPY --from=builder /app/example/jt1078/docker_config.yaml ./config.yaml
EXPOSE 12079 12081 12051 12052
CMD ["./jt1078"]