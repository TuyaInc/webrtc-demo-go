'use strict'

const callButton = document.getElementById('callButton')
const hangupButton = document.getElementById('hangupButton')

callButton.addEventListener('click', fetchWebRTCConfigs)
hangupButton.addEventListener('click', sendDisconnect)

// Web端WebSocket连接对象
let ws = null

// webRTC的启动时间
let startTime

let iceServers    // ICE服务器地址列表
let configuration // webRTC配置
let pc            // PeerConnection对象，处理webRTC业务
let localStream   // 本地音视频流
let gAgentId      // 浏览器网页agent id，作为客户端唯一标识符
let gSessionId    // 每次点击Call生成的新webRTC会话id

// 网页加载后连接WebSocket服务
window.onload = connectWS()

const remoteVideo = document.getElementById('remoteVideo')
const remoteAudio = document.getElementById('remoteAudio')

remoteVideo.addEventListener('loadedmetadata', function () {
    console.log(`Remote video videoWidth: ${this.videoWidth}px,  videoHeight: ${this.videoHeight}px`)
})

remoteVideo.addEventListener('resize', () => {
    console.log(`Remote video size changed to ${remoteVideo.videoWidth}x${remoteVideo.videoHeight}`)
    // We'll use the first onsize callback as an indication that video has started playing out
    if (startTime) {
        const elapsedTime = window.performance.now() - startTime
        console.log('Setup time: ' + elapsedTime.toFixed(3) + 'ms')
        startTime = null
    }
})

const offerOptions = {
    offerToReceiveAudio: 1,
    offerToReceiveVideo: 1
}

function uuid()
{
    return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, c =>
        (c ^ (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))).toString(16)
    )
}

// 连接WebSocket服务
async function connectWS() {
    if (ws == null) {
        ws = new WebSocket("ws://127.0.0.1:5555/webrtc", "json")
        console.log('***CREATED WEBSOCKET')
    }

    // 生成网页对应全局唯一agent标识符
    gAgentId = uuid()

    ws.onopen = function (evt) {
        console.log('***ONOPEN')
    }

    // 注册WebSocket的消息回调处理函数
    ws.onmessage = function (evt) {
        console.log('***ONMESSAGE')
        console.log(evt.data)
        console.log(JSON.parse(evt.data))

        parseResponse(JSON.parse(evt.data))
    }
}

// 发送数据到WebSocket服务
function sendWS(type, payload) {
    console.log('***SEND')

    var data = {}

    data["agentId"] = gAgentId
    data["sessionId"] = gSessionId
    data["type"] = type
    data["payload"] = payload

    console.log(JSON.stringify(data))
    ws.send(JSON.stringify(data))
}

// 关闭到WebSocket服务的连接
function closeWS() {
    console.log('***CLOSE')
    ws.close()
}

// 每次点击Call时都拉取最新的webrtc configs，并生成新会话的id
async function fetchWebRTCConfigs() {
    console.log("fetch webrtc configs")

    gSessionId = uuid()

    sendFetchWebRTCConfigs()
}

// 启动webRTC业务
async function call() {
    console.log('call')

    startTime = window.performance.now()

    configuration = {
        "iceServers": iceServers
    }

    console.log('RTCPeerConnection configuration:', configuration)

    pc = new RTCPeerConnection(configuration)
    console.log('Created remote peer connection object pc')

    pc.addEventListener('icecandidate', e => onIceCandidate(pc, e))
    pc.addEventListener('iceconnectionstatechange', e => onIceStateChange(pc, e))
    pc.addEventListener('track', gotRemoteStream)

    pc.addTransceiver('audio', {'direction': 'recvonly'})
    pc.addTransceiver('video', {'direction': 'recvonly'})

    try {
        const stream = await navigator.mediaDevices.getUserMedia({audio: true, video: false})
        console.log('Received local stream')
        localStream = stream
    } catch (e) {
        alert(`getUserMedia() error: ${e.name}`)
    }

    localStream.getTracks().forEach(track => pc.addTrack(track, localStream))
    console.log('Added local stream to pc')

    // Since the remote side has no media stream, we need to pass in the right constraints,
    // in order for it to accept the incoming offer with audio and video.
    try {
        console.log('pc createOffer start')
        const offer = await pc.createOffer(offerOptions)
        console.log('Original Offer:', offer)
        await onCreateOfferSuccess(offer)
    } catch (e) {
        onCreateSessionDescriptionError(e)
    }
}

