package providers

import (
	"github.com/Sanchous98/elementum/bittorrent"
	"github.com/Sanchous98/elementum/tmdb"
)

// Searcher ...
type Searcher interface {
	SearchLinks(query string) []*bittorrent.TorrentFile
}

// MovieSearcher ...
type MovieSearcher interface {
	SearchMovieLinks(movie *tmdb.Movie) []*bittorrent.TorrentFile
}

// SeasonSearcher ...
type SeasonSearcher interface {
	SearchSeasonLinks(show *tmdb.Show, season *tmdb.Season) []*bittorrent.TorrentFile
}

// EpisodeSearcher ...
type EpisodeSearcher interface {
	SearchEpisodeLinks(show *tmdb.Show, episode *tmdb.Episode) []*bittorrent.TorrentFile
}
