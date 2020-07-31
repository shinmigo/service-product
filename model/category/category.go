package category

import (
	"fmt"
	"goshop/service-product/pkg/db"
	"time"

	"github.com/shinmigo/pb/productpb"
)

type Category struct {
	CategoryId uint64 `gorm:"PRIMARY_KEY"`
	StoreId    uint64
	ParentId   uint64
	Name       string
	Icon       string
	Status     productpb.CategoryStatus
	Sort       uint64
	CreatedBy  uint64
	UpdatedBy  uint64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func GetTableName() string {
	return "category"
}

type CategoryInfo struct {
	CategoryId uint64                   `json:"category_id"`
	ParentId   uint64                   `json:"parent_id"`
	Name       string                   `json:"name"`
	Icon       string                   `json:"icon"`
	Status     productpb.CategoryStatus `json:"status"`
	Sort       uint64                   `json:"sort"`
}

func GetField() []string {
	return []string{
		"category_id", "parent_id", "name", "icon", "status", "sort",
	}
}

func GetOneByCategoryId(categoryId uint64) (*CategoryInfo, error) {
	if categoryId == 0 {
		return nil, fmt.Errorf("category_id is null")
	}
	row := CategoryInfo{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("category_id = ?", categoryId).
		First(&row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return &row, nil
}

func GetCategories(page, pageSize int64) ([]*CategoryInfo, error) {
	rows := []*CategoryInfo{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Order("category_id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(rows).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return rows, nil
}
