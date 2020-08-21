package rpc

import (
	"bytes"
	"context"
	"errors"
	"goshop/service-product/model/category"
	"goshop/service-product/model/kind"
	"goshop/service-product/model/param"
	"goshop/service-product/model/product"
	"goshop/service-product/model/product_image"
	"goshop/service-product/model/product_param"
	"goshop/service-product/model/product_spec"
	"goshop/service-product/model/product_tag"
	"goshop/service-product/model/tag"
	"goshop/service-product/pkg/db"
	"goshop/service-product/pkg/utils"

	"github.com/unknwon/com"

	"github.com/shinmigo/pb/basepb"
	"github.com/shinmigo/pb/productpb"
)

type Product struct {
}

func NewProduct() *Product {
	return &Product{}
}

func (p *Product) AddProduct(ctx context.Context, req *productpb.Product) (*basepb.AnyRes, error) {
	var err error

	tx := db.Conn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}

		if err != nil {
			tx.Rollback()
		}
	}()

	if _, ok := category.ExistCategoriesByIds([]uint64{req.CategoryId}); !ok {
		return nil, errors.New("商品分类不存在")
	}

	if ok := kind.ExistKindById(req.KindId); !ok {
		return nil, errors.New("商品类型不存在")
	}

	if _, ok := tag.ExistTagsByIds(req.Tags); !ok {
		return nil, errors.New("商品标签不存在")
	}

	paramIds := make([]uint64, 0, len(req.Param))
	for i := range req.Param {
		paramIds = append(paramIds, req.Param[i].ParamId)
	}
	if _, ok := param.ExistParamsByIds(paramIds); !ok {
		return nil, errors.New("商品参数不存在")
	}

	product := product.Product{
		StoreId:          req.StoreId,
		CategoryId:       req.CategoryId,
		KindId:           req.KindId,
		Image:            req.Images[0],
		Name:             req.Name,
		SpecType:         req.SpecType,
		Price:            req.Spec[0].Price,
		Unit:             req.Unit,
		ShortDescription: req.ShortDescription,
		Description:      req.Description,
		Status:           req.Status,
		CreatedBy:        req.AdminId,
		UpdatedBy:        req.AdminId,
	}

	//商品表保存
	if err := tx.Create(&product).Error; err != nil {
		return nil, err
	}

	//商品图片
	var imageValues [][]interface{}
	for i := range req.Images {
		imageValues = append(imageValues, []interface{}{product.ProductId, req.Images[i], i})
	}
	if len(imageValues) > 0 {
		if err = db.BatchInsert(tx, product_image.GetTableName(), []string{"product_id", "image", "sort"}, imageValues); err != nil {
			return nil, err
		}
	}

	//商品标签
	var tagValues [][]interface{}
	for i := range req.Tags {
		tagValues = append(tagValues, []interface{}{product.ProductId, req.Tags[i]})
	}
	if len(tagValues) > 0 {
		if err = db.BatchInsert(tx, product_tag.GetTableName(), []string{"product_id", "tag_id"}, tagValues); err != nil {
			return nil, err
		}
	}

	//商品参数
	var paramValues [][]interface{}
	for i := range req.Param {
		paramValues = append(paramValues, []interface{}{product.ProductId, req.Param[i].ParamId, req.Param[i].Value})
	}
	if len(paramValues) > 0 {
		if err = db.BatchInsert(tx, product_param.GetTableName(), []string{"product_id", "param_id", "param_value"}, paramValues); err != nil {
			return nil, err
		}
	}

	//商品规格
	var now = utils.JSONTime{
		Time: utils.GetNow(),
	}
	var specValues [][]interface{}
	for i := range req.Spec {
		var spec bytes.Buffer
		for j := range req.Spec[i].SpecValueId {
			spec.WriteString(com.ToStr(req.Spec[i].SpecValueId[j]))
			if j < len(req.Spec[i].SpecValueId)-1 {
				spec.WriteString(",")
			}
		}
		specValues = append(specValues, []interface{}{
			product.ProductId,
			req.Spec[i].Sku,
			req.Spec[i].Image,
			req.Spec[i].Price,
			req.Spec[i].OldPrice,
			req.Spec[i].CostPrice,
			req.Spec[i].Stock,
			req.Spec[i].Weight,
			req.Spec[i].Volume,
			spec.String(),
			req.AdminId,
			req.AdminId,
			now,
			now,
		})
	}
	if len(specValues) > 0 {
		if err = db.BatchInsert(tx, product_spec.GetTableName(), []string{"product_id", "sku", "image", "price", "old_price",
			"cost_price", "stock", "weight", "volume", "spec", "created_by", "updated_by", "created_at", "updated_at"},
			specValues); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	return &basepb.AnyRes{
		Id:    product.ProductId,
		State: 1,
	}, nil
}

