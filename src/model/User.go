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
	ID       *primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`                                // id
	Name     *string             `json:"name,omitempty" bson:"name,omitempty" validate:"required"`         // 用户昵称
	Pwd      *string             `json:"pwd,omitempty" bson:"pwd,omitempty" validate:"required,min=8"`     // 密码
	Email    *string             `json:"email,omitempty" bson:"email,omitempty" validate:"required,email"` // 邮箱
	CreateAt *time.Time          `json:"createAt,omitempty" bson:"createAt,omitempty"`                     // 创建时间
	UpdateAt *time.Time          `json:"updateAt,omitempty" bson:"updateAt,omitempty"`                     // 更新时间
	Roles    []string            `json:"roles,omitempty" bson:"roles,omitempty"`                           // 角色列表
}
