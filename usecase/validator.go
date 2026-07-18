package usecase

import (
	"errors"
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func stringRequiredRule(name string) validation.Rule {
	return validation.By(func(value any) error {
		title, _ := value.(string)
		if strings.TrimSpace(title) == "" {
			msg := fmt.Sprintf("%s is required", name)
			return errors.New(msg)
		}
		return nil
	})
}
