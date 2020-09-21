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

func (p *Param) EditParam(ctx context.Context, req *productpb.EditParamReq) (*basepb.AnyRes, error) {
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

	contentLen := len(req.Contents)
	newIds := make([]uint64, 0, contentLen)
	if contentLen > 0 {
		for k := range req.Contents {
			if req.Contents[k].ParamValueId > 0 {
				// 更新
				if err = tx.Table(param_value.GetTableName()).
					Where("param_value_id = ? and param_id = ?", req.Contents[k].ParamValueId, req.ParamId).
					Update("content", req.Contents[k].Content).Error; err != nil {
					return nil, err
				}
				newIds = append(newIds, req.Contents[k].ParamValueId)
			} else {
				// 新增
				aul := param_value.ParamValue{
					ParamId:   req.ParamId,
					Content:   req.Contents[k].Content,
					CreatedBy: req.AdminId,
					UpdatedBy: req.AdminId,
				}
				if err = tx.Table(param_value.GetTableName()).Create(&aul).Error; err != nil {
					return nil, err
				}
				newIds = append(newIds, aul.ParamValueId)
			}
		}
	} else {
		if err = tx.Table(param_value.GetTableName()).Where("param_id = ?", paramInfo.ParamId).Delete(param_value.ParamValue{}).Error; err != nil {
			return nil, err
		}
	}

	deleteParamValueIds := make([]uint64, 0, 32)
	for k := range paramInfo.Contents {
		if utils.InArrayForUint64(paramInfo.Contents[k].ParamValueId, newIds) == false {
			deleteParamValueIds = append(deleteParamValueIds, paramInfo.Contents[k].ParamValueId)
		}
	}

	if len(deleteParamValueIds) > 0 {
		if err = tx.Table(param_value.GetTableName()).Where("param_value_id in (?) and param_id = ?", deleteParamValueIds, paramInfo.ParamId).Delete(param_value.ParamValue{}).Error; err != nil {
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

	if err = tx.Table(param.GetTableName()).Where("param_id in (?)", req.ParamId).Delete(param.Param{}).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(param_value.GetTableName()).Where("param_id in (?)", req.ParamId).Delete(param_value.ParamValue{}).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.ParamId[0],
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

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	list := make([]*productpb.ParamDetail, 0, len(rows))
	for k := range rows {
		paramValueList := make([]*productpb.ParamValue, 0, 8)
		if len(rows[k].Contents) > 0 {
			for i := range rows[k].Contents {
				buf := &productpb.ParamValue{
					ParamValueId: rows[k].Contents[i].ParamValueId,
					Content:      rows[k].Contents[i].Content,
				}
				paramValueList = append(paramValueList, buf)
			}
		}

		buf1, _ := jsonLib.Marshal(rows[k])
		buf2 := &productpb.ParamDetail{}
		_ = jsonLib.Unmarshal(buf1, buf2)
		buf2.Contents = paramValueList
		list = append(list, buf2)
	}

	return &productpb.ListParamRes{
		Total:  total,
		Params: list,
	}, nil
}

func (p *Param) GetBindParamAll(ctx context.Context, req *productpb.BindParamAllReq) (*productpb.BindParamAllRes, error) {
	rows := make([]struct {
		ParamId uint64
		Name    string
	}, 0, 32)

	query := db.Conn.Table(param.GetTableName()).Select("param_id, name").Where("kind_id = 0")

	if len(req.Name) > 0 {
		query = query.Where("name like ?", req.Name+"%")
	}

	query.Scan(&rows)

	rowLen := len(rows)
	if rowLen == 0 {
		return nil, nil
	}

	list := make([]*productpb.BindParamAll, 0, rowLen)
	for k := range rows {
		list = append(list, &productpb.BindParamAll{
			ParamId: rows[k].ParamId,
			Name:    rows[k].Name,
		})
	}

	return &productpb.BindParamAllRes{
		Data: list,
	}, nil
}
