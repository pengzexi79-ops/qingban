package ai

// 群聊轮次调度(第一阶段"同步执行一轮")。
// 语义(PHASE1_API P3 + BACKEND_HANDOFF §4.2):选人(≤maxSpeakers)→ 逐个调用 → 收尾;
// 消息逐条广播(round_start/round_message/round_end);防自激:人数上限固定+失败立即收尾;
// 重复触发受幂等键与冷却时间约束。伪代码草稿:调度细节以函数体内伪代码注释占位。

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"qingban/model"
)

// RoundResult:轮次执行结果(openapi RoundResult)。
type RoundResult struct {
	// RoundID:本轮唯一 id。
	RoundID string `json:"roundId"`
	// Status:completed/cancelled。
	Status string `json:"status"`
	// Messages:本轮产生的 AI 消息(completed 时返回)。
	Messages []model.Message `json:"messages,omitempty"`
	// CancelReason:取消原因(cancelled 非空,如 COOLDOWN_ACTIVE)。
	CancelReason *string `json:"cancelReason,omitempty"`
}

// GroupTurnMessage:轮次内"送某位成员的上下文消息"(群聊可见上下文)。
// 隔离要求:每位 AI 只读 用户提问+群公告+本轮其他成员发言+自己的私聊记忆,
// 不互相透传私密记忆。
type GroupTurnMessage struct {
	// SenderID:发言者(空串=用户提问)。
	SenderID string `json:"senderId"`
	// Content:纯文本。
	Content string `json:"content"`
}

// GroupRoundArgs:触发一轮入参(server/groups.go 组装)。
type GroupRoundArgs struct {
	// Group:群实体(策略来源)。
	Group model.Group
	// Members:群内角色详情(各自 memorySettings/apiProfileId;逐成员独立调用)。
	Members []model.Companion
	// UserPrompt:用户本轮提问(可空=由调度成员起话题)。
	UserPrompt string
	// ConversationID:消息归属会话(=group.ID)。
	ConversationID string
	// Now:当前时刻(注入便于测试冷却)。
	Now time.Time
}

// groupLock:进程内群互斥(防止连点触发与冷却竞态并发跑两轮)。
// 桌面本地单进程形态下即终态;若日后出现多进程形态再另行评估。
var groupLock = newGroupLocks()

// RunGroupRound:同步执行一轮群聊(调用点:POST /groups/{id}/rounds)。
func RunGroupRound(ctx context.Context, args GroupRoundArgs) (*RoundResult, error) {
	// ---- 前置校验(失败 → cancelled + 对应错误码) ----
	// members := filterEnabled(args.Members)                       // ① 剔除已删/禁用角色
	// if len(members) < 2 { return cancelled("成员不足") }
	// last := max(rounds 表 triggered_at, args.Group.LastRoundAt)  // ② 冷却(进程重启后依赖 DB 兜底)
	// if args.Now.Sub(last) < strategy.cooldownSeconds { return err(CodeCooldownActive) }
	// if !userSettings.AutoMessages(手动触发可不查,注释:手动触发豁免用户总开关) { ... }
	// lock := groupLock.get(args.Group.ID); lock.Lock(); defer lock.Unlock()  // ④ 加群锁(覆盖整个轮次)
	//
	// round := Round{RoundID: "round-" + uuid4(), GroupID: args.Group.ID,
	//                TriggeredAt: args.Now, Status: running}          // 落 Round(running)
	// hub.Publish(EventRoundStart, {roundId, groupId})               // ⑤
	//
	// speakers := SelectSpeakers(members, args.Group.Strategy, randSource())  // ⑥ 选人
	// prompt := args.UserPrompt
	// if prompt == "" { prompt = lastUserMessage(args.ConversationID) }   // ★ 补漏:前端先发消息再触发轮次时,
	//                                                                      //   回读该会话最近一条 role=user 消息作为本轮话题
	// var msgs []model.Message; failAll := true
	// for i, sp := range speakers {                                  // ⑦ 逐位串行发言(顺序即发言序)
	//     seen := [此前已发言消息]                                     //    a) 自己记忆召回(隔离)+
	//     context := buildMemberContext(args, sp, prompt, seen)      //    b) 用户提问+公告+本轮他人发言(截断)
	//     result, err := ChatOnce({model: profileOf(sp).ChatModel,    //    c) 独立调用;失败成员跳过不阻断
	//                              Messages: composeMemberMsgs(...), Temperature})
	//     if err != nil { log(err); continue }
	//     msg := Message{ID: message-, ConversationType: group, ConversationId: args.ConversationID,
	//                    Role: assistant, SenderId: sp.ID, Content: result.Content,
	//                    ContentType: text, Timestamp: now, UsageID: 落库后回填}   //    d) 落库+广播
	//     db.Insert(&msg); usageRec 落一行(success)
	//     hub.Publish(EventRoundMessage, {roundId, message}); hub.Publish(EventNewMessage, {conversationId, msg})
	//     round.Speakers = append(round.Speakers, {sp.ID, msg.ID}); msgs = append(msgs, msg)
	//     failAll = false
	// }
	// round.Status, round.EndedAt = completed, &now                      // ⑧ 收尾
	// group.LastRoundAt = &now; db.Save(group); db.Save(round)
	// hub.Publish(EventRoundEnd, {roundId, messages: msgs})
	// return &RoundResult{round.RoundID, completed, msgs, nil}, nil
	//
	// if failAll { round.Status = cancelled; round.CancelReason = &原因; db.Save(round)   // ⑨ 取消路径
	//     return &RoundResult{round.RoundID, cancelled, msgs, reason}, nil }
	return nil, nil // TODO(实现):见函数注释 ①~⑨
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

// groupLocks:群互斥锁集合(懒加载,map[groupID]*sync.Mutex)。
type groupLocks struct {
	mu   sync.Mutex
	keys map[string]*sync.Mutex
}

// newGroupLocks:构造锁集合。
func newGroupLocks() *groupLocks { return &groupLocks{keys: map[string]*sync.Mutex{}} }

// get:取某群锁(不存在则建)。
func (g *groupLocks) get(groupID string) *sync.Mutex {
	g.mu.Lock()
	defer g.mu.Unlock()
	if lk, ok := g.keys[groupID]; ok {
		return lk
	}
	lk := &sync.Mutex{}
	g.keys[groupID] = lk
	return lk
}
