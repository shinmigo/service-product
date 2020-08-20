package rpc

import (
	"goshop/service-product/model/spec"
	"goshop/service-product/model/spec_value"
	"goshop/service-product/pkg/db"
	"goshop/service-product/pkg/utils"

	"github.com/shinmigo/pb/basepb"
	"github.com/shinmigo/pb/productpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Spec struct {
}

func NewSpec() *Spec {
	return &Spec{}
}

func (s *Spec) AddSpec(ctx context.Context, req *productpb.Spec) (*basepb.AnyRes, error) {
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

	aul := spec.Spec{
		StoreId:   req.StoreId,
		KindId:    req.KindId,
		Name:      req.Name,
		Sort:      req.Sort,
		CreatedBy: req.AdminId,
		UpdatedBy: req.AdminId,
	}
	if err = tx.Table(spec.GetTableName()).Create(&aul).Error; err != nil {
		return nil, err
	}

	contentLen := len(req.Contents)
	if contentLen > 0 {
		now := utils.JSONTime{}
		now.Time = utils.GetNow()
		specs := make([]*spec_value.SpecValue, 0, contentLen)
		for k := range req.Contents {
			buf := &spec_value.SpecValue{
				SpecId:    aul.SpecId,
				Content:   req.Contents[k],
				CreatedBy: req.AdminId,
				UpdatedBy: req.AdminId,
				CreatedAt: now,
				UpdatedAt: now,
			}
			specs = append(specs, buf)
		}
		if err = spec_value.BatchInsert(tx, specs); err != nil {
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
		Id:    aul.SpecId,
		State: 1,
	}, nil
}

func (s *Spec) EditSpec(ctx context.Context, req *productpb.Spec) (*basepb.AnyRes, error) {
	var err error
	var specInfo *spec.Spec
	if specInfo, err = spec.GetOneBySpecId(req.SpecId, req.StoreId); err != nil {
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

	aul := spec.Spec{
		StoreId:   req.StoreId,
		KindId:    req.KindId,
		Name:      req.Name,
		Sort:      req.Sort,
		UpdatedBy: req.AdminId,
	}

	if err = tx.Table(spec.GetTableName()).Model(&spec.Spec{SpecId: specInfo.SpecId}).Updates(aul).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(spec_value.GetTableName()).Where("spec_id = ?", specInfo.SpecId).Delete(spec_value.SpecValue{}).Error; err != nil {
		return nil, err
	}

	contentLen := len(req.Contents)
	if contentLen > 0 {
		now := utils.JSONTime{}
		now.Time = utils.GetNow()
		specs := make([]*spec_value.SpecValue, 0, contentLen)
		for k := range req.Contents {
			buf := &spec_value.SpecValue{
				SpecId:    specInfo.SpecId,
				Content:   req.Contents[k],
				CreatedBy: specInfo.CreatedBy,
				UpdatedBy: req.AdminId,
				CreatedAt: specInfo.CreatedAt,
				UpdatedAt: now,
			}
			specs = append(specs, buf)
		}
		if err = spec_value.BatchInsert(tx, specs); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.SpecId,
		State: 1,
	}, nil
}

func (s *Spec) DelSpec(ctx context.Context, req *productpb.DelSpecReq) (*basepb.AnyRes, error) {
	var err error
	var specInfo *spec.Spec
	if specInfo, err = spec.GetOneBySpecId(req.SpecId, req.StoreId); err != nil {
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

	if err = tx.Table(spec.GetTableName()).Delete(&spec.Spec{SpecId: specInfo.SpecId}).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(spec_value.GetTableName()).Where("spec_id = ?", specInfo.SpecId).Delete(spec_value.SpecValue{}).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.SpecId,
		State: 1,
	}, nil
}

func (s *Spec) GetSpecList(ctx context.Context, req *productpb.ListSpecReq) (*productpb.ListSpecRes, error) {
	var page uint64 = 1
	if req.Page > 0 {
		page = req.Page
	}

	var pageSize uint64 = 10
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	rows, total, err := spec.GetSpecList(req.Id, req.Name, page, pageSize, req.StoreId)
	if err != nil {
		return nil, err
	}

	rowLen := len(rows)
	specIds := make([]uint64, 0, rowLen)
	for k := range rows {
		specIds = append(specIds, rows[k].SpecId)
	}

	getContents, _ := spec_value.GetContentsBySpecIds(specIds)

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	list := make([]*productpb.SpecDetail, 0, rowLen)
	for k := range rows {
		contents := make([]string, 0, 8)
		if _, ok := getContents[rows[k].SpecId]; ok {
			contents = getContents[rows[k].SpecId]
		}
		list = append(list, &productpb.SpecDetail{
			SpecId:    rows[k].SpecId,
			StoreId:   rows[k].StoreId,
			KindId:    rows[k].KindId,
			Name:      rows[k].Name,
			Sort:      rows[k].Sort,
			CreatedBy: rows[k].CreatedBy,
			UpdatedBy: rows[k].UpdatedBy,
			CreatedAt: rows[k].CreatedAt.Format(utils.TIME_STD_FORMART),
			UpdatedAt: rows[k].UpdatedAt.Format(utils.TIME_STD_FORMART),
			Contents:  contents,
		})
	}

	return &productpb.ListSpecRes{
		Total: total,
		Specs: list,
	}, nil
}
