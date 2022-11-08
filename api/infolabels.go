package api

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"math/rand"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"github.com/Sanchous98/elementum/bittorrent"
	"github.com/Sanchous98/elementum/config"
	"github.com/Sanchous98/elementum/library"
	"github.com/Sanchous98/elementum/tmdb"
	"github.com/Sanchous98/elementum/xbmc"
	"github.com/sanity-io/litter"
)

var (
	infoLabels = []string{
		"ListItem.DBID",
		"ListItem.DBTYPE",
		"ListItem.Mediatype",
		"ListItem.TMDB",
		"ListItem.UniqueId",

		"ListItem.Label",
		"ListItem.Label2",
		"ListItem.ThumbnailImage",
		"ListItem.Title",
		"ListItem.OriginalTitle",
		"ListItem.TVShowTitle",
		"ListItem.Season",
		"ListItem.Episode",
		"ListItem.Premiered",
		"ListItem.Plot",
		"ListItem.PlotOutline",
		"ListItem.Tagline",
		"ListItem.Year",
		"ListItem.Trailer",
		"ListItem.Studio",
		"ListItem.MPAA",
		"ListItem.Genre",
		"ListItem.Mediatype",
		"ListItem.Writer",
		"ListItem.Director",
		"ListItem.Rating",
		"ListItem.Votes",
		"ListItem.IMDBNumber",
		"ListItem.Code",
		"ListItem.ArtFanart",
		"ListItem.ArtBanner",
		"ListItem.ArtPoster",
		"ListItem.ArtTvshowPoster",
	}
)

func saveEncoded(encoded string) {
	xbmc.SetWindowProperty("ListItem.Encoded", encoded)
}

func encodeItem(item *xbmc.ListItem) string {
	data, _ := json.Marshal(item)

	return *(*string)(unsafe.Pointer(&data))
}

// InfoLabelsStored ...
func InfoLabelsStored(ctx *fiber.Ctx) error {
	labelsString := "{}"

	if listLabel := xbmc.InfoLabel("ListItem.Label"); len(listLabel) > 0 {
		labels := xbmc.InfoLabels(infoLabels...)

		listItemLabels := make(map[string]string, len(labels))
		for k, v := range labels {
			key := strings.Replace(k, "ListItem.", "", 1)
			listItemLabels[key] = v
		}

		b, _ := json.Marshal(listItemLabels)
		labelsString = *(*string)(unsafe.Pointer(&b))
		saveEncoded(labelsString)
	} else if encoded := xbmc.GetWindowProperty("ListItem.Encoded"); len(encoded) > 0 {
		labelsString = encoded
	}

	ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Status(fiber.StatusOK)
	return ctx.SendString(labelsString)
}

// InfoLabelsEpisode ...
func InfoLabelsEpisode() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		tmdbID := ctx.Params("showId")
		showID, _ := strconv.Atoi(tmdbID)
		seasonNumber, _ := strconv.Atoi(ctx.Params("season"))
		episodeNumber, _ := strconv.Atoi(ctx.Params("episode"))

		if item, err := GetEpisodeLabels(showID, seasonNumber, episodeNumber); err == nil {
			saveEncoded(encodeItem(item))
			ctx.Status(fiber.StatusOK)
			return ctx.JSON(item)
		} else {
			ctx.Status(fiber.StatusNotFound)
			return err
		}
	}
}

// InfoLabelsMovie ...
func InfoLabelsMovie(ctx *fiber.Ctx) error {
	tmdbID := ctx.Params("tmdbId")

	if item, err := GetMovieLabels(tmdbID); err == nil {
		saveEncoded(encodeItem(item))

		ctx.Status(fiber.StatusOK)
		return ctx.JSON(item)
	} else {
		ctx.Status(fiber.StatusNotFound)
		return err
	}
}

// InfoLabelsSearch ...
func InfoLabelsSearch(btService *bittorrent.BTService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		tmdbID := ctx.Params("tmdbId")

		if item, err := GetSearchLabels(btService, tmdbID); err == nil {
			saveEncoded(encodeItem(item))
			ctx.Status(fiber.StatusOK)
			return ctx.JSON(item)
		} else {
			ctx.Status(fiber.StatusNotFound)
			return err
		}
	}
}

// GetEpisodeLabels returnes listitem for an episode
func GetEpisodeLabels(showID, seasonNumber, episodeNumber int) (item *xbmc.ListItem, err error) {
	show := tmdb.GetShow(showID, config.Get().Language)
	if show == nil {
		return nil, errors.New("Unable to find show")
	}

	season := tmdb.GetSeason(showID, seasonNumber, config.Get().Language)
	if season == nil {
		return nil, errors.New("Unable to find season")
	}

	episode := tmdb.GetEpisode(showID, seasonNumber, episodeNumber, config.Get().Language)
	if episode == nil {
		return nil, errors.New("Unable to find episode")
	}

	item = episode.ToListItem(show, season)
	if ls, err := library.GetShowByTMDB(show.ID); ls != nil && err == nil {
		log.Debugf("Found show in library: %s", litter.Sdump(ls.UIDs))
		if le := ls.GetEpisode(episode.SeasonNumber, episodeNumber); le != nil {
			item.Info.DBID = le.UIDs.Kodi
		}
	}
	if item.Art.FanArt == "" {
		fanarts := make([]string, 0, len(show.Images.Backdrops))
		for _, backdrop := range show.Images.Backdrops {
			fanarts = append(fanarts, tmdb.ImageURL(backdrop.FilePath, "w1280"))
		}
		if len(fanarts) > 0 {
			item.Art.FanArt = fanarts[rand.Intn(len(fanarts))]
		}
	}
	item.Art.Poster = tmdb.ImageURL(season.Poster, "w500")

	return
}

// GetMovieLabels returnes listitem for a movie
func GetMovieLabels(tmdbID string) (item *xbmc.ListItem, err error) {
	movie := tmdb.GetMovieByID(tmdbID, config.Get().Language)
	if movie == nil {
		return nil, errors.New("Unable to find movie")
	}

	item = movie.ToListItem()
	if lm, err := library.GetMovieByTMDB(movie.ID); lm != nil && err == nil {
		log.Debugf("Found movie in library: %s", litter.Sdump(lm))
		item.Info.DBID = lm.UIDs.Kodi
	}

	return
}

// GetSearchLabels returnes listitem for a search query
func GetSearchLabels(btService *bittorrent.BTService, tmdbID string) (item *xbmc.ListItem, err error) {
	torrent := btService.GetTorrentByFakeID(tmdbID)
	if torrent == nil || torrent.DBItem == nil {
		return nil, errors.New("Unable to find the torrent")
	}
	chosenFileNames := make([]string, 0, len(torrent.ChosenFiles))
	for _, f := range torrent.ChosenFiles {
		chosenFileNames = append(chosenFileNames, filepath.Base(f.DisplayPath()))
	}
	sort.Sort(sort.StringSlice(chosenFileNames))
	subtitle := strings.Join(chosenFileNames, ", ")

	item = &xbmc.ListItem{
		Label:  torrent.DBItem.Query,
		Label2: subtitle,
		Info: &xbmc.ListItemInfo{
			Title:         torrent.DBItem.Query,
			OriginalTitle: torrent.DBItem.Query,
			TVShowTitle:   subtitle,
			DBTYPE:        "episode",
			Mediatype:     "episode",
		},
		Art: &xbmc.ListItemArt{},
	}

	return
}
