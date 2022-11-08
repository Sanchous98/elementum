package api

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"strings"

	"github.com/Sanchous98/elementum/bittorrent"
	"github.com/Sanchous98/elementum/config"
	"github.com/Sanchous98/elementum/database"
	"github.com/Sanchous98/elementum/providers"
	"github.com/Sanchous98/elementum/xbmc"

	"github.com/cespare/xxhash"
	"github.com/op/go-logging"
)

var searchLog = logging.MustGetLogger("search")

// Search ...
func Search(btService *bittorrent.BTService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
		query := ctx.Query("q")
		keyboard := ctx.Query("keyboard")

		if len(query) == 0 {
			historyType := ""
			if len(keyboard) > 0 {
				if query = xbmc.Keyboard("", "LOCALIZE[30206]"); len(query) == 0 {
					return nil
				}
				searchHistoryAppend(ctx, historyType, query)
			}

			return searchHistoryList(ctx, historyType)
		}

		fakeTmdbID := strconv.FormatUint(xxhash.Sum64String(query), 10)
		existingTorrent := btService.HasTorrentByQuery(query)
		if existingTorrent != "" && (config.Get().SilentStreamStart || xbmc.DialogConfirmFocused("Elementum", "LOCALIZE[30270]")) {
			xbmc.PlayURLWithTimeout(URLQuery(
				URLForXBMC("/play"),
				"resume", existingTorrent,
				"query", query,
				"tmdb", fakeTmdbID,
				"type", "search"))
			return nil
		}

		if torrent := InTorrentsMap(fakeTmdbID); torrent != nil {
			xbmc.PlayURLWithTimeout(URLQuery(
				URLForXBMC("/play"), "uri", torrent.URI,
				"query", query,
				"tmdb", fakeTmdbID,
				"type", "search"))
			return nil
		}

		var torrents []*bittorrent.TorrentFile
		var err error

		if torrents, err = GetCachedTorrents(fakeTmdbID); err != nil || len(torrents) == 0 {
			searchLog.Infof("Searching providers for: %s", query)

			searchers := providers.GetSearchers()
			torrents = providers.Search(searchers, query)

			SetCachedTorrents(fakeTmdbID, torrents)
		}

		if len(torrents) == 0 {
			xbmc.Notify("Elementum", "LOCALIZE[30205]", config.AddonIcon())
			return nil
		}

		choices := make([]string, 0, len(torrents))
		for _, torrent := range torrents {
			resolution := ""
			if torrent.Resolution > 0 {
				resolution = fmt.Sprintf("[B][COLOR %s]%s[/COLOR][/B] ", bittorrent.Colors[torrent.Resolution], bittorrent.Resolutions[torrent.Resolution])
			}

			info := make([]string, 0)
			if torrent.Size != "" {
				info = append(info, fmt.Sprintf("[B][%s][/B]", torrent.Size))
			}
			if torrent.RipType > 0 {
				info = append(info, bittorrent.Rips[torrent.RipType])
			}
			if torrent.VideoCodec > 0 {
				info = append(info, bittorrent.Codecs[torrent.VideoCodec])
			}
			if torrent.AudioCodec > 0 {
				info = append(info, bittorrent.Codecs[torrent.AudioCodec])
			}
			if torrent.Provider != "" {
				info = append(info, fmt.Sprintf(" - [B]%s[/B]", torrent.Provider))
			}

			multi := ""
			if torrent.Multi {
				multi = multiType
			}

			label := fmt.Sprintf("%s(%d / %d) %s\n%s\n%s%s",
				resolution,
				torrent.Seeds,
				torrent.Peers,
				strings.Join(info, " "),
				torrent.Name,
				torrent.Icon,
				multi,
			)
			choices = append(choices, label)
		}

		choice := xbmc.ListDialogLarge("LOCALIZE[30228]", query, choices...)
		if choice >= 0 {
			AddToTorrentsMap(fakeTmdbID, torrents[choice])

			xbmc.PlayURLWithTimeout(URLQuery(
				URLForXBMC("/play"),
				"uri", torrents[choice].URI,
				"query", query,
				"tmdb", fakeTmdbID,
				"type", "search"))
		}

		return nil
	}
}

func searchHistoryAppend(ctx *fiber.Ctx, historyType string, query string) {
	database.Get().AddSearchHistory(historyType, query)

	go xbmc.UpdatePath(searchHistoryGetXbmcURL(historyType, query))
	ctx.Status(fiber.StatusOK)
	return
}

func searchHistoryList(ctx *fiber.Ctx, historyType string) error {
	historyList := []string{}
	rows, err := database.Get().Query(`SELECT query FROM history_queries WHERE type = ? ORDER BY dt DESC`, historyType)
	if err != nil {
		return err
	}

	query := ""
	for rows.Next() {
		rows.Scan(&query)
		historyList = append(historyList, query)
	}
	rows.Close()

	urlPrefix := ""
	if len(historyType) > 0 {
		urlPrefix = "/" + historyType
	}

	items := make(xbmc.ListItems, 0, len(historyList)+1)
	items = append(items, &xbmc.ListItem{
		Label:     "LOCALIZE[30323]",
		Path:      URLQuery(URLForXBMC(urlPrefix+"/search"), "keyboard", "1"),
		Thumbnail: config.AddonResource("img", "search.png"),
		Icon:      config.AddonResource("img", "search.png"),
	})

	for _, query := range historyList {
		items = append(items, &xbmc.ListItem{
			Label: query,
			Path:  searchHistoryGetXbmcURL(historyType, query),
			ContextMenu: [][]string{
				{"LOCALIZE[30406]", fmt.Sprintf("XBMC.RunPlugin(%s)",
					URLQuery(URLForXBMC("/search/remove"),
						"query", query,
						"type", historyType,
					))},
			},
		})
	}

	ctx.Status(fiber.StatusOK)
	ctx.JSON(xbmc.NewView("", items))

	return nil
}

// SearchRemove ...
func SearchRemove(ctx *fiber.Ctx) error {
	query := ctx.Query("query", "")
	historyType := ctx.Query("type", "")

	if len(query) == 0 {
		return nil
	}

	log.Debugf("Removing query '%s' with history type '%s'", query, historyType)
	database.Get().Exec("DELETE FROM history_queries WHERE query = ? AND type = ?", query, historyType)
	xbmc.Refresh()

	ctx.Status(fiber.StatusOK)
	return nil
}

// SearchClear ...
func SearchClear(ctx *fiber.Ctx) error {
	historyType := ctx.Query("type", "")

	log.Debugf("Cleaning queries with history type %s", historyType)
	database.Get().Exec("DELETE FROM history_queries WHERE type = ?", historyType)
	xbmc.Refresh()

	ctx.Status(fiber.StatusOK)
	return nil
}

func searchHistoryGetXbmcURL(historyType string, query string) string {
	urlPrefix := ""
	if len(historyType) > 0 {
		urlPrefix = "/" + historyType
	}

	return URLQuery(URLForXBMC(urlPrefix+"/search"), "q", query)
}
