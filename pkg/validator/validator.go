package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"time"
)

func CustomValidator() (*Validator, error) {
	var err error

	v := NewValidator()
	if err = v.RegisterDefaultRules(); err != nil {
		return nil, err
	}
	if err = v.Register("comma_array", ValidateCommaArray); err != nil {
		return nil, err
	}

	return v, nil
}

// ValidateCommaArray is a custom validation function for validates comma separated array of integers.
func ValidateCommaArray(fl validator.FieldLevel) bool {
	var err error

	ids := fl.Field().String()
	for _, id := range strings.Split(ids, `,`) {
		if _, err = strconv.Atoi(id); err != nil {
			return false
		}
	}

	return true
}

func NewValidator() *Validator {
	return &Validator{validator.New()}
}

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) RegisterDefaultRules() error {
	var err error

	if err = v.validator.RegisterValidation("uuid", ValidateUUID); err != nil {
		return err
	}
	if err = v.validator.RegisterValidation("timestamp", ValidateTimestamp); err != nil {
		return err
	}

	return nil
}

func (v *Validator) Validate(i any) error {
	return v.validator.Struct(i)
}

func (v *Validator) Register(tag string, fn validator.Func) error {
	return v.validator.RegisterValidation(tag, fn)
}

// ValidateUUID is a custom validation function for UUIDs.
func ValidateUUID(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

// ValidateTimestamp is a custom validation function for timestamps.
func ValidateTimestamp(fl validator.FieldLevel) bool {
	_, err := time.Parse(time.RFC3339, fl.Field().String())
	return err == nil
}
