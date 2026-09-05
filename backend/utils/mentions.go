package utils

// 群聊 "@点名" 解析工具 —— 伪代码草稿,解析细节以注释占位(实现顺序见文件尾注)。
// 语义依据:docs/batch_dispatch_design.md §3.1(设计决策记录 #2):
//   - 发送时按 "@角色名" 精确匹配会话成员(长名优先),命中结果写 Message.Mentions;
//   - 请求体可显式带 mentionIds 覆盖(前端已解析的场景);
//   - @ 谁 → 调度层只立即投喂谁;解析不到的 @xxx 视为普通文本(不影响攒批);
//   - 保留字 @所有人 / @all 展开为点名全体(结果 id = MentionEveryoneID)。

// Mention:一次 @ 命中。
type Mention struct {
	// Name:命中的成员名(@ 后原文;保留字时为原文,如 "all")。
	Name string `json:"name"`
	// ID:命中的成员 id;保留字 @所有人/@all 时为 MentionEveryoneID。
	ID string `json:"id"`
	// Start/End:@ 符号在原文中的 rune 下标区间(前端高亮/后端替换用;End 为名字尾后)。
	Start int `json:"start"`
	End   int `json:"end"`
}

// MentionEveryoneID:点名全员的保留 id(与 model.MentionEveryone 同值;
// 引擎按此展开为群内全部成员,见 ai.MentionReaderIDs)。
const MentionEveryoneID = "everyone"

// MentionEveryoneNames:保留字(与角色同名时角色名优先,见 ParseMentions ③)。
var MentionEveryoneNames = []string{"所有人", "all"}

// ParseMentions:从 text 提取 @点名命中,返回按出现顺序的命中列表。
// 参数 names:群成员名 → 成员 id 的映射(由调用方按会话成员组装)。语义:
//
//	① 候选名按字符数降序预排序(长名优先,防 "mumu" 被 "mu" 截胡);
//	② 逐 rune 扫描 '@',取其后"已知名最长字长"窗口做最长匹配;命中后游标越过名字
//	   (允许同一文本多次 @);名字后紧跟字母/数字视为名字未结束(防 @mumu2 误中 @mumu);
//	③ 角色名优先于保留字:names 不含保留字时,才尝试 @所有人/@all(ASCII 大小写不敏感),
//	   命中 id = MentionEveryoneID;
//	④ 均未命中 → 普通文本,不产出命中。
func ParseMentions(text string, names map[string]string) []Mention {
	// hits := []Mention{}
	// cands := sortByRuneLenDesc(names 的键)                        // ①
	// for i, r := range []rune(text) {
	//     if r != '@' { continue }
	//     window := runes(i+1 : min(len, i+1+maxNameLen))            // ② 逐候选最长匹配
	//     if hit { hits = append(hits, {name, byName[id], i, i+1+len(name)}); i += len(name) }
	//     else { 保留字匹配,见 ③ }
	// }
	// return hits
	return nil // TODO(实现):见函数注释 ①~④
}

// 注:实现顺序建议——先在 tests 子目录(仿既有 utils 测试形态)准备样例文本:
//   中文名("沐沐")/英文名(mumu)/长名优先(mumu vs mumu2)/标点边界/重复 @/@所有人;
//   全部通过后按注释 ①~④ 落地,并把命中结果写回 Message.Mentions(见 server 发送链路)。
