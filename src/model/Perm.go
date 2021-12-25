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
	IsMenu     *bool   `json:"isMenu,omitempty" bson:"isMenu,omitempty" validate:"required"`          // 是否菜单
	IsHide     *bool   `json:"isHide,omitempty" bson:"isHide,omitempty"`                              // 是否不在菜单中显视
	IsMicroApp *bool   `json:"isMicroApp,omitempty" bson:"isMicroApp,omitempty"`                      // 是否微应用入口
	PID        *string `json:"pID,omitempty" bson:"pID,omitempty"`                                    // 父级id
	Url        *string `json:"url,omitempty" bson:"url,omitempty" validate:"required_if=IsMenu true"` // url
	Icon       *string `json:"icon,omitempty" bson:"icon,omitempty" `                                 // icon
}
