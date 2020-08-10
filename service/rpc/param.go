package rpc

import (
	"goshop/service-product/model/param"
	"goshop/service-product/model/param_value"
	"goshop/service-product/pkg/db"
	"goshop/service-product/pkg/utils"

	"github.com/shinmigo/pb/basepb"
	"github.com/shinmigo/pb/productpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Param struct {
}

func NewParam() *Param {
	return &Param{}
}

func (p *Param) AddParam(ctx context.Context, req *productpb.Param) (*basepb.AnyRes, error) {
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
		StoreId:   req.StoreId,
		KindId:    req.KindId,
		Name:      req.Name,
		Type:      int32(req.Type),
		Sort:      req.Sort,
		CreatedBy: req.AdminId,
		UpdatedBy: req.AdminId,
	}
	if err = tx.Table(param.GetTableName()).Create(&aul).Error; err != nil {
		return nil, err
	}

	contentLen := len(req.Contents)
	if contentLen > 0 {
		now := utils.JSONTime{}
		now.Time = utils.GetNow()
		params := make([]*param_value.ParamValue, 0, contentLen)
		for k := range req.Contents {
			buf := &param_value.ParamValue{
				ParamId:   aul.ParamId,
				Content:   req.Contents[k],
				CreatedBy: req.AdminId,
				UpdatedBy: req.AdminId,
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

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    aul.ParamId,
		State: 1,
	}, nil
}

func (p *Param) EditParam(ctx context.Context, req *productpb.Param) (*basepb.AnyRes, error) {
	var err error
	var paramInfo *param.Param
	if paramInfo, err = param.GetOneByParamId(req.ParamId); err != nil {
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
		StoreId:   req.StoreId,
		KindId:    req.KindId,
		Name:      req.Name,
		Type:      int32(req.Type),
		Sort:      req.Sort,
		UpdatedBy: req.AdminId,
	}
	if err = tx.Table(param.GetTableName()).Model(&param.Param{ParamId: paramInfo.ParamId}).Updates(aul).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(param_value.GetTableName()).Where("param_id = ?", paramInfo.ParamId).Delete(param_value.ParamValue{}).Error; err != nil {
		return nil, err
	}

	contentLen := len(req.Contents)
	if contentLen > 0 {
		now := utils.JSONTime{}
		now.Time = utils.GetNow()
		params := make([]*param_value.ParamValue, 0, contentLen)
		for k := range req.Contents {
			buf := &param_value.ParamValue{
				ParamId:   paramInfo.ParamId,
				Content:   req.Contents[k],
				CreatedBy: paramInfo.CreatedBy,
				UpdatedBy: req.AdminId,
				CreatedAt: paramInfo.CreatedAt,
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

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.ParamId,
		State: 1,
	}, nil
}

func (p *Param) DelParam(ctx context.Context, req *productpb.DelParamReq) (*basepb.AnyRes, error) {
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

	if err = tx.Table(param.GetTableName()).Delete(param.Param{ParamId: req.ParamId}).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(param_value.GetTableName()).Where("param_id = ?", req.ParamId).Delete(param_value.ParamValue{}).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.ParamId,
		State: 1,
	}, nil
}

func (p *Param) GetParamList(ctx context.Context, req *productpb.ListParamReq) (*productpb.ListParamRes, error) {
	var page uint64 = 1
	if req.Page > 0 {
		page = req.Page
	}

	var pageSize uint64 = 10
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	rows, total, err := param.GetParamList(req.Id, req.Name, page, pageSize)
	if err != nil {
		return nil, err
	}

	rowLen := len(rows)
	paramIds := make([]uint64, 0, rowLen)
	for k := range rows {
		paramIds = append(paramIds, rows[k].ParamId)
	}

	getContents, _ := param_value.GetContentsByParamIds(paramIds)

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	list := make([]*productpb.ParamDetail, 0, rowLen)
	for k := range rows {
		contents := make([]string, 0, 8)
		if _, ok := getContents[rows[k].ParamId]; ok {
			contents = getContents[rows[k].ParamId]
		}
		list = append(list, &productpb.ParamDetail{
			ParamId:   rows[k].ParamId,
			StoreId:   rows[k].StoreId,
			KindId:    rows[k].KindId,
			Name:      rows[k].Name,
			Type:      productpb.ParamType(rows[k].Type),
			Sort:      rows[k].Sort,
			CreatedBy: rows[k].CreatedBy,
			UpdatedBy: rows[k].UpdatedBy,
			CreatedAt: rows[k].CreatedAt.Format(utils.TIME_STD_FORMART),
			UpdatedAt: rows[k].UpdatedAt.Format(utils.TIME_STD_FORMART),
			Contents:  contents,
		})
	}

	return &productpb.ListParamRes{
		Total:  total,
		Params: list,
	}, nil
}
