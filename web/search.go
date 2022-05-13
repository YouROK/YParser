package web

import (
	"YParser/db"
	"github.com/gofiber/fiber/v2"
	"log"
	"net/url"
	"strings"
)

func search(c *fiber.Ctx) error {
	query := c.Params("query")
	query, err := url.PathUnescape(query)
	if err != nil {
		return err
	}

	torrs := db.Search("*" + query + "*")

	if len(torrs) == 0 {
		query = "*" + strings.ReplaceAll(query, " ", "*") + "*"
		query = strings.ReplaceAll(query, "**", "*")
		torrs = db.Search(query)
	}

	log.Println("Query:", query, "| Torrents:", len(torrs))

	err = c.JSON(torrs)
	c.Set("Content-type", "application/json; charset=utf-8")
	return err
}
