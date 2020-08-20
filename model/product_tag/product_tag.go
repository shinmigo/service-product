package product_tag

import "github.com/jinzhu/gorm"

type ProductTag struct {
	ProductTagId uint64 `gorm:"PRIMARY_KEY"`
	ProductId    uint64
	TagId        uint64
}

func GetTableName() string {
	return "product_tag"
}

func EditTags(db *gorm.DB, productId uint64, tags []uint64) error {
	var (
		err         error
		productTags = make([]*ProductTag, 0, 8)
	)
	db.Model(ProductTag{}).Where("product_id = ?", productId).Find(&productTags)

	var maxLen, tagsLen, productTagsLen int
	tagsLen = len(tags)
	productTagsLen = len(productTags)
	if tagsLen > productTagsLen {
		maxLen = tagsLen
	} else {
		maxLen = productTagsLen
	}

	for i := 0; i < maxLen; i++ {
		if i < tagsLen && i < productTagsLen {
			err = db.Model(&ProductTag{}).
				Where("product_tag_id = ?", productTags[i].ProductTagId).
				Update(map[string]interface{}{
					"tag_id": tags[i],
				}).Error
		} else if i >= tagsLen {
			err = db.Where("product_tag_id IN (?)", productTags[i].ProductTagId).Delete(&ProductTag{}).Error
		} else {
			err = db.Create(&ProductTag{
				ProductId: productId,
				TagId:     tags[i],
			}).Error
		}

		if err != nil {
			return err
		}
	}

	return nil
}
