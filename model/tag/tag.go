package tag

import (
	"fmt"
	"time"

	"goshop/service-product/pkg/db"
)

type Tag struct {
	TagId     uint64    `json:"tag_id" gorm:"PRIMARY_KEY"`
	StoreId   uint64    `json:"store_id"`
	Name      string    `json:"name"`
	CreatedBy uint64    `json:"created_by"`
	UpdatedBy uint64    `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetTableName() string {
	return "tag"
}

func GetField() []string {
	return []string{
		"tag_id", "store_id", "name",
		"created_by", "updated_by", "created_at", "updated_at",
	}
}

func GetOneByTagId(tagId uint64) (*Tag, error) {
	if tagId == 0 {
		return nil, fmt.Errorf("tag_id is null")
	}
	row := &Tag{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("tag_id = ?", tagId).
		First(row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return row, nil
}

func GetTagList(tagId uint64, tagName string, page, pageSize uint64) ([]*Tag, uint64, error) {
	var total uint64

	rows := make([]*Tag, 0, pageSize)

	query := db.Conn.Table(GetTableName()).Select(GetField())
	if tagId > 0 {
		query = query.Where("tag_id = ?", tagId)
	}

	if tagName != "" {
		query = query.Where("name like ?", "%"+tagName+"%")
	}

	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	if err != nil {
		return nil, total, err
	}

	query.Count(&total)

	return rows, total, nil
}
