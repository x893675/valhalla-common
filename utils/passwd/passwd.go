package passwd

import (
	"golang.org/x/crypto/bcrypt"
)

func EncryptPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func IsPasswordMatch(encodePW string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encodePW), []byte(password))
	return err == nil
}
