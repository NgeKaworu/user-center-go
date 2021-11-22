package validator

import (
	"reflect"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// NewValidator 工厂方法
func NewValidator() *validator.Validate {
	return validator.New()
}

// NewValidatorTranslator 工厂方法
func NewValidatorTranslator(v *validator.Validate) *ut.Translator {
	uni := ut.New(zh.New())
	trans, _ := uni.GetTranslator("zh")

	//注册一个函数，获取struct tag里自定义的label作为字段名
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("label")
	})

	//注册翻译器
	err := zh_translations.RegisterDefaultTranslations(v, trans)
	if err != nil {
		panic(err)
	}

	return &trans

}
