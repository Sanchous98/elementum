package memory

import (
	"errors"

	"github.com/anacrolix/sync"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"

	"github.com/dustin/go-humanize"
	"github.com/op/go-logging"

	"github.com/elgatito/elementum/bittorrent/reader"
	"github.com/elgatito/elementum/config"
	estorage "github.com/elgatito/elementum/storage"
	"github.com/elgatito/elementum/xbmc"
)

const (
	chunkSize = 1024 * 16
	// readaheadRatio = 0.33
)

var (
	log = logging.MustGetLogger("memory")
)

// Storage main object
type Storage struct {
	Type     int
	mu       *sync.Mutex
	items    map[string]*Cache
	capacity int64
}

// NewMemoryStorage initializer function
func NewMemoryStorage(maxMemorySize int64) *Storage {
	log.Infof("Initializing memory storage of size: %s", humanize.Bytes(uint64(maxMemorySize)))
	s := &Storage{
		mu:       &sync.Mutex{},
		capacity: maxMemorySize,
		items:    map[string]*Cache{},
	}

	return s
}

// GetTorrentStorage ...
func (s *Storage) GetTorrentStorage(hash string) estorage.TorrentStorage {
	if i, ok := s.items[hash]; ok {
		return i
	}

	return nil
}

// Close ...
func (s *Storage) Close() error {
	return nil
}

// GetReadaheadSize ...
func (s *Storage) GetReadaheadSize() int64 {
	return s.capacity
}

// SetReadaheadSize ...
func (s *Storage) SetReadaheadSize(size int64) {}

// SetReaders ...
func (s *Storage) SetReaders(readers []*reader.PositionReader) {
}

// OpenTorrent ...
func (s *Storage) OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (storage.TorrentImpl, error) {
	if !s.haveAvailableMemory() {
		xbmc.Notify("Elementum", "LOCALIZE[30356]", config.AddonIcon())
		return nil, errors.New("Not enough free memory")
	}

	c := &Cache{
		s:        s,
		capacity: s.capacity,
		id:       infoHash.HexString(),
	}
	c.Init(info)
	go c.Start()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[c.id] = c

	return c, nil
}

// TODO: add checker for memory usage.
// Use physical free memory or try to detect free to allocate?
func (s *Storage) haveAvailableMemory() bool {
	return true
}
