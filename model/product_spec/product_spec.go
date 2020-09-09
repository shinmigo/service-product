package product_spec

import (
	"fmt"

	"goshop/service-product/pkg/db"
	"goshop/service-product/pkg/utils"

	"github.com/jinzhu/gorm"
)

type ProductSpec struct {
	ProductSpecId uint64          `gorm:"PRIMARY_KEY" json:"product_spec_id"`
	ProductId     uint64          `json:"product_id"`
	Sku           string          `json:"sku"`
	Image         string          `json:"image"`
	Price         float64         `json:"price"`
	OldPrice      float64         `json:"old_price"`
	CostPrice     float64         `json:"cost_price"`
	Stock         uint64          `json:"stock"`
	Weight        float64         `json:"weight"`
	Volume        float64         `json:"volume"`
	Spec          string          `json:"spec"`
	CreatedBy     uint64          `json:"-"`
	UpdatedBy     uint64          `json:"-"`
	CreatedAt     utils.JSONTime  `json:"-"`
	UpdatedAt     utils.JSONTime  `json:"-"`
	DeletedAt     *utils.JSONTime `json:"-"`
}

func GetTableName() string {
	return "product_spec"
}

func GetField() []string {
	return []string{
		"product_spec_id", "product_id", "sku", "image",
		"price", "old_price", "cost_price", "stock", "weight", "volume", "spec",
	}
}

func GetProductSpecListByProductSpecId(productSpecIds []uint64) ([]*ProductSpec, error) {
	productSpecIdLen := len(productSpecIds)
	if productSpecIdLen == 0 {
		return nil, nil
	}
	rows := make([]*ProductSpec, 0, productSpecIdLen)
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("product_spec_id in (?)", productSpecIds).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return rows, nil
}

func EditProductSpec(db *gorm.DB, productId uint64, specs []map[string]interface{}) error {
	var (
		err     error
		specIds = make([]uint64, 0, 8)
	)

	for _, spec := range specs {
		if spec["product_spec_id"].(uint64) == 0 {
			err = db.Create(&ProductSpec{
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
			}).Error
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
