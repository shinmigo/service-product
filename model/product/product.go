package product

import (
	"fmt"

	"goshop/service-product/model/category"
	"goshop/service-product/model/kind"
	"goshop/service-product/model/product_image"
	"goshop/service-product/model/product_param"
	"goshop/service-product/model/product_spec"
	"goshop/service-product/model/product_tag"
	"goshop/service-product/pkg/db"
	"goshop/service-product/pkg/utils"

	"github.com/jinzhu/gorm"

	"github.com/shinmigo/pb/productpb"
)

type Product struct {
	ProductId        uint64 `gorm:"PRIMARY_KEY"`
	StoreId          uint64
	CategoryId       uint64
	KindId           uint64
	Image            string
	Name             string
	SpecType         productpb.ProductSpecType
	Price            float64
	Unit             string
	ShortDescription string
	Description      string
	Status           productpb.ProductStatus
	SpecDescription  string
	ParamDescription string
	CreatedBy        uint64
	UpdatedBy        uint64
	CreatedAt        utils.JSONTime
	UpdatedAt        utils.JSONTime
	DeletedAt        *utils.JSONTime
	Kind             kind.Kind                    `gorm:"foreignkey:KindId;association_foreignkey:KindId"`
	ProductImage     []product_image.ProductImage `gorm:"foreignkey:ProductId"`
	Category         category.Category            `gorm:"foreignkey:CategoryId;association_foreignkey:CategoryId"`
	ProductTag       []product_tag.ProductTag     `gorm:"foreignkey:ProductId"`
	ProductParam     []product_param.ProductParam `gorm:"foreignkey:ProductId"`
	ProductSpec      []product_spec.ProductSpec   `gorm:"foreignkey:ProductId"`
}

func GetTableName() string {
	return "product"
}

func GetField() []string {
	return []string{
		"product_id", "category_id", "kind_id", "image", "name", "spec_type", "price", "unit", "short_description", "description", "status", "spec_description", "param_description",
	}
}

func ExistProductById(id uint64, storeId uint64) bool {
	product := Product{}
	db.Conn.Select("product_id").Where("product_id = ? and store_id = ?", id, storeId).First(&product)

	return product.ProductId > 0
}

//根据分类ID确定是否存在商品
func ExistProductByCategoriesId(ids []uint64) bool {
	product := Product{}
	db.Conn.Select("product_id").Where("category_id in (?)", ids).First(&product)

	return product.ProductId > 0
}

func EditProduct(db *gorm.DB, product map[string]interface{}) error {
	err := db.Model(&Product{}).Where("product_id = ?", product["product_id"]).Update(product).Error

	return err
}

func GetProducts(isAll uint8, req *productpb.ListProductReq, productSpecIds []uint64) (list []*Product, total uint64, err error) {
	pageSize := uint64(32)
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}
	rows := make([]*Product, 0, pageSize)

	query := db.Conn.Model(Product{}).Select(GetField()).
		Preload("Category").
		Preload("Kind").
		Preload("ProductImage").
		Preload("ProductTag").
		Preload("ProductParam")

	if len(productSpecIds) > 0 {
		query = query.Preload("ProductSpec", "product_spec_id in (?)", productSpecIds)
	} else {
		query = query.Preload("ProductSpec")
	}
	query = query.Order("product_id desc")

	conditions := make([]func(db *gorm.DB) *gorm.DB, 0, 4)
	if req.Name != "" {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("name like ?", fmt.Sprintf("%%%v%%", req.Name))
		})
	}

	if len(req.ProductId) > 0 {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("product_id in (?)", req.ProductId)
		})
	}

	if req.Status > 0 {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", req.Status)
		})
	}

	if req.StoreId > 0 {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("store_id = ?", req.StoreId)
		})
	}

	if req.CategoryId > 0 {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			var (
				childrenId []uint64
				err        error
			)
			if childrenId, err = category.GetBottomChildrenId([]uint64{req.CategoryId}); err != nil || len(childrenId) == 0 {
				childrenId = []uint64{req.CategoryId}
			}
			return db.Where("category_id in (?)", childrenId)
		})
	}

	if req.TagId > 0 {
		var pids []uint64
		db.Conn.Table(product_tag.GetTableName()).
			Select("product_id").
			Where("tag_id = ?", req.TagId).
			Pluck("product_id", &pids)

		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("product_id in (?)", pids)
		})
	}

	if isAll == 1 {
		err = query.Scopes(conditions...).Find(&rows).Error
		total = uint64(len(rows))
	} else {
		err = query.Scopes(conditions...).
			Offset((req.Page - 1) * pageSize).
			Limit(req.PageSize).Find(&rows).Error

		query.Scopes(conditions...).Count(&total)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("err: %v", err)
	}

	return rows, total, nil
}
