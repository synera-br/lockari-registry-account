package utils

import (
	"errors"
	"net/mail"
)

func IsEmail(email *string) error {
	if email == nil || *email == "" {
		return errors.New("invalid email")
	}

	_, err := mail.ParseAddress(*email)
	if err != nil {
		return errors.New("invalid email")
	}
	return nil
}
