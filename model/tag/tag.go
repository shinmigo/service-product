package tag

import (
	"fmt"

	"goshop/service-product/pkg/db"
	"goshop/service-product/pkg/utils"

	"github.com/jinzhu/gorm"

	"github.com/shinmigo/pb/productpb"
)

type Tag struct {
	TagId     uint64         `json:"tag_id" gorm:"PRIMARY_KEY"`
	StoreId   uint64         `json:"store_id"`
	Name      string         `json:"name"`
	Display   int32          `json:"display"`
	Sort      uint64         `json:"sort"`
	CreatedBy uint64         `json:"created_by"`
	UpdatedBy uint64         `json:"updated_by"`
	CreatedAt utils.JSONTime `json:"created_at"`
	UpdatedAt utils.JSONTime `json:"updated_at"`
}

func GetTableName() string {
	return "tag"
}

func GetField() []string {
	return []string{
		"tag_id", "store_id", "name", "display", "sort",
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

//不是所有的id都存在记录，将返回false及不存在的id
func ExistTagsByIds(ids []uint64) ([]uint64, bool) {
	rows := make([]*struct{ TagId uint64 }, 0, len(ids))
	db.Conn.Model(&Tag{}).Select("tag_id").Where("tag_id in (?)", ids).Scan(&rows)

	queryIds := make([]uint64, 0, len(rows))
	for i := range rows {
		queryIds = append(queryIds, rows[i].TagId)
	}
	diffIds := utils.SliceDiffUint64(ids, queryIds)
	return diffIds, len(diffIds) == 0
}

func GetTagList(req *productpb.ListTagReq) ([]*Tag, uint64, error) {
	var total uint64

	var page uint64 = 1
	if req.Page > 0 {
		page = req.Page
	}

	var pageSize uint64 = 10
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	rows := make([]*Tag, 0, pageSize)

	conditions := make([]func(db *gorm.DB) *gorm.DB, 0, 4)
	if req.Id > 0 {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("tag_id = ?", req.Id)
		})
	}

	if req.Name != "" {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("name like ?", "%"+req.Name+"%")
		})
	}

	if req.Display > 0 {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("display = ?", int32(req.Display))
		})
	}

	query := db.Conn.Table(GetTableName()).Select(GetField()).Scopes(conditions...)

	err := query.Order("sort desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	if err != nil {
		return nil, total, err
	}

	query.Count(&total)
	return rows, total, nil
}
