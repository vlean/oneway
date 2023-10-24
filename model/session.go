package model

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-session/session/v3"
	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	SessionId string    `gorm:"column:sid"`
	Value     string    `gorm:"column:value;size:2048;"`
	ExpiredAt time.Time `gorm:"column:expired_at"`
}

func (s *Session) TableName() string {
	return "session"
}

func (s *Session) Encrypt(val map[string]any) {
	v, _ := json.Marshal(val)
	s.Value = string(v)
}

func (s *Session) Decrypt() map[string]any {
	data := make(map[string]any)
	_ = json.Unmarshal([]byte(s.Value), &data)
	return data
}

func (s *Session) Save() error {
	return DB().Save(s).Error
}

func NewSessionManager() *SessionManager {
	return &SessionManager{cache: &sync.Map{}}
}

type SessionManager struct {
	cache *sync.Map
}

func (s *SessionManager) Check(ctx context.Context, sid string) (bool, error) {
	_, ok := s.cache.Load(sid)
	if ok {
		return ok, nil
	}
	ss := &Session{}
	err := Q(ctx).Where("sid=? and expired_at>?", sid, time.Now()).First(ss).Error
	if err != nil {
		return false, nil
	}
	data := make(map[string]any)
	_ = json.Unmarshal([]byte(ss.Value), &data)
	s.cache.Store(sid, &SessionStore{
		data:    ss.Decrypt(),
		ctx:     ctx,
		Session: ss,
		db:      Q(ctx),
	})
	return true, nil
}

func (s *SessionManager) Create(ctx context.Context, sid string, expired int64) (session.Store, error) {
	store := &SessionStore{
		data: make(map[string]any),
		ctx:  ctx,
		Session: &Session{
			SessionId: sid,
			Value:     "{}",
			ExpiredAt: time.Unix(expired, 0),
		},
		db: Q(ctx),
	}
	s.cache.Store(sid, store)
	return store, nil
}

func (s *SessionManager) Update(ctx context.Context, sid string, expired int64) (session.Store, error) {
	if store, ok := s.cache.Load(sid); ok {
		if st, ok := store.(*SessionStore); ok {
			st.Session.ExpiredAt = time.Unix(expired, 0)
			return st, nil
		}
	}
	panic("not found store")
}

func (s *SessionManager) Delete(ctx context.Context, sid string) error {
	s.cache.Delete(sid)
	return nil
}

func (s *SessionManager) Refresh(ctx context.Context, oldsid, sid string, expired int64) (session.Store, error) {
	st, ok := s.cache.LoadAndDelete(oldsid)
	if ok {
		if store, ok := st.(*SessionStore); ok {
			store.SessionId = sid
			s.cache.Store(sid, store)
			return store, nil
		}
	}
	panic("not found store")
}

func (s *SessionManager) Close() error {
	s.cache.Range(func(key, value any) bool {
		k := key.(string)
		val := value.(*SessionStore)
		val.SessionId = k
		return val.Save() == nil
	})
	return nil
}

type SessionStore struct {
	data map[string]any
	ctx  context.Context
	*Session
	db *gorm.DB
}

func (s *SessionStore) Context() context.Context {
	return s.ctx
}

func (s *SessionStore) SessionID() string {
	return s.SessionId
}

func (s *SessionStore) Set(key string, value interface{}) {
	s.data[key] = value
}

func (s *SessionStore) Get(key string) (interface{}, bool) {
	v, ok := s.data[key]
	return v, ok
}

func (s *SessionStore) Delete(key string) interface{} {
	v := s.data[key]
	delete(s.data, key)
	return v
}

func (s *SessionStore) Save() error {
	s.Encrypt(s.data)
	return s.Session.Save()
}

func (s *SessionStore) Flush() error {
	return s.Save()
}
