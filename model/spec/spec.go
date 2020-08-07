package spec

import (
	"fmt"
	"goshop/service-product/pkg/utils"

	"goshop/service-product/pkg/db"
)

type Spec struct {
	SpecId    uint64         `json:"spec_id" gorm:"PRIMARY_KEY"`
	StoreId   uint64         `json:"store_id"`
	KindId    uint64         `json:"kind_id"`
	Name      string         `json:"name"`
	Sort      uint64         `json:"sort"`
	CreatedBy uint64         `json:"created_by"`
	UpdatedBy uint64         `json:"updated_by"`
	CreatedAt utils.JSONTime `json:"created_at"`
	UpdatedAt utils.JSONTime `json:"updated_at"`
}

func GetTableName() string {
	return "spec"
}

func GetField() []string {
	return []string{
		"spec_id", "store_id", "kind_id", "name", "sort",
		"created_by", "updated_by", "created_at", "updated_at",
	}
}

func GetOneBySpecId(SpecId uint64) (*Spec, error) {
	if SpecId == 0 {
		return nil, fmt.Errorf("spec_id is null")
	}
	row := &Spec{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("spec_id = ?", SpecId).
		First(row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return row, nil
}

func GetSpecList(specId uint64, specName string, page, pageSize uint64) ([]*Spec, uint64, error) {
	var total uint64

	rows := make([]*Spec, 0, pageSize)

	query := db.Conn.Table(GetTableName()).Select(GetField())
	if specId > 0 {
		query = query.Where("spec_id = ?", specId)
	}

	if specName != "" {
		query = query.Where("name like ?", "%"+specName+"%")
	}

	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	if err != nil {
		return nil, total, err
	}

	query.Count(&total)

	return rows, total, nil
}
