FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
# 不用数据库也可以跑 只是默认页面会报错 https://github.com/cuteLittleDevil/m7s-jt1078/issues/4
RUN cd ./example/jt1078 && go build -tags sqlite -o jt1078

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
# 12079是http页面 12051是jt1078的实时音视频 12052是jt1078的历史音视频
EXPOSE 12079 12051 12052
CMD ["./jt1078"]