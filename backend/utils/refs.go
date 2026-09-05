package utils

// 消息内容里的"文件引用标记"解析工具(纯函数,单测目标)。
// 语法:图片 ![alt](fileID)、附件 [文件名](fileID);fileID 为 files 表主键(十进制数字)。
// 双轨附件模型:①content 原样保存标记文本(前端按语法渲染);②解析出的 id 供
// Message.Files 多对多落库关联。content 与解析结果幂等可重建;重复引用同一 id 允许(落库去重)。
// 注意:标记内 id 必须为纯数字(自增 uint 主键);非数字视为普通文本不解析。

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode/utf8"
)

// 上限常量(单条消息引用数,与前端契约一致):
const (
	// MaxImageRefsPerMsg:图片引用上限(9 张)。
	MaxImageRefsPerMsg = 9
	// MaxFileRefsPerMsg:附件引用上限(20 个)。
	MaxFileRefsPerMsg = 20
	// MaxMsgContentLen:content 最大长度(5000)。
	MaxMsgContentLen = 5000
	// MaxUserTextLen:纯文本(去引用标记后)用户输入上限(500)。
	MaxUserTextLen = 500
)

// refPattern:引用标记正则。形如 ![alt](123) 或 [名字](456)。
// 分组:1=图片 alt,2=图片 id,3=附件名字,4=附件 id。
var refPattern = regexp.MustCompile(`!\[([^\]]*)\]\(([0-9]+)\)|\[([^\]]+)\]\(([0-9]+)\)`)

// ParseRefs:解析 content 的引用标记。
// 返回:
//
//	ids:标记内文件主键(按出现顺序;图片与附件混排按文本顺序)。
//	plain:去除全部标记后的纯文本。
//	err:content>5000 字 / 纯文本>500 字 / 图片>9 / 附件>20(按字符计数,非字节)。
func ParseRefs(content string) (ids []uint64, plain string, err error) {
	if utf8.RuneCountInString(content) > MaxMsgContentLen {
		return nil, "", fmt.Errorf("内容超过 %d 字上限", MaxMsgContentLen)
	}

	locs := refPattern.FindAllStringSubmatchIndex(content, -1)
	var sb []byte // 手工拼接纯文本,避免重复 alloc
	prev := 0
	imgCnt, fileCnt := 0, 0
	for _, loc := range locs {
		fullStart, fullEnd := loc[0], loc[1]
		sb = append(sb, content[prev:fullStart]...)
		prev = fullEnd

		var idStr string
		if loc[2] >= 0 {
			// 图片 ![alt](id):组1 alt、组2 id
			imgCnt++
			if imgCnt > MaxImageRefsPerMsg {
				return nil, "", fmt.Errorf("图片超过 %d 张上限", MaxImageRefsPerMsg)
			}
			idStr = content[loc[4]:loc[5]]
		} else {
			// 附件 [名字](id):组3 名字、组4 id
			fileCnt++
			if fileCnt > MaxFileRefsPerMsg {
				return nil, "", fmt.Errorf("附件超过 %d 个上限", MaxFileRefsPerMsg)
			}
			idStr = content[loc[8]:loc[9]]
		}
		id, convErr := strconv.ParseUint(idStr, 10, 64)
		if convErr != nil {
			return nil, "", fmt.Errorf("引用标记含非法文件 id: %q", idStr)
		}
		ids = append(ids, id)
	}
	sb = append(sb, content[prev:]...)

	plain = string(sb)
	if utf8.RuneCountInString(plain) > MaxUserTextLen {
		return nil, "", fmt.Errorf("纯文本超过 %d 字上限(不含引用标记)", MaxUserTextLen)
	}
	return ids, plain, nil
}

// StripRefs:仅做"去除引用标记"的轻量处理(历史摘要/搜索/字数统计用;不校验数量)。
func StripRefs(content string) string {
	return refPattern.ReplaceAllString(content, "")
}
