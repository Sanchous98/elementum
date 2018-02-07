package main

import (
	// _ "net/http/pprof"

	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/op/go-logging"

	"github.com/elgatito/elementum/api"
	"github.com/elgatito/elementum/bittorrent"
	"github.com/elgatito/elementum/config"
	"github.com/elgatito/elementum/database"
	"github.com/elgatito/elementum/library"
	"github.com/elgatito/elementum/lockfile"
	"github.com/elgatito/elementum/trakt"
	"github.com/elgatito/elementum/util"
	"github.com/elgatito/elementum/xbmc"
)

var log = logging.MustGetLogger("main")

func ensureSingleInstance(conf *config.Configuration) (lock *lockfile.LockFile, err error) {
	file := filepath.Join(conf.Info.Path, ".lockfile")
	lock, err = lockfile.New(file)
	if err != nil {
		log.Critical("Unable to initialize lockfile:", err)
		return
	}
	var pid int
	var p *os.Process
	pid, err = lock.Lock()
	if err != nil {
		log.Warningf("Unable to acquire lock %q: %v, killing...", lock.File, err)
		p, err = os.FindProcess(pid)
		if err != nil {
			log.Warning("Unable to find other process:", err)
			return
		}
		if err = p.Kill(); err != nil {
			log.Critical("Unable to kill other process:", err)
			return
		}
		if err = os.Remove(lock.File); err != nil {
			log.Critical("Unable to remove lockfile")
			return
		}
		_, err = lock.Lock()
	}
	return
}

func main() {
	// Make sure we are properly multithreaded.
	runtime.GOMAXPROCS(runtime.NumCPU())

	logging.SetFormatter(logging.MustStringFormatter(
		`%{color}%{level:.4s}  %{module:-12s} ▶ %{shortfunc:-15s}  %{color:reset}%{message}`,
	))
	logging.SetBackend(logging.NewLogBackend(ioutil.Discard, "", 0), logging.NewLogBackend(os.Stdout, "", 0))

	log.Infof("Starting Elementum daemon")
	log.Infof("Version: %s Go: %s", util.GetVersion(), runtime.Version())

	conf := config.Reload()

	log.Infof("Addon: %s v%s", conf.Info.ID, conf.Info.Version)

	lock, err := ensureSingleInstance(conf)
	defer lock.Unlock()
	if err != nil {
		log.Warningf("Unable to acquire lock %q: %v, exiting...", lock.File, err)
		os.Exit(1)
	}

	wasFirstRun := Migrate()

	db, err := database.InitSqliteDB(conf)
	if err != nil {
		log.Error(err)
		return
	}

	boltDb, err := database.InitBoltDB(conf)
	if err != nil {
		log.Error(err)
		return
	}

	cacheDb, errCache := database.InitCacheDB(conf)
	if errCache != nil {
		log.Error(errCache)
		return
	}

	btService := bittorrent.NewBTService()

	var shutdown = func(fromSignal bool) {
		log.Info("Shutting down...")
		library.CloseLibrary()
		btService.Close(true)

		db.Close()
		boltDb.Close()
		cacheDb.Close()

		log.Info("Goodbye")

		// If we don't give an exit code - python treat as well done and not
		// restarting the daemon. So when we come here from Signal -
		// we should properly exit with non-0 exitcode.
		if !fromSignal {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	var watchParentProcess = func() {
		for {
			if os.Getppid() == 1 {
				log.Warning("Parent shut down, shutting down too...")
				go shutdown(false)
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
	go watchParentProcess()

	http.Handle("/", api.Routes(btService))
	http.Handle("/info", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		btService.ClientInfo(w)
	}))
	http.Handle("/files/", bittorrent.ServeTorrent(btService, config.Get().DownloadPath))
	http.Handle("/reload", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		btService.Reconfigure()
	}))
	http.Handle("/notification", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Notification(w, r, btService)
	}))
	http.Handle("/shutdown", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shutdown(false)
	}))

	xbmc.Notify("Elementum", "LOCALIZE[30208]", config.AddonIcon())

	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-sigc
		shutdown(true)
	}()

	go func() {
		if !wasFirstRun {
			log.Info("Updating Kodi add-on repositories...")
			xbmc.UpdateAddonRepos()
		}

		xbmc.DialogProgressBGCleanup()
		xbmc.ResetRPC()
	}()

	go library.Init()
	go trakt.TokenRefreshHandler()
	go db.MaintenanceRefreshHandler()
	go boltDb.MaintenanceRefreshHandler()
	go cacheDb.MaintenanceRefreshHandler()

	http.ListenAndServe(":"+strconv.Itoa(config.ListenPort), nil)

	// s := &http.Server{
	// 	Addr:         ":" + strconv.Itoa(config.ListenPort),
	// 	Handler:      nil,
	// 	// ReadTimeout:  60 * time.Second,
	// 	// WriteTimeout: 60 * time.Second,
	//
	// 	MaxHeaderBytes: 1 << 20,
	// }
	// s.ListenAndServe()
}
