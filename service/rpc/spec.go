package rpc

import (
	"goshop/service-product/model/spec"
	"goshop/service-product/pkg/db"
	"time"

	"goshop/service-product/model/spec_value"

	"github.com/shinmigo/pb/productpb"
	"golang.org/x/net/context"
)

type Spec struct {
}

func NewSpec() *Spec {
	return &Spec{}
}

func (s *Spec) AddSpec(ctx context.Context, req *productpb.AddSpecReq) (*productpb.AddSpecRes, error) {
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
		StoreId:   req.Spec.StoreId,
		KindId:    req.Spec.KindId,
		Name:      req.Spec.Name,
		Sort:      req.Spec.Sort,
		CreatedBy: req.Spec.AdminId,
		UpdatedBy: req.Spec.AdminId,
	}
	if err = tx.Table(spec.GetTableName()).Create(&aul).Error; err != nil {
		return nil, err
	}

	contentLen := len(req.Spec.Contents)
	if contentLen > 0 {
		now := time.Now()
		specs := make([]*spec_value.SpecValue, 0, contentLen)
		for k := range req.Spec.Contents {
			buf := &spec_value.SpecValue{
				SpecId:    aul.SpecId,
				Content:   req.Spec.Contents[k],
				CreatedBy: req.Spec.AdminId,
				UpdatedBy: req.Spec.AdminId,
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

	return &productpb.AddSpecRes{
		SpecId: aul.SpecId,
	}, nil
}

func (s *Spec) EditSpec(ctx context.Context, req *productpb.EditSpecReq) (*productpb.EditSpecRes, error) {
	var err error
	var specInfo *spec.SpecInfo
	specInfo, err = spec.GetOneBySpecId(req.Spec.SpecId)
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

	aul := spec.Spec{
		StoreId:   req.Spec.StoreId,
		KindId:    req.Spec.KindId,
		Name:      req.Spec.Name,
		Sort:      req.Spec.Sort,
		UpdatedBy: req.Spec.AdminId,
	}
	if err = tx.Table(spec.GetTableName()).Where("spec_id = ?", specInfo.SpecId).Updates(&aul).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(spec_value.GetTableName()).Where("spec_id = ?", specInfo.SpecId).Delete(spec_value.SpecValue{}).Error; err != nil {
		return nil, err
	}

	contentLen := len(req.Spec.Contents)
	if contentLen > 0 {
		now := time.Now()
		specs := make([]*spec_value.SpecValue, 0, contentLen)
		for k := range req.Spec.Contents {
			buf := &spec_value.SpecValue{
				SpecId:    specInfo.SpecId,
				Content:   req.Spec.Contents[k],
				CreatedBy: req.Spec.AdminId,
				UpdatedBy: req.Spec.AdminId,
				CreatedAt: now,
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

	return &productpb.EditSpecRes{
		Updated: 1,
	}, nil
}

func (s *Spec) DelSpec(ctx context.Context, req *productpb.DelSpecReq) (*productpb.DelSpecRes, error) {
	var err error
	var specInfo *spec.SpecInfo

	specInfo, err = spec.GetOneBySpecId(req.SpecId)
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

	if err = tx.Table(spec.GetTableName()).Where("spec_id = ?", specInfo.SpecId).Delete(spec.Spec{}).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(spec_value.GetTableName()).Where("spec_id = ?", specInfo.SpecId).Delete(spec_value.SpecValue{}).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &productpb.DelSpecRes{
		Deleted: 1,
	}, nil
}

func (s *Spec) ReadSpec(ctx context.Context, req *productpb.ReadSpecReq) (*productpb.ReadSpecRes, error) {
	row, err := spec.GetOneBySpecId(req.SpecId)
	if err != nil {
		return nil, err
	}

	getContents, err := spec_value.GetContentsBySpecIds([]uint64{row.SpecId})

	contents := make([]string, len(getContents))
	if _, ok := getContents[row.SpecId]; ok {
		contents = getContents[row.SpecId]
	}

	return &productpb.ReadSpecRes{
		Spec: &productpb.SpecInfo{
			SpecId:   row.SpecId,
			Name:     row.Name,
			Sort:     row.Sort,
			Contents: contents,
		},
	}, nil
}

func (s *Spec) ReadSpecs(ctx context.Context, req *productpb.ReadSpecsReq) (*productpb.ReadSpecsRes, error) {
	var page uint64 = 1
	if req.Page > 0 {
		page = req.Page
	}

	var pageSize uint64 = 10
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	rows, err := spec.GetSpecs(page, pageSize)
	if err != nil {
		return nil, err
	}

	rowLen := len(rows)
	specIds := make([]uint64, 0, rowLen)
	for k := range rows {
		specIds = append(specIds, rows[k].SpecId)
	}

	getContents, err := spec_value.GetContentsBySpecIds(specIds)

	list := make([]*productpb.SpecInfo, 0, rowLen)
	for k := range rows {
		contents := make([]string, 0, 8)
		if _, ok := getContents[rows[k].SpecId]; ok {
			contents = getContents[rows[k].SpecId]
		}
		list = append(list, &productpb.SpecInfo{
			SpecId:   rows[k].SpecId,
			Name:     rows[k].Name,
			Sort:     rows[k].Sort,
			Contents: contents,
		})
	}

	return &productpb.ReadSpecsRes{
		Specs: list,
	}, nil
}
