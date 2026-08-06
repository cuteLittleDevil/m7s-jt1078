## 默认页面

| 页面 | HTTP | HTTPS |
|------|------|-------|
| 首页 | `http://127.0.0.1:12079` | `https://127.0.0.1:12080` |
| 查看流 | `http://127.0.0.1:12079/preview` | `https://127.0.0.1:12080/preview` |

### 模拟流地址

支持 `flv` / `mp4` / `webrtc` / `rtsp`。需要其他格式时，在代码里初始化修改。

**实时** `live/jt1078-295696659617-1`

| 格式 | 地址 |
|------|------|
| mp4 | `http://127.0.0.1:12079/mp4/live/jt1078-295696659617-1.mp4` |
| mp4 (https) | `https://127.0.0.1:12080/mp4/live/jt1078-295696659617-1.mp4` |
| flv | `http://127.0.0.1:12079/flv/live/jt1078-295696659617-1.flv` |
| flv (https) | `https://127.0.0.1:12080/flv/live/jt1078-295696659617-1.flv` |
| webrtc | `webrtc://127.0.0.1:12080/webrtc/play/live/jt1078-295696659617-1` |
| rtsp | `rtsp://127.0.0.1:8554/live/jt1078-295696659617-1` |

**回放** `live/jt1079-156987000796-1`

| 格式 | 地址 |
|------|------|
| flv | `http://127.0.0.1:12079/flv/live/jt1079-156987000796-1.flv` |
| flv (https) | `https://127.0.0.1:12080/flv/live/jt1079-156987000796-1.flv` |
| mp4 | `http://127.0.0.1:12079/mp4/live/jt1079-156987000796-1.mp4` |
| mp4 (https) | `https://127.0.0.1:12080/mp4/live/jt1079-156987000796-1.mp4` |
| webrtc | `webrtc://127.0.0.1:12080/webrtc/play/live/jt1079-156987000796-1` |
| rtsp | `rtsp://127.0.0.1:8554/live/jt1079-156987000796-1` |

## Docker

- 拉取镜像

```
docker pull cdcddcdc/m7s-jt1078:latest
```

### 音视频启动

| 服务 | 端口 |
|------|------|
| HTTP | 12079 |
| 实时视频流 | 12051 |
| 回放视频流 | 12052 |
| RTSP | 8554 |

```
docker run -it -d \
-v /home/m7s-jt1078/config.yaml:/app/config.yaml \
--network host \
cdcddcdc/m7s-jt1078:latest
```

### 增加对讲功能

| 服务 | 端口 / 说明 |
|------|-------------|
| HTTP | 12079 |
| HTTPS | 12080 |
| 实时视频流 | 12051 |
| 回放视频流 | 12052 |
| RTSP | 8554 |
| Webrtc 外网 IP | 124.221.30.46 |
| Webrtc 外网 UDP | 12020 |
| 音频端口组 | 12021-12050 |

```
docker run -it -d \
-v /home/m7s-jt1078/go-jt808.online.crt:/app/go-jt808.online.crt \
-v /home/m7s-jt1078/go-jt808.online.key:/app/go-jt808.online.key \
-v /home/m7s-jt1078/config.yaml:/app/config.yaml \
--network host \
cdcddcdc/m7s-jt1078:latest
```

---

## 配置说明

```yaml
global:
  publish:
    publish_timeout: 30s  # 将无数据超时改为30秒
    idle_timeout: 60s     # 空闲(无订阅)超时
  db: # 这个 jt1078 不用 db 也可以，只是默认页面会报错
    db_type: sqlite
    dsn: m7s.db
  http:
    listen_addr: ":12079"
#    listen_addr_tls: ":12080" # HTTPS 访问 API
#    cert_file: "go-jt808.online.crt"
#    key_file: "go-jt808.online.key"
  tcp:
    listen_addr: ":12081"

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
    overtime_second: 60 # 多久没有下发对讲语音数据就关闭这个链接

  realtime: # 实时视频
    addr: '0.0.0.0:12051'
    on_join_url: "https://127.0.0.1:12000/api/v1/jt808/event/real-time-join" # 设备连接到实时视频端口的回调
    on_leave_url: "https://127.0.0.1:12000/api/v1/jt808/event/real-time-leave" # 设备断开实时视频端口的回调
    prefix: "live/jt1078" # 默认自定义前缀-手机号-通道，如：live/jt1078-295696659617-1
    overtime_second: 0 # 无人订阅时多久关闭链接（<=0 不启用，默认 0，推荐用 9102 指令关闭）
    debug: # 有流时保存一段时间的流文件，文件名为流名称-debug.txt
      enable: false
      dir: "./save/"
      save_time_second: 30

  playback: # 回放视频
    addr: '0.0.0.0:12052'
    on_join_url: "https://127.0.0.1:12000/api/v1/play-back-join"
    on_leave_url: "https://127.0.0.1:12000/api/v1/play-back-leave"
    prefix: "live/jt1079" # 默认自定义前缀-手机号-通道，如：live/jt1079-295696659617-1
    overtime_second: 0
    debug:
      enable: false
      dir: "./save/"
      save_time_second: 30

  simulations:
    # jt1078 文件，默认循环发送
    - name: ./data.txt
      addr: 127.0.0.1:12051 # 模拟实时
    - name: ./audio_data.txt
      addr: 127.0.0.1:12052 # 模拟回放

mp4:
  enable: true

rtsp:
  enable: true
  tcp:
    listen_addr: ":8554"
```
