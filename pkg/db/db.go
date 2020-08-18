package db

import (
	"bytes"
	"flag"
	"fmt"
	"goshop/service-product/pkg/utils"
	"net/url"

	"github.com/unknwon/com"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var (
	Conn *gorm.DB
)

func GetDbConnect() (*gorm.DB, error) {
	dbConStr := fmt.Sprintf("%s:%s@tcp(%s)/%s%s?charset=utf8&parseTime=True&loc=%s", utils.C.Mysql.User, utils.C.Mysql.Password, utils.C.Mysql.Host, utils.C.Mysql.Database, getConnectDbName(), url.QueryEscape("Asia/Shanghai"))
	db, err := gorm.Open("mysql", dbConStr)
	Conn = db

	if gin.Mode() == "debug" || gin.Mode() == "test" {
		db.LogMode(true)
	}

	return db, err
}

func getConnectDbName() string {
	dbName := flag.String("db", "", "")
	flag.Parse()

	return *dbName
}

func BatchInsert(db *gorm.DB, tableName string, fields []string, params [][]interface{}) error {
	var (
		buf bytes.Buffer
	)

	buf.WriteString("INSERT INTO ")
	buf.WriteString(tableName)
	buf.WriteString("(")
	for i, field := range fields {
		buf.WriteString(field)
		if i != len(fields)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString(") VALUES ")
	for i, param := range params {
		buf.WriteString("(")
		for j, value := range param {
			buf.WriteString("'")
			buf.WriteString(com.ToStr(value))
			buf.WriteString("'")
			if j != len(param)-1 {
				buf.WriteString(",")
			}
		}
		buf.WriteString(")")
		if i != len(params)-1 {
			buf.WriteString(",")
		}
	}

	return db.Exec(buf.String()).Error
}
