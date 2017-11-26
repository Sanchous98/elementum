package storage

import (
	"github.com/anacrolix/torrent/storage"

	fat32storage "github.com/iamacarpet/go-torrent-storage-fat32"
)

const (
	StorageFile = iota
	StorageMemory
	StorageFat32
)

type ElementumStorage interface {
	storage.ClientImpl

	Start()
	Stop()
	SyncPieces(map[int]bool)
	RemovePiece(int)
}

type DummyStorage struct {
	storage.ClientImpl
	Type int
}

func NewFat32Storage(path string) ElementumStorage {
	return &DummyStorage{fat32storage.NewFat32Storage(path), StorageFat32}
}

func NewFileStorage(path string, pc storage.PieceCompletion) ElementumStorage {
	return &DummyStorage{storage.NewFileWithCompletion(path, pc), StorageFile}
}

func (me *DummyStorage) Start()                    {}
func (me *DummyStorage) Stop()                     {}
func (me *DummyStorage) SyncPieces(a map[int]bool) {}
func (me *DummyStorage) RemovePiece(idx int)       {}
