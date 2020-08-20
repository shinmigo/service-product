package product_param

import "github.com/jinzhu/gorm"

type ProductParam struct {
	ProductParamId uint64 `gorm:"PRIMARY_KEY"`
	ProductId      uint64
	ParamId        uint64
	ParamValue     string
}

func GetTableName() string {
	return "product_param"
}

func EditParams(db *gorm.DB, productId uint64, params [][]interface{}) error {
	var (
		err           error
		productParams = make([]*ProductParam, 0, 8)
	)
	db.Model(ProductParam{}).Where("product_id = ?", productId).Find(&productParams)

	var maxLen, paramsLen, productParamsLen int
	paramsLen = len(params)
	productParamsLen = len(productParams)
	if paramsLen > productParamsLen {
		maxLen = paramsLen
	} else {
		maxLen = productParamsLen
	}

	for i := 0; i < maxLen; i++ {
		if i < paramsLen && i < productParamsLen {
			err = db.Model(&ProductParam{}).
				Where("product_param_id = ?", productParams[i].ProductParamId).
				Update(map[string]interface{}{
					"param_id":    params[i][0].(uint64),
					"param_value": params[i][1].(string),
				}).Error
		} else if i >= paramsLen {
			err = db.Where("product_param_id IN (?)", productParams[i].ProductParamId).Delete(&ProductParam{}).Error
		} else {
			err = db.Create(&ProductParam{
				ProductId:  productId,
				ParamId:    params[i][0].(uint64),
				ParamValue: params[i][1].(string),
			}).Error
		}

		if err != nil {
			return err
		}
	}

	return nil
}
