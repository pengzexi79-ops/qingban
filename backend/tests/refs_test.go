package tests

// 目标:对"纯逻辑层"做第一批驱动测试(红→绿):
//   utils(refs 引用解析 / 密钥盒 / id)、common(分页 / 错误码映射)、
//   core(SSE 事件总线 / 幂等表)、AI(群聊选人纯函数)。
// 运行:go test ./tests(模块根 backend,包路径 qingban/tests)。
// 说明:server/model 依赖后续引入 gorm/zap 后逐步加入;本包只 import 已可编译的包。

import (
	"encoding/json"
	"strings"
	"testing"

	"qingban/model"
	"qingban/utils"
)

func TestParseRefs_Mixed(t *testing.T) {
	content := "看图 ![设计稿](file-img-1) 和附件 [报告.pdf](file-doc-2),再看 ![照片](file-img-3)"
	refs, plain, err := utils.ParseRefs(content)
	if err != nil {
		t.Fatalf("ParseRefs 不应报错: %v", err)
	}
	if len(refs) != 3 {
		t.Fatalf("期望 3 条引用,得到 %d", len(refs))
	}
	wantKind := []string{"image", "file", "image"}
	wantFile := []string{"file-img-1", "file-doc-2", "file-img-3"}
	wantName := []string{"设计稿", "报告.pdf", "照片"}
	for i, r := range refs {
		if r.Kind != wantKind[i] {
			t.Errorf("refs[%d].Kind=%s 期望 %s", i, r.Kind, wantKind[i])
		}
		if r.FileID != wantFile[i] {
			t.Errorf("refs[%d].FileID=%s 期望 %s", i, r.FileID, wantFile[i])
		}
		if r.FileName != wantName[i] {
			t.Errorf("refs[%d].FileName=%s 期望 %s", i, r.FileName, wantName[i])
		}
	}
	if want := "看图  和附件 ,再看 "; plain != want {
		t.Errorf("plain=%q 期望 %q", plain, want)
	}
}

func TestParseRefs_ImageNoAlt(t *testing.T) {
	// alt 为空时 FileName 回落为 fileId(引用可渲染)
	refs, plain, err := utils.ParseRefs("见 ![](file-img-x)")
	if err != nil {
		t.Fatalf("不应报错: %v", err)
	}
	if len(refs) != 1 || refs[0].Kind != "image" || refs[0].FileID != "file-img-x" {
		t.Fatalf("解析异常: %+v", refs)
	}
	if refs[0].FileName != "file-img-x" {
		t.Errorf("FileName 应回落 fileId,得到 %q", refs[0].FileName)
	}
	if plain != "见 " {
		t.Errorf("plain=%q", plain)
	}
}

func TestParseRefs_Limits(t *testing.T) {
	// 图片超 9 → 报错
	var manyImg string
	for i := 0; i < 10; i++ {
		manyImg += "![p](f-i)"
	}
	if _, _, err := utils.ParseRefs(manyImg); err == nil {
		t.Error("10 张图片应报错")
	}

	// 附件超 20 → 报错
	var manyFile string
	for i := 0; i < 21; i++ {
		manyFile += "[f](f-d)"
	}
	if _, _, err := utils.ParseRefs(manyFile); err == nil {
		t.Error("21 个附件应报错")
	}

	// content >5000 → 报错
	long := strings.Repeat("a", 5001)
	if _, _, err := utils.ParseRefs(long); err == nil {
		t.Error("content 超 5000 应报错")
	}
}

func TestParseRefs_PlainTextLimit(t *testing.T) {
	// 纯文本(去引用标记)>500 → 报错;引用标记本身不计入
	text := strings.Repeat("好", 499)
	content := text + "![图](file-x)" + strings.Repeat("啊", 2) // 纯文本 501>500
	if _, _, err := utils.ParseRefs(content); err == nil {
		t.Error("去标记纯文本 501 字应报错")
	}
	contentOK := text + "![图](file-x)" + "啊" // 纯文本 500,标记不计入 → 应通过
	if _, _, err := utils.ParseRefs(contentOK); err != nil {
		t.Errorf("499+1 字+标记应通过,得到 %v", err)
	}
}

func TestStripRefs(t *testing.T) {
	got := utils.StripRefs("a ![图](f1) b [名](f2) c")
	if got != "a  b  c" {
		t.Errorf("StripRefs=%q", got)
	}
}

func TestParseRefs_DuplicateFileAllowed(t *testing.T) {
	refs, _, err := utils.ParseRefs("![a](f-x)![b](f-x)")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 2 {
		t.Errorf("重复引用应各自成条(渲染层去重),得到 %d 条", len(refs))
	}
}

func TestMessageRef_JSONShape(t *testing.T) {
	// 契约字段名回显检查(json tag 不得随意改名,前端按此渲染)
	refs, _, err := utils.ParseRefs("![设计稿](f1)")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(refs[0])
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	_ = json.Unmarshal(data, &m)
	for _, k := range []string{"kind", "fileId", "fileName"} {
		if _, ok := m[k]; !ok {
			t.Errorf("MessageRef 缺少 json 字段 %q: %s", k, data)
		}
	}
	_ = model.MessageRef{} // 显式引用 model,保证实体编译被本包测试覆盖
}
