package rpc

import (
	"goshop/service-product/model/spec"
	"goshop/service-product/pkg/db"
	"strings"
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
	tx := db.Conn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, err
	}

	aul := spec.Spec{
		StoreId:   req.Spec.StoreId,
		TypeId:    req.Spec.TypeId,
		Name:      req.Spec.Name,
		Sort:      req.Spec.Sort,
		CreatedBy: req.Spec.AdminId,
		UpdatedBy: req.Spec.AdminId,
	}
	if err := tx.Table(spec.GetTableName()).Create(&aul).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(req.Spec.Contents) > 0 {
		now := time.Now()
		sqlStr := "INSERT INTO spec_value (spec_id, content, created_by, updated_by, created_at, updated_at) VALUES "
		vals := []interface{}{}
		const rowSQL = "(?, ?, ?, ?, ?, ?)"
		var inserts []string
		for k := range req.Spec.Contents {
			inserts = append(inserts, rowSQL)
			vals = append(vals, aul.SpecId, req.Spec.Contents[k], req.Spec.AdminId, req.Spec.AdminId, now, now)
		}
		sqlStr = sqlStr + strings.Join(inserts, ",")
		if err := tx.Exec(sqlStr, vals...).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &productpb.AddSpecRes{
		SpecId: aul.SpecId,
	}, nil
}

func (s *Spec) EditSpec(ctx context.Context, req *productpb.EditSpecReq) (*productpb.EditSpecRes, error) {
	specInfo, err := spec.GetOneBySpecId(req.Spec.SpecId)
	if err != nil {
		return nil, err
	}

	tx := db.Conn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, err
	}

	aul := spec.Spec{
		StoreId:   req.Spec.StoreId,
		TypeId:    req.Spec.TypeId,
		Name:      req.Spec.Name,
		Sort:      req.Spec.Sort,
		UpdatedBy: req.Spec.AdminId,
	}
	if err := tx.Table(spec.GetTableName()).Where("spec_id = ?", specInfo.SpecId).Updates(&aul).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Table(spec_value.GetTableName()).Where("spec_id = ?", specInfo.SpecId).Delete(spec_value.SpecValue{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(req.Spec.Contents) > 0 {
		now := time.Now()
		sqlStr := "INSERT INTO spec_value (spec_id, content, created_by, updated_by, created_at, updated_at) VALUES "
		vals := []interface{}{}
		const rowSQL = "(?, ?, ?, ?, ?, ?)"
		var inserts []string
		for k := range req.Spec.Contents {
			inserts = append(inserts, rowSQL)
			vals = append(vals, specInfo.SpecId, req.Spec.Contents[k], specInfo.CreatedBy, req.Spec.AdminId, specInfo.CreatedAt, now)
		}
		sqlStr = sqlStr + strings.Join(inserts, ",")
		if err := tx.Exec(sqlStr, vals...).Error; err != nil {
			tx.Rollback()
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
	specInfo, err := spec.GetOneBySpecId(req.SpecId)
	if err != nil {
		return nil, err
	}
	tx := db.Conn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, err
	}

	if err := db.Conn.Table(spec.GetTableName()).Where("spec_id = ?", specInfo.SpecId).Delete(spec.Spec{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Table(spec_value.GetTableName()).Where("spec_id = ?", specInfo.SpecId).Delete(spec_value.SpecValue{}).Error; err != nil {
		tx.Rollback()
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

	var contents []string
	getContents, err := spec_value.GetContentsBySpecIds([]uint64{row.SpecId})
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
	list := []*productpb.SpecInfo{}

	rows, err := spec.GetSpecs(1, 10)
	if err != nil {
		return nil, err
	}

	var specIds []uint64
	for k := range rows {
		specIds = append(specIds, rows[k].SpecId)
	}

	getContents, err := spec_value.GetContentsBySpecIds(specIds)

	for k := range rows {
		var contents []string
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
		Spec: list,
	}, nil
}
