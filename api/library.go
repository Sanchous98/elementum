package api

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"strconv"

	"github.com/Sanchous98/elementum/bittorrent"
	"github.com/Sanchous98/elementum/config"
	"github.com/Sanchous98/elementum/library"
	"github.com/Sanchous98/elementum/trakt"
	"github.com/Sanchous98/elementum/xbmc"
)

const (
	playLabel  = "LOCALIZE[30023]"
	linksLabel = "LOCALIZE[30202]"

	statusQueued = "Queued"

	statusSeeding    = "Seeding"
	statusFinished   = "Finished"
	statusPaused     = "Paused"
	statusFinding    = "Finding"
	statusBuffering  = "Buffering"
	statusAllocating = "Allocating"
	statusStalled    = "Stalled"
	statusChecking   = "Checking"
	trueType         = "true"
	falseType        = "false"
	movieType        = "movie"
	multiType        = "\nmulti"
)

// AddMovie ...
func AddMovie(ctx *fiber.Ctx) error {
	tmdbID := ctx.Params("tmdbId")
	force := ctx.Query("force", falseType) == trueType

	movie, err := library.AddMovie(tmdbID, force)
	if err != nil {
		ctx.Status(fiber.StatusOK)
		return ctx.SendString(err.Error())
	}
	if config.Get().TraktToken != "" && config.Get().TraktSyncAddedMovies {
		go trakt.SyncAddedItem("movies", tmdbID, config.Get().TraktSyncAddedMoviesLocation)
	}

	label := "LOCALIZE[30277]"
	logMsg := "%s (%s) added to library"
	if force {
		label = "LOCALIZE[30286]"
		logMsg = "%s (%s) merged to library"
	}

	log.Noticef(logMsg, movie.Title, tmdbID)
	if config.Get().LibraryUpdate == 0 || (config.Get().LibraryUpdate == 1 && xbmc.DialogConfirm("Elementum", fmt.Sprintf("%s;;%s", label, movie.Title))) {
		xbmc.VideoLibraryScanDirectory(library.MoviesLibraryPath, true)
	} else {
		library.ClearPageCache()
	}

	return nil
}

// AddMoviesList ...
func AddMoviesList(ctx *fiber.Ctx) error {
	listID := ctx.Params("listId")
	updatingStr := ctx.Query("updating", falseType)

	return library.SyncMoviesList(listID, updatingStr != falseType)
}

// RemoveMovie ...
func RemoveMovie(ctx *fiber.Ctx) error {
	tmdbID, _ := strconv.Atoi(ctx.Params("tmdbId"))
	tmdbStr := ctx.Params("tmdbId")
	movie, err := library.RemoveMovie(tmdbID)
	if err != nil {
		ctx.Status(fiber.StatusOK)
		return ctx.SendString(err.Error())
	}
	if config.Get().TraktToken != "" && config.Get().TraktSyncRemovedMovies {
		go trakt.SyncRemovedItem("movies", tmdbStr, config.Get().TraktSyncRemovedMoviesLocation)
	}

	if ctx != nil {
		if movie != nil && xbmc.DialogConfirm("Elementum", fmt.Sprintf("LOCALIZE[30278];;%s", movie.Title)) {
			xbmc.VideoLibraryClean()
		} else {
			library.ClearPageCache()
		}
	}

	return nil
}

//
// Shows externals
//

// AddShow ...
func AddShow(ctx *fiber.Ctx) error {
	tmdbID := ctx.Params("tmdbId")
	force := ctx.Query("force", falseType) == trueType

	show, err := library.AddShow(tmdbID, force)
	if err != nil {
		ctx.Status(fiber.StatusOK)
		return ctx.SendString(err.Error())
	}
	if config.Get().TraktToken != "" && config.Get().TraktSyncAddedShows {
		go trakt.SyncAddedItem("shows", tmdbID, config.Get().TraktSyncAddedShowsLocation)
	}

	label := "LOCALIZE[30277]"
	logMsg := "%s (%s) added to library"
	if force {
		label = "LOCALIZE[30286]"
		logMsg = "%s (%s) merged to library"
	}

	log.Noticef(logMsg, show.Name, tmdbID)
	if config.Get().LibraryUpdate == 0 || (config.Get().LibraryUpdate == 1 && xbmc.DialogConfirm("Elementum", fmt.Sprintf("%s;;%s", label, show.Name))) {
		xbmc.VideoLibraryScanDirectory(library.ShowsLibraryPath, true)
	} else {
		library.ClearPageCache()
	}

	return nil
}

// AddShowsList ...
func AddShowsList(ctx *fiber.Ctx) error {
	listID := ctx.Params("listId")
	updatingStr := ctx.Query("updating", falseType)

	return library.SyncShowsList(listID, updatingStr != falseType)
}

// RemoveShow ...
func RemoveShow(ctx *fiber.Ctx) error {
	tmdbID := ctx.Params("tmdbId")
	show, err := library.RemoveShow(tmdbID)
	if err != nil {
		ctx.Status(fiber.StatusOK)
		return ctx.SendString(err.Error())
	}
	if config.Get().TraktToken != "" && config.Get().TraktSyncRemovedShows {
		go trakt.SyncRemovedItem("shows", tmdbID, config.Get().TraktSyncRemovedShowsLocation)
	}

	if ctx != nil {
		if show != nil && xbmc.DialogConfirm("Elementum", fmt.Sprintf("LOCALIZE[30278];;%s", show.Name)) {
			xbmc.VideoLibraryClean()
		} else {
			library.ClearPageCache()
		}
	}

	return nil
}

// UpdateLibrary ...
func UpdateLibrary(ctx *fiber.Ctx) error {
	if err := library.Refresh(); err != nil {
		ctx.Status(fiber.StatusOK)
		return ctx.SendString(err.Error())
	}
	if config.Get().LibraryUpdate == 0 || (config.Get().LibraryUpdate == 1 && xbmc.DialogConfirm("Elementum", "LOCALIZE[30288]")) {
		xbmc.VideoLibraryScan()
	}

	return nil
}

// UpdateTrakt ...
func UpdateTrakt(ctx *fiber.Ctx) error {
	xbmc.Notify("Elementum", "LOCALIZE[30358]", config.AddonIcon())
	ctx.Status(fiber.StatusOK)

	go func() {
		library.RefreshTrakt()
		if config.Get().LibraryUpdate == 0 || (config.Get().LibraryUpdate == 1 && xbmc.DialogConfirm("Elementum", "LOCALIZE[30288]")) {
			xbmc.VideoLibraryScan()
		}
	}()

	return nil
}

// PlayMovie ...
func PlayMovie(btService *bittorrent.BTService) fiber.Handler {
	if config.Get().ChooseStreamAuto {
		return MoviePlay(btService)
	}
	return MovieLinks(btService)
}

// PlayShow ...
func PlayShow(btService *bittorrent.BTService) fiber.Handler {
	if config.Get().ChooseStreamAuto {
		return ShowEpisodePlay(btService)
	}
	return ShowEpisodeLinks(btService)
}
