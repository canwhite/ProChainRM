package database

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserCredit 用户积分模型
type UserCredit struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	Credit        int                `bson:"credit" json:"credit"`
	TotalUsed     int                `bson:"total_used" json:"total_used"`
	TotalRecharge int                `bson:"total_recharge" json:"total_recharge"`
	IsActive      bool               `bson:"is_active" json:"is_active"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// Novel 小说模型
type Novel struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Author      string             `bson:"author" json:"author"`
	Category    string             `bson:"category" json:"category"`
	Content     string             `bson:"content" json:"content"`
	Description string             `bson:"description" json:"description"`
	Tags        []string           `bson:"tags" json:"tags"`
	Price       float64            `bson:"price" json:"price"`
	IsPublished bool               `bson:"is_published" json:"is_published"`
	ViewCount   int                `bson:"view_count" json:"view_count"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// UserNovelPurchase 用户小说购买记录
type UserNovelPurchase struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID   string             `bson:"user_id" json:"user_id"`
	NovelID  string             `bson:"novel_id" json:"novel_id"`
	Price    float64            `bson:"price" json:"price"`
	PaidAt   time.Time          `bson:"paid_at" json:"paid_at"`
	Status   string             `bson:"status" json:"status"` // "completed", "pending", "failed"
}

// UserActivity 用户活动日志
type UserActivity struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	Action    string             `bson:"action" json:"action"` // "login", "purchase", "read"
	TargetID  string             `bson:"target_id" json:"target_id"`
	TargetType string            `bson:"target_type" json:"target_type"` // "novel", "user"
	Metadata  map[string]interface{} `bson:"metadata" json:"metadata"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}