package param_value

import (
	"bytes"
	"fmt"

	"goshop/service-product/pkg/utils"

	"github.com/jinzhu/gorm"

	"goshop/service-product/pkg/db"
)

type ParamValue struct {
	ParamValueId uint64 `json:"param_value_id" gorm:"PRIMARY_KEY"`
	ParamId      uint64
	Content      string
	CreatedBy    uint64
	UpdatedBy    uint64
	CreatedAt    utils.JSONTime
	UpdatedAt    utils.JSONTime
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
	paramIdLen := len(paramIds)
	if paramIdLen == 0 {
		return nil, nil
	}

	rows := make([]*ParamContent, 0, paramIdLen)
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("param_id in (?)", paramIds).
		Find(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}

	list := make(map[uint64][]string, paramIdLen)
	for k := range rows {
		list[rows[k].ParamId] = append(list[rows[k].ParamId], rows[k].Content)
	}
	return list, nil
}

func BatchInsert(db *gorm.DB, params []*ParamValue) error {
	var buf bytes.Buffer
	sql := "INSERT INTO param_value (param_id, content, created_by, updated_by, created_at, updated_at) VALUES "
	if _, err := buf.WriteString(sql); err != nil {
		return err
	}

	for k := range params {
		if k == len(params)-1 {
			buf.WriteString(fmt.Sprintf("(%d, '%s', %d, %d, '%s', '%s');",
				params[k].ParamId,
				params[k].Content,
				params[k].CreatedBy,
				params[k].UpdatedBy,
				params[k].CreatedAt,
				params[k].UpdatedAt,
			))
		} else {
			buf.WriteString(fmt.Sprintf("(%d, '%s', %d, %d, '%s', '%s'),",
				params[k].ParamId,
				params[k].Content,
				params[k].CreatedBy,
				params[k].UpdatedBy,
				params[k].CreatedAt,
				params[k].UpdatedAt,
			))
		}
	}
	return db.Exec(buf.String()).Error
}
