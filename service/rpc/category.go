package rpc

import (
	"context"
	"errors"
	"fmt"
	"goshop/service-product/model/category"
	"goshop/service-product/model/product"
	"goshop/service-product/pkg/db"

	"github.com/unknwon/com"

	"github.com/jinzhu/gorm"

	"github.com/shinmigo/pb/basepb"

	"github.com/shinmigo/pb/productpb"
)

type Category struct {
}

func NewCategory() *Category {
	return &Category{}
}

func (c *Category) AddCategory(ctx context.Context, req *productpb.Category) (*basepb.AnyRes, error) {
	var (
		parentCategory *category.Category
		path           = ""
		err            error
	)

	if req.ParentId != 0 {
		if parentCategory, err = category.GetOneByCategoryId(req.ParentId, req.StoreId); err != nil {
			return nil, err
		}
		path = parentCategory.Path
	}

	row := category.Category{
		StoreId:   req.StoreId,
		ParentId:  req.ParentId,
		Name:      req.Name,
		Path:      path,
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
	var (
		parentCategory *category.Category
		oldCategory    *category.Category
		path           = ""
		err            error
	)

	tx := db.Conn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	if req.ParentId == req.CategoryId {
		return nil, fmt.Errorf("parent_id cannot equal to category_id")
	}
	if req.ParentId != 0 {
		if parentCategory, err = category.GetOneByCategoryId(req.ParentId, req.StoreId); err != nil {
			return nil, err
		}
		path = parentCategory.Path
	}
	if oldCategory, err = category.GetOneByCategoryId(req.CategoryId, req.StoreId); err != nil {
		return nil, err
	}
	if oldCategory.CategoryId == 0 {
		return nil, errors.New("category isn't exist")
	}

	if err := tx.Model(category.Category{}).Where(map[string]interface{}{
		"category_id": req.CategoryId,
		"store_id":    req.StoreId,
	}).Update(map[string]interface{}{
		"store_id":   req.StoreId,
		"parent_id":  req.ParentId,
		"name":       req.Name,
		"path":       path,
		"icon":       req.Icon,
		"status":     req.Status,
		"sort":       req.Sort,
		"updated_by": req.AdminId,
	}).Error; err != nil {
		return nil, err
	}

	//更新分类所属需同步更新相关数据
	if req.ParentId != oldCategory.ParentId {
		if oldCategory.ParentId > 0 {
			if err = tx.Model(category.Category{}).Where("category_id = ?", oldCategory.ParentId).
				Update(map[string]interface{}{
					"children_count": gorm.Expr("children_count - ?", 1),
				}).
				Error; err != nil {
				return nil, err
			}
		}
		if req.ParentId > 0 {
			if err = tx.Model(category.Category{}).Where("category_id = ?", req.ParentId).
				Update(map[string]interface{}{
					"children_count": gorm.Expr("children_count + ?", 1),
				}).
				Error; err != nil {
				return nil, err
			}
			path += "," + com.ToStr(req.CategoryId)
		} else {
			path = com.ToStr(req.CategoryId)
		}

		if err = tx.Model(category.Category{}).Where("category_id = ?", req.CategoryId).
			Update("path", path).
			Error; err != nil {
			return nil, err
		}

		//更新子类的path
		if oldCategory.ChildrenCount > 0 {
			expr := "CONCAT('" + path + ",', SUBSTRING(path,POSITION('" + oldCategory.Path + ",' in path)+length('" + oldCategory.Path + ",')))"
			if err = tx.Model(category.Category{}).Where("path like ?", oldCategory.Path+",%").
				Updates(map[string]interface{}{
					"path": gorm.Expr(expr),
				}).Error; err != nil {
				return nil, err
			}
		}
	}

	tx.Commit()

	return &basepb.AnyRes{
		Id:    req.CategoryId,
		State: 1,
	}, nil
}

func (c *Category) EditCategoryStatus(ctx context.Context, req *productpb.EditCategoryStatusReq) (*basepb.AnyRes, error) {
	db.Conn.Table(category.GetTableName()).Where("category_id in (?)", req.CategoryId).Updates(map[string]interface{}{
		"status":     req.Status,
		"updated_by": req.AdminId,
	})

	return &basepb.AnyRes{
		Id:    0,
		State: 1,
	}, nil
}

func (c *Category) DelCategory(ctx context.Context, req *productpb.DelCategoryReq) (*basepb.AnyRes, error) {
	//存在子目录是类目不能删除
	var (
		count       int
		categories  []*category.Category
		parentCount = make(map[uint64]uint64)
		err         error
	)
	db.Conn.Model(&category.Category{}).Where("parent_id in (?)", req.CategoryId).Count(&count)
	if count > 0 {
		return nil, errors.New("some category exist children")
	}

	//存在商品的类目不能删除
	if product.ExistProductByCategoriesId(req.CategoryId) {
		return nil, errors.New("some category exist products")
	}

	tx := db.Conn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	tx.Model(&category.Category{}).Where("category_id in (?)", req.CategoryId).Find(&categories)
	for i := range categories {
		if categories[i].ParentId == 0 {
			continue
		}

		if _, ok := parentCount[categories[i].ParentId]; ok {
			parentCount[categories[i].ParentId]++
		} else {
			parentCount[categories[i].ParentId] = 1
		}
	}

	tx.Where("category_id IN (?)", req.CategoryId).Delete(&category.Category{})
	for categoryId, childrenCount := range parentCount {
		if err := tx.Model(category.Category{}).Where("category_id = ?", categoryId).
			Update("children_count", gorm.Expr("children_count - ?", childrenCount)).
			Error; err != nil {
			return nil, err
		}
	}

	tx.Commit()

	return &basepb.AnyRes{
		Id:    0,
		State: 1,
	}, nil
}

func (c *Category) GetCategoryList(ctx context.Context, req *productpb.ListCategoryReq) (*productpb.ListCategoryRes, error) {
	rows, total, err := category.GetCategories(req)
	if err != nil {
		return nil, err
	}

	categories := make([]*productpb.CategoryDetail, 0, req.PageSize)
	for _, row := range rows {
		categories = append(categories, &productpb.CategoryDetail{
			CategoryId: row.CategoryId,
			ParentId:   row.ParentId,
			Name:       row.Name,
			Path:       row.Path,
			Icon:       row.Icon,
			Status:     row.Status,
			Sort:       row.Sort,
		})
	}

	return &productpb.ListCategoryRes{
		Total:      total,
		Categories: categories,
	}, nil
}
