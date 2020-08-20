package product_image

import (
	"github.com/jinzhu/gorm"
)

type ProductImage struct {
	ProductImageId uint64 `gorm:"PRIMARY_KEY"`
	ProductId      uint64
	Image          string
	Sort           int
}

func GetTableName() string {
	return "product_image"
}

func EditImages(db *gorm.DB, productId uint64, images []string) error {
	var (
		err           error
		productImages = make([]*ProductImage, 0, 8)
	)
	db.Model(ProductImage{}).Where("product_id = ?", productId).Find(&productImages)

	var maxLen, imagesLen, productImagesLen int
	imagesLen = len(images)
	productImagesLen = len(productImages)
	if imagesLen > productImagesLen {
		maxLen = imagesLen
	} else {
		maxLen = productImagesLen
	}

	for i := 0; i < maxLen; i++ {
		if i < imagesLen && i < productImagesLen {
			err = db.Model(&ProductImage{}).
				Where("product_image_id = ?", productImages[i].ProductImageId).
				Update(map[string]interface{}{
					"image": images[i],
				}).Error
		} else if i >= imagesLen {
			err = db.Where("product_image_id IN (?)", productImages[i].ProductImageId).Delete(&ProductImage{}).Error
		} else {
			err = db.Create(&ProductImage{
				ProductId: productId,
				Image:     images[i],
				Sort:      i,
			}).Error
		}

		if err != nil {
			return err
		}
	}

	return nil
}
