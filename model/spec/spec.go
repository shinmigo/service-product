package spec

import (
	"fmt"
	"time"

	"goshop/service-product/pkg/db"
)

type Spec struct {
	SpecId    uint64 `gorm:"PRIMARY_KEY"`
	StoreId   uint64
	KindId    uint64
	Name      string
	Sort      uint64
	CreatedBy uint64
	UpdatedBy uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SpecInfo struct {
	SpecId    uint64    `json:"spec_id"`
	Name      string    `json:"name"`
	Sort      uint64    `json:"sort"`
	CreatedBy uint64    `json:"-"`
	CreatedAt time.Time `json:"-"`
}

func GetTableName() string {
	return "spec"
}

func GetField() []string {
	return []string{
		"spec_id", "name", "sort", "created_by", "created_at",
	}
}

func GetOneBySpecId(SpecId uint64) (*SpecInfo, error) {
	if SpecId == 0 {
		return nil, fmt.Errorf("spec_id is null")
	}
	row := &SpecInfo{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("spec_id = ?", SpecId).
		First(row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return row, nil
}

func GetSpecs(page, pageSize uint64) ([]*SpecInfo, error) {
	rows := make([]*SpecInfo, 0, pageSize)
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
