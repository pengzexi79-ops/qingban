package core

// 本地实时事件总线(SSE)。GET /events 订阅端直接消费;业务层通过 Publish 广播。
// 设计要点:
//   - 单机多窗口(浏览器多标签/Wails 窗口)经本总线保持一致;
//   - 事件带全局自增 seq,支持 Last-Event-ID 断线重连补发(环形缓冲保留最近 hubRetain 条);
//   - 订阅者写入缓慢(通道满)时直接断开,避免阻塞业务协程(聊天生成主链路不被慢消费者拖死)。
// 本文件为可运行的基础设施实现(阶段 P0 验收直接落地)。

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// SSE 事件名(与 docs/PHASE1_API.md §2.1 对齐;proactive_message/moment_published/backup_done 属后续阶段暂不推送)。
const (
	// EventNewMessage:新消息(用户/AI/群聊),data 携带 {conversationId, message}
	EventNewMessage = "new_message"
	// EventTyping:输入状态(回复开始前),data 携带 {conversationId, senderId}
	EventTyping = "typing"
	// EventRead:会话已读,data 携带 {conversationId, unreadTotal}
	EventRead = "read"
	// EventPresence:在线状态变更,data 携带 {companionId, online}
	EventPresence = "presence"
	// EventMemoryCandidates:AI 回复完成后的待确认记忆候选,data 携带 {conversationId, candidates}
	EventMemoryCandidates = "memory_candidates"
	// EventRoundStart/EventRoundMessage/EventRoundEnd:群聊轮次 开始/逐条消息/结束,data 见 groups 业务
	EventRoundStart   = "round_start"
	EventRoundMessage = "round_message"
	EventRoundEnd     = "round_end"
	// EventSettingsChanged:设置(用户/角色/API 配置)变更,data 携带 {scope,id} 供前端局部刷新
	EventSettingsChanged = "settings_changed"
	// EventThreadsChanged:会话列表整体变更提示(建/删会话、置顶、删除好友后),前端可触发 /refresh
	EventThreadsChanged = "threads_changed"
)

// hubBufferSize:每个订阅者事件通道缓冲长度。
// 说明:前端消费慢于生成时先缓冲,再满则直接断连(客户端靠 Last-Event-ID 重连补发)。
const hubBufferSize = 256

// hubRetain:环形缓冲保留的历史事件条数(供 Last-Event-ID 断线补发)。
// 说明:本地单机事件量小,1024 条足够覆盖秒级断线场景。
const hubRetain = 1024

// Event:总线上的一个事件。
type Event struct {
	// Seq:全局自增序号(同一 Hub 内单调递增),作为 SSE 帧的 id,供 Last-Event-ID 使用。
	Seq int64
	// Name:事件名(上方 const 之一)。
	Name string
	// Data:业务载荷(任意结构体),由 Hub 序列化为 JSON 写入 data: 行。
	Data any
}

// Subscriber:单个前端连接(一个 GET /events 请求)。
type Subscriber struct {
	ch   chan []byte // 已序列化的完整 SSE 帧(含 id:/event:/data: 行与结尾空行)
	done chan struct{}
}

// SSEHub:事件总线核心。
type SSEHub struct {
	mu       sync.RWMutex             // 保护 subs/ring 并发读写
	subs     map[*Subscriber]struct{} // 活跃订阅集合
	ring     []Event                  // 环形缓冲(断线补发),逻辑容量 hubRetain
	ringBase int64                    // ring[0] 对应的全局 seq(溢出裁剪后前移)
	seq      int64                    // 已分发的最新全局 seq
	closeCh  chan struct{}            // Hub 关闭信号
	once     sync.Once                // 保证 Close 只执行一次
}

// NewSSEHub:创建事件总线(调用点:init.NewApp(),先于路由注册)。
func NewSSEHub() *SSEHub {
	return &SSEHub{subs: make(map[*Subscriber]struct{}), closeCh: make(chan struct{})}
}

// Publish:向全部订阅者广播一个事件(调用点:server 业务消息落库后/轮次推进/设置变更、AI 回调)。
func (h *SSEHub) Publish(name string, data any) {
	h.mu.Lock()
	h.seq++
	ev := Event{Seq: h.seq, Name: name, Data: data}
	h.ring = append(h.ring, ev)
	// 裁剪:只保留最新 hubRetain 条,并同步前移 ringBase
	if len(h.ring) > hubRetain {
		h.ring = h.ring[len(h.ring)-hubRetain:]
		h.ringBase = h.ring[0].Seq
	}
	// 帧必须携带已分配 seq,故在锁内编码(编码失败只记日志丢弃,不 panic 主链路)
	frame, encErr := encodeFrame(ev)
	if encErr == nil {
		for s := range h.subs {
			select {
			case s.ch <- frame:
			default:
				// 慢消费者:断开,由客户端 Last-Event-ID 重连补发
				delete(h.subs, s)
				close(s.done)
			}
		}
	}
	h.mu.Unlock()
}

// Subscribe:注册新订阅者。
// 参数 lastEventID:客户端 Last-Event-ID(-1=全新连接,不补历史);≥0 时先补发其后事件。
// 调用点:server/infra.go 的 GET /events handler。
func (h *SSEHub) Subscribe(lastEventID int64) *Subscriber {
	sub := &Subscriber{ch: make(chan []byte, hubBufferSize), done: make(chan struct{})}

	h.mu.Lock()
	if lastEventID >= 0 && len(h.ring) > 0 {
		// 收集 seq>lastEventID 的历史事件(取时间正序的最新 hubBufferSize 条,防止通道溢出)
		hist := make([]Event, 0, len(h.ring))
		for _, e := range h.ring {
			if e.Seq > lastEventID {
				hist = append(hist, e)
			}
		}
		if len(hist) > hubBufferSize {
			hist = hist[len(hist)-hubBufferSize:]
		}
		for _, e := range hist {
			if f, err := encodeFrame(e); err == nil {
				select {
				case sub.ch <- f:
				default:
				}
			}
		}
	}
	h.subs[sub] = struct{}{}
	h.mu.Unlock()
	return sub
}

// Events:订阅者事件帧通道(handler 侧 for-range 输出 SSE)。
func (s *Subscriber) Events() <-chan []byte {
	return s.ch
}

// Done:订阅者关闭信号(连接断开)。
func (s *Subscriber) Done() <-chan struct{} {
	return s.done
}

// Close:关闭整个总线(进程退出):逐一关闭订阅者并通知;仅执行一次。
func (h *SSEHub) Close() {
	h.once.Do(func() {
		h.mu.Lock()
		for s := range h.subs {
			close(s.done)
		}
		h.subs = make(map[*Subscriber]struct{})
		h.mu.Unlock()
		close(h.closeCh)
	})
}

// encodeFrame:把事件序列化为一条完整 SSE 帧:
//
//	id: <seq>
//	event: <name>
//	data: <json>
//	<空行>
func encodeFrame(e Event) ([]byte, error) {
	data, err := json.Marshal(e.Data)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "id: %d\nevent: %s\ndata: %s\n\n", e.Seq, e.Name, data)
	return []byte(b.String()), nil
}
