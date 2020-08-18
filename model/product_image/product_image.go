package product_image

type ProductImage struct {
	ProductImageId uint64 `gorm:"PRIMARY_KEY"`
	ProductId      uint64
	Image          string
	Sort           int
}

func GetTableName() string {
	return "product_image"
}
