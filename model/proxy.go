package model

import (
	"context"
	"sync"
	"time"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Model struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Forward struct {
	Model
	From   string `gorm:"column:from" json:"from"`
	To     string `gorm:"column:to" json:"to"`
	Schema string `gorm:"column:schema;default:http" json:"schema"`
	Client string `gorm:"column:client;default:default" json:"client"`
	Status int    `gorm:"column:status;default:1" json:"status"`
}

func (f *Forward) TableName() string {
	return "forward"
}

var (
	_forwardDao  *ForwardDao
	_forwardOnce sync.Once
)

type ForwardDao struct {
	db     *gorm.DB
	lock   sync.RWMutex
	proxys map[string]*Forward
}

func NewForwardDao() *ForwardDao {
	_forwardOnce.Do(func() {
		_forwardDao = &ForwardDao{db: DB()}
		_forwardDao.load()
	})
	return _forwardDao
}

func (f *ForwardDao) load() {
	res := f.FindAll()
	_fmap := make(map[string]*Forward)
	for _, re := range res {
		tmp := re
		_fmap[tmp.From] = tmp
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	f.proxys = _fmap
}

func (f *ForwardDao) FindAll() []*Forward {
	res := make([]*Forward, 0)
	err := Q(context.Background()).Find(&res).Error
	if err != nil {
		return nil
	}
	return res
}

func (f *ForwardDao) Save(fw *Forward) error {
	err := Q(context.Background()).Save(fw).Error
	if err == nil {
		f.load()
	}
	return err
}
func (f *ForwardDao) Delete(id []int) error {
	err := Q(context.Background()).Where("id in ?", id).Delete(&Forward{}).Error
	if err == nil {
		f.load()
	}
	return err
}

func (f *ForwardDao) Proxy(from string) *Forward {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.proxys[from]
}

func (f *ForwardDao) Domains() []string {
	res := f.FindAll()
	return lo.Uniq(lo.Map(res, func(item *Forward, index int) string {
		return item.From
	}))
}
