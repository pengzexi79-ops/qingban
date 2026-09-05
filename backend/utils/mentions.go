package utils

// 群聊 "@点名" 解析工具 —— 伪代码草稿,解析细节以注释占位(实现顺序见文件尾注)。
// 语义依据:docs/batch_dispatch_design.md §3.1(设计决策记录 #2)+ model/message.go:
//   - 发送时按 "@角色名" 精确匹配会话成员(长名优先),命中结果写 Message.Mentions(关系表);
//   - @ 谁 → 调度层只立即投喂谁;解析不到的 @xxx 视为普通文本(不影响攒批);
//   - 保留字 @所有人 / @all → 置 Message.MentionAll(全员是动态的,展开在运行时做,
//     不再使用"everyone"伪 id);请求体显式带 mention_ids 时可覆盖。

// Mention:一次 @ 命中。
type Mention struct {
	// Name:命中的成员名(@ 后原文)。
	Name string `json:"name"`
	// ID:命中的成员 id(companions.id)。
	ID uint `json:"id"`
	// Start/End:@ 符号在原文中的 rune 下标区间(前端高亮/后端替换用;End 为名字尾后)。
	Start int `json:"start"`
	End   int `json:"end"`
}

// MentionEveryoneNames:点名全员的保留字(与角色同名时角色名优先,见 ParseMentions ③)。
var MentionEveryoneNames = []string{"所有人", "all"}

// ParseMentions:从 text 提取 @点名命中。
// 返回:(all, mentions)——all=命中 @所有人/@all;mentions=点到的具体成员(按出现顺序)。
// 参数 names:群成员名 → 成员 id 的映射(由调用方按会话成员组装)。语义:
//
//	① 候选名按字符数降序预排序(长名优先,防 "mumu" 被 "mu" 截胡);
//	② 逐 rune 扫描 '@',取其后再长名最长窗口做最长匹配;命中后游标越过名字
//	   (允许同一文本多次 @);名字后紧跟字母/数字视为名字未结束(防 @mumu2 误中 @mumu);
//	③ 角色名优先于保留字:names 不含保留字时,才尝试 @所有人/@all(ASCII 大小写不敏感),
//	   命中 → all=true;
//	④ 均未命中 → 普通文本,不产出命中。
func ParseMentions(text string, names map[string]uint) (all bool, mentions []Mention) {
	// all, hits := false, []Mention{}
	// cands := sortByRuneLenDesc(keysOf(names))                         // ①
	// for i, r := range []rune(text) {
	//     if r != '@' { continue }
	//     window := runes(i+1 : min(len, i+1+maxNameLen))               // ② 逐候选最长匹配
	//     if hit { hits = append(hits, {name, names[name], i, i+1+len(name)}); i += len(name) }
	//     else if !角色名冲突 { if 保留字匹配 { all = true; i += len(保留字) } }  // ③
	// }
	// return all, hits                                                // ④
	return false, nil // TODO(实现):见函数注释 ①~④
}

// 注:实现顺序建议——先在 tests 子目录(仿既有 utils 测试形态)准备样例文本:
//   中文名("沐沐")/英文名(mumu)/长名优先(mumu vs mumu2)/标点边界/重复 @/@所有人;
//   全部通过后按注释 ①~④ 落地,并把结果写回 Message.MentionAll/Mentions(见 server 发送链路)。
