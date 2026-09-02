package response

import "github.com/labstack/echo/v4"

type Meta struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type Envelope struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data,omitempty"`
}

func Success(c echo.Context, code int, message string, data interface{}) error {
	return c.JSON(code, Envelope{
		Meta: Meta{Success: true, Message: message},
		Data: data,
	})
}

func Error(c echo.Context, code int, message string) error {
	return c.JSON(code, Envelope{
		Meta: Meta{Success: false, Message: message},
	})
}
