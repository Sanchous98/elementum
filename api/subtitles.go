package api

import (
	"compress/gzip"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Sanchous98/elementum/scrape"
	"github.com/Sanchous98/elementum/tmdb"

	"github.com/Sanchous98/elementum/config"
	"github.com/Sanchous98/elementum/osdb"
	"github.com/Sanchous98/elementum/util"
	"github.com/Sanchous98/elementum/xbmc"
	"github.com/op/go-logging"
)

var subLog = logging.MustGetLogger("subtitles")

func appendLocalFilePayloads(playingFile string, payloads *[]osdb.SearchPayload) error {
	file, err := os.Open(playingFile)
	if err != nil {
		return err
	}
	defer file.Close()

	hashPayload := osdb.SearchPayload{}
	if h, err := osdb.HashFile(file); err == nil {
		hashPayload.Hash = h
	}
	if s, err := file.Stat(); err == nil {
		hashPayload.Size = s.Size()
	}

	*payloads = append(*payloads, []osdb.SearchPayload{
		hashPayload,
		{Query: strings.Replace(filepath.Base(playingFile), filepath.Ext(playingFile), "", -1)},
	}...)

	return nil
}

func appendMoviePayloads(labels map[string]string, payloads *[]osdb.SearchPayload) error {
	title := labels["VideoPlayer.OriginalTitle"]
	if title == "" {
		title = labels["VideoPlayer.Title"]
	}
	*payloads = append(*payloads, osdb.SearchPayload{
		Query: fmt.Sprintf("%s %s", title, labels["VideoPlayer.Year"]),
	})
	return nil
}

func appendEpisodePayloads(labels map[string]string, payloads *[]osdb.SearchPayload) error {
	season := -1
	if labels["VideoPlayer.Season"] != "" {
		if s, err := strconv.Atoi(labels["VideoPlayer.Season"]); err == nil {
			season = s
		}
	}
	episode := -1
	if labels["VideoPlayer.Episode"] != "" {
		if e, err := strconv.Atoi(labels["VideoPlayer.Episode"]); err == nil {
			episode = e
		}
	}
	if season >= 0 && episode > 0 {
		searchString := fmt.Sprintf("%s S%02dE%02d", labels["VideoPlayer.TVshowtitle"], season, episode)
		*payloads = append(*payloads, osdb.SearchPayload{
			Query: searchString,
		})
	}
	return nil
}

