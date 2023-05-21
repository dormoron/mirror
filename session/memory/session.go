package memory

import (
	"context"
	"errors"
	"fmt"
	"github.com/nothingZero/mirror/session"
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

var (
	errorKeyNotFound     = errors.New("session: 找不到 key")
	errorSessionNotFound = errors.New("session: 找不到 session")
)

type Store struct {
	mutex      sync.RWMutex
	sessions   *cache.Cache
	expiration time.Duration
}

func InitStore(expiration time.Duration) *Store {
	return &Store{
		sessions:   cache.New(expiration, time.Second),
		expiration: expiration,
	}
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	sess := &Session{
		id:     id,
		values: sync.Map{},
	}
	s.sessions.Set(id, sess, s.expiration)
	return sess, nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	val, ok := s.sessions.Get(id)
	if !ok {
		return errors.New("session: 该 id 对应的 session 不存在")
	}
	s.sessions.Set(id, val, s.expiration)
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.sessions.Delete(id)
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	s.mutex.RLock()
	defer s.mutex.Unlock()
	sess, ok := s.sessions.Get(id)
	if !ok {
		return nil, errorSessionNotFound
	}
	return sess.(*Session), nil
}

type Session struct {
	id     string
	values sync.Map
}

func (s *Session) Get(ctx context.Context, key string) (any, error) {
	val, ok := s.values.Load(key)
	if !ok {
		return nil, fmt.Errorf("%w, key %s", errorKeyNotFound, key)
	}
	return val, nil
}

func (s *Session) Set(ctx context.Context, key string, value any) error {
	s.values.Store(key, value)
	return nil
}

func (s *Session) ID() string {
	return s.id
}
