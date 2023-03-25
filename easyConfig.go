package easyConfig

import (
	"encoding/json"
	"github.com/voyager-hang/go-easy-config/cast"
	"github.com/voyager-hang/go-easy-config/file_conf"
	"github.com/voyager-hang/go-easy-config/nacos_conf"
	"github.com/voyager-hang/go-easy-config/tool"
	"reflect"
	"strings"
	"time"
)

type ConfigType string

const (
	ConfigTypeFile  ConfigType = "file"
	ConfigTypeNacos ConfigType = "nacos"
)

type EasyConfInter interface {
	Load() error
	GetConfig() map[string]interface{}
}

type NacosConf struct {
}

type EasyConfig struct {
	configType ConfigType
	FileConf   file_conf.FileConfBox
	NacosConf  nacos_conf.ConfBox
	confObj    EasyConfInter
	config     map[string]interface{}
}

func New(confType ...ConfigType) *EasyConfig {
	ec := new(EasyConfig)
	ct := ConfigTypeFile
	if len(confType) > 0 {
		ct = confType[0]
	}
	ec.configType = ct
	ec.config = make(map[string]interface{})
	return ec
}

func (ec *EasyConfig) SetType(confType ConfigType) {
	ec.configType = confType
}
func (ec *EasyConfig) SetFileConf(fc file_conf.FileConfBox) {
	ec.FileConf = fc
}

func (ec *EasyConfig) SetNacosConf(nc nacos_conf.ConfBox) {
	ec.NacosConf = nc
}

func (ec *EasyConfig) Load() error {
	switch ec.configType {
	case ConfigTypeFile:
		f := file_conf.New()
		if ec.FileConf.ConfigPaths != nil {
			f.AddConfigPaths(ec.FileConf.ConfigPaths...)
		}
		if ec.FileConf.ConfigName != nil {
			f.AddConfigName(ec.FileConf.ConfigName...)
		}
		if ec.FileConf.ConfigExt != nil {
			f.AddConfigExt(ec.FileConf.ConfigExt...)
		}
		ec.confObj = f
	case ConfigTypeNacos:
		n := nacos_conf.New()
		c := ec.NacosConf
		n.Host = c.Host
		n.HostYaml = c.HostYaml
		n.ConfInfo = c.ConfInfo
		n.TimeoutMs = c.TimeoutMs
		n.NotLoadCacheAtStart = c.NotLoadCacheAtStart
		n.LogDir = c.LogDir
		n.CacheDir = c.CacheDir
		n.LogLevel = c.LogLevel
		ec.confObj = n
	}
	err := ec.confObj.Load()
	if err != nil {
		return err
	}
	ec.config = ec.confObj.GetConfig()
	return nil
}

func (ec *EasyConfig) Find(key string) *EasyConfig {
	findV := ec
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
	value = tool.ToCaseInsensitiveValue(value)

	path := strings.Split(key, ".")
	lastKey := path[len(path)-1]
	deepestMap := tool.DeepSearch(ec.config, path[0:len(path)-1])

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
