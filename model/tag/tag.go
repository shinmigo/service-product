package tag

import (
	"fmt"
	"goshop/service-product/pkg/db"
	"time"
)

type Tag struct {
	TagId     uint64 `gorm:"PRIMARY_KEY"`
	StoreId   uint64
	Name      string
	CreatedBy uint64
	UpdatedBy uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TagInfo struct {
	TagId uint64 `json:"tag_id"`
	Name  string `json:"name"`
}

func GetTableName() string {
	return "tag"
}

func GetField() []string {
	return []string{
		"tag_id", "name",
	}
}

func GetOneByTagId(tagId uint64) (*TagInfo, error) {
	if tagId == 0 {
		return nil, fmt.Errorf("tag_id is null")
	}
	row := &TagInfo{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("tag_id = ?", tagId).
		First(row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return row, nil
}

func GetTags(page, pageSize int64) ([]*TagInfo, error) {
	rows := []*TagInfo{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Order("tag_id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return rows, nil
}