// SubtitlesIndex ...
func SubtitlesIndex(ctx *fiber.Ctx) error {
	searchString := ctx.Query("searchstring")
	languages := strings.Split(ctx.Query("languages"), ",")
	preferredLanguage := ctx.Query("preferredlanguage")

	// First of all, we get Subtitles language settings from Kodi
	// (there is a separate setting for that) in Player settings.
	if !config.Get().OSDBAutoLanguage && config.Get().OSDBLanguage != "" {
		languages = []string{config.Get().OSDBLanguage}
	}

	// If there is preferred language - we should use it
	if preferredLanguage != "" && preferredLanguage != "Unknown" && !contains(languages, preferredLanguage) {
		languages = append([]string{preferredLanguage}, languages...)
		preferredLanguage = strings.ToLower(preferredLanguage)
	} else {
		preferredLanguage = ""
	}

	labels := xbmc.InfoLabels(
		"VideoPlayer.Title",
		"VideoPlayer.OriginalTitle",
		"VideoPlayer.Year",
		"VideoPlayer.TVshowtitle",
		"VideoPlayer.Season",
		"VideoPlayer.Episode",
		"VideoPlayer.IMDBNumber",
	)
	playingFile := xbmc.PlayerGetPlayingFile()

	// Check if we are reading a file from Elementum
	if strings.HasPrefix(playingFile, util.GetContextHTTPHost(ctx)) {
		playingFile = strings.Replace(playingFile, util.GetContextHTTPHost(ctx)+"/files", config.Get().DownloadPath, 1)
		playingFile, _ = url.QueryUnescape(playingFile)
	}

	for i, lang := range languages {
		if lang == "Portuguese (Brazil)" {
			languages[i] = "pob"
		} else {
			isoLang := xbmc.ConvertLanguage(lang, xbmc.Iso639_2)
			if isoLang == "gre" {
				isoLang = "ell"
			}
			languages[i] = isoLang
		}
	}

	payloads := []osdb.SearchPayload{}
	if searchString != "" {
		payloads = append(payloads, osdb.SearchPayload{
			Query:     searchString,
			Languages: strings.Join(languages, ","),
		})
	} else {
		// If player ListItem has IMDBNumber specified - we try to get TMDB item from it.
		// If not - we can use localized show/movie name - which is not always found on OSDB.
		if strings.HasPrefix(labels["VideoPlayer.IMDBNumber"], "tt") {
			if labels["VideoPlayer.TVshowtitle"] != "" {
				r := tmdb.Find(labels["VideoPlayer.IMDBNumber"], "imdb_id")
				if r != nil && len(r.TVResults) > 0 {
					labels["VideoPlayer.TVshowtitle"] = r.TVResults[0].OriginalName
				}
			} else {
				r := tmdb.Find(labels["VideoPlayer.IMDBNumber"], "imdb_id")
				if r != nil && len(r.MovieResults) > 0 {
					labels["VideoPlayer.OriginalTitle"] = r.MovieResults[0].OriginalTitle
				}
			}
		}

		if strings.HasPrefix(playingFile, "http://") == false && strings.HasPrefix(playingFile, "https://") == false {
			appendLocalFilePayloads(playingFile, &payloads)
		}

		if labels["VideoPlayer.TVshowtitle"] != "" {
			appendEpisodePayloads(labels, &payloads)
		} else {
			appendMoviePayloads(labels, &payloads)
		}
	}

	for i, payload := range payloads {
		payload.Languages = strings.Join(languages, ",")
		payloads[i] = payload
	}

	subLog.Infof("Subtitles payload: %+v", payloads)

	client, err := osdb.NewClient()
	if err != nil {
		subLog.Error(err)
		ctx.Status(fiber.StatusOK)
		return err
	}
	if err := client.LogIn(config.Get().OSDBUser, config.Get().OSDBPass, config.Get().OSDBLanguage); err != nil {
		subLog.Error(err)
		ctx.Status(fiber.StatusOK)
		return err
	}

	items := make(xbmc.ListItems, 0)
	results, _ := client.SearchSubtitles(payloads)

	// If needed - try to manually sort items
	if preferredLanguage != "" {
		sort.Slice(results, func(i, j int) bool {
			id := strings.ToLower(results[i].LanguageName) == preferredLanguage
			return id
		})
	}

	for _, sub := range results {
		rating, _ := strconv.ParseFloat(sub.SubRating, 64)
		subLang := sub.LanguageName
		if subLang == "Brazilian" {
			subLang = "Portuguese (Brazil)"
		}
		item := &xbmc.ListItem{
			Label:     subLang,
			Label2:    sub.SubFileName,
			Icon:      strconv.Itoa(int((rating / 2) + 0.5)),
			Thumbnail: sub.ISO639,
			Path: URLQuery(URLForXBMC("/subtitle/%s", sub.IDSubtitleFile),
				"file", sub.SubFileName,
				"lang", sub.SubLanguageID,
				"fmt", sub.SubFormat,
				"dl", sub.SubDownloadLink),
			Properties: make(map[string]string),
		}
		if sub.MatchedBy == "moviehash" {
			item.Properties["sync"] = trueType
		}
		if sub.SubHearingImpaired == "1" {
			item.Properties["hearing_imp"] = trueType
		}
		items = append(items, item)
	}

	ctx.Status(fiber.StatusOK)
	return ctx.JSON(xbmc.NewView("", items))
}

// SubtitleGet ...
func SubtitleGet(ctx *fiber.Ctx) error {
	file := ctx.Query("file")
	dl := ctx.Query("dl")
	resp, err := scrape.GetClient().Get(dl)

	if err != nil {
		subLog.Error(err)
		ctx.Status(fiber.StatusOK)
		return err
	}
	defer resp.Body.Close()

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		subLog.Error(err)
		ctx.Status(fiber.StatusOK)
		return err
	}
	defer reader.Close()

	subtitlesPath := filepath.Join(config.Get().DownloadPath, "Subtitles")
	if config.Get().DownloadPath == "." {
		subtitlesPath = filepath.Join(config.Get().TemporaryPath, "Subtitles")
	}
	if _, errStat := os.Stat(subtitlesPath); os.IsNotExist(errStat) {
		if errMk := os.Mkdir(subtitlesPath, 0755); errMk != nil {
			subLog.Error("Unable to create Subtitles folder")
		}
	}

	outFile, err := os.Create(filepath.Join(subtitlesPath, file))
	if err != nil {
		subLog.Error(err)
		ctx.Status(fiber.StatusOK)
		return err
	}
	defer outFile.Close()

	io.Copy(outFile, reader)

	ctx.Status(fiber.StatusOK)
	return ctx.JSON(xbmc.NewView("", xbmc.ListItems{
		{Label: file, Path: outFile.Name()},
	}))
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
