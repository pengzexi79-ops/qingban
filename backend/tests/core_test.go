package tests

// core 包测试:SSE 事件总线(广播/断线补发/关闭)与幂等表。
// 事件总线为阶段 P0 验收核心("进程能起、前端能连、能收事件"),先行锁定语义。

import (
	"strings"
	"testing"
	"time"

	"qingban/core"
)

// mustRecv:带超时读取一个事件帧;超时即失败。
func mustRecv(t *testing.T, sub *core.Subscriber, timeout time.Duration) []byte {
	t.Helper()
	select {
	case frame := <-sub.Events():
		return frame
	case <-time.After(timeout):
		t.Fatal("等待事件超时")
		return nil
	}
}

func TestHub_PublishDelivers(t *testing.T) {
	hub := core.NewSSEHub()
	defer hub.Close()
	sub := hub.Subscribe(-1) // 全新连接,不补历史

	hub.Publish(core.EventNewMessage, map[string]any{"conversationId": "conv-1", "text": "你好"})
	frame := mustRecv(t, sub, time.Second)
	s := string(frame)
	if !strings.Contains(s, "event: new_message") {
		t.Errorf("缺少 event 行: %q", s)
	}
	if !strings.Contains(s, `"conversationId":"conv-1"`) {
		t.Errorf("data 载荷缺失: %q", s)
	}
	if !strings.Contains(s, "id: 1") {
		t.Errorf("缺少递增 seq: %q", s)
	}
}

func TestHub_SubscribeReplayByLastEventID(t *testing.T) {
	hub := core.NewSSEHub()
	defer hub.Close()

	hub.Publish(core.EventTyping, "e1") // seq=1
	hub.Publish(core.EventRead, "e2")   // seq=2

	// 携带 Last-Event-ID=1 → 只补发 seq=2
	sub := hub.Subscribe(1)
	frame := mustRecv(t, sub, time.Second)
	s := string(frame)
	if !strings.Contains(s, "id: 2") || !strings.Contains(s, "event: read") {
		t.Errorf("补发帧错误: %q", s)
	}
	// 不应有多余补发(缓冲区应空)
	select {
	case extra := <-sub.Events():
		t.Errorf("补发超出范围: %q", extra)
	default:
	}
	// 之后新事件正常推送(seq=3)
	hub.Publish(core.EventRead, "e3")
	if f := mustRecv(t, sub, time.Second); !strings.Contains(string(f), "id: 3") {
		t.Errorf("新事件错误: %q", f)
	}
}

func TestHub_SeqMonotonicAndRingRetain(t *testing.T) {
	hub := core.NewSSEHub()
	defer hub.Close()
	const n = 1100 // 超过环形保留窗口(1024)+ 缓冲(256),验证溢出裁剪不 panic、seq 单调
	for i := 0; i < n; i++ {
		hub.Publish(core.EventRead, i)
	}
	sub := hub.Subscribe(0)
	// 首个补发帧 seq 应大于 0(保留窗口尾部),且 event 行正确
	f := mustRecv(t, sub, time.Second)
	if !strings.Contains(string(f), "event: read") {
		t.Fatalf("帧异常: %q", f)
	}
}

func TestHub_CloseNotifiesSubscriber(t *testing.T) {
	hub := core.NewSSEHub()
	sub := hub.Subscribe(-1)
	hub.Close()
	select {
	case <-sub.Done():
	case <-time.After(time.Second):
		t.Fatal("Close 后订阅者未收到 done 通知")
	}
}

func TestIdem_RegisterGetReplay(t *testing.T) {
	store := core.NewIdemStore()
	if _, ok := store.Get("k-1"); ok {
		t.Fatal("未登记的 key 不应命中")
	}
	store.Register("k-1", 201, []byte(`{"ok":true}`))
	rec, ok := store.Get("k-1")
	if !ok {
		t.Fatal("登记后应命中")
	}
	if rec.StatusCode != 201 || string(rec.Response) != `{"ok":true}` {
		t.Errorf("重放数据不符: %+v", rec)
	}
}

func TestIdem_RepeatedRegisterOverwrites(t *testing.T) {
	store := core.NewIdemStore()
	store.Register("k", 200, []byte("v1"))
	store.Register("k", 201, []byte("v2"))
	rec, ok := store.Get("k")
	if !ok || rec.StatusCode != 201 || string(rec.Response) != "v2" {
		t.Errorf("同 key 重复登记应覆盖: %+v", rec)
	}
}
