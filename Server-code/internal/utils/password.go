package utils

import (
	"errors"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func HashPasswordWithCost(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidatePasswordComplexity(password string, minLength int, requireUpper, requireLower, requireDigit, requireSpecial bool) error {
	if len(password) < minLength {
		return errors.New("password too short")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()-_=+[]{}|;:'\",.<>?/`~", ch):
			hasSpecial = true
		}
	}

	if requireUpper && !hasUpper {
		return errors.New("password must contain uppercase letter")
	}
	if requireLower && !hasLower {
		return errors.New("password must contain lowercase letter")
	}
	if requireDigit && !hasDigit {
		return errors.New("password must contain digit")
	}
	if requireSpecial && !hasSpecial {
		return errors.New("password must contain special character")
	}

	return nil
}
