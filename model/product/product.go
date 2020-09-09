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

func GetField() []string {
	return []string{
		"product_id", "category_id", "kind_id", "image", "name", "spec_type", "price", "unit", "short_description", "description", "status",
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

func GetProducts(req *productpb.ListProductReq) ([]*Product, uint64, error) {
	var total uint64
	rows := make([]*Product, 0, req.PageSize)

	query := db.Conn.Model(Product{}).Select(GetField()).Preload("Category").
		Preload("Kind").
		Preload("ProductImage").
		Preload("ProductTag").
		Preload("ProductParam").
		Preload("ProductSpec").
		Order("product_id desc")

	conditions := make([]func(db *gorm.DB) *gorm.DB, 0, 4)
	if req.Name != "" {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("name like ?", fmt.Sprintf("%%%v%%", req.Name))
		})
	}

	if req.Id > 0 {
		conditions = append(conditions, func(db *gorm.DB) *gorm.DB {
			return db.Where("product_id = ?", req.Id)
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

	err := query.Scopes(conditions...).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).Find(&rows).Error

	if err != nil {
		return nil, 0, fmt.Errorf("err: %v", err)
	}

	query.Scopes(conditions...).Count(&total)

	return rows, total, nil
}
