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
