package model

import (
	"time"
)

// TPerm 权限表
const TPerm = "t_perm"

// undo init

// Perm 权限schema
type Perm struct {
	ID       *string    `json:"id,omitempty" bson:"_id,omitempty"`                        // id
	Name     *string    `json:"name,omitempty" bson:"name,omitempty" validate:"required"` // 权限名
	CreateAt *time.Time `json:"createAt,omitempty" bson:"createAt,omitempty"`             // 创建时间
	UpdateAt *time.Time `json:"updateAt,omitempty" bson:"updateAt,omitempty"`             // 更新时间

	// menu
	PID *string `json:"pID,omitempty" bson:"pID,omitempty"` // 父级id
	Url *string `json:"url,omitempty" bson:"url,omitempty"` // url
}
