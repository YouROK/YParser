package web

import (
	"YParser/config"
	"github.com/gofiber/fiber/v2"
	"os"
	"path/filepath"
)

func WebInit() {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	dir := filepath.Dir(os.Args[0])

	app.Static("/", filepath.Join(dir, "public"))

	app.Get("/search/:query", search)

	app.Listen(config.Config.WebAddr + ":" + config.Config.WebPort)
}
