package v0

import "github.com/gofiber/fiber/v2"

func BadRequest(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "request body is not a valid JSON string",
	})
}

func NotFound(c *fiber.Ctx, resourceName string, detail string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"errors": fiber.Map{
			resourceName: detail,
		},
	})
}

func Unauthorized(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusUnauthorized)
}

func UnprocessableEntity(c *fiber.Ctx, entityErrors fiber.Map) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"errors": entityErrors,
	})
}
