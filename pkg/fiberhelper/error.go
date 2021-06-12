package fiberhelper

import "github.com/gofiber/fiber/v2"

func HandleErrorJSONResp(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(errorJSONResp{
		Message: message,
	})
}

type errorJSONResp struct {
	Message string `json:"message"`
}
