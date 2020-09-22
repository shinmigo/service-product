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
	specValueIdMap := make(map[uint64]map[uint64]interface{}, specValueListLen)
	for i := range specValueList {
		specIdList = append(specIdList, specValueList[i].SpecId)

		if _, ok := specValueIdMap[specValueList[i].SpecId]; ok {
			specValueIdMap[specValueList[i].SpecId][specValueList[i].SpecValueId] = map[string]interface{}{
				"spec_value_id": specValueList[i].SpecValueId,
				"spec_id":       specValueList[i].SpecId,
				"content":       specValueList[i].Content,
			}
		} else {
			buf := make(map[uint64]interface{})
			buf[specValueList[i].SpecValueId] = map[string]interface{}{
				"spec_value_id": specValueList[i].SpecValueId,
				"spec_id":       specValueList[i].SpecId,
				"content":       specValueList[i].Content,
			}
			specValueIdMap[specValueList[i].SpecId] = buf
		}
	}
	specList := make([]*spec.Spec, 0, len(specIdList))
	err = tx.Table(spec.GetTableName()).Select([]string{"spec_id", "name"}).
		Where("spec_id in (?)", specIdList).Order("sort asc").
		Find(&specList).Error
	if err != nil {
		return nil, err
	}

	specDescriptionList := make([]map[string]interface{}, 0, len(specIdList))
	for i := range specList {
		if _, ok := specValueIdMap[specList[i].SpecId]; !ok {
			continue
		}
		buf := make(map[string]interface{})
		buf["spec_id"] = specList[i].SpecId
		buf["name"] = specList[i].Name
		buf["children"] = specValueIdMap[specList[i].SpecId]
		specDescriptionList = append(specDescriptionList, buf)
	}

	specDesByte, err := json.Marshal(specDescriptionList)
	if err != nil {
		return nil, err
	}

	//处理param_description
	paramValueIdList := make([]uint64, 0, 8)
	for k := range req.Param {
		valueId, _ := strconv.ParseUint(req.Param[k].Value, 10, 64)
		paramValueIdList = append(paramValueIdList, valueId)
	}

	paramValueList := make([]*param_value.ParamValue, 0, 8)
	err = tx.Table(param_value.GetTableName()).Select([]string{"param_value_id", "param_id", "content"}).
		Where("param_value_id in (?)", paramValueIdList).Find(&paramValueList).Error
	if err != nil {
		return nil, err
	}

	paramValueListLen := len(paramValueList)
	paramIdList := make([]uint64, 0, 8)
	paramValueIdMap := make(map[uint64]map[uint64]interface{}, paramValueListLen)
	for i := range paramValueList {
		paramIdList = append(paramIdList, paramValueList[i].ParamId)

		if _, ok := paramValueIdMap[paramValueList[i].ParamId]; ok {
			paramValueIdMap[paramValueList[i].ParamId][paramValueList[i].ParamValueId] = map[string]interface{}{
				"param_value_id": paramValueList[i].ParamValueId,
				"param_id":       paramValueList[i].ParamId,
				"content":        paramValueList[i].Content,
			}
		} else {
			buf := make(map[uint64]interface{})
			buf[paramValueList[i].ParamValueId] = map[string]interface{}{
				"param_value_id": paramValueList[i].ParamValueId,
				"param_id":       paramValueList[i].ParamId,
				"content":        paramValueList[i].Content,
			}
			paramValueIdMap[paramValueList[i].ParamId] = buf
		}
	}

	paramList := make([]*param.Param, 0, len(paramIdList))
	err = tx.Table(param.GetTableName()).Select([]string{"param_id", "name"}).
		Where("param_id in (?)", paramIdList).Order("sort asc").
		Find(&paramList).Error
	if err != nil {
		return nil, err
	}

	paramDescriptionList := make([]map[string]interface{}, 0, len(paramIdList))
	for i := range paramList {
		if _, ok := paramValueIdMap[paramList[i].ParamId]; !ok {
			continue
		}
		buf := make(map[string]interface{})
		buf["param_id"] = paramList[i].ParamId
		buf["name"] = paramList[i].Name
		buf["children"] = paramValueIdMap[paramList[i].ParamId]
		paramDescriptionList = append(paramDescriptionList, buf)
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
	specValueIdMap := make(map[uint64]map[uint64]interface{}, specValueListLen)
	for i := range specValueList {
		specIdList = append(specIdList, specValueList[i].SpecId)

		if _, ok := specValueIdMap[specValueList[i].SpecId]; ok {
			specValueIdMap[specValueList[i].SpecId][specValueList[i].SpecValueId] = map[string]interface{}{
				"spec_value_id": specValueList[i].SpecValueId,
				"spec_id":       specValueList[i].SpecId,
				"content":       specValueList[i].Content,
			}
		} else {
			buf := make(map[uint64]interface{})
			buf[specValueList[i].SpecValueId] = map[string]interface{}{
				"spec_value_id": specValueList[i].SpecValueId,
				"spec_id":       specValueList[i].SpecId,
				"content":       specValueList[i].Content,
			}
			specValueIdMap[specValueList[i].SpecId] = buf
		}
	}

	specList := make([]*spec.Spec, 0, len(specIdList))
	err = tx.Table(spec.GetTableName()).Select([]string{"spec_id", "name"}).
		Where("spec_id in (?)", specIdList).Order("sort asc").
		Find(&specList).Error
	if err != nil {
		return nil, err
	}

	specDescriptionList := make([]map[string]interface{}, 0, len(specIdList))
	for i := range specList {
		if _, ok := specValueIdMap[specList[i].SpecId]; !ok {
			continue
		}
		buf := make(map[string]interface{})
		buf["spec_id"] = specList[i].SpecId
		buf["name"] = specList[i].Name
		buf["children"] = specValueIdMap[specList[i].SpecId]
		specDescriptionList = append(specDescriptionList, buf)
	}

	specDesByte, err := json.Marshal(specDescriptionList)
	if err != nil {
		return nil, err
	}

	//处理param_description
	paramValueIdList := make([]uint64, 0, 8)
	for k := range req.Param {
		valueId, _ := strconv.ParseUint(req.Param[k].Value, 10, 64)
		paramValueIdList = append(paramValueIdList, valueId)
	}

	paramValueList := make([]*param_value.ParamValue, 0, 8)
	err = tx.Table(param_value.GetTableName()).Select([]string{"param_value_id", "param_id", "content"}).
		Where("param_value_id in (?)", paramValueIdList).Find(&paramValueList).Error
	if err != nil {
		return nil, err
	}

	paramValueListLen := len(paramValueList)
	paramIdList := make([]uint64, 0, 8)
	paramValueIdMap := make(map[uint64]map[uint64]interface{}, paramValueListLen)
	for i := range paramValueList {
		paramIdList = append(paramIdList, paramValueList[i].ParamId)

		if _, ok := paramValueIdMap[paramValueList[i].ParamId]; ok {
			paramValueIdMap[paramValueList[i].ParamId][paramValueList[i].ParamValueId] = map[string]interface{}{
				"param_value_id": paramValueList[i].ParamValueId,
				"param_id":       paramValueList[i].ParamId,
				"content":        paramValueList[i].Content,
			}
		} else {
			buf := make(map[uint64]interface{})
			buf[paramValueList[i].ParamValueId] = map[string]interface{}{
				"param_value_id": paramValueList[i].ParamValueId,
				"param_id":       paramValueList[i].ParamId,
				"content":        paramValueList[i].Content,
			}
			paramValueIdMap[paramValueList[i].ParamId] = buf
		}
	}

	paramList := make([]*param.Param, 0, len(paramIdList))
	err = tx.Table(param.GetTableName()).Select([]string{"param_id", "name"}).
		Where("param_id in (?)", paramIdList).Order("sort asc").
		Find(&paramList).Error
	if err != nil {
		return nil, err
	}

	paramDescriptionList := make([]map[string]interface{}, 0, len(paramIdList))
	for i := range paramList {
		if _, ok := paramValueIdMap[paramList[i].ParamId]; !ok {
			continue
		}
		buf := make(map[string]interface{})
		buf["param_id"] = paramList[i].ParamId
		buf["name"] = paramList[i].Name
		buf["children"] = paramValueIdMap[paramList[i].ParamId]
		paramDescriptionList = append(paramDescriptionList, buf)
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
