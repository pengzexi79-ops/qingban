// tests/utils:验证 utils 包纯函数可用性(引用解析/密钥盒/id 生成/掩码)。
// 运行:在 D:\开源项目\青伴\tests 下执行 `go run ./utils`。
// 断言失败会打印 [FAIL] 并以非零码退出;全部通过打印 [PASS]。

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"qingban/utils"
)

// fails:累计失败数。
var fails int

// check:断言辅助;cond 为 false 时记失败。
func check(name string, cond bool, detail string) {
	if cond {
		fmt.Printf("[PASS] %s\n", name)
	} else {
		fails++
		fmt.Printf("[FAIL] %s: %s\n", name, detail)
	}
}

func main() {
	// ---- ParseRefs:混排(图片+附件),返回文件主键序列与纯文本 ----
	ids, plain, err := utils.ParseRefs("看图 ![设计稿](12) 和附件 [报告.pdf](34)")
	check("refs.mixed.len", err == nil && len(ids) == 2 && ids[0] == 12 && ids[1] == 34, fmt.Sprint(err, ids))
	check("refs.mixed.plain", plain == "看图  和附件 ", fmt.Sprintf("plain=%q", plain))

	// ---- ParseRefs:标记内容非纯数字 id 视为普通文本(不解析、不报错) ----
	r, _, e2 := utils.ParseRefs("见 ![](file-x) 或 ![图](abc)")
	check("refs.nonNumeric.skipped", e2 == nil && len(r) == 0, fmt.Sprint(r, e2))

	// ---- ParseRefs:上限(9 图 / 20 附件 / 5000 内容 / 500 纯文本) ----
	var imgs, files string
	for i := 0; i < 10; i++ {
		imgs += fmt.Sprintf("![p](%d)", i)
	}
	_, _, e3 := utils.ParseRefs(imgs)
	check("refs.limit.img10", e3 != nil, "10 张图片应报错")
	for i := 0; i < 21; i++ {
		files += fmt.Sprintf("[f](%d)", i)
	}
	_, _, e4 := utils.ParseRefs(files)
	check("refs.limit.file21", e4 != nil, "21 个附件应报错")
	_, _, e5 := utils.ParseRefs(strings.Repeat("a", 5001))
	check("refs.limit.len5000", e5 != nil, "超 5000 应报错")
	text := strings.Repeat("好", 499)
	_, _, e6 := utils.ParseRefs(text + "![图](1)" + strings.Repeat("啊", 2)) // 纯文本 501
	check("refs.limit.plain501", e6 != nil, "纯文本 501 应报错(500 上限按字符)")
	_, _, e7 := utils.ParseRefs(text + "![图](1)" + "啊") // 纯文本 500
	check("refs.limit.plain500", e7 == nil, fmt.Sprint(e7))

	// ---- ParseRefs:重复引用同一文件允许(落库去重) ----
	r2, _, _ := utils.ParseRefs("![a](7)![b](7)")
	check("refs.duplicate", len(r2) == 2, fmt.Sprint(len(r2)))

	// ---- StripRefs(仅图片/附件数字标记被剥离) ----
	check("stripRefs", utils.StripRefs("a ![图](1) b [名](2) c") == "a  b  c",
		utils.StripRefs("a ![图](1) b [名](2) c"))

	// ---- SecretBox:加密往返 + 重载同钥 ----
	dir, _ := os.MkdirTemp("", "qingban-utils-key")
	defer os.RemoveAll(dir)
	box, err := utils.LoadOrCreateBox(dir)
	if err != nil {
		check("secret.create", false, err.Error())
	} else {
		enc, _ := box.Encrypt("sk-abcdef123456")
		dec, dErr := box.Decrypt(enc)
		check("secret.roundtrip", dErr == nil && dec == "sk-abcdef123456", fmt.Sprint(dErr, dec))
		box2, bErr := utils.LoadOrCreateBox(dir)
		dec2, d2 := box2.Decrypt(enc)
		check("secret.reload", bErr == nil && d2 == nil && dec2 == "sk-abcdef123456", fmt.Sprint(bErr, d2))
		otherDir, _ := os.MkdirTemp("", "qingban-utils-other")
		defer os.RemoveAll(otherDir)
		other, oErr := utils.LoadOrCreateBox(otherDir)
		check("secret.other.create", oErr == nil, fmt.Sprint(oErr))
		_, cross := other.Decrypt(enc)
		check("secret.crossKey", cross != nil, "错误主密钥应解密失败")
	}

	// ---- MaskKey ----
	check("maskKey.long", utils.MaskKey("sk-abcdefgh1234abcd") == "sk-a****abcd", utils.MaskKey("sk-abcdefgh1234abcd"))
	check("maskKey.short", utils.MaskKey("abc") == "***", utils.MaskKey("abc"))

	// ---- UUID4:格式(版本 4 + variant)/ 唯一性 ----
	pat := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	allOK := true
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		id := utils.UUID4()
		if !pat.MatchString(id) || seen[id] {
			allOK = false
			break
		}
		seen[id] = true
	}
	check("uuid4.format.uniqueness", allOK, "UUID4 格式/唯一性检查失败")

	// ---- PrefixedID ----
	pid := utils.PrefixedID("message")
	check("prefixedID", strings.HasPrefix(pid, "message-") && pat.MatchString(strings.TrimPrefix(pid, "message-")), pid)

	if fails > 0 {
		fmt.Printf("\n%d 项失败\n", fails)
		os.Exit(1)
	}
	fmt.Println("\nutils:全部可用性检查通过")
}
