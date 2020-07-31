package spec_value

import (
	"fmt"
	"time"

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
	rows := []*SpecContent{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("spec_id in (?)", specIds).
		Find(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}

	list := make(map[uint64][]string)
	for k := range rows {
		list[rows[k].SpecId] = append(list[rows[k].SpecId], rows[k].Content)
	}
	return list, nil
}
