package memory

import (
	"errors"

	"github.com/anacrolix/sync"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"

	"github.com/op/go-logging"

	"github.com/Sanchous98/elementum/bittorrent/reader"
	"github.com/Sanchous98/elementum/config"
	estorage "github.com/Sanchous98/elementum/storage"
	"github.com/Sanchous98/elementum/xbmc"
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
func (s *Storage) SetReadaheadSize(int64) {}

// SetReaders ...
func (s *Storage) SetReaders([]*reader.PositionReader) {
}

// OpenTorrent ...
func (s *Storage) OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (storage.TorrentImpl, error) {
	if !s.haveAvailableMemory() {
		xbmc.Notify("Elementum", "LOCALIZE[30356]", config.AddonIcon())
		return storage.TorrentImpl{}, errors.New("Not enough free memory")
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
