package helper

import (
	"image/color"
	"strings"

	base64Captcha "github.com/mojocn/base64Captcha"
)

var loginCaptchaStore = base64Captcha.DefaultMemStore

func GenerateLoginCaptcha() (string, string, error) {
	driver := base64Captcha.NewDriverString(
		60,  // tinggi gambar
		220, // lebar gambar
		0,   // tanpa noise tambahan
		0,   // tanpa garis pengganggu
		5,   // panjang kode
		"23456789abcdefghjkmnpqrstuvwxyz",
		&color.RGBA{
			R: 245,
			G: 247,
			B: 250,
			A: 255,
		},
		nil, // gunakan font bawaan
		nil, // gunakan kumpulan font default
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
