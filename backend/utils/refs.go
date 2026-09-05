package utils

// 消息内容里的"文件引用标记"解析工具(纯函数,单测目标)。
// 语法(见 openapi.phase1.yaml 文件头):图片 ![alt](fileId)、附件 [文件名](fileId)。
// 职责:按文本位置抽取引用 → refs 数组;同时给出"去除标记后的纯文本"(供 500 字校验/搜索/截断)。
// content 原样保存,refs 与 content 幂等可重建。重复引用同一 fileId 允许(渲染层去重)。

import (
	"errors"
	"fmt"
	"regexp"

	"qingban/model"
)

// 上限常量(与契约一致,可调):
const (
	// MaxImageRefsPerMsg:单条消息图片引用上限(9 张)。
	MaxImageRefsPerMsg = 9
	// MaxFileRefsPerMsg:单条消息附件引用上限(20 个)。
	MaxFileRefsPerMsg = 20
	// MaxMsgContentLen:content 最大长度(5000)。
	MaxMsgContentLen = 5000
	// MaxUserTextLen:纯文本(去引用标记后)用户输入上限(500)。
	MaxUserTextLen = 500
)

// 引用标记的 kind 取值(与消息 refs 契约一致):
const (
	// refKindImage:图片引用。
	refKindImage = "image"
	// refKindFile:附件引用。
	refKindFile = "file"
)

// refPattern:引用标记正则。形如 ![alt](fileId) 或 [名字](fileId)。
// 分组:1=图片 alt,2=图片 fileId,3=附件名字,4=附件 fileId。
var refPattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)\)|\[([^\]]+)\]\(([^)\s]+)\)`)

// ParseRefs:解析 content 的引用标记。
// 返回 refs(按出现顺序;kind=image|file)、plain(去掉全部标记的纯文本)。
// 错误:content >5000、纯文本(去标记)>500、图片 >9、附件 >20。
func ParseRefs(content string) (refs []model.MessageRef, plain string, err error) {
	if len(content) > MaxMsgContentLen {
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

		ref := model.MessageRef{}
		if loc[2] >= 0 {
			// 图片 ![alt](fileId):组1 alt、组2 fileId
			imgCnt++
			if imgCnt > MaxImageRefsPerMsg {
				return nil, "", fmt.Errorf("图片超过 %d 张上限", MaxImageRefsPerMsg)
			}
			ref.Kind = refKindImage
			alt := content[loc[2]:loc[3]]
			ref.FileID = content[loc[4]:loc[5]]
			ref.FileName = alt
			if ref.FileName == "" {
				ref.FileName = ref.FileID
			}
		} else {
			// 附件 [名字](fileId):组3 名字、组4 fileId
			fileCnt++
			if fileCnt > MaxFileRefsPerMsg {
				return nil, "", fmt.Errorf("附件超过 %d 个上限", MaxFileRefsPerMsg)
			}
			ref.Kind = refKindFile
			ref.FileName = content[loc[6]:loc[7]]
			ref.FileID = content[loc[8]:loc[9]]
		}
		refs = append(refs, ref)
	}
	sb = append(sb, content[prev:]...)

	plain = string(sb)
	if len(plain) > MaxUserTextLen {
		return nil, "", fmt.Errorf("纯文本超过 %d 字上限(不含引用标记)", MaxUserTextLen)
	}
	return refs, plain, nil
}

// StripRefs:仅做"去除引用标记"的轻量处理(历史摘要/搜索/字数统计用;不校验数量)。
func StripRefs(content string) string {
	return refPattern.ReplaceAllString(content, "")
}
