package model

import (
	"context"
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
	_forwardGlobal map[string]*Forward
)

type ForwardDao struct {
	db *gorm.DB
}

func NewForwardDao() *ForwardDao {
	return &ForwardDao{db: DB()}
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
	return Q(context.Background()).Save(fw).Error
}
func (f *ForwardDao) Delete(id []int) error {
	return Q(context.Background()).Where("id in ?", id).Delete(&Forward{}).Error
}

func (f *ForwardDao) Proxy(from string) *Forward {
	res := f.FindAll()
	_fmap := make(map[string]*Forward)
	for _, re := range res {
		tmp := re
		_fmap[tmp.From] = tmp
	}
	return _fmap[from]
}

func (f *ForwardDao) Domains() []string {
	res := f.FindAll()
	return lo.Uniq(lo.Map(res, func(item *Forward, index int) string {
		return item.From
	}))
}
