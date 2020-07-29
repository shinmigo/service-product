package param

import (
	"fmt"
	"goshop/service-product/pkg/db"
	"time"
)

const (
	ParamTypeText     = 1
	ParamTypeRadio    = 2
	ParamTypeCheckbox = 3
)

type Param struct {
	ParamId   uint64 `gorm:"PRIMARY_KEY"`
	StoreId   uint64
	TypeId    uint64
	Name      string
	Type      int
	Sort      uint64
	CreatedBy uint64
	UpdatedBy uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ParamInfo struct {
	ParamId   uint64    `json:"param_id"`
	Name      string    `json:"name"`
	Type      int       `json:"type"`
	Sort      uint64    `json:"sort"`
	CreatedBy uint64    `json:"-"`
	CreatedAt time.Time `json:"-"`
}

func GetTableName() string {
	return "param"
}

func GetField() []string {
	return []string{
		"param_id", "name", "type", "sort", "created_by", "created_at",
	}
}

func GetOneByParamId(ParamId uint64) (*ParamInfo, error) {
	if ParamId == 0 {
		return nil, fmt.Errorf("param_id is null")
	}
	row := new(ParamInfo)
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("param_id = ?", ParamId).
		First(&row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return row, nil
}

func GetParams(page, pageSize int64) ([]*ParamInfo, error) {
	rows := []*ParamInfo{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return rows, nil
}
