<h1 id="m7s"> m7s-jt1078 </h1>

- [m7s官方地址](https://monibuca.com)

---

| 例子 |  测试页面  | 代码 |
|----------|-----|-------------------|
| 音视频 | http://124.221.30.46:11000 | [详情点击](./example/video) |
| 对讲 | https://go-jt808.online:12000 | [详情点击](./example/intercom)  |
| 模拟流 | 视频 http://124.221.30.46:8088/preview/live/jt1078-295696659617-1?type=mp4 <br/> 音视频 http://124.221.30.46:8088/preview/live/jt1079-156987000796-1| [详情点击](./example/simulation)  |

---

<h2>1. 仅运行当jt1078使用</h2>

```
启动 (win系统的话 双击exe)
cd ./jt1078 && ./jt1078
默认查看流页面
http://127.0.0.1:12079/preview
https://127.0.0.1:12080/preview
默认模拟流地址 默认只使用flv和mp4的插件
需要增加其他格式的话 在代码里面初始化修改
实时
http://127.0.0.1:12079/mp4/live/jt1078-295696659617-1.mp4
https://127.0.0.1:12080/mp4/live/jt1078-295696659617-1.mp4
回放
http://127.0.0.1:12079/flv/live/jt1079-156987000796-1.flv
https://127.0.0.1:12080/flv/live/jt1079-156987000796-1.flv
```

默认配置文件如下 可自行修改
```
实时视频 12051 回放视频 12052
m7s http端口 12079 https端口12080 tcp端口12081
webrtc对讲 外网ip 124.221.30.46 外网udp端口 12020
对讲让设备连接的端口范围 [12021-12050]
```


<h2>2. 模拟流</h2>

```
启动 (win系统的话 双击exe)
cd ./simulation && ./simulation
"模拟实时视频流地址": "http://127.0.0.1:8080/preview/live/jt1078-295696659617-1?type=mp4"
"模拟回放音视频流地址(音频G711A)": "http://127.0.0.7:8080/preview/live/jt1079-156987000796-1"
```

<h2>3. 音视频</h2>

```
启动 (win系统的话 双击exe)
cd ./video && ./video
默认首页 http://127.0.0.1:11000
已经部署的在线网页 http://124.221.30.46:11000/
```

- 需要设备连到jt808服务 并且下发9101指令
- 内置的jt808服务监听的端口是11001

``` curl
使用内置jt808下发9101指令如下 key是sim卡号 用自己的测试设备
POST http://127.0.0.1:11000/api/v1/jt808/9101
Content-Type: application/json

{
  "key": "10088",
  "data": {
    "serverIPLen": 13,
    "serverIPAddr": "124.221.30.46",
    "tcpPort": 11051,
    "udpPort": 0,
    "channelNo": 1,
    "dataType": 0,
    "streamType": 0
  }
}
```

```
触发9101指令后 在如下地址可以看到流存在情况
http://127.0.0.1:11080/preview/
播放地址如 http://124.221.30.46:11080/flv/live/jt1078-10086-1.flv
```

<h2>4. 对讲</h2>

```
需要https 因为浏览器调用设备需要https
启动 (win系统的话 双击exe)
cd ./intercom && ./intercom
默认首页 https://127.0.0.1:12000
已经部署的在线网页 https://124.221.30.46:12000/
```

详情参考 https://github.com/cuteLittleDevil/m7s-jt1078#对讲流程参考

<h2> 配置说明 </h2>

``` yaml
jt1078:
  enable: true # 是否启用

  intercom:
    enable: true # 是否启用 用于双向对讲
    jt1078webrtc:
      port: 12020 # 外网UDP端口 用于浏览器webrtc把音频数据推到这个端口
      ip: 124.221.30.46 # 外网ip 用于SDP协商修改
    audioports: [12021, 12050] # 音频端口 [min,max]
    onjoinurl: "https://127.0.0.1:12000/api/v1/jt808/event/join-audio" # 设备连接到音频端口的回调
    onleaveurl: "https://127.0.0.1:12000/api/v1/jt808/event/leave-audio" # 设备断开了音频端口的回调
    overtimesecond: 60 # 多久没有下发对讲语音的数据 就关闭这个链接

  realtime: # 实时视频
    addr: '0.0.0.0:12051'
    onjoinurl: "https://127.0.0.1:12000/api/v1/jt808/event/real-time-join" # 设备连接到了实时视频指定端口的回调
    onleaveurl: "https://127.0.0.1:12000/api/v1/jt808/event/real-time-leave" # 设备断开了实时视频指定端口的回调
    prefix: "live/jt1078" # 默认自定义前缀-手机号-通道 如：live/jt1078-295696659617-1
    overtime_second: 0 # 无人订阅的情况 多久就关闭这个链接（小于等于0则不启用 默认0 推荐还是使用9102指令去触发关闭)

  playback: # 回放视频
    addr: '0.0.0.0:12052'
    onjoinurl: "https://127.0.0.1:12000/api/v1/play-back-join" # 设备连接到了回放视频指定端口的回调
    onleaveurl: "https://127.0.0.1:12000/api/v1/play-back-leave" # 设备断开了回放视频指定端口的回调
    prefix: "live/jt1079" # 默认自定义前缀-手机号-通道 如：live/jt1079-295696659617-1
    overtime_second: 0 # 无人订阅的情况 多久就关闭这个链接（小于等于0则不启用 默认0 推荐还是使用9102指令去触发关闭)

  simulations:
    # jt1078文件 默认循环发送
      - name: ../testdata/data.txt
        addr: 127.0.0.1:12051 # 模拟实时
      - name: ../testdata/audio_data.txt
        addr: 127.0.0.1:12052 # 模拟回放

```