package product_param

type ProductParam struct {
	ProductParamId uint64 `gorm:"PRIMARY_KEY"`
	ProductId      uint64
	ParamId        uint64
	ParamValue     string
}

func GetTableName() string {
	return "product_param"
}
