package product_tag

type ProductTag struct {
	ProductTagId uint64 `gorm:"PRIMARY_KEY"`
	ProductId    uint64
	TagId        uint64
}

func GetTableName() string {
	return "product_tag"
}
