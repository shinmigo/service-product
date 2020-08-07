package kind

import (
	"fmt"
	"goshop/service-product/pkg/db"
	"goshop/service-product/pkg/utils"
)

type Kind struct {
	KindId    uint64         `json:"kind_id" gorm:"PRIMARY_KEY"`
	StoreId   uint64         `json:"store_id"`
	Name      string         `json:"name"`
	ParamQty  uint64         `json:"param_qty"`
	SpecQty   uint64         `json:"spec_qty"`
	CreatedBy uint64         `json:"created_by"`
	UpdatedBy uint64         `json:"updated_by"`
	CreatedAt utils.JSONTime `json:"created_at"`
	UpdatedAt utils.JSONTime `json:"updated_at"`
}

func GetTableName() string {
	return "kind"
}

func GetField() []string {
	return []string{
		"kind_id", "store_id", "name", "param_qty", "spec_qty",
		"created_by", "updated_by", "created_at", "updated_at",
	}
}

func GetOneByKindId(KindId uint64) (*Kind, error) {
	if KindId == 0 {
		return nil, fmt.Errorf("kind_id is null")
	}
	row := &Kind{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("kind_id = ?", KindId).
		First(row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return row, nil
}

func GetKindList(kindId uint64, kindName string, page, pageSize uint64) ([]*Kind, uint64, error) {
	var total uint64

	rows := make([]*Kind, 0, pageSize)

	query := db.Conn.Table(GetTableName()).Select(GetField())
	if kindId > 0 {
		query = query.Where("kind_id = ?", kindId)
	}

	if kindName != "" {
		query = query.Where("name like ?", "%"+kindName+"%")
	}

	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	if err != nil {
		return nil, total, err
	}

	query.Count(&total)

	return rows, total, nil
}
