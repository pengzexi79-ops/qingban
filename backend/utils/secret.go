package utils

// 密钥盒:API Key 等敏感字符串的本地加密存储(AES-256-GCM,主密钥落数据目录 secret.key)。
// 原则(架构文档):明文只在创建/更新那一次 HTTP 请求出现;导出/备份/列表永不返回明文,
// 只返回 secretConfigured 与 maskedKey。
// 生产加固方向(属于桌面端本地能力选型,当前阶段不实现;待引入 Wails 原生能力时评估):
// Windows DPAPI / macOS Keychain / Linux libsecret。当前"数据目录 secret.key"已满足
// 本地单用户威胁模型(防明文文件散步),升级路径不变、不影响任何调用方。

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// secretKeyFile:主密钥文件名(位于数据目录)。
const secretKeyFile = "secret.key"

// secretKeyBytes:AES-256 主密钥长度。
const secretKeyBytes = 32

// SecretBox:主密钥容器(进程内存持有,不序列化)。
type SecretBox struct {
	key []byte // 32 字节 AES-256 主密钥
}

// LoadOrCreateBox:从数据目录装载主密钥(不存在则生成并落盘,权限 0600)。
// 调用点:init.NewApp()(早于任何 API 配置读写)。
func LoadOrCreateBox(dataDir string) (*SecretBox, error) {
	path := filepath.Join(dataDir, secretKeyFile)
	raw, err := os.ReadFile(path)
	if err == nil {
		key, decErr := decodeKey(raw)
		if decErr != nil {
			return nil, decErr
		}
		return &SecretBox{key: key}, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	// 首次:生成 32B 随机密钥落盘
	key := make([]byte, secretKeyBytes)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(base64.StdEncoding.EncodeToString(key)), 0o600); err != nil {
		return nil, err
	}
	return &SecretBox{key: key}, nil
}

// decodeKey:解析 secret.key 内容(base64 → 32 字节)。
func decodeKey(raw []byte) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, errors.New("密钥文件损坏:不是合法 base64")
	}
	if len(key) != secretKeyBytes {
		return nil, errors.New("密钥文件损坏:长度非 32 字节")
	}
	return key, nil
}

// Encrypt:加密明文 → base64 密文(格式:12 字节随机 nonce ‖ 密文)。
// 使用点:ApiProfile 保存 apiKey。
func (b *SecretBox) Encrypt(plain string) (string, error) {
	block, err := aes.NewCipher(b.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt:解密 base64 密文 → 明文(仅进程内使用,不入任何响应)。
// 失败(密钥文件已更换等)→ 返回错误,调用方提示重新录入密钥。
func (b *SecretBox) Decrypt(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("密文损坏:不是合法 base64")
	}
	block, err := aes.NewCipher(b.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("密文损坏:长度不足")
	}
	nonce, sealed := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", errors.New("解密失败(密钥可能已更换),请重新录入密钥")
	}
	return string(plain), nil
}

// MaskKey:展示用掩码,如 "sk-****abcd"(前 4 + **** + 末 4;长度 ≤8 整段打星)。
// 使用点:ApiProfile 的 maskedKey 字段。
func MaskKey(plain string) string {
	r := []rune(plain)
	if len(r) <= 8 {
		return strings.Repeat("*", len(r))
	}
	return string(r[:4]) + "****" + string(r[len(r)-4:])
}
