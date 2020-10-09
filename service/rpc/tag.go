package rpc

import (
	"goshop/service-product/pkg/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/shinmigo/pb/basepb"

	"github.com/shinmigo/pb/productpb"

	"goshop/service-product/model/tag"
	"goshop/service-product/pkg/db"

	"golang.org/x/net/context"
)

type Tag struct {
}

func NewTag() *Tag {
	return &Tag{}
}

func (t *Tag) AddTag(ctx context.Context, req *productpb.Tag) (*basepb.AnyRes, error) {
	aul := tag.Tag{
		StoreId:   req.StoreId,
		Name:      req.Name,
		Display:   int32(req.Display),
		Sort:      req.Sort,
		CreatedBy: req.AdminId,
		UpdatedBy: req.AdminId,
	}

	if err := db.Conn.Table(tag.GetTableName()).Create(&aul).Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    aul.TagId,
		State: 1,
	}, nil
}

func (t *Tag) EditTag(ctx context.Context, req *productpb.Tag) (*basepb.AnyRes, error) {
	if _, err := tag.GetOneByTagId(req.TagId); err != nil {
		return nil, err
	}

	aul := tag.Tag{
		StoreId:   req.StoreId,
		Name:      req.Name,
		Display:   int32(req.Display),
		Sort:      req.Sort,
		UpdatedBy: req.AdminId,
	}

	if err := db.Conn.Table(tag.GetTableName()).Model(&tag.Tag{TagId: req.TagId}).Updates(aul).Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.TagId,
		State: 1,
	}, nil
}

func (t *Tag) DelTag(ctx context.Context, req *productpb.DelTagReq) (*basepb.AnyRes, error) {
	if err := db.Conn.Table(tag.GetTableName()).Where("tag_id in (?)", req.TagId).Delete(&tag.Tag{}).Error; err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	return &basepb.AnyRes{
		Id:    req.TagId[0],
		State: 1,
	}, nil
}

func (t *Tag) GetTagList(ctx context.Context, req *productpb.ListTagReq) (*productpb.ListTagRes, error) {
	rows, total, err := tag.GetTagList(req)
	if err != nil {
		return nil, err
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "The client canceled the request")
	}

	list := make([]*productpb.TagDetail, 0, len(rows))
	for k := range rows {
		list = append(list, &productpb.TagDetail{
			TagId:     rows[k].TagId,
			StoreId:   rows[k].StoreId,
			Name:      rows[k].Name,
			Display:   productpb.TagDisplay(rows[k].Display),
			Sort:      rows[k].Sort,
			CreatedBy: rows[k].CreatedBy,
			UpdatedBy: rows[k].UpdatedBy,
			CreatedAt: rows[k].CreatedAt.Format(utils.TIME_STD_FORMART),
			UpdatedAt: rows[k].UpdatedAt.Format(utils.TIME_STD_FORMART),
		})
	}

	return &productpb.ListTagRes{
		Total: total,
		Tags:  list,
	}, nil
}
