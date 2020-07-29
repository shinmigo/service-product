package param_value

import (
	"fmt"
	"goshop/service-product/pkg/db"
	"time"
)

type ParamValue struct {
	ParamId   uint64
	Content   string
	CreatedBy uint64
	UpdatedBy uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ParamContent struct {
	ParamId uint64 `json:"param_id"`
	Content string `json:"content"`
}

func GetTableName() string {
	return "param_value"
}

func GetField() []string {
	return []string{
		"param_id", "content",
	}
}

func GetContentsByParamIds(paramIds []uint64) (map[uint64][]string, error) {
	rows := []*ParamContent{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("param_id in (?)", paramIds).
		Find(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}

	list := make(map[uint64][]string)
	for k := range rows {
		list[rows[k].ParamId] = append(list[rows[k].ParamId], rows[k].Content)
	}
	return list, nil
}
