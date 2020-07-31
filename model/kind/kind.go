package kind

import (
	"fmt"
	"time"

	"goshop/service-product/pkg/db"
)

type Kind struct {
	KindId    uint64 `gorm:"PRIMARY_KEY"`
	StoreId   uint64
	Name      string
	ParamQty  uint64
	SpecQty   uint64
	CreatedBy uint64
	UpdatedBy uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type KindInfo struct {
	KindId   uint64 `json:"kind_id"`
	Name     string `json:"name"`
	ParamQty uint64 `json:"param_qty"`
	SpecQty  uint64 `json:"spec_qty"`
}

func GetTableName() string {
	return "kind"
}

func GetField() []string {
	return []string{
		"kind_id", "name", "param_qty", "spec_qty",
	}
}

func GetOneByKindId(KindId uint64) (*KindInfo, error) {
	if KindId == 0 {
		return nil, fmt.Errorf("kind_id is null")
	}
	row := &KindInfo{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("kind_id = ?", KindId).
		First(row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return row, nil
}

func GetKinds(page, pageSize int64) ([]*KindInfo, error) {
	rows := []*KindInfo{}
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
