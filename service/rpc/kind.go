package rpc

import (
	"goshop/service-product/model/kind"
	"goshop/service-product/model/param"
	"goshop/service-product/model/spec"
	"goshop/service-product/pkg/db"

	"github.com/shinmigo/pb/productpb"
	"golang.org/x/net/context"
)

type Kind struct {
}

func NewKind() *Kind {
	return &Kind{}
}

func (k *Kind) AddKind(ctx context.Context, req *productpb.AddKindReq) (*productpb.AddKindRes, error) {
	aul := kind.Kind{
		StoreId:   req.Kind.StoreId,
		Name:      req.Kind.Name,
		CreatedBy: req.Kind.AdminId,
		UpdatedBy: req.Kind.AdminId,
	}

	if err := db.Conn.Table(kind.GetTableName()).Create(&aul).Error; err != nil {
		return nil, err
	}

	return &productpb.AddKindRes{
		KindId: aul.KindId,
	}, nil
}

func (k *Kind) EditKind(ctx context.Context, req *productpb.EditKindReq) (*productpb.EditKindRes, error) {
	if _, err := kind.GetOneByKindId(req.Kind.KindId); err != nil {
		return nil, err
	}

	aul := kind.Kind{
		StoreId:   req.Kind.StoreId,
		Name:      req.Kind.Name,
		UpdatedBy: req.Kind.AdminId,
	}

	if err := db.Conn.Table(kind.GetTableName()).Model(&kind.Kind{KindId: req.Kind.KindId}).Updates(aul).Error; err != nil {
		return nil, err
	}

	return &productpb.EditKindRes{
		Updated: 1,
	}, nil
}

func (k *Kind) DelKind(ctx context.Context, req *productpb.DelKindReq) (*productpb.DelKindRes, error) {
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

	return &productpb.DelKindRes{
		Deleted: 1,
	}, nil
}

func (k *Kind) ReadKind(ctx context.Context, req *productpb.ReadKindReq) (*productpb.ReadKindRes, error) {
	row, err := kind.GetOneByKindId(req.KindId)
	if err != nil {
		return nil, err
	}

	return &productpb.ReadKindRes{
		Kind: &productpb.KindInfo{
			KindId:   row.KindId,
			Name:     row.Name,
			ParamQty: row.ParamQty,
			SpecQty:  row.SpecQty,
		},
	}, nil
}

func (k *Kind) ReadKinds(ctx context.Context, req *productpb.ReadKindsReq) (*productpb.ReadKindsRes, error) {
	var page uint64 = 1
	if req.Page > 0 {
		page = req.Page
	}

	var pageSize uint64 = 10
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	rows, err := kind.GetKinds(page, pageSize)
	if err != nil {
		return nil, err
	}
	rowLen := len(rows)
	list := make([]*productpb.KindInfo, 0, rowLen)
	for k := range rows {
		list = append(list, &productpb.KindInfo{
			KindId:   rows[k].KindId,
			Name:     rows[k].Name,
			ParamQty: rows[k].ParamQty,
			SpecQty:  rows[k].SpecQty,
		})
	}
	return &productpb.ReadKindsRes{
		Kinds: list,
	}, nil
}

func (k *Kind) BindParam(ctx context.Context, req *productpb.BindParamReq) (*productpb.BindParamRes, error) {
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

	return &productpb.BindParamRes{
		Updated: 1,
	}, nil
}

func (k *Kind) BindSpec(ctx context.Context, req *productpb.BindSpecReq) (*productpb.BindSpecRes, error) {
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

	return &productpb.BindSpecRes{
		Updated: 1,
	}, nil
}
