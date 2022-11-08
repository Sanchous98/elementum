package api

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/op/go-logging"

	"github.com/Sanchous98/elementum/bittorrent"
	"github.com/Sanchous98/elementum/config"
	"github.com/Sanchous98/elementum/database"
	"github.com/Sanchous98/elementum/util"
	"github.com/Sanchous98/elementum/xbmc"
)

var (
	torrentsLog = logging.MustGetLogger("torrents")
)

// TorrentsWeb ...
type TorrentsWeb struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Size          string  `json:"size"`
	Status        string  `json:"status"`
	Progress      float64 `json:"progress"`
	Ratio         float64 `json:"ratio"`
	TimeRatio     float64 `json:"time_ratio"`
	SeedingTime   string  `json:"seeding_time"`
	SeedTime      float64 `json:"seed_time"`
	SeedTimeLimit int     `json:"seed_time_limit"`
	DownloadRate  float64 `json:"download_rate"`
	UploadRate    float64 `json:"upload_rate"`
	Seeders       int     `json:"seeders"`
	SeedersTotal  int     `json:"seeders_total"`
	Peers         int     `json:"peers"`
	PeersTotal    int     `json:"peers_total"`
}

// AddToTorrentsMap ...
func AddToTorrentsMap(tmdbID string, torrent *bittorrent.TorrentFile) {
	if strings.HasPrefix(torrent.URI, "magnet") {
		torrentsLog.Debugf("Saving torrent entry for TMDB: %#v", tmdbID)
		if b, err := torrent.MarshalJSON(); err == nil {
			database.Get().AddTorrentHistory(tmdbID, torrent.InfoHash, b)
		}

		return
	}

	b, err := ioutil.ReadFile(torrent.URI)
	if err != nil {
		return
	}

	torrentsLog.Debugf("Saving torrent entry for TMDB: %#v", tmdbID)
	database.Get().AddTorrentHistory(tmdbID, torrent.InfoHash, b)
}

// InTorrentsMap ...
func InTorrentsMap(tmdbID string) *bittorrent.TorrentFile {
	if !config.Get().UseCacheSelection {
		return nil
	}

	var infohash string
	var infohashID int64
	var b []byte
	database.Get().QueryRow(`SELECT l.infohash_id, i.infohash, i.metainfo FROM thistory_assign l LEFT JOIN thistory_metainfo i ON i.rowid = l.infohash_id WHERE l.item_id = ?`, tmdbID).Scan(&infohashID, &infohash, &b)

	if len(infohash) > 0 && len(b) > 0 {
		torrent := &bittorrent.TorrentFile{}
		if b[0] == '{' {
			torrent.UnmarshalJSON(b)
		} else {
			torrent.LoadFromBytes(b)
		}

		if len(torrent.URI) > 0 && (config.Get().SilentStreamStart || xbmc.DialogConfirmFocused("Elementum", fmt.Sprintf("LOCALIZE[30260];;[COLOR gold]%s[/COLOR]", torrent.Title))) {
			return torrent
		}

		database.Get().Exec(`DELETE FROM thistory_assign WHERE item_id = ?`, tmdbID)
		var left int
		database.Get().QueryRow(`SELECT COUNT(*) FROM thistory_assign WHERE infohash_id = ?`, infohashID).Scan(&left)
		if left == 0 {
			database.Get().Exec(`DELETE FROM thistory_metainfo WHERE rowid = ?`, infohashID)
		}
	}

	return nil
}

// GetCachedTorrents searches for torrent entries in the cache
func GetCachedTorrents(tmdbID string) ([]*bittorrent.TorrentFile, error) {
	if !config.Get().UseCacheSearch {
		return nil, fmt.Errorf("Caching is disabled")
	}

	cacheDB := database.GetCache()

	var ret []*bittorrent.TorrentFile
	err := cacheDB.GetCachedObject(database.CommonBucket, tmdbID, &ret)
	if len(ret) > 0 {
		for _, t := range ret {
			if !strings.HasPrefix(t.URI, "magnet:") {
				if _, err = os.Open(t.URI); err != nil {
					return nil, fmt.Errorf("Cache is not up to date")
				}
			}
		}
	}

	return ret, err
}

