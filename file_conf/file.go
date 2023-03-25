package file_conf

import (
	"errors"
	"github.com/voyager-hang/go-easy-config/cast"
	"github.com/voyager-hang/go-easy-config/tool"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

type FileConfBox struct {
	ConfigPaths []string
	ConfigName  []string
	ConfigExt   []string
}

type FileConf struct {
	FileConfBox
	ConfigFilePath    []string
	configFileContent map[string][]byte
	config            map[string]interface{}
}

func New() *FileConf {
	f := new(FileConf)
	f.ConfigName = []string{}
	f.ConfigPaths = []string{}
	f.ConfigExt = []string{}
	f.ConfigFilePath = []string{}
	f.configFileContent = map[string][]byte{}
	f.config = make(map[string]interface{})
	return f
}

func (f *FileConf) GetConfig() map[string]interface{} {
	return f.config
}
func (f *FileConf) AddConfigName(cn ...string) {
	for _, v := range cn {
		n := v
		if strings.Contains(v, ".") {
			nArr := strings.Split(v, ".")
			if !tool.InArray(nArr[len(nArr)-1], f.ConfigExt) {
				f.ConfigExt = append(f.ConfigExt, nArr[len(nArr)-1])
			}
			extLen := strings.Count(nArr[len(nArr)-1], "") + 1
			n = v[:strings.Count(v, "")-extLen]
		}
		if !tool.InArray(n, f.ConfigName) {
			f.ConfigName = append(f.ConfigName, n)
		}
	}
}

func (f *FileConf) AddConfigPaths(cp ...string) {
	for _, v := range cp {
		if strings.HasPrefix(v, "./") {
			v = strings.TrimRight(v[2:], "/")
		}
		if v == "." {
			v = ""
		}
		if !tool.InArray(v, f.ConfigPaths) {
			f.ConfigPaths = append(f.ConfigPaths, v)
		}
	}
}

func (f *FileConf) AddConfigExt(ce ...string) {
	for _, v := range ce {
		if !tool.InArray(v, f.ConfigExt) {
			f.ConfigExt = append(f.ConfigExt, v)
		}
	}
}

func (f *FileConf) Load() error {
	if len(f.ConfigPaths) == 0 {
		return errors.New("configPaths is empty")
	}
	if len(f.ConfigName) == 0 {
		return errors.New("configName is empty")
	}
	if len(f.ConfigExt) == 0 {
		return errors.New("configExt is empty")
	}
	for _, vp := range f.ConfigPaths {
		if vp != "" {
			vp += "/"
		}
		for _, vn := range f.ConfigName {
			vn += "."
			for _, ve := range f.ConfigExt {
				f.ConfigFilePath = append(f.ConfigFilePath, vp+vn+ve)
			}
		}
	}
	// 读取文件内容
	nowPath := tool.GetPwd()
	for _, v := range f.ConfigFilePath {
		readPath := nowPath + "/" + v
		if tool.Exists(readPath) && tool.IsFile(readPath) {
			content, err := os.ReadFile(readPath)
			if err != nil {
				return errors.New("read :" + v + err.Error())
			}
			f.configFileContent[v] = content
			f.setConfig(content)
		}
	}
	f.setConfigLink()
	return nil
}

func (f *FileConf) setConfig(content []byte) {
	confMap := make(map[string]interface{})
	err := yaml.Unmarshal(content, &confMap)
	if err != nil {
		log.Fatalln(err)
	}
	for ck, cm := range confMap {
		f.config[ck] = cm
	}
}

func (f *FileConf) setConfigLink() {
	for _, key := range f.AllKeys() {
		val := f.GetString(key)
		val = strings.ToLower(val)
		if strings.HasPrefix(val, "this.") {
			f.Set(key, f.Get(f.GetString(key)[5:]))
		}
	}
}

func (f *FileConf) Get(key string) interface{} {
	return f.find(key, true)
}

func (f *FileConf) find(key string, flagDefault bool) interface{} {
	var (
		val  interface{}
		path = strings.Split(key, ".")
	)
	// Set() override first
	val = f.searchMap(f.config, path)
	return val
}

func (f *FileConf) searchMap(source map[string]interface{}, path []string) interface{} {
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
			return f.searchMap(cast.ToStringMap(next), path[1:])
		case map[string]interface{}:
			// Type assertion is safe here since it is only reached
			// if the type of `next` is the same as the type being asserted
			return f.searchMap(next.(map[string]interface{}), path[1:])
		default:
			// got a value but nested key expected, return "nil" for not found
			return nil
		}
	}
	return nil
}

func (f *FileConf) IsSet(key string) bool {
	val := f.find(key, false)
	return val != nil
}

func (f *FileConf) Set(key string, value interface{}) {
	// If alias passed in, then set the proper override
	value = tool.ToCaseInsensitiveValue(value)

	path := strings.Split(key, ".")
	lastKey := path[len(path)-1]
	deepestMap := tool.DeepSearch(f.config, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

func (f *FileConf) AllKeys() []string {
	return f.getAllKey(f.config, "")
}

func (f *FileConf) getAllKey(data map[string]interface{}, key string) []string {
	keyArr := []string{}
	for k, v := range data {
		if key != "" {
			k = key + "." + k
		}
		keyArr = append(keyArr, k)
		vm, ok := v.(map[string]interface{})
		if ok {
			keyArr = append(keyArr, f.getAllKey(vm, k)...)
		}
	}
	return keyArr
}

func (f *FileConf) GetString(key string) string {
	return cast.ToString(f.Get(key))
}
