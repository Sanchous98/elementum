package api

import (
	"github.com/Sanchous98/elementum/api/repository"
	"github.com/Sanchous98/elementum/bittorrent"
	"github.com/Sanchous98/elementum/providers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("api")

// Routes ...
func Routes(app *fiber.App, btService *bittorrent.BTService) {
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Output: os.Stdout,
	}))
	//r.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/torrents/list", "/notification"))

	app.Get("/", Index)
	app.Get("/playtorrent", PlayTorrent)
	app.Get("/infolabels", InfoLabelsStored)
	app.Get("/changelog", Changelog)
	app.Get("/status", Status)

	search := app.Group("/search")
	{
		search.Get("", Search(btService))
		search.Get("/remove", SearchRemove)
		search.Get("/clear", SearchClear)
		search.Get("/infolabels/:tmdbId", InfoLabelsSearch(btService))
	}

	//r.LoadHTMLGlob(filepath.Join(config.Get().Info.Path, "resources", "web", "*.html"))
	//web := app.Group("/web")
	//{
	//	web.Get("/", func(c *fiber.Ctx) error {
	//		c.HTML(http.StatusOK, "index.html", nil)
	//	})
	//	web.Static("/static", filepath.Join(config.Get().Info.Path, "resources", "web", "static"))
	//	web.Static("/favicon.ico", filepath.Join(config.Get().Info.Path, "resources", "web", "favicon.ico"))
	//}

	torrents := app.Group("/torrents")
	{
		torrents.Get("/", ListTorrents(btService))
		torrents.All("/add", AddTorrent(btService))
		torrents.Get("/pause", PauseSession)
		torrents.Get("/resume", ResumeSession)
		torrents.Get("/move/:torrentId", MoveTorrent(btService))
		torrents.Get("/pause/:torrentId", PauseTorrent(btService))
		torrents.Get("/resume/:torrentId", ResumeTorrent(btService))
		torrents.Get("/delete/:torrentId", RemoveTorrent(btService))

		// Web UI json
		torrents.Get("/list", ListTorrentsWeb(btService))
	}

	movies := app.Group("/movies")
	{
		movies.Get("/", MoviesIndex)
		movies.Get("/search", SearchMovies)
		movies.Get("/popular", PopularMovies)
		movies.Get("/popular/genre/:genre", PopularMovies)
		movies.Get("/popular/language/:language", PopularMovies)
		movies.Get("/popular/country/:country", PopularMovies)
		movies.Get("/recent", RecentMovies)
		movies.Get("/recent/genre/:genre", RecentMovies)
		movies.Get("/recent/language/:language", RecentMovies)
		movies.Get("/recent/country/:country", RecentMovies)
		movies.Get("/top", TopRatedMovies)
		movies.Get("/imdb250", IMDBTop250)
		movies.Get("/mostvoted", MoviesMostVoted)
		movies.Get("/genres", MovieGenres)
		movies.Get("/languages", MovieLanguages)
		movies.Get("/countries", MovieCountries)

		trakt := movies.Group("/trakt")
		{
			trakt.Get("/watchlist", WatchlistMovies)
			trakt.Get("/collection", CollectionMovies)
			trakt.Get("/popular", TraktPopularMovies)
			trakt.Get("/recommendations", TraktRecommendationsMovies)
			trakt.Get("/trending", TraktTrendingMovies)
			trakt.Get("/toplists", TopTraktLists)
			trakt.Get("/played", TraktMostPlayedMovies)
			trakt.Get("/watched", TraktMostWatchedMovies)
			trakt.Get("/collected", TraktMostCollectedMovies)
			trakt.Get("/anticipated", TraktMostAnticipatedMovies)
			trakt.Get("/boxoffice", TraktBoxOffice)
			trakt.Get("/history", TraktHistoryMovies)

			lists := trakt.Group("/lists")
			{
				lists.Get("/", MoviesTraktLists)
				lists.Get("/:user/:listId", UserlistMovies)
			}

			calendars := trakt.Group("/calendars")
			{
				calendars.Get("/", CalendarMovies)
				calendars.Get("/movies", TraktMyMovies)
				calendars.Get("/releases", TraktMyReleases)
				calendars.Get("/allmovies", TraktAllMovies)
				calendars.Get("/allreleases", TraktAllReleases)
			}
		}
	}
	movie := app.Group("/movie")
	{
		movie.Get("/:tmdbId/infolabels", InfoLabelsMovie)
		movie.Get("/:tmdbId/links", MoviePlaySelector("links", btService))
		movie.Get("/:tmdbId/forcelinks", MoviePlaySelector("forcelinks", btService))
		movie.Get("/:tmdbId/play", MoviePlaySelector("play", btService))
		movie.Get("/:tmdbId/forceplay", MoviePlaySelector("forceplay", btService))
		movie.Get("/:tmdbId/watchlist/add", AddMovieToWatchlist)
		movie.Get("/:tmdbId/watchlist/remove", RemoveMovieFromWatchlist)
		movie.Get("/:tmdbId/collection/add", AddMovieToCollection)
		movie.Get("/:tmdbId/collection/remove", RemoveMovieFromCollection)
	}

	shows := app.Group("/shows")
	{
		shows.Get("/", TVIndex)
		shows.Get("/search", SearchShows)
		shows.Get("/popular", PopularShows)
		shows.Get("/popular/genre/:genre", PopularShows)
		shows.Get("/popular/language/:language", PopularShows)
		shows.Get("/popular/country/:country", PopularShows)
		shows.Get("/recent/shows", RecentShows)
		shows.Get("/recent/shows/genre/:genre", RecentShows)
		shows.Get("/recent/shows/language/:language", RecentShows)
		shows.Get("/recent/shows/country/:country", RecentShows)
		shows.Get("/recent/episodes", RecentEpisodes)
		shows.Get("/recent/episodes/genre/:genre", RecentEpisodes)
		shows.Get("/recent/episodes/language/:language", RecentEpisodes)
		shows.Get("/recent/episodes/country/:country", RecentEpisodes)
		shows.Get("/top", TopRatedShows)
		shows.Get("/mostvoted", TVMostVoted)
		shows.Get("/genres", TVGenres)
		shows.Get("/languages", TVLanguages)
		shows.Get("/countries", TVCountries)

		trakt := shows.Group("/trakt")
		{
			trakt.Get("/watchlist", WatchlistShows)
			trakt.Get("/collection", CollectionShows)
			trakt.Get("/popular", TraktPopularShows)
			trakt.Get("/recommendations", TraktRecommendationsShows)
			trakt.Get("/trending", TraktTrendingShows)
			trakt.Get("/played", TraktMostPlayedShows)
			trakt.Get("/watched", TraktMostWatchedShows)
			trakt.Get("/collected", TraktMostCollectedShows)
			trakt.Get("/anticipated", TraktMostAnticipatedShows)
			trakt.Get("/progress", TraktProgressShows)
			trakt.Get("/history", TraktHistoryShows)

			lists := trakt.Group("/lists")
			{
				lists.Get("/", TVTraktLists)
				lists.Get("/:user/:listId", UserlistShows)
			}

			calendars := trakt.Group("/calendars")
			{
				calendars.Get("/", CalendarShows)
				calendars.Get("/shows", TraktMyShows)
				calendars.Get("/newshows", TraktMyNewShows)
				calendars.Get("/premieres", TraktMyPremieres)
				calendars.Get("/allshows", TraktAllShows)
				calendars.Get("/allnewshows", TraktAllNewShows)
				calendars.Get("/allpremieres", TraktAllPremieres)
			}
		}
	}
	show := app.Group("/show")
	{
		show.Get("/:showId/seasons", ShowSeasons)
		show.Get("/:showId/season/:season/links", ShowSeasonLinks(btService))
		show.Get("/:showId/season/:season/play", ShowSeasonPlay(btService))
		show.Get("/:showId/season/:season/episodes", ShowEpisodes)
		show.Get("/:showId/season/:season/episode/:episode/infolabels", InfoLabelsEpisode())
		show.Get("/:showId/season/:season/episode/:episode/play", ShowEpisodePlaySelector("play", btService))
		show.Get("/:showId/season/:season/episode/:episode/forceplay", ShowEpisodePlaySelector("forceplay", btService))
		show.Get("/:showId/season/:season/episode/:episode/links", ShowEpisodePlaySelector("links", btService))
		show.Get("/:showId/season/:season/episode/:episode/forcelinks", ShowEpisodePlaySelector("forcelinks", btService))
		show.Get("/:showId/watchlist/add", AddShowToWatchlist)
		show.Get("/:showId/watchlist/remove", RemoveShowFromWatchlist)
		show.Get("/:showId/collection/add", AddShowToCollection)
		show.Get("/:showId/collection/remove", RemoveShowFromCollection)
	}
	// TODO
	// episode := r.Group("/episode")
	// {
	// 	episode.GET("/:episodeId/watchlist/add", AddEpisodeToWatchlist)
	// }

	library := app.Group("/library")
	{
		library.Get("/movie/add/:tmdbId", AddMovie)
		library.Get("/movie/remove/:tmdbId", RemoveMovie)
		library.Get("/movie/list/add/:listId", AddMoviesList)
		library.Get("/movie/play/:tmdbId", PlayMovie(btService))
		library.Get("/show/add/:tmdbId", AddShow)
		library.Get("/show/remove/:tmdbId", RemoveShow)
		library.Get("/show/list/add/:listId", AddShowsList)
		library.Get("/show/play/:showId/:season/:episode", PlayShow(btService))

		library.Get("/update", UpdateLibrary)

		// DEPRECATED
		library.Get("/play/movie/:tmdbId", PlayMovie(btService))
		library.Get("/play/show/:showId/season/:season/episode/:episode", PlayShow(btService))
	}

	context := app.Group("/context")
	{
		context.Get("/:media/:kodiID/play", ContextPlaySelector())
	}

	provider := app.Group("/provider")
	{
		provider.Get("/", ProviderList)
		provider.Get("/:provider/check", ProviderCheck)
		provider.Get("/:provider/enable", ProviderEnable)
		provider.Get("/:provider/disable", ProviderDisable)
		provider.Get("/:provider/failure", ProviderFailure)
		provider.Get("/:provider/settings", ProviderSettings)

		provider.Get("/:provider/movie/:tmdbId", ProviderGetMovie)
		provider.Get("/:provider/show/:showId/season/:season/episode/:episode", ProviderGetEpisode)
	}

	allproviders := app.Group("/providers")
	{
		allproviders.Get("/enable", ProvidersEnableAll)
		allproviders.Get("/disable", ProvidersDisableAll)
	}

	repo := app.Group("/repository")
	{
		repo.Get("/:user/:repository/*filepath", repository.GetAddonFiles)
		repo.Head("/:user/:repository/*filepath", repository.GetAddonFilesHead)
	}

	trakt := app.Group("/trakt")
	{
		trakt.Get("/authorize", AuthorizeTrakt)
		trakt.Get("/select_list/:action/:media", SelectTraktUserList)
		trakt.Get("/update", UpdateTrakt)
	}

	app.Get("/migrate/:plugin", MigratePlugin)

	app.Get("/setviewmode/:content_type", SetViewMode)

	app.Get("/subtitles", SubtitlesIndex)
	app.Get("/subtitle/:id", SubtitleGet)

	app.Get("/play", Play(btService))
	app.All("/playuri", PlayURI(btService))

	app.Post("/callbacks/:cid", providers.CallbackHandler)

	// r.GET("/notification", Notification(btService))

	app.Get("/versions", Versions(btService))

	cmd := app.Group("/cmd")
	{
		cmd.Get("/clear_cache_key/:key", ClearCache)
		cmd.Get("/clear_page_cache", ClearPageCache)
		cmd.Get("/clear_trakt_cache", ClearTraktCache)
		cmd.Get("/clear_tmdb_cache", ClearTmdbCache)

		cmd.Get("/reset_path", ResetPath)

		cmd.Get("/paste/:type", Pastebin)

		cmd.Get("/select_network_interface", SelectNetworkInterface)
		cmd.Get("/select_strm_language", SelectStrmLanguage)

		database := cmd.Group("/database")
		{
			database.Get("/clear_movies", ClearDatabaseMovies)
			database.Get("/clear_shows", ClearDatabaseShows)
			database.Get("/clear_torrent_history", ClearDatabaseTorrentHistory)
			database.Get("/clear_search_history", ClearDatabaseSearchHistory)
			database.Get("/clear_database", ClearDatabase)
		}

		cache := cmd.Group("/cache")
		{
			cache.Get("/clear_tmdb", ClearCacheTMDB)
			cache.Get("/clear_trakt", ClearCacheTrakt)
			cache.Get("/clear_cache", ClearCache)
		}
	}
}
