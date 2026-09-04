package dao

func GenCaptchaPrefix(email string) string {
	return "verify:" + email
}
