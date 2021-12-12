package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TPerm 权限表
const TPerm = "t_perm"

// undo init

// Perm 权限schema
type Perm struct {
	ID       *primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`                        // id
	Name     *string             `json:"name,omitempty" bson:"name,omitempty" validate:"required"` // 权限名
	Key      *string             `json:"key,omitempty" bson:"key,omitempty" validate:"required"`   // 权限标识
	CreateAt *time.Time          `json:"createAt,omitempty" bson:"createAt,omitempty"`             // 创建时间
	UpdateAt *time.Time          `json:"updateAt,omitempty" bson:"updateAt,omitempty"`             // 更新时间

	// menu
	PID *primitive.ObjectID `json:"pid,omitempty" bson:"pid,omitempty"` // 父id
	Url *string             `json:"url,omitempty" bson:"url,omitempty"` // url
}
