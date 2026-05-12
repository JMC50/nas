package auth

import (
	"fmt"
	"regexp"

	"github.com/JMC50/nas/internal/config"
)

var (
	upperRE   = regexp.MustCompile(`[A-Z]`)
	lowerRE   = regexp.MustCompile(`[a-z]`)
	digitRE   = regexp.MustCompile(`[0-9]`)
	specialRE = regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)
)

func ValidatePassword(plaintext string, requirements config.PasswordRequirements) error {
	if len(plaintext) < requirements.MinLength {
		return fmt.Errorf("password must be at least %d characters long", requirements.MinLength)
	}
	if requirements.RequireUppercase && !upperRE.MatchString(plaintext) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if requirements.RequireLowercase && !lowerRE.MatchString(plaintext) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if requirements.RequireNumber && !digitRE.MatchString(plaintext) {
		return fmt.Errorf("password must contain at least one number")
	}
	if requirements.RequireSpecial && !specialRE.MatchString(plaintext) {
		return fmt.Errorf("password must contain at least one special character")
	}
	return nil
}
