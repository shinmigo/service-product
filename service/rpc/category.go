package rpc

import (
	"context"
	"fmt"
	"goshop/service-product/model/category"
	"goshop/service-product/pkg/db"

	"github.com/shinmigo/pb/basepb"

	"github.com/shinmigo/pb/productpb"
)

type Category struct {
}

func NewCategory() *Category {
	return &Category{}
}

func (c *Category) AddCategory(ctx context.Context, req *productpb.Category) (*basepb.AnyRes, error) {
	if req.ParentId != 0 {
		if _, err := category.GetOneByCategoryId(req.ParentId); err != nil {
			return nil, err
		}
	}

	row := category.Category{
		StoreId:   req.StoreId,
		ParentId:  req.ParentId,
		Name:      req.Name,
		Icon:      req.Icon,
		Status:    req.Status,
		Sort:      req.Sort,
		CreatedBy: req.AdminId,
		UpdatedBy: req.AdminId,
	}
	if err := db.Conn.Create(&row).Error; err != nil {
		return nil, err
	}

	return &basepb.AnyRes{
		Id:    row.CategoryId,
		State: 1,
	}, nil
}

func (c *Category) EditCategory(ctx context.Context, req *productpb.Category) (*basepb.AnyRes, error) {
	if req.ParentId == req.CategoryId {
		return nil, fmt.Errorf("parent_id cannot equal to category_id")
	}
	if req.ParentId != 0 {
		_, err := category.GetOneByCategoryId(req.ParentId)
		if err != nil {
			return nil, err
		}
	}
	if err := db.Conn.Table(category.GetTableName()).Where("category_id = ?", req.CategoryId).Updates(map[string]interface{}{
		"store_id":   req.StoreId,
		"parent_id":  req.ParentId,
		"name":       req.Name,
		"icon":       req.Icon,
		"status":     req.Status,
		"sort":       req.Sort,
		"updated_by": req.AdminId,
	}).Error; err != nil {
		return nil, err
	}

	return &basepb.AnyRes{
		Id:    req.CategoryId,
		State: 1,
	}, nil
}

func (c *Category) EditCategoryStatus(ctx context.Context, req *productpb.EditCategoryStatusReq) (*basepb.AnyRes, error) {
	db.Conn.Table(category.GetTableName()).Updates(map[string]interface{}{
		"status": req.Status,
	})

	return &basepb.AnyRes{
		Id:    0,
		State: 1,
	}, nil
}

func (c *Category) DelCategory(ctx context.Context, req *productpb.DelCategoryReq) (*basepb.AnyRes, error) {
	db.Conn.Where("category_id IN (?)", req.CategoryId).Delete(&category.Category{})

	return &basepb.AnyRes{
		Id:    0,
		State: 1,
	}, nil
}

func (c *Category) GetCategoryList(ctx context.Context, req *productpb.ListCategoryReq) (*productpb.ListCategoryRes, error) {
	rows, err := category.GetCategories(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	categories := make([]*productpb.CategoryDetail, 0, req.PageSize)
	for _, row := range rows {
		categories = append(categories, &productpb.CategoryDetail{
			CategoryId: row.CategoryId,
			ParentId:   row.ParentId,
			Name:       row.Name,
			Icon:       row.Icon,
			Status:     row.Status,
			Sort:       row.Sort,
		})
	}

	return &productpb.ListCategoryRes{
		Total:      uint64(len(categories)),
		Categories: categories,
	}, nil
}
