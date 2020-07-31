package rpc

import (
	"context"
	"fmt"
	"goshop/service-product/model/category"
	"goshop/service-product/pkg/db"

	"github.com/shinmigo/pb/productpb"
)

type Category struct {
}

func NewCategory() *Category {
	return &Category{}
}

func (c *Category) AddCategory(ctx context.Context, req *productpb.AddCategoryReq) (*productpb.AddCategoryRes, error) {
	if req.Category.ParentId != 0 {
		_, err := category.GetOneByCategoryId(req.Category.ParentId)
		if err != nil {
			return nil, err
		}
	}
	row := category.Category{
		StoreId:   req.Category.StoreId,
		ParentId:  req.Category.ParentId,
		Name:      req.Category.Name,
		Icon:      req.Category.Icon,
		Status:    req.Category.Status,
		Sort:      req.Category.Sort,
		CreatedBy: req.Category.AdminId,
		UpdatedBy: req.Category.AdminId,
	}

	if err := db.Conn.Table(category.GetTableName()).Create(&row).Error; err != nil {
		return nil, err
	}

	return &productpb.AddCategoryRes{
		CategoryId: row.CategoryId,
	}, nil
}

func (c *Category) EditCategory(ctx context.Context, req *productpb.EditCategoryReq) (*productpb.EditCategoryRes, error) {
	if req.Category.ParentId == req.Category.CategoryId {
		return nil, fmt.Errorf("parent_id cannot equal to category_id")
	}
	if req.Category.ParentId != 0 {
		_, err := category.GetOneByCategoryId(req.Category.ParentId)
		if err != nil {
			return nil, err
		}
	}
	if err := db.Conn.Table(category.GetTableName()).Where("category_id = ?", req.Category.CategoryId).Updates(map[string]interface{}{
		"store_id":   req.Category.StoreId,
		"parent_id":  req.Category.ParentId,
		"name":       req.Category.Name,
		"icon":       req.Category.Icon,
		"status":     req.Category.Status,
		"sort":       req.Category.Sort,
		"updated_by": req.Category.AdminId,
	}).Error; err != nil {
		return nil, err
	}

	return &productpb.EditCategoryRes{
		Updated: 1,
	}, nil
}

func (c *Category) DelCategory(ctx context.Context, req *productpb.DelCategoryReq) (*productpb.DelCategoryRes, error) {
	db.Conn.Delete(&category.Category{CategoryId: req.CategoryId})

	return &productpb.DelCategoryRes{
		Deleted: 1,
	}, nil
}

func (c *Category) ReadCategory(ctx context.Context, req *productpb.ReadCategoryReq) (*productpb.ReadCategoryRes, error) {
	row, err := category.GetOneByCategoryId(req.CategoryId)
	if err != nil {
		return nil, err
	}

	return &productpb.ReadCategoryRes{
		Category: &productpb.CategoryInfo{
			CategoryId: row.CategoryId,
			ParentId:   row.ParentId,
			Name:       row.Name,
			Icon:       row.Icon,
			Status:     row.Status,
			Sort:       row.Sort,
		},
	}, nil
}

func (c *Category) ReadCategories(ctx context.Context, req *productpb.ReadCategoriesReq) (*productpb.ReadCategoriesRes, error) {
	rows, err := category.GetCategories(1, 10)
	if err != nil {
		return nil, err
	}

	categories := []*productpb.CategoryInfo{}
	for _, row := range rows {
		categories = append(categories, &productpb.CategoryInfo{
			CategoryId: row.CategoryId,
			ParentId:   row.ParentId,
			Name:       row.Name,
			Icon:       row.Icon,
			Status:     row.Status,
			Sort:       row.Sort,
		})
	}

	return &productpb.ReadCategoriesRes{
		Categories: categories,
	}, nil
}
