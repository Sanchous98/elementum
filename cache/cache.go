package cache

import (
	"errors"
	"net/http"
	"time"
)

//go:generate msgp -o msgp.go -io=false -tests=false

var errNotSupported = errors.New("cache: not supported")

// CStore ...
type CStore interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}, expire time.Duration) error
	Add(key string, value interface{}, expire time.Duration) error
	Replace(key string, data interface{}, expire time.Duration) error
	Delete(key string) error
	Increment(key string, data uint64) (uint64, error)
	Decrement(key string, data uint64) (uint64, error)
	Flush() error
}

// ResponseCache ...
type ResponseCache struct {
	Status int
	Header http.Header
	Data   []byte
}