function onCreateSessionDescriptionError(error) {
    console.log(`Failed to create session description: ${error.toString()}`)
}

// PeerConnection生成offer成功，发送到WebSocket服务
async function onCreateOfferSuccess(desc) {
    console.log(`Offer from pc`)
    console.log(JSON.stringify(desc))
    console.log('pc setLocalDescription start')

    try {
        await pc.setLocalDescription(desc)
        onSetLocalSuccess(pc)
    } catch (e) {
        onSetSessionDescriptionError()
    }

    sendOffer(desc.sdp)
}

// WebSocket消息回调处理函数
function parseResponse(response) {
    console.log("Response:", response)

    if (response.success !== true) {
        console.log("response not success")
        return
    }

    if (response.type === 'answer') {
        console.log("get answer from OpenAPI")

        var answer = {}
        answer["type"] = "answer"
        answer["sdp"] = response.payload
        pc.setRemoteDescription(answer)
    } else if (response.type === 'candidate') {
        console.log("get candidate from OpenAPI")

        var candidate = {}
        candidate["candidate"] = response.payload
        candidate["sdpMid"] = '0'
        candidate["sdpMLineIndex"] = 0

        pc.addIceCandidate(candidate).catch(e => {
            console.log("addIceCandidate fail: " + e.name)
        })
    } else if (response.type === 'disconnect') {
        console.log("get disconnect from OpenAPI")

        sendDisconnect()
    } else if (response.type === 'webrtcConfigs') {
        console.log("get iceServers from OpenAPI")
        console.log(response.payload)
        console.log(JSON.parse(response.payload))

        iceServers = JSON.parse(response.payload)

        call()
    }
}

async function sendFetchWebRTCConfigs() {
    try {
        sendWS("webRTCConfigs", "")
    } catch (e) {
        console.log("sendUpdateWebRTCConfigs fail: " + e.name)
    }
}

async function sendOffer(sdp) {
    // shorter sdp, remove a=extmap... line, device ONLY allow 8KB json payload
    sdp = sdp.replace(/\r\na=extmap[^\r\n]*/g, '')

    console.log("send offer: " + sdp)

    try {
        sendWS("offer", sdp)
    } catch (e) {
        console.log("send offer via WebSocket fail: " + e.name)
    }
}

async function sendCandidate(candidate) {
    try {
        sendWS("candidate", candidate)
    } catch (e) {
        console.log("sendCandidate fail: " + e.name)
    }
}

async function sendDisconnect() {
    console.log("hangup")

    pc.close()

    try {
        sendWS("disconnect", "")
    } catch (e) {
        console.log("hangup the call fail: " + e.name)
    }
}

function onSetLocalSuccess(pc) {
    console.log(`setLocalDescription complete`)
}

function onSetSessionDescriptionError(error) {
    console.log(`Failed to set session description: ${error.toString()}`)
}

// 获取对端音视频流，绑定到网页上的播放控件
function gotRemoteStream(e) {
    console.log('Debug........ ', e.track.kind)
    if (e.track.kind === 'audio') {
        remoteAudio.srcObject = e.streams[0]
    } else if (e.track.kind === 'video') {
        remoteVideo.srcObject = e.streams[0]
    }
}

// 采集到本地candidate候选地址，发送到WebSocket服务
async function onIceCandidate(pc, event) {
    console.log(`ICE candidate:\n${event.candidate ? event.candidate.candidate : '(null)'}`)

    try {
        if (event.candidate != null) {
            sendCandidate("a=" + event.candidate.candidate)
        } else {
            sendCandidate("")
        }
    } catch (e) {
        onAddIceCandidateError(pc, e)
    }
}

function onAddIceCandidateError(pc, error) {
    console.log(`failed to add ICE Candidate: ${error.toString()}`)
}

// ICE连接状态变更处理函数
function onIceStateChange(pc, event) {
    if (pc) {
        console.log(`ICE state: ${pc.iceConnectionState}`)
        console.log('ICE state change event: ', event)

        if (pc.iceConnectionState === 'connected') {
            console.log("webRTC connected")
        }
    }
}
