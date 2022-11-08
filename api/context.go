package api

import (
	"github.com/gofiber/fiber/v2"
	"strconv"

	"github.com/Sanchous98/elementum/config"
	"github.com/Sanchous98/elementum/library"
)

// ContextPlaySelector ...
func ContextPlaySelector(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")

	id := ctx.Params("kodiID")
	kodiID, _ := strconv.Atoi(id)
	media := ctx.Params("media")

	action := "forcelinks"
	if config.Get().ChooseStreamAuto {
		action = "forceplay"
	}

	if kodiID == 0 {
		return ctx.Redirect(URLQuery(URLForXBMC("/search"), "q", id))
	} else if media == "movie" {
		if m := library.GetLibraryMovie(kodiID); m != nil && m.UIDs.TMDB != 0 {
			return ctx.Redirect(URLQuery(URLForXBMC("/movie/%d/%s", m.UIDs.TMDB, action)))
		}
	} else if media == "episode" {
		if s, e := library.GetLibraryEpisode(kodiID); s != nil && e != nil && e.UIDs.TMDB != 0 {
			return ctx.Redirect(URLQuery(URLForXBMC("/show/%d/season/%d/episode/%d/%s", s.UIDs.TMDB, e.Season, e.Episode, action)))
		}
	}

	log.Debugf("Cound not find TMDB entry for requested Kodi item %d of type %s", kodiID, media)
	ctx.Status(404)
	return ctx.SendString("Cannot find TMDB for selected Kodi item")
}
