package product

import (
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
