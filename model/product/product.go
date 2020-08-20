package product

import (
	"goshop/service-product/pkg/db"
	"goshop/service-product/pkg/utils"

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
}

func ExistProductById(id uint64) bool {
	product := Product{}
	db.Conn.Select("product_id").Where("product_id in (?)", id).First(&product)

	return product.ProductId > 0
}

func EditProduct(product map[string]interface{}) error {
	err := db.Conn.Model(&Product{}).Where("product_id = ?", product["product_id"]).Update(product).Error

	return err
}
