package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TUser 用户表
const TUser = "t_user"

// undo init
// User 用户schema
type User struct {
	ID       *primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`            // id
	Name     *string             `json:"name,omitempty" bson:"name,omitempty"`         // 用户昵称
	Pwd      *string             `json:"pwd,omitempty" bson:"pwd,omitempty"`           // 密码
	Email    *string             `json:"email,omitempty" bson:"email,omitempty"`       // 邮箱
	CreateAt *time.Time          `json:"createAt,omitempty" bson:"createAt,omitempty"` // 创建时间
	UpdateAt *time.Time          `json:"updateAt,omitempty" bson:"updateAt,omitempty"` // 更新时间
	IsAdmin  bool                `json:"isAdmin,omitempty" bson:"isAdmin,omitempty"`   // 是否管理员
	Roles    []Role              `json:"roles,omitempty" bson:"roles,omitempty"`       // 角色列表
}
