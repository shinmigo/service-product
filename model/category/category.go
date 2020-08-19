package category

import (
	"fmt"
	"goshop/service-product/pkg/db"
	"goshop/service-product/pkg/utils"

	"github.com/jinzhu/gorm"

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
	CreatedAt  utils.JSONTime
	UpdatedAt  utils.JSONTime
	DeletedAt  *utils.JSONTime
}

func GetTableName() string {
	return "category"
}

type Info struct {
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

func GetOneByCategoryId(categoryId uint64) (*Category, error) {
	if categoryId == 0 {
		return nil, fmt.Errorf("category_id is null")
	}
	row := &Category{}
	err := db.Conn.
		Select(GetField()).
		Where("category_id = ?", categoryId).
		First(row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return row, nil
}

//不是所有的id都存在记录，将返回false及不存在的id
func ExistCategoriesByIds(ids []uint64) ([]uint64, bool) {
	rows := make([]*struct{ CategoryId uint64 }, 0, len(ids))
	db.Conn.Model(&Category{}).Select("category_id").Where("category_id in (?)", ids).Scan(&rows)

	queryIds := make([]uint64, 0, len(rows))
	for i := range rows {
		queryIds = append(queryIds, rows[i].CategoryId)
	}
	diffIds := utils.SliceDiffUint64(ids, queryIds)
	return diffIds, len(diffIds) == 0
}

func GetCategories(req *productpb.ListCategoryReq) ([]*Category, uint64, error) {
	var total uint64
	rows := make([]*Category, 0, req.PageSize)

	query := db.Conn.Model(Category{}).
		Select(GetField()).
		Order("category_id desc")

	conditions := make([]func(db *gorm.DB) *gorm.DB, 0, 4)
	if req.Name != "" {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("name = ?", req.Name)
		})
	}
	if req.GetStatusPresent() != nil {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", req.GetStatus())
		})
	}
	if req.Id > 0 {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("category_id = ?", req.Id)
		})
	}
	if req.StoreId > 0 {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("store_id = ?", req.StoreId)
		})
	}

	err := query.Scopes(conditions...).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).Find(&rows).Error

	if err != nil {
		return nil, 0, fmt.Errorf("err: %v", err)
	}

	query.Scopes(conditions...).Count(&total)

	return rows, total, nil
}

func EditCategory(id uint64, data interface{}) bool {
	db.Conn.Model(&Category{}).Where("category_id = ?", id).Update(data)

	return true
}
