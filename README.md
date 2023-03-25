# go-easy-config
可重用读取yaml  后续:json,yml，ETCD,consul
支持内部复用，写来自己用的，别太在意细节

大部分代码来自 [viper](github.com/spf13/viper)
本来用viper 他key转小写了没办法，只能自己造轮子了

>fike支持读取多文件，多目录
```go
    ec := easyConfig.New()
    ec.SetType(easyConfig.ConfigTypeFile)
    ec.SetFileConf(file_conf.FileConfBox{
        ConfigPaths: []string{".", "./config", "./bbb/ddd/"},
        ConfigName:  []string{"conf.yaml", "etcd.yaml"},
    })
    err := ec.Load()
    if err != nil {
        return
    }
    fmt.Println(ec.GetAll())
```
>Nacos
```go
    ec := easyConfig.New()
	ec.SetType(easyConfig.ConfigTypeNacos)
	ec.SetNacosConf(nacos_conf.ConfBox{
		Host: []constant.ServerConfig{
			constant.ServerConfig{
				IpAddr: "127.0.0.1",
				Port:   8848,
			},
		},
		HostYaml: "./config/nacos.yaml", // Host 和 HostYaml 配置一个就可以
		ConfInfo: []nacos_conf.ConfInfo{
			nacos_conf.ConfInfo{
				Namespace: "575856a7-be79-4142-a289-b013a9dcfcdf",
				ConfKey: []nacos_conf.ConfKey{
					nacos_conf.ConfKey{
						Group: "TOKER_GROUP",
						DataId: []string{
							"admin_srv",
							"basic_api_srv",
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
	fmt.Println(ec.GetAll())
```

>支持内部复用,支持夸文件复用 语法 this.key.key
```yaml
testServer:
  Name: apiDiscovery
  ListenOn: 0.0.0.0:8080
  Timeout: 10000
  Etcd:
    Hosts: this.Etcd.Hosts
    Key: pdftodocx.rpc
  Consul:
    Servers: this.Consul.Servers
Consul:
  Servers:
    - "http://127.0.0.1:8500"
```