package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS returns a middleware that handles OPTIONS requests and sets the CORS
// headers on the response.
func CORS(allowOrigins string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: allowOrigins,
		AllowHeaders: "Origin,Authorization", // CORS-safelisted headers omitted
		AllowMethods: "DELETE,GET,OPTIONS,PATCH,POST,PUT",
	})
}
