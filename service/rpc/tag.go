package rpc

import (
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

func (t *Tag) AddTag(ctx context.Context, req *productpb.AddTagReq) (*productpb.AddTagRes, error) {
	aul := tag.Tag{
		StoreId:   req.Tag.StoreId,
		Name:      req.Tag.Name,
		CreatedBy: req.Tag.AdminId,
		UpdatedBy: req.Tag.AdminId,
	}

	if err := db.Conn.Table(tag.GetTableName()).Create(&aul).Error; err != nil {
		return nil, err
	}

	return &productpb.AddTagRes{
		TagId: aul.TagId,
	}, nil
}

func (t *Tag) EditTag(ctx context.Context, req *productpb.EditTagReq) (*productpb.EditTagRes, error) {
	tagInfo, err := tag.GetOneByTagId(req.Tag.TagId)
	if err != nil {
		return nil, err
	}

	aul := tag.Tag{
		StoreId:   req.Tag.StoreId,
		Name:      req.Tag.Name,
		UpdatedBy: req.Tag.AdminId,
	}

	if err := db.Conn.Table(tag.GetTableName()).Where("tag_id = ?", tagInfo.TagId).Updates(&aul).Error; err != nil {
		return nil, err
	}

	return &productpb.EditTagRes{
		Updated: 1,
	}, nil
}

func (t *Tag) DelTag(ctx context.Context, req *productpb.DelTagReq) (*productpb.DelTagRes, error) {
	tagInfo, err := tag.GetOneByTagId(req.TagId)
	if err != nil {
		return nil, err
	}

	if err := db.Conn.Table(tag.GetTableName()).Where("tag_id = ?", tagInfo.TagId).Delete(tag.Tag{}).Error; err != nil {
		return nil, err
	}

	return &productpb.DelTagRes{
		Deleted: 1,
	}, nil
}

func (t *Tag) ReadTag(ctx context.Context, req *productpb.ReadTagReq) (*productpb.ReadTagRes, error) {
	row, err := tag.GetOneByTagId(req.TagId)
	if err != nil {
		return nil, err
	}

	return &productpb.ReadTagRes{
		Tag: &productpb.TagInfo{
			TagId: row.TagId,
			Name:  row.Name,
		},
	}, nil
}

func (t *Tag) ReadTags(ctx context.Context, req *productpb.ReadTagsReq) (*productpb.ReadTagsRes, error) {
	list := []*productpb.TagInfo{}

	rows, err := tag.GetTags(1, 10)
	if err != nil {
		return nil, err
	}

	for k := range rows {
		list = append(list, &productpb.TagInfo{
			TagId: rows[k].TagId,
			Name:  rows[k].Name,
		})
	}

	return &productpb.ReadTagsRes{
		Tags: list,
	}, nil
}
