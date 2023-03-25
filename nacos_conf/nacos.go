package nacos_conf

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/voyager-hang/go-easy-config/cast"
	"github.com/voyager-hang/go-easy-config/tool"
	"gopkg.in/yaml.v3"
	"log"
	"strings"
)

type nHost struct {
	Scheme      string `yaml:"Scheme"`
	ContextPath string `yaml:"ContextPath"`
	IpAddr      string `yaml:"IpAddr"`
	Port        uint64 `yaml:"Port"`
}
type nacosConf struct {
	Host []nHost `yaml:"Host"`
}

type ConfKey struct {
	Group  string
	DataId []string
}

type ConfInfo struct {
	Namespace string
	ConfKey   []ConfKey
}

type ConfBox struct {
	Host                []constant.ServerConfig
	HostYaml            string
	ConfInfo            []ConfInfo
	TimeoutMs           uint64
	NotLoadCacheAtStart bool
	LogDir              string
	CacheDir            string
	LogLevel            string
}

type NacosConf struct {
	ConfBox
	configContent map[string]string
	config        map[string]interface{}
}

func New() *NacosConf {
	n := new(NacosConf)
	n.Host = []constant.ServerConfig{}
	n.ConfInfo = []ConfInfo{}
	n.configContent = map[string]string{}
	n.config = make(map[string]interface{})
	return n
}

func (n *NacosConf) GetConfig() map[string]interface{} {
	return n.config
}

func (n *NacosConf) Load() error {
	if n.HostYaml != "" {
		content := tool.GetFileContent(n.HostYaml)
		if content != "" {
			confData := nacosConf{}
			err := yaml.Unmarshal([]byte(content), &confData)
			if err != nil {
				log.Println(err)
			} else {
				if confData.Host != nil {
					for _, v := range confData.Host {
						n.Host = append(n.Host, constant.ServerConfig{
							Scheme:      v.Scheme,
							ContextPath: v.ContextPath,
							IpAddr:      v.IpAddr,
							Port:        v.Port,
						})
					}
				}
			}
		}
	}
	var content string
	for _, v := range n.ConfInfo {
		cont, err := n.getContent(v)
		if err != nil {
			return err
		}
		content += cont
	}
	n.setConfigLink()
	return nil
	//err = configClient.ListenConfig(vo.ConfigParam{
	//	DataId: "data_id",
	//	Group:  "group",
	//	OnChange: func(namespace, group, dataId, data string) {
	//		fmt.Println("配置文件发生了变化...")
	//		fmt.Println("group:" + group + ", dataId:" + dataId + ", data:" + data)
	//	},
	//})
	//time.Sleep(300 * time.Second)
}

func (n *NacosConf) getContent(conf ConfInfo) (string, error) {
	cc := constant.ClientConfig{
		NamespaceId:         conf.Namespace,
		TimeoutMs:           n.TimeoutMs,
		NotLoadCacheAtStart: n.NotLoadCacheAtStart,
		LogDir:              n.LogDir,
		CacheDir:            n.CacheDir,
		LogLevel:            n.LogLevel,
	}
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": n.Host,
		"clientConfig":  cc,
	})
	if err != nil {
		return "", err
	}
	var allContent string
	var content string
	for _, gd := range conf.ConfKey {
		for _, dId := range gd.DataId {
			content, err = configClient.GetConfig(vo.ConfigParam{
				DataId: dId,
				Group:  gd.Group,
			})
			if err != nil {
				return "", err
			}
			n.configContent[gd.Group+":"+dId] = content
			n.setConfig([]byte(content))
			allContent += content
		}
	}
	return allContent, nil
}

func (n *NacosConf) setConfig(content []byte) {
	confMap := make(map[string]interface{})
	err := yaml.Unmarshal(content, &confMap)
	if err != nil {
		log.Fatalln(err)
	}
	for ck, cm := range confMap {
		n.config[ck] = cm
	}
}

func (n *NacosConf) setConfigLink() {
	for _, key := range n.AllKeys() {
		val := n.GetString(key)
		val = strings.ToLower(val)
		if strings.HasPrefix(val, "this.") {
			n.Set(key, n.Get(n.GetString(key)[5:]))
		}
	}
}

func (n *NacosConf) Get(key string) interface{} {
	return n.find(key, true)
}

func (n *NacosConf) find(key string, flagDefault bool) interface{} {
	var (
		val  interface{}
		path = strings.Split(key, ".")
	)
	// Set() override first
	val = n.searchMap(n.config, path)
	return val
}

func (n *NacosConf) searchMap(source map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	next, ok := source[path[0]]
	if ok {
		// Fast path
		if len(path) == 1 {
			return next
		}

		// Nested case
		switch next.(type) {
		case map[interface{}]interface{}:
			return n.searchMap(cast.ToStringMap(next), path[1:])
		case map[string]interface{}:
			// Type assertion is safe here since it is only reached
			// if the type of `next` is the same as the type being asserted
			return n.searchMap(next.(map[string]interface{}), path[1:])
		default:
			// got a value but nested key expected, return "nil" for not found
			return nil
		}
	}
	return nil
}

func (n *NacosConf) IsSet(key string) bool {
	val := n.find(key, false)
	return val != nil
}

func (n *NacosConf) Set(key string, value interface{}) {
	// If alias passed in, then set the proper override
	value = tool.ToCaseInsensitiveValue(value)

	path := strings.Split(key, ".")
	lastKey := path[len(path)-1]
	deepestMap := tool.DeepSearch(n.config, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

func (n *NacosConf) AllKeys() []string {
	return n.getAllKey(n.config, "")
}

func (n *NacosConf) getAllKey(data map[string]interface{}, key string) []string {
	keyArr := []string{}
	for k, v := range data {
		if key != "" {
			k = key + "." + k
		}
		keyArr = append(keyArr, k)
		vm, ok := v.(map[string]interface{})
		if ok {
			keyArr = append(keyArr, n.getAllKey(vm, k)...)
		}
	}
	return keyArr
}

func (n *NacosConf) GetString(key string) string {
	return cast.ToString(n.Get(key))
}
