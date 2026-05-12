package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost matches bcryptjs default (10) so re-hashing during password change stays consistent.
const bcryptCost = 10

func HashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func VerifyPassword(plaintext, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}
