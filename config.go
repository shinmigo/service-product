package main

import (
	"fmt"
	"goshop/service-product/pkg/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

//初始化配置文件
func InitConfig() {
	buf := &utils.Config{}
	UnmarshalYaml(fmt.Sprintf("./conf/%s.app.yaml", gin.Mode()), buf)
	utils.C = buf

	//app解析
	baseInfo := &utils.Base{}
	UnmarshalYaml("./conf/app.yaml", baseInfo)
	utils.C.Base = baseInfo
}

func UnmarshalYaml(fileName string, data interface{}) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(fileName)
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败, err: %v", err))
	}

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("%s文件发生改动哦 \n", e.Name)
		if err := v.Unmarshal(data); err != nil {
			panic(fmt.Sprintf("解析配置文件出错, err: %v", err))
		}
	})

	if err := v.Unmarshal(data); err != nil {
		panic(fmt.Sprintf("解析配置文件出错, err: %v", err))
	}

}
