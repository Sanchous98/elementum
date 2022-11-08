package api

import (
	"github.com/Sanchous98/elementum/config"
	"github.com/Sanchous98/elementum/xbmc"
	"github.com/gofiber/fiber/v2"
)

// Index ...
func Index(ctx *fiber.Ctx) error {
	action := ctx.Query("action")
	if action == "search" || action == "manualsearch" {
		return SubtitlesIndex(ctx)
	}

	ctx.Status(fiber.StatusOK)
	return ctx.JSON(xbmc.NewView("", xbmc.ListItems{
		{Label: "LOCALIZE[30214]", Path: URLForXBMC("/movies/"), Thumbnail: config.AddonResource("img", "movies.png")},
		{Label: "LOCALIZE[30215]", Path: URLForXBMC("/shows/"), Thumbnail: config.AddonResource("img", "tv.png")},
		{Label: "LOCALIZE[30209]", Path: URLForXBMC("/search"), Thumbnail: config.AddonResource("img", "search.png")},
		{Label: "LOCALIZE[30229]", Path: URLForXBMC("/torrents/"), Thumbnail: config.AddonResource("img", "cloud.png")},
		{Label: "LOCALIZE[30216]", Path: URLForXBMC("/playtorrent"), Thumbnail: config.AddonResource("img", "magnet.png")},
		{Label: "LOCALIZE[30239]", Path: URLForXBMC("/provider/"), Thumbnail: config.AddonResource("img", "shield.png")},
		{Label: "LOCALIZE[30355]", Path: URLForXBMC("/changelog"), Thumbnail: config.AddonResource("img", "faq8.png")},
		{Label: "LOCALIZE[30393]", Path: URLForXBMC("/status"), Thumbnail: config.AddonResource("img", "clock.png")},
	}))
}
