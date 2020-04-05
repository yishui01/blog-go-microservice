package utils

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func GeneratePassword(userPass string) ([]byte, error) {
	s, err := bcrypt.GenerateFromPassword([]byte(userPass), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "generate err:")
	}
	return s, nil
}

//ValidatePassword 密码比对
func ValidatePassword(userPassword string, hashed string) error {
	return errors.WithStack(bcrypt.CompareHashAndPassword([]byte(hashed), []byte(userPassword)))
}
