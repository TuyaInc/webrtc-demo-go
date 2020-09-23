package bootstrap

import (
	"errors"
	"github.com/webrtc-demo-go/types"
	"sync"
)

// WsLinkMgr 管理后端浏览器用户agent id与WebSocket连接的映射关系，增、删、查
// mqtt通信模式为异步，需要有一个映射关系，使得收到mqtt消息后知道该往哪个WebSocket客户端发送Response
type WsLinkMgr struct {
	rwMutex sync.RWMutex

	session2Agent map[string]string

	wsLink map[string]*types.WsMessage // agent id -> WebSocket连接
}

var wsLinkMgr *WsLinkMgr

func init() {
	wsLinkMgr = &WsLinkMgr{
		session2Agent: make(map[string]string),

		wsLink: make(map[string]*types.WsMessage),
	}
}

// AddLink 增加浏览器用户agent id关联的WebSocket连接
func AddLink(agentID, sessionID string, msg *types.WsMessage) {
	wsLinkMgr.rwMutex.Lock()
	defer wsLinkMgr.rwMutex.Unlock()

	// 一个agent id对应一个网页，一个网页同时只能有一个会话，清理agent id之前关联的会话
	for session, agent := range wsLinkMgr.session2Agent {
		if agent == agentID {
			delete(wsLinkMgr.session2Agent, session)
		}
	}

	wsLinkMgr.session2Agent[sessionID] = agentID

	wsLinkMgr.wsLink[agentID] = msg
}

// GetLink 查询浏览器用户agent id关联的WebSocket连接
func GetLink(sessionID string) (link *types.WsMessage, err error) {
	wsLinkMgr.rwMutex.RLock()
	defer wsLinkMgr.rwMutex.RUnlock()

	agentID, ok := wsLinkMgr.session2Agent[sessionID]
	if !ok {
		return nil, errors.New("get agent fail")
	}

	link, ok = wsLinkMgr.wsLink[agentID]
	if !ok {
		return nil, errors.New("getLink fail")
	}

	return
}

func GetLinkByAgent(agentID string) (link *types.WsMessage, err error) {
	wsLinkMgr.rwMutex.RLock()
	defer wsLinkMgr.rwMutex.RUnlock()

	ok := false

	link, ok = wsLinkMgr.wsLink[agentID]
	if !ok {
		return nil, errors.New("getLink fail")
	}

	return
}

// RemoveLink 删除浏览器用户agent id关联的WebSocket连接
func RemoveLink(sessionID string) {
	wsLinkMgr.rwMutex.Lock()
	defer wsLinkMgr.rwMutex.Unlock()

	agentID, ok := wsLinkMgr.session2Agent[sessionID]
	if !ok {
		return
	}

	delete(wsLinkMgr.session2Agent, sessionID)

	_, ok = wsLinkMgr.wsLink[agentID]
	if ok {
		delete(wsLinkMgr.wsLink, agentID)
	}
}

// RemoveLinkByConnLost 浏览器网页断开WebSocket连接时，清空相关记录
func RemoveLinkByConnLost(agentID string) {
	wsLinkMgr.rwMutex.Lock()
	defer wsLinkMgr.rwMutex.Unlock()

	for session, agent := range wsLinkMgr.session2Agent {
		if agent == agentID {
			delete(wsLinkMgr.session2Agent, session)
		}
	}

	if _, ok := wsLinkMgr.wsLink[agentID]; ok {
		delete(wsLinkMgr.wsLink, agentID)
	}
}
