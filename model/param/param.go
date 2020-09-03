package param

import (
	"fmt"

	"goshop/service-product/pkg/utils"

	"goshop/service-product/pkg/db"
)

type Param struct {
	ParamId   uint64         `json:"param_id" gorm:"PRIMARY_KEY"`
	StoreId   uint64         `json:"store_id"`
	KindId    uint64         `json:"kind_id"`
	Name      string         `json:"name"`
	Type      int32          `json:"type"`
	Sort      uint64         `json:"sort"`
	CreatedBy uint64         `json:"created_by"`
	UpdatedBy uint64         `json:"updated_by"`
	CreatedAt utils.JSONTime `json:"created_at"`
	UpdatedAt utils.JSONTime `json:"updated_at"`
}

func GetTableName() string {
	return "param"
}

func GetField() []string {
	return []string{
		"param_id", "store_id", "kind_id", "name", "type", "sort",
		"created_by", "updated_by", "created_at", "updated_at",
	}
}

func GetOneByParamId(ParamId uint64) (*Param, error) {
	if ParamId == 0 {
		return nil, fmt.Errorf("param_id is null")
	}
	row := &Param{}
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("param_id = ?", ParamId).
		First(row).Error

	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}
	return row, nil
}

func GetParamList(paramId uint64, paramName string, page, pageSize uint64) ([]*Param, uint64, error) {
	var total uint64

	rows := make([]*Param, 0, pageSize)

	query := db.Conn.Table(GetTableName()).Select(GetField())
	if paramId > 0 {
		query = query.Where("param_id = ?", paramId)
	}

	if paramName != "" {
		query = query.Where("name like ?", "%"+paramName+"%")
	}

	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	if err != nil {
		return nil, total, err
	}

	query.Count(&total)

	return rows, total, nil
}

func GetParamsByKindId(kindIds []uint64) (map[uint64][]*Param, error) {
	rows := make([]*Param, 0, len(kindIds))
	err := db.Conn.Table(GetTableName()).
		Select(GetField()).
		Where("kind_id in (?)", kindIds).
		Find(&rows).Error

	if err != nil {
		return nil, err
	}

	list := make(map[uint64][]*Param, len(kindIds))
	for k := range rows {
		list[rows[k].KindId] = append(list[rows[k].KindId], rows[k])
	}
	return list, nil
}

//不是所有的id都存在记录，将返回false及不存在的id
func ExistParamsByIds(ids []uint64) ([]uint64, bool) {
	rows := make([]*struct{ ParamId uint64 }, 0, len(ids))
	db.Conn.Model(&Param{}).Select("param_id").Where("param_id in (?)", ids).Scan(&rows)

	queryIds := make([]uint64, 0, len(rows))
	for i := range rows {
		queryIds = append(queryIds, rows[i].ParamId)
	}
	diffIds := utils.SliceDiffUint64(ids, queryIds)
	return diffIds, len(diffIds) == 0
}
