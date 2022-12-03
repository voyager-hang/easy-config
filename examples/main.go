package main

import (
	"fmt"
	easyConfig "github.com/voyager-hang/go-easy-config"
)

func main() {
	ec := easyConfig.New()
	ec.AddConfigPaths(".", "./config", "./bbb/ddd/")
	ec.AddConfigName("conf.yaml", "etcd.yaml") // 带后缀名自动设置AddConfigExt
	//ec.AddConfigExt()
	ec.Load()
	fmt.Println(ec.GetAll())
}
