// tests/core:验证 core 包基础设施可用性(SSE 事件总线广播/断线补发/关闭、幂等表)。
// 运行:在 D:\开源项目\青伴\tests 下执行 `go run ./core`。
// 这是阶段 P0 验收线("进程能起、前端能连、能收事件")的最小化验证。

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"qingban/core"
)

// fails:累计失败数。
var fails int

// check:断言辅助。
func check(name string, cond bool, detail string) {
	if cond {
		fmt.Printf("[PASS] %s\n", name)
	} else {
		fails++
		fmt.Printf("[FAIL] %s: %s\n", name, detail)
	}
}

// mustRecv:超时读取一个事件帧(超时返回空并记失败)。
func mustRecv(name string, sub *core.Subscriber, timeout time.Duration) []byte {
	select {
	case f := <-sub.Events():
		return f
	case <-time.After(timeout):
		fails++
		fmt.Printf("[FAIL] %s: 等待事件超时\n", name)
		return nil
	}
}

func main() {
	// ---- 广播:帧携带递增 id/事件名/载荷 ----
	hub := core.NewSSEHub()
	sub := hub.Subscribe(-1) // 全新连接
	hub.Publish(core.EventNewMessage, map[string]any{"conversationId": "conv-1", "text": "你好"})
	f1 := string(mustRecv("hub.publish", sub, time.Second))
	check("hub.frame.event", strings.Contains(f1, "event: new_message"), f1)
	check("hub.frame.data", strings.Contains(f1, `"conversationId":"conv-1"`), f1)
	check("hub.frame.seq1", strings.Contains(f1, "id: 1"), f1)

	// ---- Last-Event-ID 断线补发:seq=1 之后的事件才补 ----
	hub.Publish(core.EventTyping, "e2") // seq=2
	sub2 := hub.Subscribe(1)            // 断线前已收 seq=1 → 只补 seq=2
	f2 := string(mustRecv("hub.replay", sub2, time.Second))
	check("hub.replay.frame", strings.Contains(f2, "id: 2") && strings.Contains(f2, "event: typing"), f2)
	select {
	case extra := <-sub2.Events():
		fails++
		fmt.Printf("[FAIL] hub.replay.range: 补发越界 %q\n", extra)
	default:
		fmt.Println("[PASS] hub.replay.range")
	}
	hub.Publish(core.EventRead, "e3")
	f3 := string(mustRecv("hub.afterReplay", sub2, time.Second))
	check("hub.afterReplay.seq3", strings.Contains(f3, "id: 3"), f3)

	// ---- 环形溢出(1100 条 > 保留 1024):不 panic、seq 单调 ----
	big := core.NewSSEHub()
	for i := 0; i < 1100; i++ {
		big.Publish(core.EventRead, i)
	}
	sub3 := big.Subscribe(0)
	f4 := string(mustRecv("hub.ringOverflow", sub3, time.Second))
	check("hub.ringOverflow.frame", strings.Contains(f4, "event: read"), f4)
	big.Close()

	// ---- Close:通知全部订阅者断开 ----
	hub.Close()
	select {
	case <-sub.Done():
		fmt.Println("[PASS] hub.close.notify")
	case <-time.After(time.Second):
		fails++
		fmt.Println("[FAIL] hub.close.notify: 订阅者未收到 done")
	}

	// ---- 幂等表:登记/命中/重放(状态码+响应体)/覆盖 ----
	idem := core.NewIdemStore()
	_, ok0 := idem.Get("k-1")
	check("idem.empty", !ok0, "未登记 key 不应命中")
	idem.Register("k-1", 201, []byte(`{"ok":true}`))
	rec, ok := idem.Get("k-1")
	check("idem.hit", ok && rec.StatusCode == 201 && string(rec.Response) == `{"ok":true}`, fmt.Sprint(rec))
	idem.Register("k-1", 201, []byte(`{"v":2}`))
	rec2, _ := idem.Get("k-1")
	check("idem.overwrite", string(rec2.Response) == `{"v":2}`, string(rec2.Response))

	if fails > 0 {
		fmt.Printf("\n%d 项失败\n", fails)
		os.Exit(1)
	}
	fmt.Println("\ncore:全部可用性检查通过")
}
