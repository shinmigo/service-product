package product_spec

import "goshop/service-product/pkg/utils"

type ProductSpec struct {
	ProductSpecId uint64 `gorm:"PRIMARY_KEY"`
	ProductId     uint64
	Sku           string
	Image         string
	Price         float64
	OldPrice      float64
	CostPrice     float64
	Stock         int
	Weight        float64
	Volume        float64
	Spec          string
	CreatedBy     uint64
	UpdatedBy     uint64
	CreatedAt     utils.JSONTime
	UpdatedAt     utils.JSONTime
	DeletedAt     *utils.JSONTime
}

func GetTableName() string {
	return "product_spec"
}
