package db

import (
	"flag"
	"fmt"
	"goshop/service-product/pkg/utils"
	"net/url"

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
