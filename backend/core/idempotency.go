package core

// 幂等键登记表(进程内实现;多实例/云同步阶段再迁移 SQLite 或独立缓存,见架构文档 §6.3)。
// 语义(PHASE1 §2):发送类接口携带 Idempotency-Key,同一 key 只执行一次;
// 命中后按首次响应的状态码与响应体原样重放(不重复产生副作用)。

import (
	"sync"
	"time"
)

// idemTTL:幂等记录保留时长(覆盖"前端超时重试"窗口,默认 15 分钟)。
const idemTTL = 15 * time.Minute

// idemCleanInterval:惰性清理扫描间隔(只在写入时顺带清理,无常驻定时器)。
const idemCleanInterval = 5 * time.Minute

// IdemRecord:一次幂等执行的登记结果。
type IdemRecord struct {
	// Key:请求头 Idempotency-Key 原文(≤128 字符,超长在中间件即校验失败)。
	Key string
	// StatusCode:首次执行的 HTTP 状态码(重放时保持一致,如 201 不能变 200)。
	StatusCode int
	// Response:首次执行的响应体快照(≤64KB;重放时原样返回)。
	Response []byte
	// CreatedAt:首次执行时刻(用于 TTL 清理)。
	CreatedAt time.Time
}

// IdemStore:幂等登记表(内存 map + 互斥锁)。
type IdemStore struct {
	mu    sync.Mutex            // 保护 items
	items map[string]IdemRecord // key → 首次执行记录
	last  time.Time             // 上次清理时刻(惰性清理)
}

// NewIdemStore:创建幂等登记表(调用点:init.NewApp())。
func NewIdemStore() *IdemStore {
	return &IdemStore{items: make(map[string]IdemRecord), last: time.Now()}
}

// Get:查询 key 是否已执行;命中(且未过期)返回记录。
func (s *IdemStore) Get(key string) (IdemRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.items[key]
	if !ok {
		return IdemRecord{}, false
	}
	if time.Since(rec.CreatedAt) > idemTTL {
		delete(s.items, key) // 过期即失效,允许重新触发
		return IdemRecord{}, false
	}
	return rec, true
}

// Register:登记某 key 首次执行结果(幂等语义完成点;业务失败不应调用,允许用户修参重试)。
func (s *IdemStore) Register(key string, statusCode int, resp []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = IdemRecord{Key: key, StatusCode: statusCode, Response: resp, CreatedAt: time.Now()}
	s.sweepIfNeededLocked()
}

// sweepIfNeededLocked:惰性清理(调用方已持锁):距上次清理超过 idemCleanInterval 时删除全部过期项。
func (s *IdemStore) sweepIfNeededLocked() {
	now := time.Now()
	if now.Sub(s.last) < idemCleanInterval {
		return
	}
	s.last = now
	for k, rec := range s.items {
		if now.Sub(rec.CreatedAt) > idemTTL {
			delete(s.items, k)
		}
	}
}
