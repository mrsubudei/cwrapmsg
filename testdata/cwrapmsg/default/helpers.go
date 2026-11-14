package defaults

import (
	"context"

	"github.com/pkg/errors"
)

var errStub = errors.New("error")

type Store struct{}

func (s *Store) GetByIDs(ctx context.Context, ids []int) ([]int, error) {
	return ids, errStub
}

type Conn struct{}

func (c Conn) Execx(query string) error {
	return errStub
}

type Service struct {
	store *Store
	items []*Store
}

func (s *Service) readConn() Conn {
	return Conn{}
}

func getItemIDs() ([]uint64, error) {
	return []uint64{}, errStub
}

func load_items() error {
	return errStub
}

func getError() error {
	return errStub
}

func getStore() *Store {
	return &Store{}
}

func generic[T any](in T) (T, error) {
	return in, errStub
}

func generic2[K comparable, V any](key K, value V) (V, error) {
	return value, errStub
}

func run(fn func() error) error {
	return fn()
}
