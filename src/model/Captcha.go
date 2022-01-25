package model

type Captcha struct {
	Captcha *string `query:"captcha,omitempty" validate:"required"`
	Email   *string `query:"email,omitempty" validate:"required,email"`
}
