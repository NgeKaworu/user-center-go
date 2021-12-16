package model

import (
	"time"
)

// TRole 角色表
const TRole = "t_role"

// undo init
// Role 角色schema
type Role struct {
	ID       *string    `json:"id,omitempty" bson:"_id,omitempty"`                        // id
	Name     *string    `json:"name,omitempty" bson:"name,omitempty" validtor:"required"` // 角色名
	CreateAt *time.Time `json:"createAt,omitempty" bson:"createAt,omitempty"`             // 创建时间
	UpdateAt *time.Time `json:"updateAt,omitempty" bson:"updateAt,omitempty"`             // 更新时间
	Perms    []*string  `json:"perms,omitempty" bson:"perms,omitempty"`                   // 权限列表
}
