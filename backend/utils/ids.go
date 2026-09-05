package utils

// ID 生成工具(与全局无关的纯工具)。
// 约定(PHASE1_API.md §2):ID 一律 uuid4 短横线字符串,允许带可读前缀(companion-/group-/conv-…),
// 前缀只用于日志/调试可读性,id 本身不承载业务含义。

import (
	"crypto/rand"
	"encoding/hex"
)

// UUID4:生成一条标准 uuid4(36 字符,带短横线)。
// 实现:读取 16 随机字节,置 version=0100、variant=10xx 位后按 8-4-4-4-12 拼装。
// 说明:优先"计划依赖 github.com/google/uuid",当前占位实现与之一致且无外部依赖。
func UUID4() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// 理论不可达(crypto/rand 失败时 panic 防静默重复 id)
		panic("uuid: random read failed: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	raw := hex.EncodeToString(b[:])
	// 拼装 8-4-4-4-12(前缀可读、整体仍为 uuid4 形式)
	return raw[0:8] + "-" + raw[8:12] + "-" + raw[12:16] + "-" + raw[16:20] + "-" + raw[20:32]
}

// PrefixedID:带业务前缀的 id(如 "message-" + uuid4)。
// 参数 prefix:业务名(companion/group/conversation/message/memory/file/usage/round/profile/…)。
// 使用场景:库表主键生成;前缀便于日志与导出数据肉眼定位归属类型。
func PrefixedID(prefix string) string {
	return prefix + "-" + UUID4()
}
