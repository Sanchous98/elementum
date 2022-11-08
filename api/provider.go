package api

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"strconv"

	"github.com/Sanchous98/elementum/config"
	"github.com/Sanchous98/elementum/providers"
	"github.com/Sanchous98/elementum/tmdb"
	"github.com/Sanchous98/elementum/xbmc"
)

type providerDebugResponse struct {
	Payload any `json:"payload"`
	Results any `json:"results"`
}

// ProviderGetMovie ...
func ProviderGetMovie(ctx *fiber.Ctx) error {
	tmdbID := ctx.Params("tmdbId")
	provider := ctx.Params("provider")
	log.Infof("Searching links for:", tmdbID)
	movie := tmdb.GetMovieByID(tmdbID, config.Get().Language)
	log.Infof("Resolved %s to %s", tmdbID, movie.Title)

	searcher := providers.NewAddonSearcher(provider)
	torrents := searcher.SearchMovieLinks(movie)
	if ctx.Query("resolve") == "true" {
		for _, torrent := range torrents {
			torrent.Resolve()
		}
	}
	data, err := json.MarshalIndent(providerDebugResponse{
		Payload: searcher.GetMovieSearchObject(movie),
		Results: torrents,
	}, "", "    ")
	if err != nil {
		xbmc.AddonFailure(provider)
		return err
	}

	ctx.Status(fiber.StatusOK)
	ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
	return ctx.Send(data)
}

// ProviderGetEpisode ...
func ProviderGetEpisode(ctx *fiber.Ctx) error {
	provider := ctx.Params("provider")
	showID, _ := strconv.Atoi(ctx.Params("showId"))
	seasonNumber, _ := strconv.Atoi(ctx.Params("season"))
	episodeNumber, _ := strconv.Atoi(ctx.Params("episode"))

	log.Infof("Searching links for TMDB Id:", showID)

	show := tmdb.GetShow(showID, config.Get().Language)
	season := tmdb.GetSeason(showID, seasonNumber, config.Get().Language)
	if season == nil {
		return fmt.Errorf("Unable to get season %d", seasonNumber)
	}
	episode := season.Episodes[episodeNumber-1]

	log.Infof("Resolved %d to %s", showID, show.Name)

	searcher := providers.NewAddonSearcher(provider)
	torrents := searcher.SearchEpisodeLinks(show, episode)
	if ctx.Query("resolve") == "true" {
		for _, torrent := range torrents {
			torrent.Resolve()
		}
	}
	data, err := json.MarshalIndent(providerDebugResponse{
		Payload: searcher.GetEpisodeSearchObject(show, episode),
		Results: torrents,
	}, "", "    ")
	if err != nil {
		xbmc.AddonFailure(provider)
		return err
	}

	ctx.Status(fiber.StatusOK)
	ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
	return ctx.Send(data)
}
