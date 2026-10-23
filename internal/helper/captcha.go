package helper

import (
	"strings"

	base64Captcha "github.com/mojocn/base64Captcha"
)

var loginCaptchaStore = base64Captcha.DefaultMemStore

func GenerateLoginCaptcha() (string, string, error) {
	driver := base64Captcha.NewDriverString(
		48,
		180,
		2,
		base64Captcha.OptionShowSineLine|
			base64Captcha.OptionShowSlimeLine,
		5,
		"23456789abcdefghjkmnpqrstuvwxyz",
		nil,
		nil,
		nil,
	)

	generator := base64Captcha.NewCaptcha(
		driver,
		loginCaptchaStore,
	)

	id, image, _, err := generator.Generate()
	if err != nil {
		return "", "", err
	}

	return id, image, nil
}

func VerifyLoginCaptcha(id, answer string) bool {
	id = strings.TrimSpace(id)
	answer = strings.ToLower(strings.TrimSpace(answer))

	if id == "" || answer == "" {
		return false
	}

	return loginCaptchaStore.Verify(id, answer, true)
}