// SetCachedTorrents caches torrent search results in cache
func SetCachedTorrents(tmdbID string, torrents []*bittorrent.TorrentFile) error {
	cacheDB := database.GetCache()

	return cacheDB.SetCachedObject(database.CommonBucket, config.Get().CacheSearchDuration, tmdbID, torrents)
}

// ListTorrents ...
func ListTorrents(btService *bittorrent.BTService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		items := make(xbmc.ListItems, 0, len(btService.Torrents))
		if len(btService.Torrents) == 0 {
			ctx.Status(fiber.StatusOK)
			return ctx.JSON(xbmc.NewView("", items))
		}

		// torrentsLog.Debug("Currently downloading:")
		for i, torrent := range btService.Torrents {
			if torrent == nil {
				continue
			}

			torrentName := torrent.Name()
			progress := torrent.GetProgress()
			status := torrent.GetStateString()

			torrentAction := []string{"LOCALIZE[30231]", fmt.Sprintf("XBMC.RunPlugin(%s)", URLForXBMC("/torrents/pause/%s", i))}
			sessionAction := []string{"LOCALIZE[30233]", fmt.Sprintf("XBMC.RunPlugin(%s)", URLForXBMC("/torrents/pause"))}

			if status == "Paused" {
				sessionAction = []string{"LOCALIZE[30234]", fmt.Sprintf("XBMC.RunPlugin(%s)", URLForXBMC("/torrents/resume"))}
			} else if status != "Finished" {
				torrentAction = []string{"LOCALIZE[30235]", fmt.Sprintf("XBMC.RunPlugin(%s)", URLForXBMC("/torrents/resume/%s", i))}
			}

			color := "white"
			switch status {
			case statusPaused:
				fallthrough
			case statusFinished:
				color = "grey"
			case statusSeeding:
				color = "green"
			case statusBuffering:
				color = "blue"
			case statusFinding:
				color = "orange"
			case statusChecking:
				color = "teal"
			case statusQueued:
			case statusAllocating:
				color = "black"
			case statusStalled:
				color = "red"
			}

			// TODO: Add seeding time and ratio getter/output
			// torrentsLog.Debugf("- %.2f%% - %s - %s", progress, status, torrentName)

			var (
				tmdb        string
				show        string
				season      string
				episode     string
				contentType string
			)

			if torrent.DBItem != nil && torrent.DBItem.Type != "" {
				contentType = torrent.DBItem.Type
				if contentType == movieType {
					tmdb = strconv.Itoa(torrent.DBItem.ID)
				} else {
					show = strconv.Itoa(torrent.DBItem.ShowID)
					season = strconv.Itoa(torrent.DBItem.Season)
					episode = strconv.Itoa(torrent.DBItem.Episode)
				}
			}

			playURL := URLQuery(URLForXBMC("/play"),
				"resume", i,
				"type", contentType,
				"tmdb", tmdb,
				"show", show,
				"season", season,
				"episode", episode)

			item := xbmc.ListItem{
				Label: fmt.Sprintf("%.2f%% - [COLOR %s]%s[/COLOR] - %s", progress, color, status, torrentName),
				Path:  playURL,
				Info: &xbmc.ListItemInfo{
					Title: torrentName,
				},
			}
			item.ContextMenu = [][]string{
				{"LOCALIZE[30230]", fmt.Sprintf("XBMC.PlayMedia(%s)", playURL)},
				torrentAction,
				{"LOCALIZE[30232]", fmt.Sprintf("XBMC.RunPlugin(%s)", URLForXBMC("/torrents/delete/%s", i))},
				{"LOCALIZE[30276]", fmt.Sprintf("XBMC.RunPlugin(%s)", URLForXBMC("/torrents/delete/%s?files=1", i))},
				{"LOCALIZE[30308]", fmt.Sprintf("XBMC.RunPlugin(%s)", URLForXBMC("/torrents/move/%s", i))},
				sessionAction,
			}
			item.IsPlayable = true
			items = append(items, &item)
		}

		ctx.Status(fiber.StatusOK)
		return ctx.JSON(xbmc.NewView("", items))
	}
}

