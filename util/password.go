package util

import "encoding/base64"

func EncryptPassword(phoneNumber string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(phoneNumber))
	return encoded
}
