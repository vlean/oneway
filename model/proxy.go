package model

import (
	"context"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Forward struct {
	gorm.Model
	From   string `gorm:"column:from"`
	To     string `gorm:"column:to"`
	Schema string `gorm:"column:schema;default:http"`
	Client string `gorm:"column:client;default:default"`
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
	err := Q(context.Background()).Where("deleted_at is null").Find(&res).Error
	if err != nil {
		return nil
	}
	return res
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

