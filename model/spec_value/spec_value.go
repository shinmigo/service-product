package spec_value

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"goshop/service-product/pkg/db"
)

type SpecValue struct {
	SpecId    uint64
	Content   string
	CreatedBy uint64
	UpdatedBy uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SpecContent struct {
	SpecId  uint64 `json:"spec_id"`
	Content string `json:"content"`
}

func GetTableName() string {
	return "spec_value"
}

func GetField() []string {
	return []string{
		"spec_id", "content",
	}
}

func GetContentsBySpecIds(specIds []uint64) (map[uint64][]string, error) {
	specIdLen := len(specIds)
	if specIdLen == 0 {
		return nil, nil
	}

	rows := make([]*SpecContent, 0, specIdLen)
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("spec_id in (?)", specIds).
		Find(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}

	list := make(map[uint64][]string, specIdLen)
	for k := range rows {
		list[rows[k].SpecId] = append(list[rows[k].SpecId], rows[k].Content)
	}
	return list, nil
}

func BatchInsert(db *gorm.DB, specs []*SpecValue) error {
	var buf bytes.Buffer
	sql := "INSERT INTO spec_value (spec_id, content, created_by, updated_by, created_at, updated_at) VALUES "
	if _, err := buf.WriteString(sql); err != nil {
		return err
	}

	for k := range specs {
		if k == len(specs)-1 {
			buf.WriteString(fmt.Sprintf("(%d, '%s', %d, %d, '%s', '%s');",
				specs[k].SpecId,
				specs[k].Content,
				specs[k].CreatedBy,
				specs[k].UpdatedBy,
				specs[k].CreatedAt,
				specs[k].UpdatedAt,
			))
		} else {
			buf.WriteString(fmt.Sprintf("(%d, '%s', %d, %d, '%s', '%s'),",
				specs[k].SpecId,
				specs[k].Content,
				specs[k].CreatedBy,
				specs[k].UpdatedBy,
				specs[k].CreatedAt,
				specs[k].UpdatedAt,
			))
		}
	}
	return db.Exec(buf.String()).Error
}
