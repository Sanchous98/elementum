package bittorrent

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Sanchous98/elementum/config"
	"github.com/Sanchous98/elementum/xbmc"
)

// DebugAll ...
func DebugAll(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "text/plain")

	writeHeader(ctx, "Torrent Client")
	writeResponse(ctx, "/info")

	writeHeader(ctx, "Debug Perf")
	writeResponse(ctx, "/debug/perf")

	writeHeader(ctx, "Debug LockTimes")
	writeResponse(ctx, "/debug/lockTimes")

	writeHeader(ctx, "Debug Vars")
	writeResponse(ctx, "/debug/vars")
	return nil
}

// DebugBundle ...
func DebugBundle(ctx *fiber.Ctx) error {
	logPath := xbmc.TranslatePath("special://logpath/kodi.log")
	logFile, err := os.Open(logPath)
	if err != nil {
		log.Debugf("Could not open kodi.log: %#v", err)
		return nil
	}
	defer logFile.Close()

	now := time.Now()
	fileName := fmt.Sprintf("bundle_%d_%d_%d_%d_%d.log", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
	ctx.Response().Header.Set("Content-Disposition", "attachment; filename="+fileName)
	ctx.Response().Header.Set("Content-Type", "text/plain")

	writeHeader(ctx, "Torrent Client")
	writeResponse(ctx, "/info")

	writeHeader(ctx, "Debug Perf")
	writeResponse(ctx, "/debug/perf")

	writeHeader(ctx, "Debug LockTimes")
	writeResponse(ctx, "/debug/lockTimes")

	writeHeader(ctx, "Debug Vars")
	writeResponse(ctx, "/debug/vars")

	writeHeader(ctx, "kodi.log")
	io.Copy(ctx, logFile)

	return nil
}

func writeHeader(ctx *fiber.Ctx, title string) {
	ctx.Write([]byte("\n\n" + strings.Repeat("-", 70) + "\n"))
	ctx.Write([]byte(title))
	ctx.Write([]byte("\n" + strings.Repeat("-", 70) + "\n\n"))
}

func writeResponse(ctx *fiber.Ctx, url string) {
	ctx.Write([]byte("Response for url: " + url + "\n\n"))

	resp, err := http.Get(fmt.Sprintf("http://%s:%d%s", config.Args.LocalHost, config.Args.LocalPort, url))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	io.Copy(ctx, resp.Body)
}
