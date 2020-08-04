package rpc

import (
	"time"

	"github.com/shinmigo/pb/productpb"

	"goshop/service-product/model/param"
	"goshop/service-product/pkg/db"

	"golang.org/x/net/context"

	"goshop/service-product/model/param_value"
)

type Param struct {
}

func NewParam() *Param {
	return &Param{}
}

func (p *Param) AddParam(ctx context.Context, req *productpb.AddParamReq) (*productpb.AddParamRes, error) {
	var err error
	tx := db.Conn.Begin()
	if err = tx.Error; err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}

		if err != nil {
			tx.Rollback()
		}
	}()

	aul := param.Param{
		StoreId:   req.Param.StoreId,
		KindId:    req.Param.KindId,
		Name:      req.Param.Name,
		Type:      int32(req.Param.Type),
		Sort:      req.Param.Sort,
		CreatedBy: req.Param.AdminId,
		UpdatedBy: req.Param.AdminId,
	}
	if err = tx.Table(param.GetTableName()).Create(&aul).Error; err != nil {
		return nil, err
	}

	contentLen := len(req.Param.Contents)
	if contentLen > 0 {
		now := time.Now()
		params := make([]*param_value.ParamValue, 0, contentLen)
		for k := range req.Param.Contents {
			buf := &param_value.ParamValue{
				ParamId:   aul.ParamId,
				Content:   req.Param.Contents[k],
				CreatedBy: req.Param.AdminId,
				UpdatedBy: req.Param.AdminId,
				CreatedAt: now,
				UpdatedAt: now,
			}
			params = append(params, buf)
		}
		if err = param_value.BatchInsert(tx, params); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	return &productpb.AddParamRes{
		ParamId: aul.ParamId,
	}, nil
}

func (p *Param) EditParam(ctx context.Context, req *productpb.EditParamReq) (*productpb.EditParamRes, error) {
	var err error
	var paramInfo *param.ParamInfo

	paramInfo, err = param.GetOneByParamId(req.Param.ParamId)
	if err != nil {
		return nil, err
	}

	tx := db.Conn.Begin()
	if err = tx.Error; err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}

		if err != nil {
			tx.Rollback()
		}
	}()

	aul := param.Param{
		StoreId:   req.Param.StoreId,
		KindId:    req.Param.KindId,
		Name:      req.Param.Name,
		Type:      int32(req.Param.Type),
		Sort:      req.Param.Sort,
		UpdatedBy: req.Param.AdminId,
	}
	if err = tx.Table(param.GetTableName()).Where("param_id = ?", paramInfo.ParamId).Updates(&aul).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(param_value.GetTableName()).Where("param_id = ?", paramInfo.ParamId).Delete(param_value.ParamValue{}).Error; err != nil {
		return nil, err
	}

	contentLen := len(req.Param.Contents)
	if contentLen > 0 {
		now := time.Now()
		params := make([]*param_value.ParamValue, 0, contentLen)
		for k := range req.Param.Contents {
			buf := &param_value.ParamValue{
				ParamId:   paramInfo.ParamId,
				Content:   req.Param.Contents[k],
				CreatedBy: req.Param.AdminId,
				UpdatedBy: req.Param.AdminId,
				CreatedAt: now,
				UpdatedAt: now,
			}
			params = append(params, buf)
		}
		if err = param_value.BatchInsert(tx, params); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	return &productpb.EditParamRes{
		Updated: 1,
	}, nil
}

func (p *Param) DelParam(ctx context.Context, req *productpb.DelParamReq) (*productpb.DelParamRes, error) {
	var err error
	if _, err = param.GetOneByParamId(req.ParamId); err != nil {
		return nil, err
	}
	tx := db.Conn.Begin()
	if err = tx.Error; err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}

		if err != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Table(param.GetTableName()).Where("param_id = ?", req.ParamId).Delete(param.Param{}).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(param_value.GetTableName()).Where("param_id = ?", req.ParamId).Delete(param_value.ParamValue{}).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &productpb.DelParamRes{
		Deleted: 1,
	}, nil
}

func (p *Param) ReadParam(ctx context.Context, req *productpb.ReadParamReq) (*productpb.ReadParamRes, error) {
	row, err := param.GetOneByParamId(req.ParamId)
	if err != nil {
		return nil, err
	}

	getContents, err := param_value.GetContentsByParamIds([]uint64{row.ParamId})
	contents := make([]string, 0, len(getContents))
	if _, ok := getContents[row.ParamId]; ok {
		contents = getContents[row.ParamId]
	}

	return &productpb.ReadParamRes{
		Param: &productpb.ParamInfo{
			ParamId:  row.ParamId,
			Name:     row.Name,
			Type:     productpb.ParamType(row.Type),
			Sort:     row.Sort,
			Contents: contents,
		},
	}, nil
}

func (p *Param) ReadParams(ctx context.Context, req *productpb.ReadParamsReq) (*productpb.ReadParamsRes, error) {
	var page uint64 = 1
	if req.Page > 0 {
		page = req.Page
	}

	var pageSize uint64 = 10
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	rows, err := param.GetParams(page, pageSize)
	if err != nil {
		return nil, err
	}

	rowLen := len(rows)
	paramIds := make([]uint64, 0, rowLen)
	for k := range rows {
		paramIds = append(paramIds, rows[k].ParamId)
	}

	getContents, _ := param_value.GetContentsByParamIds(paramIds)

	list := make([]*productpb.ParamInfo, 0, rowLen)
	for k := range rows {
		contents := make([]string, 0, 8)
		if _, ok := getContents[rows[k].ParamId]; ok {
			contents = getContents[rows[k].ParamId]
		}
		list = append(list, &productpb.ParamInfo{
			ParamId:  rows[k].ParamId,
			Name:     rows[k].Name,
			Type:     productpb.ParamType(rows[k].Type),
			Sort:     rows[k].Sort,
			Contents: contents,
		})
	}

	return &productpb.ReadParamsRes{
		Params: list,
	}, nil
}
