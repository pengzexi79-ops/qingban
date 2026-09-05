package tests

// utils 密钥盒与 id 生成测试(纯标准库实现,可直接跑绿)。

import (
	"regexp"
	"strings"
	"testing"

	"qingban/utils"
)

func TestSecretBox_RoundTrip(t *testing.T) {
	box, err := utils.LoadOrCreateBox(t.TempDir())
	if err != nil {
		t.Fatalf("创建密钥盒失败: %v", err)
	}
	secret := "sk-abcdef1234567890"
	enc, err := box.Encrypt(secret)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if enc == secret {
		t.Error("密文不应等于明文")
	}
	dec, err := box.Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if dec != secret {
		t.Errorf("往返不一致: got %q want %q", dec, secret)
	}
}

func TestSecretBox_ReloadSameKey(t *testing.T) {
	dir := t.TempDir()
	b1, _ := utils.LoadOrCreateBox(dir)
	enc, _ := b1.Encrypt("k1-secret")

	// 第二次装载必须读到同一把密钥,能解开同一段密文
	b2, err := utils.LoadOrCreateBox(dir)
	if err != nil {
		t.Fatalf("重载密钥盒失败: %v", err)
	}
	dec, err := b2.Decrypt(enc)
	if err != nil || dec != "k1-secret" {
		t.Errorf("跨装载解密失败: dec=%q err=%v", dec, err)
	}
}

func TestSecretBox_CrossKeyFail(t *testing.T) {
	b1, _ := utils.LoadOrCreateBox(t.TempDir())
	b2, _ := utils.LoadOrCreateBox(t.TempDir()) // 另一个数据目录 = 另一把主密钥
	enc, _ := b1.Encrypt("secret-a")
	if _, err := b2.Decrypt(enc); err == nil {
		t.Error("用错误密钥解密应失败(GCM 认证拦截)")
	}
}

func TestMaskKey(t *testing.T) {
	cases := []struct{ in, want string }{
		{"sk-abcdefgh1234abcd", "sk-a****abcd"},
		{"abc", "***"},
		{"12345678", "********"}, // 恰好 8 位整段打星
		{"", ""},
	}
	for _, c := range cases {
		if got := utils.MaskKey(c.in); got != c.want {
			t.Errorf("MaskKey(%q)=%q 期望 %q", c.in, got, c.want)
		}
	}
}

func TestUUID4_Format(t *testing.T) {
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	for i := 0; i < 100; i++ {
		id := utils.UUID4()
		if !pattern.MatchString(id) {
			t.Fatalf("UUID4 格式非法(版本 4 / variant 校验): %q", id)
		}
	}
}

func TestPrefixedID(t *testing.T) {
	id := utils.PrefixedID("message")
	if !strings.HasPrefix(id, "message-") {
		t.Errorf("缺少前缀: %q", id)
	}
	// 前缀后必须是合法 uuid4
	if !regexp.MustCompile(`^message-[0-9a-f-]{36}$`).MatchString(id) {
		t.Errorf("前缀后段非法: %q", id)
	}
}

func TestUUID4_Uniqueness(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		id := utils.UUID4()
		if seen[id] {
			t.Fatalf("重复 UUID4: %q", id)
		}
		seen[id] = true
	}
}
