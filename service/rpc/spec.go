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

func (s *Spec) EditSpec(ctx context.Context, req *productpb.EditSpecReq) (*basepb.AnyRes, error) {
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

	contentLen := len(req.Contents)
	if contentLen > 0 {
		for k := range req.Contents {
			if req.Contents[k].SpecValueId > 0 {
				// 更新
				if err = tx.Table(spec_value.GetTableName()).
					Where("spec_value_id = ? and spec_id = ?", req.Contents[k].SpecValueId, req.SpecId).
					Update("content", req.Contents[k].Content).Error; err != nil {
					return nil, err
				}
			} else {
				// 新增
				aul := spec_value.SpecValue{
					SpecId:    req.SpecId,
					Content:   req.Contents[k].Content,
					CreatedBy: req.AdminId,
					UpdatedBy: req.AdminId,
				}
				if err = tx.Table(spec_value.GetTableName()).Create(&aul).Error; err != nil {
					return nil, err
				}
			}
		}
	} else {
		// 内容长度等于0 全删了
		if err = tx.Table(spec_value.GetTableName()).Where("spec_id = ?", specInfo.SpecId).Delete(spec_value.SpecValue{}).Error; err != nil {
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

	if err = tx.Table(spec.GetTableName()).Where("spec_id in (?) AND store_id = ?", req.SpecId, req.StoreId).Delete(&spec.Spec{}).Error; err != nil {
		return nil, err
	}

	if err = tx.Table(spec_value.GetTableName()).Where("spec_id in (?)", req.SpecId).Delete(spec_value.SpecValue{}).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.SpecId[0],
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

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	list := make([]*productpb.SpecDetail, 0, len(rows))
	for k := range rows {
		specValueList := make([]*productpb.SpecValue, 0, 8)
		if len(rows[k].Contents) > 0 {
			for i := range rows[k].Contents {
				buf := &productpb.SpecValue{
					SpecValueId: rows[k].Contents[i].SpecValueId,
					Content:     rows[k].Contents[i].Content,
				}
				specValueList = append(specValueList, buf)
			}
		}

		buf1, _ := jsonLib.Marshal(rows[k])
		buf2 := &productpb.SpecDetail{}
		_ = jsonLib.Unmarshal(buf1, buf2)
		buf2.Contents = specValueList
		list = append(list, buf2)
	}

	return &productpb.ListSpecRes{
		Total: total,
		Specs: list,
	}, nil
}

func (p *Spec) GetBindSpecAll(ctx context.Context, req *productpb.BindSpecAllReq) (*productpb.BindSpecAllRes, error) {
	rows := make([]struct {
		SpecId uint64
		Name   string
	}, 0, 32)

	query := db.Conn.Table(spec.GetTableName()).Select("spec_id, name").Where("kind_id = 0")

	if len(req.Name) > 0 {
		query = query.Where("name like ?", req.Name+"%")
	}

	query.Scan(&rows)

	rowLen := len(rows)
	if rowLen == 0 {
		return nil, nil
	}

	list := make([]*productpb.BindSpecAll, 0, rowLen)
	for k := range rows {
		list = append(list, &productpb.BindSpecAll{
			SpecId: rows[k].SpecId,
			Name:   rows[k].Name,
		})
	}

	return &productpb.BindSpecAllRes{
		Data: list,
	}, nil
}
