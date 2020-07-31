package rpc

import (
	"strings"
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

var paramType = map[string]int{
	"Text":     param.ParamTypeText,
	"Radio":    param.ParamTypeRadio,
	"Checkbox": param.ParamTypeCheckbox,
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
		}

		if err != nil {
			tx.Rollback()
		}
	}()

	aul := param.Param{
		StoreId:   req.Param.StoreId,
		KindId:    req.Param.KindId,
		Name:      req.Param.Name,
		Type:      paramType[req.Param.Type.String()],
		Sort:      req.Param.Sort,
		CreatedBy: req.Param.AdminId,
		UpdatedBy: req.Param.AdminId,
	}
	if err = tx.Table(param.GetTableName()).Create(&aul).Error; err != nil {
		return nil, err
	}

	if len(req.Param.Contents) > 0 {
		now := time.Now()
		sqlStr := "INSERT INTO param_value (param_id, content, created_by, updated_by, created_at, updated_at) VALUES "
		vals := []interface{}{}
		rowSQL := "(?, ?, ?, ?, ?, ?)"
		var inserts []string
		for k := range req.Param.Contents {
			inserts = append(inserts, rowSQL)
			vals = append(vals, aul.ParamId, req.Param.Contents[k], req.Param.AdminId, req.Param.AdminId, now, now)
		}
		sqlStr = sqlStr + strings.Join(inserts, ",")
		if err = tx.Exec(sqlStr, vals...).Error; err != nil {
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
		}

		if err != nil {
			tx.Rollback()
		}
	}()

	aul := param.Param{
		StoreId:   req.Param.StoreId,
		KindId:    req.Param.KindId,
		Name:      req.Param.Name,
		Type:      paramType[req.Param.Type.String()],
		Sort:      req.Param.Sort,
		UpdatedBy: req.Param.AdminId,
	}
	if err = tx.Table(param.GetTableName()).Where("param_id = ?", paramInfo.ParamId).Updates(&aul).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(param_value.GetTableName()).Where("param_id = ?", paramInfo.ParamId).Delete(param_value.ParamValue{}).Error; err != nil {
		return nil, err
	}

	if len(req.Param.Contents) > 0 {
		now := time.Now()
		sqlStr := "INSERT INTO param_value (param_id, content, created_by, updated_by, created_at, updated_at) VALUES "
		vals := []interface{}{}
		rowSQL := "(?, ?, ?, ?, ?, ?)"
		var inserts []string
		for k := range req.Param.Contents {
			inserts = append(inserts, rowSQL)
			vals = append(vals, paramInfo.ParamId, req.Param.Contents[k], paramInfo.CreatedBy, req.Param.AdminId, paramInfo.CreatedAt, now)
		}
		sqlStr = sqlStr + strings.Join(inserts, ",")
		if err = tx.Exec(sqlStr, vals...).Error; err != nil {
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
	var paramInfo *param.ParamInfo

	paramInfo, err = param.GetOneByParamId(req.ParamId)
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
		}

		if err != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Table(param.GetTableName()).Where("param_id = ?", paramInfo.ParamId).Delete(param.Param{}).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(param_value.GetTableName()).Where("param_id = ?", paramInfo.ParamId).Delete(param_value.ParamValue{}).Error; err != nil {
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

	var contents []string
	getContents, err := param_value.GetContentsByParamIds([]uint64{row.ParamId})
	if _, ok := getContents[row.ParamId]; ok {
		contents = getContents[row.ParamId]
	}

	return &productpb.ReadParamRes{
		Param: &productpb.ParamInfo{
			ParamId:  row.ParamId,
			Name:     row.Name,
			Type:     productpb.ParamType(row.Type - 1),
			Sort:     row.Sort,
			Contents: contents,
		},
	}, nil
}

func (p *Param) ReadParams(ctx context.Context, req *productpb.ReadParamsReq) (*productpb.ReadParamsRes, error) {
	list := []*productpb.ParamInfo{}

	rows, err := param.GetParams(1, 10)
	if err != nil {
		return nil, err
	}

	var paramIds []uint64
	for k := range rows {
		paramIds = append(paramIds, rows[k].ParamId)
	}

	getContents, err := param_value.GetContentsByParamIds(paramIds)

	for k := range rows {
		var contents []string
		if _, ok := getContents[rows[k].ParamId]; ok {
			contents = getContents[rows[k].ParamId]
		}
		list = append(list, &productpb.ParamInfo{
			ParamId:  rows[k].ParamId,
			Name:     rows[k].Name,
			Type:     productpb.ParamType(rows[k].Type - 1),
			Sort:     rows[k].Sort,
			Contents: contents,
		})
	}

	return &productpb.ReadParamsRes{
		Params: list,
	}, nil
}
