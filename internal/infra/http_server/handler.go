package httpserver

import "github.com/gofiber/fiber/v3"

type Handler interface {
	Register(app *fiber.App) error
}
