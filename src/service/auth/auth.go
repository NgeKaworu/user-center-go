package auth

// Auth 加解密相关
type Auth struct {
	key []byte
}

// New 工厂方法
func New(k string) *Auth {
	return &Auth{
		key: []byte(k),
	}
}