// ListTorrentsWeb ...
func ListTorrentsWeb(btService *bittorrent.BTService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		torrents := make([]*TorrentsWeb, 0, len(btService.Torrents))

		if len(btService.Torrents) == 0 {
			ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
			ctx.Status(fiber.StatusOK)
			return ctx.JSON(torrents)
		}

		// torrentsLog.Debugf("Currently downloading:")
		for _, torrent := range btService.Torrents {
			if torrent == nil {
				continue
			}

			torrentName := torrent.Name()
			progress := torrent.GetProgress()
			status := torrent.GetStateString()

			// if status != statusFinished {
			// 	if progress >= 100 {
			// 		status = statusFinished
			// 	} else {
			// 		status = statusDownloading
			// 	}
			// } else if status == statusFinished || progress >= 100 {
			// 	status = statusSeeding
			// }

			size := humanize.Bytes(uint64(torrent.Length()))
			downloadRate := float64(torrent.DownloadRate) / 1024
			uploadRate := float64(torrent.UploadRate) / 1024

			stats := torrent.Stats()
			peers := stats.ActivePeers
			peersTotal := stats.TotalPeers

			t := TorrentsWeb{
				ID:           torrent.InfoHash(),
				Name:         torrentName,
				Size:         size,
				Status:       status,
				Progress:     progress,
				DownloadRate: downloadRate,
				UploadRate:   uploadRate,
				Peers:        peers,
				PeersTotal:   peersTotal,
			}
			torrents = append(torrents, &t)

			// torrentsLog.Debugf("- %.2f%% - %s - %s", progress, status, torrentName)
		}

		ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Status(fiber.StatusOK)
		return ctx.JSON(torrents)
	}
}

// PauseSession ...
func PauseSession(ctx *fiber.Ctx) error {
	// TODO: Add Global Pause
	xbmc.Refresh()
	ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Status(fiber.StatusOK)
	return nil
}

// ResumeSession ...
func ResumeSession(ctx *fiber.Ctx) error {
	// TODO: Add Global Resume
	xbmc.Refresh()
	ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Status(fiber.StatusOK)
	return nil
}

// AddTorrent ...
func AddTorrent(btService *bittorrent.BTService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		uri := ctx.FormValue("uri")
		header, fileError := ctx.FormFile("file")

		if fileError != nil {
			ctx.Status(fiber.StatusInternalServerError)
			return fileError
		}

		file, err := header.Open()

		if err != nil {
			ctx.Status(fiber.StatusInternalServerError)
			return err
		}

		t, err := saveTorrentFile(file, header)
		if err == nil && t != "" {
			uri = t
		}

		ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")

		if uri == "" {
			ctx.Status(fiber.StatusNotFound)
			return ctx.SendString("Missing torrent URI")
		}
		torrentsLog.Infof("Adding torrent from %s", uri)

		_, err = btService.AddTorrent(uri)
		if err != nil {
			ctx.Status(fiber.StatusNotFound)
			return err
		}

		torrentsLog.Infof("Downloading %s", uri)

		xbmc.Refresh()
		ctx.Status(fiber.StatusOK)
		return nil
	}
}

// ResumeTorrent ...
func ResumeTorrent(btService *bittorrent.BTService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		torrentID := ctx.Params("torrentId")
		torrent, err := GetTorrentFromParam(btService, torrentID)
		if err != nil {
			ctx.Status(fiber.StatusNotFound)
			return fmt.Errorf("Unable to resume torrent with index %s", torrentID)
		}

		torrent.Resume()

		xbmc.Refresh()
		ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Status(fiber.StatusOK)
		return nil
	}
}

