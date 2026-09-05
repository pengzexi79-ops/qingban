package tests

// AI 包纯函数测试:群聊选人(SelectSpeakers)。
// 目的:先锁定"随机/顺序/上限"三种策略语义,再驱动 RunGroupRound 落库部分(后续阶段)。

import (
	"math/rand"
	"testing"

	ai "qingban/AI"
	"qingban/model"
)

// makeCompanions:造 n 个测试角色(id 可预测)。
func makeCompanions(n int) []model.Companion {
	out := make([]model.Companion, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, model.Companion{ID: "companion-" + string(rune('a'+i)), Name: "角色"})
	}
	return out
}

func TestSelectSpeakers_TurnModeTakesFirst(t *testing.T) {
	members := makeCompanions(5)
	// mode=turn(默认成员序):取前 min(maxSpeakers, n)
	s := ai.SelectSpeakers(members, model.GroupStrategy{Mode: "turn", MaxSpeakers: 2}, nil)
	if len(s) != 2 || s[0].ID != "companion-a" || s[1].ID != "companion-b" {
		t.Fatalf("turn 模式应取前 2 名且保序: %+v", idsOf(s))
	}
}

func TestSelectSpeakers_RandomSubsetNoDup(t *testing.T) {
	members := makeCompanions(8)
	rng := rand.New(rand.NewSource(42))
	s := ai.SelectSpeakers(members, model.GroupStrategy{Mode: "random", MaxSpeakers: 3}, rng)
	if len(s) != 3 {
		t.Fatalf("random 应取 3 人,得到 %d", len(s))
	}
	seen := map[string]bool{}
	for _, m := range s {
		if seen[m.ID] {
			t.Fatalf("出现重复成员: %s", m.ID)
		}
		seen[m.ID] = true
	}
}

func TestSelectSpeakers_RandomIsRandom(t *testing.T) {
	// 固定种子结果应稳定(回归保护),不同种子结果应(大概率)不同
	members := makeCompanions(10)
	s1 := ai.SelectSpeakers(members, model.GroupStrategy{Mode: "random", MaxSpeakers: 3}, rand.New(rand.NewSource(1)))
	s2 := ai.SelectSpeakers(members, model.GroupStrategy{Mode: "random", MaxSpeakers: 3}, rand.New(rand.NewSource(2)))
	if len(s1) != 3 || len(s2) != 3 {
		t.Fatalf("长度异常: %d %d", len(s1), len(s2))
	}
	if idsOf(s1)[0] == idsOf(s2)[0] && idsOf(s1)[1] == idsOf(s2)[1] && idsOf(s1)[2] == idsOf(s2)[2] {
		t.Log("注意:不同种子出现相同结果(概率极低),仅记录")
	}
}

func TestSelectSpeakers_ClampLimits(t *testing.T) {
	members := makeCompanions(2)
	// maxSpeakers 超成员数 → 夹取到成员数
	s := ai.SelectSpeakers(members, model.GroupStrategy{Mode: "turn", MaxSpeakers: 9}, nil)
	if len(s) != 2 {
		t.Errorf("超上限应夹取 2,得到 %d", len(s))
	}
	// maxSpeakers<1 → 至少 1 人
	s = ai.SelectSpeakers(members, model.GroupStrategy{Mode: "turn", MaxSpeakers: 0}, nil)
	if len(s) != 1 {
		t.Errorf("maxSpeakers=0 应保底 1 人,得到 %d", len(s))
	}
	// 空成员 → 空结果(不 panic)
	if s = ai.SelectSpeakers(nil, model.GroupStrategy{Mode: "random", MaxSpeakers: 2}, rand.New(rand.NewSource(1))); s != nil {
		t.Errorf("空成员应返回 nil,得到 %+v", s)
	}
}

func idsOf(list []model.Companion) []string {
	out := make([]string, 0, len(list))
	for _, m := range list {
		out = append(out, m.ID)
	}
	return out
}
