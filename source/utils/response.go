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

//SendResponse with success response
func SendResponse(c *fiber.Ctx, code int, data interface{}) error {
	return c.Status(code).JSON(&successResponse{
		Success: true,
		Code:    code,
		Data:    data,
	})
}

//SendValidationError ...
func SendValidationError(c *fiber.Ctx, errors interface{}) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(&validationErrorResponse{
		Success: false,
		Code:    fiber.StatusUnprocessableEntity,
		Errors:  errors,
	})
}

//LogAndSendError logs the error and send error response
func LogAndSendError(c *fiber.Ctx, code int, function string, err error) error {
	var message string
	//logs the error
	logrus.WithFields(logrus.Fields{
		"function": function,
	}).Error(err)

	switch code {
	case fiber.StatusNotFound:
		message = "Not Found"
	case fiber.StatusInternalServerError:
		message = "Internal Server Error"
	case fiber.StatusBadRequest:
		message = "Bad Request"
	case fiber.StatusConflict:
		message = "Conflict"
	case fiber.StatusUnauthorized:
		message = "Unauthorized"
	case fiber.StatusForbidden:
		message = "Forbidden"
	case fiber.StatusUnprocessableEntity:
		message = "Unprocessable Entity"
	case fiber.StatusServiceUnavailable:
		message = "Service Unavailable"
	}

	return c.Status(code).JSON(&errorResponse{
		Success: false,
		Code:    code,
		Message: message,
	})
}
