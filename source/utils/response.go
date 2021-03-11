package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type successResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data,omitempty"`
}

type validationErrorResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Errors  interface{} `json:"errors"`
}

type errorResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

//LogError ...
func LogError(funcName string, err error) {
	logrus.WithFields(logrus.Fields{
		"function": funcName,
	}).Error(err)
}

//SendResponse with success response
func SendResponse(c *fiber.Ctx, code int, data interface{}) error {
	return c.Status(code).JSON(successResponse{
		Success: true,
		Code:    code,
		Data:    data,
	})
}

//SendValidationError ...
func SendValidationError(c *fiber.Ctx, errors interface{}) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(validationErrorResponse{
		Success: false,
		Code:    fiber.StatusUnprocessableEntity,
		Errors:  errors,
	})
}

//SendError send error response
func SendError(c *fiber.Ctx, code int, message string) error {
	//if message in empty
	if message == "" {
		switch code {
		case fiber.StatusNotFound:
			message = "not found"
		case fiber.StatusInternalServerError:
			message = "internal server error"
		case fiber.StatusBadRequest:
			message = "bad request"
		case fiber.StatusConflict:
			message = "conflict"
		case fiber.StatusUnauthorized:
			message = "unauthorized"
		case fiber.StatusForbidden:
			message = "forbidden"
		case fiber.StatusUnprocessableEntity:
			message = "unprocessable entity"
		case fiber.StatusServiceUnavailable:
			message = "service unavailable"
		}
	}

	return c.Status(code).JSON(errorResponse{
		Success: false,
		Code:    code,
		Message: message,
	})
}
