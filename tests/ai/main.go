// tests/ai:验证 AI 包纯逻辑可用性(群聊轮次选人策略)。
// 运行:在 D:\开源项目\青伴\tests 下执行 `go run ./ai`。
// 注意:此模块 import "qingban/AI"(大写目录),包名 ai——见后端骨架目录约定。

package main

import (
	"fmt"
	"math/rand"
	"os"

	ai "qingban/AI"
	"qingban/model"
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

// makeMembers:造 n 个可预测 id 的角色。
func makeMembers(n int) []model.Companion {
	out := make([]model.Companion, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, model.Companion{ID: fmt.Sprintf("companion-%02d", i), Name: "角色"})
	}
	return out
}

// ids:提取 id 列表(便于比对)。
func ids(list []model.Companion) []string {
	out := make([]string, 0, len(list))
	for _, m := range list {
		out = append(out, m.ID)
	}
	return out
}

func main() {
	members := makeMembers(8)

	// ---- turn(顺序)模式:取前 min(maxSpeakers, n) 且保序 ----
	s := ai.SelectSpeakers(makeMembers(5), model.GroupStrategy{Mode: "turn", MaxSpeakers: 2}, nil)
	check("select.turn", len(s) == 2 && s[0].ID == "companion-00" && s[1].ID == "companion-01", fmt.Sprint(ids(s)))

	// ---- random 模式:不重复、数量正确、同种子稳定 ----
	s1 := ai.SelectSpeakers(members, model.GroupStrategy{Mode: "random", MaxSpeakers: 3}, rand.New(rand.NewSource(42)))
	uniq := map[string]bool{}
	allU := true
	for _, m := range s1 {
		if uniq[m.ID] {
			allU = false
		}
		uniq[m.ID] = true
	}
	check("select.random", len(s1) == 3 && allU, fmt.Sprint(ids(s1)))
	s2 := ai.SelectSpeakers(members, model.GroupStrategy{Mode: "random", MaxSpeakers: 3}, rand.New(rand.NewSource(42)))
	stable := fmt.Sprint(ids(s1)) == fmt.Sprint(ids(s2))
	check("select.random.stable", stable, "同种子应稳定")

	// ---- 夹取:maxSpeakers 超成员数/小于 1/空成员 ----
	check("select.clamp.high", len(ai.SelectSpeakers(makeMembers(2), model.GroupStrategy{Mode: "turn", MaxSpeakers: 9}, nil)) == 2,
		"超上限应夹取到成员数")
	check("select.clamp.low", len(ai.SelectSpeakers(makeMembers(2), model.GroupStrategy{Mode: "turn", MaxSpeakers: 0}, nil)) == 1,
		"maxSpeakers<1 应保底 1 人")
	check("select.clamp.empty", ai.SelectSpeakers(nil, model.GroupStrategy{Mode: "random", MaxSpeakers: 2}, rand.New(rand.NewSource(1))) == nil,
		"空成员应返回 nil(不 panic)")

	if fails > 0 {
		fmt.Printf("\n%d 项失败\n", fails)
		os.Exit(1)
	}
	fmt.Println("\nai:全部可用性检查通过")
}
