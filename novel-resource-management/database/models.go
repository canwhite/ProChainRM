package database

import (
	// 不再需要 primitive 包
)

// Novel 与链码中的 Novel 结构体保持一致
type Novel struct {
	ID           string `bson:"_id,omitempty" json:"id"`           // 改为string类型，与链码一致
	Author       string `bson:"author,omitempty" json:"author,omitempty"`
	StoryOutline string `bson:"storyOutline,omitempty" json:"storyOutline,omitempty"`
	Subsections  string `bson:"subsections,omitempty" json:"subsections,omitempty"`
	Characters   string `bson:"characters,omitempty" json:"characters,omitempty"`
	Items        string `bson:"items,omitempty" json:"items,omitempty"`
	TotalScenes  string `bson:"totalScenes,omitempty" json:"totalScenes,omitempty"`
	CreatedAt    string `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt    string `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

// UserCredit 与链码中的 UserCredit 结构体保持一致
type UserCredit struct {
	ID            string `bson:"_id,omitempty" json:"id"`           // 改为string类型，与链码一致
	UserID        string `bson:"userId" json:"userId"`
	Credit        int    `bson:"credit" json:"credit"`
	TotalUsed     int    `bson:"totalUsed" json:"totalUsed"`
	TotalRecharge int    `bson:"totalRecharge" json:"totalRecharge"`
	CreatedAt     string `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt     string `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

// CreditHistory 与链码中的 CreditHistory 结构体保持一致
type CreditHistory struct {
	ID          string `bson:"_id,omitempty" json:"id"`               // 改为string类型，与链码一致
	UserID      string `bson:"userId" json:"userId"`
	Amount      int    `bson:"amount" json:"amount"`                   //积分变动的数额
	Type        string `bson:"type" json:"type"`                     // "consume", "recharge", "reward"
	Description string `bson:"description" json:"description"`
	Timestamp   string `bson:"timestamp" json:"timestamp"`
	NovelID     string `bson:"novelId,omitempty" json:"novelId,omitempty"`
}

// User MongoDB users 集合的结构体
type User struct {
	ID                string   `bson:"_id,omitempty" json:"id"`
	Email             string   `bson:"email" json:"email"`
	Username          string   `bson:"username" json:"username"`
	PasswordHash      string   `bson:"passwordHash" json:"passwordHash"`
	DeviceFingerprint string   `bson:"deviceFingerprint,omitempty" json:"deviceFingerprint,omitempty"`
	IsActive          bool     `bson:"isActive" json:"isActive"`
	Role              string   `bson:"role" json:"role"`
	NovelIds          []string `bson:"novelIds,omitempty" json:"novelIds,omitempty"`
	CreatedAt         string   `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt         string   `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}