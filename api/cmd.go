package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/op/go-logging"

	"github.com/Sanchous98/elementum/config"
	"github.com/Sanchous98/elementum/database"
	"github.com/Sanchous98/elementum/library"
	"github.com/Sanchous98/elementum/xbmc"
)

var cmdLog = logging.MustGetLogger("cmd")

// ClearCache ...
func ClearCache(ctx *fiber.Ctx) error {
	key := ctx.Params("key")
	if key != "" {
		library.ClearCacheKey(key)
	} else {
		log.Debug("Removing all the cache")

		if !xbmc.DialogConfirm("Elementum", "LOCALIZE[30471]") {
			ctx.Status(fiber.StatusOK)
			return nil
		}

		database.GetCache().RecreateBucket(database.CommonBucket)
	}

	xbmc.Notify("Elementum", "LOCALIZE[30200]", config.AddonIcon())
	return nil
}

// ClearCacheTMDB ...
func ClearCacheTMDB(*fiber.Ctx) error {
	log.Debug("Removing TMDB cache")

	library.ClearTmdbCache()

	xbmc.Notify("Elementum", "LOCALIZE[30200]", config.AddonIcon())
	return nil
}

// ClearCacheTrakt ...
func ClearCacheTrakt(*fiber.Ctx) error {
	log.Debug("Removing Trakt cache")

	library.ClearTraktCache()

	xbmc.Notify("Elementum", "LOCALIZE[30200]", config.AddonIcon())
	return nil
}

// ClearPageCache ...
func ClearPageCache(*fiber.Ctx) error {
	library.ClearPageCache()
	return nil
}

// ClearTraktCache ...
func ClearTraktCache(*fiber.Ctx) error {
	library.ClearTraktCache()
	return nil
}

// ClearTmdbCache ...
func ClearTmdbCache(*fiber.Ctx) error {
	library.ClearTmdbCache()
	return nil
}

// ResetPath ...
func ResetPath(*fiber.Ctx) error {
	xbmc.SetSetting("download_path", "")
	xbmc.SetSetting("library_path", "special://temp/elementum_library/")
	xbmc.SetSetting("torrents_path", "special://temp/elementum_torrents/")
	return nil
}

// SetViewMode ...
func SetViewMode(ctx *fiber.Ctx) error {
	contentType := ctx.Params("content_type")
	viewName := xbmc.InfoLabel("Container.Viewmode")
	viewMode := xbmc.GetCurrentView()
	cmdLog.Noticef("ViewMode: %s (%s)", viewName, viewMode)
	if viewMode != "0" {
		xbmc.SetSetting("viewmode_"+contentType, viewMode)
	}

	ctx.Status(fiber.StatusOK)
	return nil
}

// ClearDatabaseMovies ...
func ClearDatabaseMovies(ctx *fiber.Ctx) error {
	log.Debug("Removing deleted movies from database")

	database.Get().Exec("DELETE FROM library_items WHERE state = ? AND mediaType = ?", library.StateDeleted, library.MovieType)

	xbmc.Notify("Elementum", "LOCALIZE[30472]", config.AddonIcon())

	ctx.Status(fiber.StatusOK)
	return nil
}

// ClearDatabaseShows ...
func ClearDatabaseShows(ctx *fiber.Ctx) error {
	log.Debug("Removing deleted shows from database")

	database.Get().Exec("DELETE FROM library_items WHERE state = ? AND mediaType = ?", library.StateDeleted, library.ShowType)

	xbmc.Notify("Elementum", "LOCALIZE[30472]", config.AddonIcon())

	ctx.Status(fiber.StatusOK)
	return nil
}

// ClearDatabaseTorrentHistory ...
func ClearDatabaseTorrentHistory(ctx *fiber.Ctx) error {
	log.Debug("Removing torrent history from database")

	database.Get().Exec("DELETE FROM thistory_assign")
	database.Get().Exec("DELETE FROM thistory_metainfo")
	database.Get().Exec("DELETE FROM tinfo")

	xbmc.Notify("Elementum", "LOCALIZE[30472]", config.AddonIcon())

	ctx.Status(fiber.StatusOK)
	return nil
}

// ClearDatabaseSearchHistory ...
func ClearDatabaseSearchHistory(ctx *fiber.Ctx) error {
	log.Debug("Removing search history from database")

	database.Get().Exec("DELETE FROM history_queries")

	xbmc.Notify("Elementum", "LOCALIZE[30472]", config.AddonIcon())

	ctx.Status(fiber.StatusOK)
	return nil
}

// ClearDatabase ...
func ClearDatabase(ctx *fiber.Ctx) error {
	log.Debug("Removing all the database")
	ctx.Status(fiber.StatusOK)

	if !xbmc.DialogConfirm("Elementum", "LOCALIZE[30471]") {
		return nil
	}

	database.Get().Exec("DELETE FROM history_queries")
	database.Get().Exec("DELETE FROM library_items")
	database.Get().Exec("DELETE FROM library_uids")
	database.Get().Exec("DELETE FROM thistory_assign")
	database.Get().Exec("DELETE FROM thistory_metainfo")
	database.Get().Exec("DELETE FROM tinfo")

	xbmc.Notify("Elementum", "LOCALIZE[30472]", config.AddonIcon())

	return nil
}