func (p *Product) EditProduct(ctx context.Context, req *productpb.Product) (*basepb.AnyRes, error) {
	var err error
	tx := db.Conn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		if err != nil {
			tx.Rollback()
		}
	}()

	if ok := product.ExistProductById(req.ProductId, req.StoreId); !ok {
		return nil, errors.New("商品不存在")
	}

	if _, ok := category.ExistCategoriesByIds([]uint64{req.CategoryId}); !ok {
		return nil, errors.New("商品分类不存在")
	}

	if ok := kind.ExistKindById(req.KindId); !ok {
		return nil, errors.New("商品类型不存在")
	}

	if _, ok := tag.ExistTagsByIds(req.Tags); !ok {
		return nil, errors.New("商品标签不存在")
	}

	paramIds := make([]uint64, 0, len(req.Param))
	for i := range req.Param {
		paramIds = append(paramIds, req.Param[i].ParamId)
	}
	if _, ok := param.ExistParamsByIds(paramIds); !ok {
		return nil, errors.New("商品参数不存在")
	}

	//商品
	productMap := map[string]interface{}{
		"product_id":        req.ProductId,
		"store_id":          req.StoreId,
		"category_id":       req.CategoryId,
		"kind_id":           req.KindId,
		"image":             req.Images[0],
		"name":              req.Name,
		"spec_type":         req.SpecType,
		"price":             req.Spec[0].Price,
		"unit":              req.Unit,
		"short_description": req.ShortDescription,
		"description":       req.Description,
		"status":            req.Status,
		"updated_by":        req.AdminId,
	}
	if err = product.EditProduct(productMap); err != nil {
		return nil, err
	}

	//商品图片
	if err = product_image.EditImages(tx, req.ProductId, req.Images); err != nil {
		return nil, err
	}

	//商品标签
	if err = product_tag.EditTags(tx, req.ProductId, req.Tags); err != nil {
		return nil, err
	}

	//商品参数
	var paramValues [][]interface{}
	for i := range req.Param {
		paramValues = append(paramValues, []interface{}{req.Param[i].ParamId, req.Param[i].Value})
	}
	if err = product_param.EditParams(tx, req.ProductId, paramValues); err != nil {
		return nil, err
	}

	//商品规格
	var specValues []map[string]interface{}
	for i := range req.Spec {
		var spec bytes.Buffer
		for j := range req.Spec[i].SpecValueId {
			spec.WriteString(com.ToStr(req.Spec[i].SpecValueId[j]))
			if j < len(req.Spec[i].SpecValueId)-1 {
				spec.WriteString(",")
			}
		}
		specValues = append(specValues, map[string]interface{}{
			"product_spec_id": req.Spec[i].ProductSpecId,
			"sku":             req.Spec[i].Sku,
			"image":           req.Spec[i].Image,
			"price":           req.Spec[i].Price,
			"old_price":       req.Spec[i].OldPrice,
			"cost_price":      req.Spec[i].CostPrice,
			"stock":           req.Spec[i].Stock,
			"weight":          req.Spec[i].Weight,
			"volume":          req.Spec[i].Volume,
			"spec":            spec.String(),
			"admin_id":        req.AdminId,
		})
	}
	if err = product_spec.EditProductSpec(tx, req.ProductId, specValues); err != nil {
		return nil, err
	}

	tx.Commit()

	return &basepb.AnyRes{
		Id:    req.ProductId,
		State: 1,
	}, nil
}

func (p *Product) DelProduct(ctx context.Context, req *productpb.DelProductReq) (*basepb.AnyRes, error) {
	if err := db.Conn.Where("product_id = ? and store_id = ?", req.ProductId, req.StoreId).Delete(&product.Product{}).Error; err != nil {
		return nil, err
	}

	return &basepb.AnyRes{
		Id:    req.ProductId,
		State: 1,
	}, nil
}

func (p *Product) GetProductList(ctx context.Context, req *productpb.ListProductReq) (*productpb.ListProductRes, error) {
	products, total, err := product.GetProducts(req)
	if err != nil {
		return nil, err
	}

	var (
		productDetails = make([]*productpb.ProductDetail, 0, req.PageSize)
	)

	for i := range products {
		productDetails = append(productDetails, &productpb.ProductDetail{
			ProductId:    products[i].ProductId,
			ProductName:  products[i].Name,
			CategoryName: products[i].Category.Name,
			KindName:     products[i].Kind.Name,
			Status:       products[i].Status,
			Price:        products[i].Price,
		})
	}

	return &productpb.ListProductRes{
		Total:    total,
		Products: productDetails,
	}, nil
}
