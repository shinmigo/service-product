package rpc

import (
	"goshop/service-product/model/kind"
	"goshop/service-product/model/param"
	"goshop/service-product/model/spec"
	"goshop/service-product/pkg/db"
	"goshop/service-product/pkg/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/shinmigo/pb/basepb"

	"github.com/shinmigo/pb/productpb"
	"golang.org/x/net/context"
)

type Kind struct {
}

func NewKind() *Kind {
	return &Kind{}
}

func (k *Kind) AddKind(ctx context.Context, req *productpb.Kind) (*basepb.AnyRes, error) {
	aul := kind.Kind{
		StoreId:   req.StoreId,
		Name:      req.Name,
		CreatedBy: req.AdminId,
		UpdatedBy: req.AdminId,
	}

	if err := db.Conn.Table(kind.GetTableName()).Create(&aul).Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    aul.KindId,
		State: 1,
	}, nil
}

func (k *Kind) EditKind(ctx context.Context, req *productpb.Kind) (*basepb.AnyRes, error) {
	if _, err := kind.GetOneByKindId(req.KindId); err != nil {
		return nil, err
	}

	aul := kind.Kind{
		StoreId:   req.StoreId,
		Name:      req.Name,
		UpdatedBy: req.AdminId,
	}

	if err := db.Conn.Table(kind.GetTableName()).Model(&kind.Kind{KindId: req.KindId}).Updates(aul).Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.KindId,
		State: 1,
	}, nil
}

func (k *Kind) DelKind(ctx context.Context, req *productpb.DelKindReq) (*basepb.AnyRes, error) {
	var err error
	if _, err = kind.GetOneByKindId(req.KindId); err != nil {
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

	if err = tx.Table(param.GetTableName()).Where("kind_id = ?", req.KindId).Update("kind_id", 0).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(spec.GetTableName()).Where("kind_id = ?", req.KindId).Update("kind_id", 0).Error; err != nil {
		return nil, err
	}

	if err := tx.Table(kind.GetTableName()).Delete(&kind.Kind{KindId: req.KindId}).Error; err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.KindId,
		State: 1,
	}, nil
}

func (k *Kind) GetKindList(ctx context.Context, req *productpb.ListKindReq) (*productpb.ListKindRes, error) {
	var page uint64 = 1
	if req.Page > 0 {
		page = req.Page
	}

	var pageSize uint64 = 10
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	rows, total, err := kind.GetKindList(req.Id, req.Name, page, pageSize)
	if err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	list := make([]*productpb.KindDetail, 0, len(rows))
	for k := range rows {
		list = append(list, &productpb.KindDetail{
			KindId:    rows[k].KindId,
			StoreId:   rows[k].StoreId,
			Name:      rows[k].Name,
			ParamQty:  rows[k].ParamQty,
			SpecQty:   rows[k].SpecQty,
			CreatedBy: rows[k].CreatedBy,
			UpdatedBy: rows[k].UpdatedBy,
			CreatedAt: rows[k].CreatedAt.Format(utils.TIME_STD_FORMART),
			UpdatedAt: rows[k].UpdatedAt.Format(utils.TIME_STD_FORMART),
		})
	}
	return &productpb.ListKindRes{
		Total: total,
		Kinds: list,
	}, nil
}

func (k *Kind) BindParam(ctx context.Context, req *productpb.BindParamReq) (*basepb.AnyRes, error) {
	var err error
	if _, err = kind.GetOneByKindId(req.KindId); err != nil {
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

	if err = tx.Table(param.GetTableName()).Where("kind_id = ?", req.KindId).Update("kind_id", 0).Error; err != nil {
		return nil, err
	}

	var total uint64
	if len(req.ParamIds) > 0 {
		if err = tx.Table(param.GetTableName()).Where("param_id in (?) and kind_id = 0", req.ParamIds).Update("kind_id", req.KindId).Error; err != nil {
			return nil, err
		}

		if err = tx.Table(param.GetTableName()).Where("kind_id = ?", req.KindId).Count(&total).Error; err != nil {
			return nil, err
		}
	}

	if err = tx.Table(kind.GetTableName()).Where("kind_id = ?", req.KindId).Update("param_qty", total).Error; err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.KindId,
		State: 1,
	}, nil
}

func (k *Kind) BindSpec(ctx context.Context, req *productpb.BindSpecReq) (*basepb.AnyRes, error) {
	var err error
	if _, err = kind.GetOneByKindId(req.KindId); err != nil {
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

	if err = tx.Table(spec.GetTableName()).Where("kind_id = ?", req.KindId).Update("kind_id", 0).Error; err != nil {
		return nil, err
	}

	var total uint64
	if len(req.SpecIds) > 0 {
		if err = tx.Table(spec.GetTableName()).Where("spec_id in (?) and kind_id = 0", req.SpecIds).Update("kind_id", req.KindId).Error; err != nil {
			return nil, err
		}

		if err = tx.Table(spec.GetTableName()).Where("kind_id = ?", req.KindId).Count(&total).Error; err != nil {
			return nil, err
		}

	}
	if err = tx.Table(kind.GetTableName()).Where("kind_id = ?", req.KindId).Update("spec_qty", total).Error; err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.KindId,
		State: 1,
	}, nil
}
