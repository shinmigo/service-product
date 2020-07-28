package rpc

import (
	"github.com/shinmigo/pb/productpb"
	
	"golang.org/x/net/context"
	"goshop/service-product/model/tag"
)

type Tag struct {
}

func NewTag() *Tag {
	return &Tag{}
}

func (t *Tag) AddTag(ctx context.Context, req *productpb.AddTagRequest) (*productpb.AddTagResponse, error) {
	aul := tag.Tag{
		StoreId:   req.Tag.StoreId,
		Name:      req.Tag.Name,
		CreatedBy: req.Tag.AdminId,
		UpdatedBy: req.Tag.AdminId,
	}
	
	tagId, err := tag.AddTag(&aul)
	
	return &productpb.AddTagResponse{
		TagId: tagId,
	}, err
}

func (t *Tag) EditTag(ctx context.Context, req *productpb.EditTagRequest) (*productpb.EditTagResponse, error) {
	aul := tag.Tag{
		StoreId:   req.Tag.StoreId,
		Name:      req.Tag.Name,
		UpdatedBy: req.Tag.AdminId,
	}
	
	res := int32(0)
	err := tag.EditTag(req.Tag.TagId, aul)
	if err == nil {
		res = 1
	}
	return &productpb.EditTagResponse{
		Updated: res,
	}, err
}

func (t *Tag) DelTag(ctx context.Context, req *productpb.DelTagRequest) (*productpb.DelTagResponse, error) {
	res := int32(0)
	err := tag.DelTag(req.TagId)
	if err == nil {
		res = 1
	}
	return &productpb.DelTagResponse{
		Deleted: res,
	}, err
}

func (t *Tag) ReadTag(ctx context.Context, req *productpb.ReadTagRequest) (*productpb.ReadTagResponse, error) {
	row, err := tag.GetOneByTagId(req.TagId)
	if err != nil {
		return nil, err
	}
	
	var td productpb.TagInfo
	td.TagId = row.TagId
	td.Name = row.Name
	
	return &productpb.ReadTagResponse{
		Tag: &td,
	}, err
}

func (t *Tag) ReadTags(ctx context.Context, req *productpb.ReadTagsRequest) (*productpb.ReadTagsResponse, error) {
	list := []*productpb.TagInfo{}
	
	rows, err := tag.GetTags(1, 10)
	if err != nil {
		return nil, err
	}
	for k := range *rows {
		list = append(list, &productpb.TagInfo{
			TagId: (*rows)[k].TagId,
			Name:  (*rows)[k].Name,
		})
	}
	
	return &productpb.ReadTagsResponse{
		Tags: list,
	}, err
}
