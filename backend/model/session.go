package model

import "gorm.io/gorm"

//会话管理的表

//单人-ai对话
type SingleSession struct {
	gorm.Model
	Name        string      //会话名称
	ModelConfig ModelConfig `json:"model_config"` //模型配置信息
	//最后一条消息
}

// 群聊
type GroupSession struct {
	gorm.Model
}

//在群聊的中间表
type SesionTOPerson struct {
	gorm.Model
	GroupID int8 //群内隐藏id消息未读的功能基于此
}
