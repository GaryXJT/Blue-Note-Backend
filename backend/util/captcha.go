package util

import (
	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.DefaultMemStore

type Captcha struct {
	ID     string `json:"captchaId"`
	Base64 string `json:"captchaImage"`
}

func GenerateCaptcha() (*Captcha, error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 6, 0.7, 80)
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64s, _, err := c.Generate()
	if err != nil {
		return nil, err
	}

	return &Captcha{
		ID:     id,
		Base64: b64s,
	}, nil
}

func VerifyCaptcha(id, code string) bool {
	return store.Verify(id, code, true)
}

func CleanExpiredCaptcha() {
	// 不需要这个方法，base64Captcha.DefaultMemStore 会自动清理过期的验证码
}
