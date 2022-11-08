package api

import (
	mQuasar "github.com/Sanchous98/elementum/migrate/quasar"
	"github.com/gofiber/fiber/v2"
)

// MigratePlugin gin proxy for /migrate/:plugin ...
func MigratePlugin(ctx *fiber.Ctx) error {
	plugin := ctx.Params("plugin")
	if plugin == "quasar" {
		mQuasar.Migrate()
	}
	return nil
}
