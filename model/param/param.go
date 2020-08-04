package param

import (
	"fmt"
	"time"

	"goshop/service-product/pkg/db"
)

type Param struct {
	ParamId   uint64 `gorm:"PRIMARY_KEY"`
	StoreId   uint64
	KindId    uint64
	Name      string
	Type      int32
	Sort      uint64
	CreatedBy uint64
	UpdatedBy uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ParamInfo struct {
	ParamId   uint64    `json:"param_id"`
	Name      string    `json:"name"`
	Type      int32     `json:"type"`
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
	row := &ParamInfo{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("param_id = ?", ParamId).
		First(row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return row, nil
}

func GetParams(page, pageSize uint64) ([]*ParamInfo, error) {
	rows := make([]*ParamInfo, 0, pageSize)
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
