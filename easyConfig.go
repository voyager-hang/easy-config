package easyConfig

import (
	"encoding/json"
	"github.com/voyager-hang/go-easy-config/cast"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

type EasyConfig struct {
	configPaths       []string
	configName        []string
	configExt         []string
	configFilePath    []string
	configFileContent map[string][]byte
	config            map[string]interface{}
}

func New() *EasyConfig {
	ec := new(EasyConfig)
	ec.configName = []string{}
	ec.configPaths = []string{}
	ec.configExt = []string{}
	ec.configFilePath = []string{}
	ec.configFileContent = map[string][]byte{}
	ec.config = make(map[string]interface{})
	return ec
}

func (ec *EasyConfig) AddConfigName(cn ...string) {
	for _, v := range cn {
		n := v
		if strings.Contains(v, ".") {
			nArr := strings.Split(v, ".")
			if !inArray(nArr[len(nArr)-1], ec.configExt) {
				ec.configExt = append(ec.configExt, nArr[len(nArr)-1])
			}
			extLen := strings.Count(nArr[len(nArr)-1], "") + 1
			n = v[:strings.Count(v, "")-extLen]
		}
		if !inArray(n, ec.configName) {
			ec.configName = append(ec.configName, n)
		}
	}
}

func (ec *EasyConfig) AddConfigPaths(cp ...string) {
	for _, v := range cp {
		if strings.HasPrefix(v, "./") {
			v = strings.TrimRight(v[2:], "/")
		}
		if v == "." {
			v = ""
		}
		if !inArray(v, ec.configPaths) {
			ec.configPaths = append(ec.configPaths, v)
		}
	}
}

func (ec *EasyConfig) AddConfigExt(ce ...string) {
	for _, v := range ce {
		if !inArray(v, ec.configExt) {
			ec.configExt = append(ec.configExt, v)
		}
	}
}

func (ec *EasyConfig) Load() {
	if len(ec.configPaths) == 0 {
		log.Fatalln("configPaths is empty")
	}
	if len(ec.configName) == 0 {
		log.Fatalln("configName is empty")
	}
	if len(ec.configExt) == 0 {
		log.Fatalln("configExt is empty")
	}
	for _, vp := range ec.configPaths {
		if vp != "" {
			vp += "/"
		}
		for _, vn := range ec.configName {
			vn += "."
			for _, ve := range ec.configExt {
				ec.configFilePath = append(ec.configFilePath, vp+vn+ve)
			}
		}
	}
	// 读取文件内容
	nowPath := getPwd()
	for _, v := range ec.configFilePath {
		readPath := nowPath + "/" + v
		if exists(readPath) && isFile(readPath) {
			content, err := os.ReadFile(readPath)
			if err != nil {
				log.Fatalln("read :", v, err)
			}
			ec.configFileContent[v] = content
			ec.setConfig(content)
		}
	}
	ec.setConfigLink()
}

func (ec *EasyConfig) setConfig(content []byte) {
	confMap := make(map[string]interface{})
	err := yaml.Unmarshal(content, &confMap)
	if err != nil {
		log.Fatalln(err)
	}
	for ck, cm := range confMap {
		ec.config[ck] = cm
	}
}

func (ec *EasyConfig) setConfigLink() {
	for _, key := range ec.AllKeys() {
		if strings.HasPrefix(ec.GetString(key), "this.") {
			ec.Set(key, ec.Get(ec.GetString(key)[5:]))
		}
	}
}

func (ec *EasyConfig) Find(key string) *EasyConfig {
	findV := New()
	data := ec.Get(key)
	if data == nil {
		return nil
	}

	if reflect.TypeOf(data).Kind() == reflect.Map {
		findV.config = cast.ToStringMap(data)
		return findV
	}
	return nil
}

func (ec *EasyConfig) Get(key string) interface{} {
	return ec.find(key, true)
}

func (ec *EasyConfig) find(key string, flagDefault bool) interface{} {
	var (
		val  interface{}
		path = strings.Split(key, ".")
	)
	// Set() override first
	val = ec.searchMap(ec.config, path)
	return val
}

func (ec *EasyConfig) searchMap(source map[string]interface{}, path []string) interface{} {
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
			return ec.searchMap(cast.ToStringMap(next), path[1:])
		case map[string]interface{}:
			// Type assertion is safe here since it is only reached
			// if the type of `next` is the same as the type being asserted
			return ec.searchMap(next.(map[string]interface{}), path[1:])
		default:
			// got a value but nested key expected, return "nil" for not found
			return nil
		}
	}
	return nil
}

