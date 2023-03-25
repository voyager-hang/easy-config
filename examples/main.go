package main

import (
	"fmt"
	easyConfig "github.com/voyager-hang/go-easy-config"
	"github.com/voyager-hang/go-easy-config/nacos_conf"
)

func main() {
	// file
	//ec := easyConfig.New()
	//ec.SetType(easyConfig.ConfigTypeFile)
	//ec.SetFileConf(file_conf.FileConfBox{
	//	ConfigPaths: []string{".", "./config", "./bbb/ddd/"},
	//	ConfigName:  []string{"conf.yaml", "etcd.yaml"},
	//})
	//err := ec.Load()
	//if err != nil {
	//	return
	//}
	//fmt.Println(ec.GetAll())

	ec := easyConfig.New()
	ec.SetType(easyConfig.ConfigTypeNacos)
	ec.SetNacosConf(nacos_conf.ConfBox{
		//Host: []constant.ServerConfig{
		//	constant.ServerConfig{
		//		IpAddr: "127.0.0.1",
		//		Port:   8848,
		//	},
		//},
		HostYaml: "./config/nacos.yaml",
		ConfInfo: []nacos_conf.ConfInfo{
			nacos_conf.ConfInfo{
				//Namespace: "575856a7-be79-4142-a289-b013a9dcfcdf",
				ConfKey: []nacos_conf.ConfKey{
					nacos_conf.ConfKey{
						Group: "TOKER_GROUP",
						DataId: []string{
							"admin_srv",
							"basic_api_srv",
							"control_center_srv",
							"consul_srv",
							"etcd_srv",
							"mysql_srv",
							"rabbitmq_srv",
							"redis_srv",
							"route_srv",
							"token_srv",
						},
					},
				},
			},
		},
		TimeoutMs:           5000,
		NotLoadCacheAtStart: false,
		LogDir:              "nacos/logs",
		CacheDir:            "nacos/cache",
		LogLevel:            "nacos/debug",
	})
	err := ec.Load()
	if err != nil {
		return
	}
	ec = ec.Find("socketIM")
	fmt.Println(ec.GetAll())
}
