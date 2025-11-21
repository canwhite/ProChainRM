package database

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Novel 与链码中的 Novel 结构体保持一致
type Novel struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Author       string             `bson:"author,omitempty" json:"author,omitempty"`
	StoryOutline string             `bson:"storyOutline,omitempty" json:"storyOutline,omitempty"`
	Subsections  string             `bson:"subsections,omitempty" json:"subsections,omitempty"`
	Characters   string             `bson:"characters,omitempty" json:"characters,omitempty"`
	Items        string             `bson:"items,omitempty" json:"items,omitempty"`
	TotalScenes  string             `bson:"totalScenes,omitempty" json:"totalScenes,omitempty"`
	CreatedAt    string             `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt    string             `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

// UserCredit 与链码中的 UserCredit 结构体保持一致
type UserCredit struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        string             `bson:"userId" json:"userId"`
	Credit        int                `bson:"credit" json:"credit"`
	TotalUsed     int                `bson:"totalUsed" json:"totalUsed"`
	TotalRecharge int                `bson:"totalRecharge" json:"totalRecharge"`
	CreatedAt     string             `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt     string             `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

// CreditHistory 与链码中的 CreditHistory 结构体保持一致
type CreditHistory struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"userId" json:"userId"`
	Amount      int                `bson:"amount" json:"amount"`       //积分变动的数额
	Type        string             `bson:"type" json:"type"`         // "consume", "recharge", "reward"
	Description string             `bson:"description" json:"description"`
	Timestamp   string             `bson:"timestamp" json:"timestamp"`
	NovelID     string             `bson:"novelId,omitempty" json:"novelId,omitempty"`
}