func (ec *EasyConfig) IsSet(key string) bool {
	val := ec.find(key, false)
	return val != nil
}

func (ec *EasyConfig) Set(key string, value interface{}) {
	// If alias passed in, then set the proper override
	value = toCaseInsensitiveValue(value)

	path := strings.Split(key, ".")
	lastKey := path[len(path)-1]
	deepestMap := deepSearch(ec.config, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

func (ec *EasyConfig) AllKeys() []string {
	return ec.getAllKey(ec.config, "")
}

func (ec *EasyConfig) getAllKey(data map[string]interface{}, key string) []string {
	keyArr := []string{}
	for k, v := range data {
		if key != "" {
			k = key + "." + k
		}
		keyArr = append(keyArr, k)
		vm, ok := v.(map[string]interface{})
		if ok {
			keyArr = append(keyArr, ec.getAllKey(vm, k)...)
		}
	}
	return keyArr
}

func (ec *EasyConfig) GetAll() map[string]interface{} {
	return ec.config
}

// GetString returns the value associated with the key as a string.
func (ec *EasyConfig) GetString(key string) string {
	return cast.ToString(ec.Get(key))
}

// GetBool returns the value associated with the key as a boolean.
func (ec *EasyConfig) GetBool(key string) bool {
	return cast.ToBool(ec.Get(key))
}

// GetInt returns the value associated with the key as an integer.
func (ec *EasyConfig) GetInt(key string) int {
	return cast.ToInt(ec.Get(key))
}

// GetInt32 returns the value associated with the key as an integer.
func (ec *EasyConfig) GetInt32(key string) int32 {
	return cast.ToInt32(ec.Get(key))
}

// GetInt64 returns the value associated with the key as an integer.
func (ec *EasyConfig) GetInt64(key string) int64 {
	return cast.ToInt64(ec.Get(key))
}

// GetUint returns the value associated with the key as an unsigned integer.
func (ec *EasyConfig) GetUint(key string) uint {
	return cast.ToUint(ec.Get(key))
}

// GetUint16 returns the value associated with the key as an unsigned integer.
func (ec *EasyConfig) GetUint16(key string) uint16 {
	return cast.ToUint16(ec.Get(key))
}

// GetUint32 returns the value associated with the key as an unsigned integer.
func (ec *EasyConfig) GetUint32(key string) uint32 {
	return cast.ToUint32(ec.Get(key))
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func (ec *EasyConfig) GetUint64(key string) uint64 {
	return cast.ToUint64(ec.Get(key))
}

// GetFloat64 returns the value associated with the key as a float64.
func (ec *EasyConfig) GetFloat64(key string) float64 {
	return cast.ToFloat64(ec.Get(key))
}

// GetTime returns the value associated with the key as time.
func (ec *EasyConfig) GetTime(key string) time.Time {
	return cast.ToTime(ec.Get(key))
}

// GetDuration returns the value associated with the key as a duration.
func (ec *EasyConfig) GetDuration(key string) time.Duration {
	return cast.ToDuration(ec.Get(key))
}

// GetIntSlice returns the value associated with the key as a slice of int values.
func (ec *EasyConfig) GetIntSlice(key string) []int {
	return cast.ToIntSlice(ec.Get(key))
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (ec *EasyConfig) GetStringSlice(key string) []string {
	return cast.ToStringSlice(ec.Get(key))
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (ec *EasyConfig) GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(ec.Get(key))
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (ec *EasyConfig) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(ec.Get(key))
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func (ec *EasyConfig) GetStringMapStringSlice(key string) map[string][]string {
	return cast.ToStringMapStringSlice(ec.Get(key))
}

func (ec *EasyConfig) ToJson() ([]byte, error) {
	return json.Marshal(ec.config)
}
