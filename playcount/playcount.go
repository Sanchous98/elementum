package playcount

import (
	"fmt"
	"sync"

	"github.com/cespare/xxhash"
)

const (
	// TMDBScraper ...
	TMDBScraper = 0
)
const (
	// MovieType ...
	MovieType = iota
	// ShowType ...
	ShowType
	// SeasonType ...
	SeasonType
	// EpisodeType ...
	EpisodeType
)

var (
	// Mu is a global lock for Playcount package
	Mu = sync.RWMutex{}

	// Watched contains uint64 hashed bools
	Watched = map[uint64]bool{}
)

// WatchedState just a simple bool with Int() conversion
type WatchedState bool

// GetWatchedMovieByTMDB checks whether item is watched
func GetWatchedMovieByTMDB(id int) (ret WatchedState) {
	Mu.RLock()
	defer Mu.RUnlock()

	_, ret = Watched[xxhash.Sum64String(fmt.Sprintf("%d_%d_%d", MovieType, TMDBScraper, id))]
	return
}

// GetWatchedMovieByIMDB checks whether item is watched

// GetWatchedMovieByTrakt checks whether item is watched

// GetWatchedShowByTMDB checks whether item is watched
func GetWatchedShowByTMDB(id int) (ret WatchedState) {
	Mu.RLock()
	defer Mu.RUnlock()

	_, ret = Watched[xxhash.Sum64String(fmt.Sprintf("%d_%d_%d", ShowType, TMDBScraper, id))]
	return
}

// GetWatchedShowByTVDB checks whether item is watched

// GetWatchedShowByTrakt checks whether item is watched

// GetWatchedSeasonByTMDB checks whether item is watched
func GetWatchedSeasonByTMDB(id int, season int) (ret WatchedState) {
	Mu.RLock()
	defer Mu.RUnlock()

	_, ret = Watched[xxhash.Sum64String(fmt.Sprintf("%d_%d_%d_%d", SeasonType, TMDBScraper, id, season))]
	return
}

// GetWatchedSeasonByTVDB checks whether item is watched

// GetWatchedSeasonByTrakt checks whether item is watched

// GetWatchedEpisodeByTMDB checks whether item is watched
func GetWatchedEpisodeByTMDB(id int, season, episode int) (ret WatchedState) {
	Mu.RLock()
	defer Mu.RUnlock()

	_, ret = Watched[xxhash.Sum64String(fmt.Sprintf("%d_%d_%d_%d_%d", EpisodeType, TMDBScraper, id, season, episode))]
	return
}

// GetWatchedEpisodeByTVDB checks whether item is watched

// GetWatchedEpisodeByTrakt checks whether item is watched

// Int converts bool to int
func (w WatchedState) Int() (r int) {
	if w {
		r = 1
	}

	return
}
