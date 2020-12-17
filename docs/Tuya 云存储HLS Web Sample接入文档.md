# Tuya 云存储HLS Web Sample接入文档

![Tuya 云存储HLS Web Sample业务流程图](./openapi_storage_hls.png)

## 模块组成
### Web前端
* 提供用于Chrome访问云存储资源的页面

### Web后端
* 托管Web页面
* 访问涂鸦云，通过HTTP协议获取需要的各种配置信息
* 访问涂鸦云，通过HTTP协议请求云存储相关信息

### 涂鸦云
* 提供开放平台各种HTTP接口

### 涂鸦Stream
* 涂鸦流媒体服务


## Step By Step
1. 注册[Tuya开放平台](https://docs.tuya.com/zh/iot/open-api/quick-start/quick-start1)，获取`clientId`和`secret`

2. 更新Sample webrtc.json中的`clientId`和`secret`

3. 认证模式分为easy简单模式和auth授权码模式
    * easy，填写uId到webrtc.json
    * auth，访问[Tuya开放平台授权](https://openapi.tuyacn.com/selectAuth?client_id={clientId}&redirect_uri=https://www.example.com/auth&state=1234)，输入涂鸦账号和密码，同意授权，截取浏览器回调URL中的授权码`code`，填写到webrtc.json

4. 涂鸦智能APP中选中一台IPC，查询设备ID，更新到Sample webrtc.json的`deviceId`

5. 在Sample源码路径，执行`go get`后执行`go build`

6. 运行`./webrtc-demo-go`

### 获取时间轴
1. Chrome打开`http://localhost:3333/api/cloud/timeline?timeGT=1607939286&timeLT=1607945715`

### 获取HLS播放资源列表
1. Chrome打开`http://localhost:3333/api/cloud/hls?timeGT=1607939286&timeLT=1607945715`

### 播放HLS资源
1. VLC打开`https://wework1.wgine.com:554/cloudrecord/6c763ce20a233d7da4qto7/bvbks0p525qb1mf6n0t0XJq6M1CNa4Fw.m3u8`

## VLC下载
### Windows
* `https://get.videolan.org/vlc/3.0.11/win64/vlc-3.0.11-win64.exe`
### Mac
* `https://get.videolan.org/vlc/3.0.11.1/macosx/vlc-3.0.11.1.dmg`