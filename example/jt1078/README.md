<h2> 默认页面 </h2>

```
默认首页 目前ip换101.35.2.3可以访问
http://127.0.0.1:12079
https://127.0.0.1:12080
默认查看流页面
http://127.0.0.1:12079/preview
https://127.0.0.1:12080/preview
默认模拟流地址 默认使用如下格式flv mp4 webrtc
需要增加其他格式的话 在代码里面初始化修改 格式如下
实时
http://127.0.0.1:12079/mp4/live/jt1078-295696659617-1.mp4
https://127.0.0.1:12080/mp4/live/jt1078-295696659617-1.mp4
webrtc://127.0.0.1:12080/webrtc/play/live/jt1078-295696659617-1
回放
http://127.0.0.1:12079/flv/live/jt1079-156987000796-1.flv
https://127.0.0.1:12080/flv/live/jt1079-156987000796-1.flv
webrtc://127.0.0.1:12080/webrtc/play/live/jt1079-156987000796-1
```

- docker拉取镜像

```
docker pull cdcddcdc/m7s-jt1078:latest
```

<h2> 音视频启动 </h2>

- HTTP服务: 12079
- 实时视频流: 12051
- 回放视频流: 12052

```
docker run -it -d \
-v /home/m7s-jt1078/config.yaml:/app/config.yaml \
--network host \
cdcddcdc/m7s-jt1078:latest
```

<h2> 增加对讲功能 </h2>

- HTTP服务: 12079
- 实时视频流: 12051
- 回放视频流: 12052
- HTTPS服务: 12080
- Webrtc外网ip: 124.221.30.46
- Webrtc外网UDP端口: 12020
- 音频端口组: [12021-12050]

```
docker run -it -d \
-v /home/m7s-jt1078/go-jt808.online.crt:/app/go-jt808.online.crt \
-v /home/m7s-jt1078/go-jt808.online.key:/app/go-jt808.online.key \
-v /home/m7s-jt1078/config.yaml:/app/config.yaml \
--network host \
cdcddcdc/m7s-jt1078:latest
```
---

<h2> 配置说明 </h2>

``` yaml
jt1078:
  enable: true # 是否启用

  intercom:
    enable: true # 是否启用 用于双向对讲
    jt1078webrtc:
      port: 12020 # 外网UDP端口 用于浏览器webrtc把音频数据推到这个端口
      ip: 124.221.30.46 # 外网ip 用于SDP协商修改
    audio_ports: [12021, 12050] # 音频端口 [min,max]
    on_join_url: "https://127.0.0.1:12000/api/v1/jt808/event/join-audio" # 设备连接到音频端口的回调
    on_leave_url: "https://127.0.0.1:12000/api/v1/jt808/event/leave-audio" # 设备断开了音频端口的回调
    overtime_second: 60 # 多久没有下发对讲语音的数据 就关闭这个链接

  realtime: # 实时视频
    addr: '0.0.0.0:12051'
    on_join_url: "https://127.0.0.1:12000/api/v1/jt808/event/real-time-join" # 设备连接到了实时视频指定端口的回调
    on_leave_url: "https://127.0.0.1:12000/api/v1/jt808/event/real-time-leave" # 设备断开了实时视频指定端口的回调
    prefix: "live/jt1078" # 默认自定义前缀-手机号-通道 如：live/jt1078-295696659617-1
    overtime_second: 0 # 无人订阅的情况 多久就关闭这个链接（小于等于0则不启用 默认0 推荐还是使用9102指令去触发关闭)

  playback: # 回放视频
    addr: '0.0.0.0:12052'
    on_join_url: "https://127.0.0.1:12000/api/v1/play-back-join" # 设备连接到了回放视频指定端口的回调
    on_leave_url: "https://127.0.0.1:12000/api/v1/play-back-leave" # 设备断开了回放视频指定端口的回调
    prefix: "live/jt1079" # 默认自定义前缀-手机号-通道 如：live/jt1079-295696659617-1
    overtime_second: 0 # 无人订阅的情况 多久就关闭这个链接（小于等于0则不启用 默认0 推荐还是使用9102指令去触发关闭)

  simulations:
    # jt1078文件 默认循环发送
      - name: ./data.txt
        addr: 127.0.0.1:12051 # 模拟实时
      - name: ./audio_data.txt
        addr: 127.0.0.1:12052 # 模拟回放

```