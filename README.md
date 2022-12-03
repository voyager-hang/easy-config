# go-easy-config
可重用读取yaml  后续:json,yml，ETCD,consul
支持内部复用，写来自己用的，别太在意细节

大部分代码来自 [viper](github.com/spf13/viper)
本来用viper 他key转小写了没办法，只能自己造轮子了

>支持读取多文件，多目录
```go
    ec := easyConfig.New()
	ec.AddConfigPaths(".", "./config")
	ec.AddConfigName("conf.yaml", "etcd.yaml") // 带后缀名自动设置AddConfigExt
	//ec.AddConfigExt()
	ec.Load()
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