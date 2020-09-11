package spec

import (
	"fmt"

	"goshop/service-product/model/spec_value"
	"goshop/service-product/pkg/utils"

	"goshop/service-product/pkg/db"
)

type Spec struct {
	SpecId    uint64                  `json:"spec_id" gorm:"PRIMARY_KEY"`
	StoreId   uint64                  `json:"store_id"`
	KindId    uint64                  `json:"kind_id"`
	Name      string                  `json:"name"`
	Sort      uint64                  `json:"sort"`
	CreatedBy uint64                  `json:"created_by"`
	UpdatedBy uint64                  `json:"updated_by"`
	CreatedAt utils.JSONTime          `json:"created_at"`
	UpdatedAt utils.JSONTime          `json:"updated_at"`
	Contents  []*spec_value.SpecValue `json:"contents" gorm:"foreignkey:SpecId;association_foreignkey:SpecId"`
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

func GetOneBySpecId(specId, storeId uint64) (*Spec, error) {
	if specId == 0 {
		return nil, fmt.Errorf("spec_id is null")
	}
	row := &Spec{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Preload("Contents").
		Where("spec_id = ? AND store_id= ?", specId, storeId).
		First(row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return row, nil
}

func GetSpecList(specId uint64, specName string, page, pageSize, storeId uint64) ([]*Spec, uint64, error) {
	var total uint64
	rows := make([]*Spec, 0, pageSize)
	query := db.Conn.Table(GetTableName()).Select(GetField()).
		Preload("Contents").
		Where("store_id = ?", storeId)
	if specId > 0 {
		query = query.Where("spec_id = ?", specId)
	}
	if specName != "" {
		query = query.Where("name like ?", "%"+specName+"%")
	}

	query.Count(&total)
	err := query.Order("spec_id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	if err != nil {
		return nil, total, err
	}

	return rows, total, nil
}

func GetSpecsByKindId(kindIds []uint64) (map[uint64][]*Spec, error) {
	rows := make([]*Spec, 0, len(kindIds))
	err := db.Conn.Table(GetTableName()).
		Preload("Contents").
		Select(GetField()).
		Where("kind_id in (?)", kindIds).
		Find(&rows).Error

	if err != nil {
		return nil, err
	}

	list := make(map[uint64][]*Spec, len(kindIds))
	for k := range rows {
		list[rows[k].KindId] = append(list[rows[k].KindId], rows[k])
	}
	return list, nil
}
