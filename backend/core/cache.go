package core

// 进程内易失 KV 缓存(bigcache/v3)。
// 用途:只放"重启即失"的运行态——用户活动/心跳时间戳、群冷却戳、运行中回合现场等;
// 与持久层(config_kvs 加密键值、各实体表)严格分工:需要跨重启的一律不放进本缓存。
// 装配:initapp.NewApp 第 7 步创建并注入(core.Mem);测试可自行构造 NewMemCache。

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"
)

// IdleWindow:用户"输入完成"判定窗口(默认 2s,编译期可注入变量)。
// 语义(见 docs/model_design.md「易失态与输入完成判定」):用户发出消息后,
// 引擎以心跳/新消息滚动重置本窗口;超过 IdleWindow 未再捕获用户活动 → 判定
// "这轮话说完了" → 才把攒批内容投喂给 AI,AI 完成一次回复后本轮终止。
// initapp/装配与测试可直接覆盖本变量。
var IdleWindow = 2 * time.Second

// Mem:进程内易失 KV 缓存实例(由 initapp.NewApp 初始化注入;nil=未装配)。
var Mem *MemCache

// MemCache:易失 KV 缓存封装(bigcache,TTL 自动淘汰)。
type MemCache struct {
	b *bigcache.BigCache
}

// NewMemCache:构造缓存。ttlSeconds<=0 时用默认 60s(心跳/冷却类条目远小于该值)。
func NewMemCache(ttlSeconds int) (*MemCache, error) {
	if ttlSeconds <= 0 {
		ttlSeconds = 60
	}
	cfg := bigcache.DefaultConfig(time.Duration(ttlSeconds) * time.Second)
	cfg.CleanWindow = time.Duration(ttlSeconds/2) * time.Second
	b, err := bigcache.New(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	return &MemCache{b: b}, nil
}

// Set:写一个键(值任意字节;内部按 bigcache 存)。
func (m *MemCache) Set(key string, value []byte) error { return m.b.Set(key, value) }

// Get:读一个键;不存在返回 ErrEntryNotFound(bigcache)。
func (m *MemCache) Get(key string) ([]byte, error) { return m.b.Get(key) }

// Delete:删一个键。
func (m *MemCache) Delete(key string) error { return m.b.Delete(key) }

// SetUnix:写一个"时刻"(UnixMilli,用于活动时间戳/冷却戳)。
func (m *MemCache) SetUnix(key string, at time.Time) error {
	var buf [8]byte
	for i := 0; i < 8; i++ {
		buf[i] = byte(at.UnixMilli() >> (8 * (7 - i)))
	}
	return m.b.Set(key, buf[:])
}

// GetUnix:读"时刻"(UnixMilli);键缺失或损坏返回零值与 false。
func (m *MemCache) GetUnix(key string) (time.Time, bool) {
	raw, err := m.b.Get(key)
	if err != nil || len(raw) != 8 {
		return time.Time{}, false
	}
	var ms int64
	for i := 0; i < 8; i++ {
		ms = ms<<8 | int64(raw[i])
	}
	return time.UnixMilli(ms), true
}

// Reset:清空全部缓存(测试/进程重置用)。
func (m *MemCache) Reset() error { return m.b.Reset() }

// ---- 键约定(集中注释,避免散落) ----
// act:<conversation_id>   最近一次用户活动/打字心跳时刻(输入完成判定)
// flush:<conversation_id> 最近一次投喂开始时刻(群冷却判断)
// round:<conversation_id> 运行中回合现场(JSON:回合 id/发言成员/已产出消息),回复完成即删
