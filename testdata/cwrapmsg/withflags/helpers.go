package withflags

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
}

func (s *Service) readConn() Conn {
	return Conn{}
}

func load_items() error {
	return errStub
}

func getError() error {
	return errStub
}