// MoveTorrent ...
func MoveTorrent(btService *bittorrent.BTService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		torrentID := ctx.Params("torrentId")
		torrent, err := GetTorrentFromParam(btService, torrentID)
		if err != nil {
			ctx.Status(fiber.StatusInternalServerError)
			return fmt.Errorf("Unable to move torrent with index %s", torrentID)
		}

		torrentsLog.Infof("Marking %s to be moved...", torrent.Name())
		btService.MarkedToMove = torrent.InfoHash()

		xbmc.Refresh()
		ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Status(fiber.StatusOK)
		return nil
	}
}

// PauseTorrent ...
func PauseTorrent(btService *bittorrent.BTService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		torrentID := ctx.Params("torrentId")
		torrent, err := GetTorrentFromParam(btService, torrentID)
		if err != nil {
			ctx.Status(fiber.StatusInternalServerError)
			return fmt.Errorf("Unable to pause torrent with index %s", torrentID)
		}

		torrent.Pause()

		xbmc.Refresh()
		ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Status(fiber.StatusOK)
		return nil
	}
}

// RemoveTorrent ...
func RemoveTorrent(btService *bittorrent.BTService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		deleteFiles := ctx.Query("files")

		torrentID := ctx.Params("torrentId")
		torrent, err := GetTorrentFromParam(btService, torrentID)
		if err != nil {
			ctx.Status(fiber.StatusNotFound)
			return fmt.Errorf("Unable to remove torrent with index %s", torrentID)
		}

		// Delete torrent file
		torrentsPath := config.Get().TorrentsPath
		infoHash := torrent.InfoHash()
		torrentFile := filepath.Join(torrentsPath, fmt.Sprintf("%s.torrent", infoHash))
		if _, err := os.Stat(torrentFile); err == nil {
			defer os.Remove(torrentFile)
		}

		torrentsLog.Infof("Removed %s from database", infoHash)

		keepSetting := config.Get().KeepFilesFinished
		deleteAnswer := false
		if keepSetting == 1 && deleteFiles == "" && xbmc.DialogConfirm("Elementum", "LOCALIZE[30269]") {
			deleteAnswer = true
		} else if keepSetting == 2 {
			deleteAnswer = true
		}

		if deleteAnswer == true || deleteFiles == trueType {
			torrentsLog.Info("Removing the torrent and deleting files from the web ...")
			btService.RemoveTorrent(torrent, true)
		} else {
			torrentsLog.Info("Removing the torrent without deleting files from the web ...")
			btService.RemoveTorrent(torrent, false)
		}

		xbmc.Refresh()
		ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Status(fiber.StatusOK)
		return nil
	}
}

// Versions ...
func Versions(btService *bittorrent.BTService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Status(fiber.StatusOK)
		return ctx.JSON(map[string]string{
			"version":    util.GetVersion(),
			"user-agent": btService.UserAgent,
		})
	}
}

// GetTorrentFromParam ...
func GetTorrentFromParam(btService *bittorrent.BTService, param string) (*bittorrent.Torrent, error) {
	if len(param) == 0 {
		return nil, errors.New("Empty param")
	}

	if t, ok := btService.Torrents[param]; ok {
		return t, nil
	}

	return nil, errors.New("Torrent not found")
}

func saveTorrentFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	if file == nil || header == nil {
		return "", fmt.Errorf("Not a valid file entry")
	}

	var err error
	path := filepath.Join(config.Get().TemporaryPath, filepath.Base(header.Filename))
	log.Debugf("Saving incoming torrent file to: %s", path)

	if _, err = os.Stat(path); err != nil && !os.IsNotExist(err) {
		err = os.Remove(path)
		if err != nil {
			return "", fmt.Errorf("Could not remove the file: %s", err)
		}
	}

	out, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("Could not create file: %s", err)
	}
	defer out.Close()
	if _, err = io.Copy(out, file); err != nil {
		return "", fmt.Errorf("Could not write file content: %s", err)
	}

	return path, nil
}
