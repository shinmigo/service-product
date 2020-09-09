package product_spec

import (
	"goshop/service-product/pkg/utils"

	"github.com/jinzhu/gorm"
)

type ProductSpec struct {
	ProductSpecId uint64 `gorm:"PRIMARY_KEY"`
	ProductId     uint64
	Sku           string
	Image         string
	Price         float64
	OldPrice      float64
	CostPrice     float64
	Stock         uint64
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

func EditProductSpec(db *gorm.DB, productId uint64, specs []map[string]interface{}) error {
	var (
		err         error
		specIds     = make([]uint64, 0, 8)
		productSpec ProductSpec
	)

	for _, spec := range specs {
		if spec["product_spec_id"].(uint64) == 0 {
			productSpec = ProductSpec{
				ProductId: productId,
				Sku:       spec["sku"].(string),
				Image:     spec["image"].(string),
				Price:     spec["price"].(float64),
				OldPrice:  spec["old_price"].(float64),
				CostPrice: spec["cost_price"].(float64),
				Stock:     spec["stock"].(uint64),
				Weight:    spec["weight"].(float64),
				Volume:    spec["volume"].(float64),
				Spec:      spec["spec"].(string),
				CreatedBy: spec["admin_id"].(uint64),
				UpdatedBy: spec["admin_id"].(uint64),
			}
			err = db.Create(&productSpec).Error
			specIds = append(specIds, productSpec.ProductSpecId)
		} else {
			specIds = append(specIds, spec["product_spec_id"].(uint64))
			spec["updated_by"] = spec["admin_id"]
			delete(spec, "admin_id")
			err = db.Model(ProductSpec{}).Where("product_spec_id = ? and product_id = ?", spec["product_spec_id"], productId).
				Update(spec).Error
		}

		if err != nil {
			return err
		}
	}

	//删除规格
	err = db.Where("product_spec_id NOT IN (?) and product_id = ?", specIds, productId).Delete(&ProductSpec{}).Error

	return err
}
