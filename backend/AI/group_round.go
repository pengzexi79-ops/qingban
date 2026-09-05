package ai

// 群聊回合(动态):手动触发入口(与批次调度引擎联动)。
// 语义:回合 = 运行期易失态,不落库(见 core/cache.go:现场 key=round:<convID>,
// 冷却 key=flush:<convID>);手动触发 POST /groups/{id}/rounds = 调度引擎"点名全员"
// 立即整批投喂(绕过攒批静默窗,仍受冷却约束);被点名消息(消息内 @)走同一引擎的
// FlushConversation(点名子集)。回合生命周期(round_* SSE 事件)由 server 装配层按引擎
// hooks(OnRoundStart/OnReply/OnSilent/OnConsumed)维护:静默成员不记入回合产出;
// AI 完成本轮回复后回合现场即删(无历史回放)。重复触发受幂等键与冷却时间约束。
// 伪代码草稿:调度细节以函数体内伪代码注释占位。

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"qingban/model"
)

// RoundResult:回合创建结果(openapi RoundResult;动态回合,响应即返回)。
type RoundResult struct {
	// RoundID:运行期回合 id(进程内自增,仅在本次响应与后续 round_* 事件流内有效)。
	RoundID uint `json:"roundId"`
	// Status:running(已交棒调度引擎,整批投喂进行中;完成/取消经 round_end 事件表达)。
	Status string `json:"status"`
	// Messages:本回合产生的 AI 消息(异步执行,响应时为空;完成后经 round_message 事件逐条推送)。
	Messages []model.Message `json:"messages,omitempty"`
	// CancelReason:取消原因(cancelled 非空,如 COOLDOWN_ACTIVE)。
	CancelReason *string `json:"cancelReason,omitempty"`
}

// GroupTurnMessage:轮次内"送某位成员的上下文消息"(群聊可见上下文)—— 旧版逐轮选人发言的
// 消息形态,保留作历史参考;批次调度语义下上下文由 dispatch.go 的 BuildBatchLines 按
// "成员未读水位"拼接(隔离要求不变:每位 AI 只见群内消息+自身短期记忆,不透传私密记忆)。
type GroupTurnMessage struct {
	// SenderID:发言者(空串=用户提问)。
	SenderID string `json:"senderId"`
	// Content:纯文本。
	Content string `json:"content"`
}

// GroupRoundArgs:触发一轮入参(server/groups.go 组装)。
type GroupRoundArgs struct {
	// Group:群实体(策略来源:冷却沿用 strategy.cooldownSeconds)。
	Group model.Group
	// Members:群内角色详情(前置校验用;实际投喂成员与各自配置由调度引擎现读)。
	Members []model.Companion
	// ConversationID:消息归属会话(conversations.id;=群会话行 id,非 group.id)。
	ConversationID uint
	// Now:当前时刻(注入便于测试冷却)。
	Now time.Time
}

// groupLock:进程内群互斥(防止连点触发与冷却竞态并发跑两轮)。
// 桌面本地单进程形态下即终态;若日后出现多进程形态再另行评估。
var groupLock = newGroupLocks()

// RunGroupRound:触发一轮群聊(调用点:POST /groups/{id}/rounds,幂等)。
// 语义:前置校验通过 → 生成运行期回合(不落库),交棒调度引擎 FlushConversation(全员),
// 立即返回 roundId(running)。
//
//	---- 前置校验(失败 → 错误码,无落库)----
//	members := filterEnabled(args.Members)                       // ① 剔除已删/禁用角色
//	if len(members) < 2 { return err("成员不足") }
//	last, _ := core.Mem.GetUnix("flush:"+convID)                // ② 冷却(易失缓存;重启即清零,可接受)
//	if args.Now.Sub(last) < strategy.cooldownSeconds { return err(CodeCooldownActive) }
//	lock := groupLock.get(args.Group.ID); lock.Lock()            // ③ 防连点双轮(引擎侧另有 running 防重入)
//	roundID := nextRoundSeq()                                    // ④ 运行期回合号(进程内递增计数器)
//	现场 JSON{roundID, targets: members, produced: []msgID} 写入 core.Mem(key=round:<convID>)
//	hub.Publish(EventRoundStart, {roundId: roundID, groupId, memberIds})  // ⑤
//
//	---- 交棒批次调度引擎(见 dispatch.go FlushConversation)----
//	err := Dispatcher.FlushConversation(ctx, args.ConversationID, true, nil)   // ⑥ 点名全员
//	//   引擎逐成员:读各自未读批(成员水位 member_cursors)→ 组上下文(短期记忆+本批)→ 调用;
//	//   回复产出经 hooks.OnReply:落库 assistant 消息 → 追加到 round 现场.produced
//	//                               → hub.Publish(round_message + new_message);
//	//   静默产出经 hooks.OnSilent:不落消息不进现场(已读不回,消费水位照常推进);
//	//   整批结束由装配层 hooks 收尾:刷新 flush 冷却戳 → hub.Publish(round_end {roundId, messages})
//	//                                → core.Mem.Delete(round:<convID>)(回复完成即终止本轮)
//	//   全员失败路径:round_end {status: cancelled, cancelReason}
//
//	---- 响应 ----
//	return &RoundResult{roundID, "running", nil, nil}, nil  // ⑦ 立即返回(异步执行)
//
// TODO(实现):见函数注释 ①~⑦
func RunGroupRound(ctx context.Context, args GroupRoundArgs) (*RoundResult, error) {
	return nil, nil // TODO(实现):见函数注释 ①~⑦
}

// SelectSpeakers:按策略选出本轮发言成员(纯函数,单测目标;返回顺序即发言顺序)。
func SelectSpeakers(members []model.Companion, strategy model.GroupStrategy, rng *rand.Rand) []model.Companion {
	if len(members) == 0 {
		return nil
	}
	// ① 上限:至少 1 人,至多 min(maxSpeakers, 成员数)
	limit := strategy.MaxSpeakers
	if limit < 1 {
		limit = 1
	}
	if limit > len(members) {
		limit = len(members)
	}
	// ② random:洗牌后取前 limit(mode 缺省按 random 处理)
	if strategy.Mode == "random" || strategy.Mode == "" {
		cp := append([]model.Companion(nil), members...)
		rng.Shuffle(len(cp), func(i, j int) { cp[i], cp[j] = cp[j], cp[i] })
		return cp[:limit]
	}
	// ③ turn/其余:按成员传入顺序取前 limit(顺序即发言序)
	return members[:limit]
}

// groupLocks:群互斥锁集合(懒加载,map[groupID]*sync.Mutex;id 为 groups.id)。
type groupLocks struct {
	mu   sync.Mutex
	keys map[uint]*sync.Mutex
}

// newGroupLocks:构造锁集合。
func newGroupLocks() *groupLocks { return &groupLocks{keys: map[uint]*sync.Mutex{}} }

// get:取某群锁(不存在则建)。
func (g *groupLocks) get(groupID uint) *sync.Mutex {
	g.mu.Lock()
	defer g.mu.Unlock()
	if lk, ok := g.keys[groupID]; ok {
		return lk
	}
	lk := &sync.Mutex{}
	g.keys[groupID] = lk
	return lk
}
