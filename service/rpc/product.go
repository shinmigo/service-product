package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"goshop/service-product/model/param_value"
	"goshop/service-product/model/spec"
	"goshop/service-product/model/spec_value"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	//处理spec_description
	specValueIdList := make([]uint64, 0, 8)
	for i := range req.Spec {
		for k := range req.Spec[i].SpecValueId {
			specValueIdList = append(specValueIdList, req.Spec[i].SpecValueId[k])
		}
	}
	specValueList := make([]*spec_value.SpecValue, 0, 8)
	err = tx.Table(spec_value.GetTableName()).Select([]string{"spec_value_id", "spec_id", "content"}).
		Where("spec_value_id in (?)", specValueIdList).Find(&specValueList).Error
	if err != nil {
		return nil, err
	}
	specValueListLen := len(specValueList)
	specIdList := make([]uint64, 0, 8)
	specValueIdMap := make(map[uint64]*spec_value.SpecValue, specValueListLen)
	for i := range specValueList {
		specIdList = append(specIdList, specValueList[i].SpecId)

		specValueIdMap[specValueList[i].SpecValueId] = &spec_value.SpecValue{
			SpecValueId: specValueList[i].SpecValueId,
			SpecId:      specValueList[i].SpecId,
			Content:     specValueList[i].Content,
		}
	}
	specList := make([]*spec.Spec, 0, len(specIdList))
	err = tx.Table(spec.GetTableName()).Select([]string{"spec_id", "name"}).
		Where("spec_id in (?)", specIdList).Find(&specList).Error
	if err != nil {
		return nil, err
	}
	specIdMap := make(map[uint64]string, len(specIdList))
	for i := range specList {
		specIdMap[specList[i].SpecId] = specList[i].Name
	}
	specDescriptionList := make([]*spec_value.SpecDescription, 0, len(specValueIdList))
	for i := range specValueIdList {
		value, ok := specValueIdMap[specValueIdList[i]]
		if !ok {
			continue
		}
		name, ok := specIdMap[value.SpecId]
		if !ok {
			continue
		}

		specDescriptionList = append(specDescriptionList, &spec_value.SpecDescription{
			SpecId:      value.SpecId,
			Name:        name,
			SpecValueId: specValueIdList[i],
			Content:     value.Content,
		})
	}
	specDesByte, err := json.Marshal(specDescriptionList)
	if err != nil {
		return nil, err
	}

	//处理param_description
	paramLen := len(req.Param)
	paramMap := make(map[uint64]string, paramLen)
	paramValueMap := make(map[uint64]string, paramLen)
	paramIdList := make([]uint64, 0, paramLen)
	paramValueIdList := make([]string, 0, paramLen)
	for i := range req.Param {
		paramIdList = append(paramIdList, req.Param[i].ParamId)
		paramValueIdList = append(paramValueIdList, req.Param[i].Value)
	}

	paramList := make([]*param.Param, paramLen)
	err = tx.Table(param.GetTableName()).Select([]string{"param_id", "name"}).
		Where("param_id in (?)", paramIdList).Find(&paramList).Error
	if err != nil {
		return nil, err
	}
	for i := range paramList {
		paramMap[paramList[i].ParamId] = paramList[i].Name
	}
	paramValueList := make([]*param_value.ParamValue, paramLen)
	err = tx.Table(param_value.GetTableName()).Select([]string{"param_value_id", "content"}).
		Where("param_value_id in (?)", paramValueIdList).Find(&paramValueList).Error
	if err != nil {
		return nil, err
	}
	for i := range paramValueList {
		paramValueMap[paramValueList[i].ParamValueId] = paramValueList[i].Content
	}

	paramDescriptionList := make([]*product_param.ParamDescription, 0, paramLen)
	for i := range req.Param {
		paramName, ok := paramMap[req.Param[i].ParamId]
		if !ok {
			continue
		}
		intValue, err := strconv.Atoi(req.Param[i].Value)
		if err != nil {
			continue
		}
		content, ok := paramValueMap[uint64(intValue)]
		if !ok {
			continue
		}

		paramDescriptionList = append(paramDescriptionList, &product_param.ParamDescription{
			ParamId:    req.Param[i].ParamId,
			ParamName:  paramName,
			ParamValue: req.Param[i].Value,
			Content:    content,
		})
	}
	paramDesByte, err := json.Marshal(paramDescriptionList)
	if err != nil {
		return nil, err
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
		SpecDescription:  string(specDesByte),
		ParamDescription: string(paramDesByte),
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

	//处理spec_description
	specValueIdList := make([]uint64, 0, 8)
	for i := range req.Spec {
		for k := range req.Spec[i].SpecValueId {
			specValueIdList = append(specValueIdList, req.Spec[i].SpecValueId[k])
		}
	}
	specValueList := make([]*spec_value.SpecValue, 0, 8)
	err = tx.Table(spec_value.GetTableName()).Select([]string{"spec_value_id", "spec_id", "content"}).
		Where("spec_value_id in (?)", specValueIdList).Find(&specValueList).Error
	if err != nil {
		return nil, err
	}
	specValueListLen := len(specValueList)
	specIdList := make([]uint64, 0, 8)
	specValueIdMap := make(map[uint64]*spec_value.SpecValue, specValueListLen)
	for i := range specValueList {
		specIdList = append(specIdList, specValueList[i].SpecId)

		specValueIdMap[specValueList[i].SpecValueId] = &spec_value.SpecValue{
			SpecValueId: specValueList[i].SpecValueId,
			SpecId:      specValueList[i].SpecId,
			Content:     specValueList[i].Content,
		}
	}
	specList := make([]*spec.Spec, 0, len(specIdList))
	err = tx.Table(spec.GetTableName()).Select([]string{"spec_id", "name"}).
		Where("spec_id in (?)", specIdList).Find(&specList).Error
	if err != nil {
		return nil, err
	}
	specIdMap := make(map[uint64]string, len(specIdList))
	for i := range specList {
		specIdMap[specList[i].SpecId] = specList[i].Name
	}
	specDescriptionList := make([]*spec_value.SpecDescription, 0, len(specValueIdList))
	for i := range specValueIdList {
		value, ok := specValueIdMap[specValueIdList[i]]
		if !ok {
			continue
		}
		name, ok := specIdMap[value.SpecId]
		if !ok {
			continue
		}

		specDescriptionList = append(specDescriptionList, &spec_value.SpecDescription{
			SpecId:      value.SpecId,
			Name:        name,
			SpecValueId: specValueIdList[i],
			Content:     value.Content,
		})
	}
	specDesByte, err := json.Marshal(specDescriptionList)
	if err != nil {
		return nil, err
	}

	//处理param_description
	paramLen := len(req.Param)
	paramMap := make(map[uint64]string, paramLen)
	paramValueMap := make(map[uint64]string, paramLen)
	paramIdList := make([]uint64, 0, paramLen)
	paramValueIdList := make([]string, 0, paramLen)
	for i := range req.Param {
		paramIdList = append(paramIdList, req.Param[i].ParamId)
		paramValueIdList = append(paramValueIdList, req.Param[i].Value)
	}

	paramList := make([]*param.Param, paramLen)
	err = tx.Table(param.GetTableName()).Select([]string{"param_id", "name"}).
		Where("param_id in (?)", paramIdList).Find(&paramList).Error
	if err != nil {
		return nil, err
	}
	for i := range paramList {
		paramMap[paramList[i].ParamId] = paramList[i].Name
	}
	paramValueList := make([]*param_value.ParamValue, paramLen)
	err = tx.Table(param_value.GetTableName()).Select([]string{"param_value_id", "content"}).
		Where("param_value_id in (?)", paramValueIdList).Find(&paramValueList).Error
	if err != nil {
		return nil, err
	}
	for i := range paramValueList {
		paramValueMap[paramValueList[i].ParamValueId] = paramValueList[i].Content
	}

	paramDescriptionList := make([]*product_param.ParamDescription, 0, paramLen)
	for i := range req.Param {
		paramName, ok := paramMap[req.Param[i].ParamId]
		if !ok {
			continue
		}
		intValue, err := strconv.Atoi(req.Param[i].Value)
		if err != nil {
			continue
		}
		content, ok := paramValueMap[uint64(intValue)]
		if !ok {
			continue
		}

		paramDescriptionList = append(paramDescriptionList, &product_param.ParamDescription{
			ParamId:    req.Param[i].ParamId,
			ParamName:  paramName,
			ParamValue: req.Param[i].Value,
			Content:    content,
		})
	}
	paramDesByte, err := json.Marshal(paramDescriptionList)
	if err != nil {
		return nil, err
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
		"spec_description":  string(specDesByte),
		"param_description": string(paramDesByte),
		"status":            req.Status,
		"updated_by":        req.AdminId,
	}
	if err = product.EditProduct(tx, productMap); err != nil {
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
	products, total, err := product.GetProducts(0, req, nil)
	if err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	productDetails := buildProductDetail(products)

	return &productpb.ListProductRes{
		Total:    total,
		Products: productDetails,
	}, nil
}

func (p *Product) GetProductListByProductSpecIds(ctx context.Context, req *productpb.ProductSpecIdsReq) (*productpb.ListProductSpecRes, error) {
	buf := &productpb.ListProductReq{
		ProductId: req.ProductId,
	}
	products, _, err := product.GetProducts(1, buf, req.ProductSpecId)
	if err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	productDetails := buildProductDetail(products)

	return &productpb.ListProductSpecRes{
		Products: productDetails,
	}, nil
}

func buildProductDetail(products []*product.Product) []*productpb.ProductDetail {
	productDetails := make([]*productpb.ProductDetail, 0, len(products))
	for i := range products {
		var (
			images []string
			tags   []uint64
			specs  []*productpb.ProductSpec
			params []*productpb.ProductParam
		)

		for _, image := range products[i].ProductImage {
			images = append(images, image.Image)
		}

		for _, t := range products[i].ProductTag {
			tags = append(tags, t.TagId)
		}

		for _, spec := range products[i].ProductSpec {
			var (
				specValueId = make([]uint64, 0, 8)
			)
			for _, id := range strings.Split(spec.Spec, ",") {
				specValueId = append(specValueId, uint64(com.StrTo(id).MustInt64()))
			}
			specs = append(specs, &productpb.ProductSpec{
				Image:         spec.Image,
				Price:         spec.Price,
				OldPrice:      spec.OldPrice,
				CostPrice:     spec.CostPrice,
				Stock:         spec.Stock,
				Sku:           spec.Sku,
				Weight:        spec.Weight,
				Volume:        spec.Volume,
				SpecValueId:   specValueId,
				ProductSpecId: spec.ProductSpecId,
			})
		}

		for _, p := range products[i].ProductParam {
			params = append(params, &productpb.ProductParam{
				ParamId: p.ParamId,
				Value:   p.ParamValue,
			})
		}

		path := make([]string, 0, 8)
		if len(products[i].Category.Path) > 0 {
			path = strings.Split(products[i].Category.Path, ",")
		}

		productDetails = append(productDetails, &productpb.ProductDetail{
			ProductId:        products[i].ProductId,
			CategoryId:       products[i].CategoryId,
			CategoryPath:     path,
			KindId:           products[i].KindId,
			Name:             products[i].Name,
			ShortDescription: products[i].ShortDescription,
			Unit:             products[i].Unit,
			Images:           images,
			SpecType:         products[i].SpecType,
			Status:           products[i].Status,
			Tags:             tags,
			Spec:             specs,
			Param:            params,
			Description:      products[i].Description,
			CategoryName:     products[i].Category.Name,
			KindName:         products[i].Kind.Name,
			Price:            products[i].Price,
			SpecDescription:  products[i].SpecDescription,
			ParamDescription: products[i].ParamDescription,
		})
	}
	return productDetails
}
