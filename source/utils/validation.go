package utils

import "github.com/go-playground/validator/v10"

//ValidationError ...
type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

//ValidateStruct data
func ValidateStruct(data interface{}) []*ValidationError {
	var errors []*ValidationError
	validate := validator.New()
	err := validate.Struct(data)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var error ValidationError
			error.Field = err.StructNamespace()
			error.Tag = err.Tag()
			error.Value = err.Param()
			errors = append(errors, &error)
		}
	}
	return errors
}
