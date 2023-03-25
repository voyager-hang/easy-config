package tool

import (
	"github.com/voyager-hang/go-easy-config/cast"
	"os"
	"strings"
)

// GetPwd 获取执行路径
func GetPwd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

func ToCaseInsensitiveValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[interface{}]interface{}:
		value = CopyAndInsensitiviseMap(cast.ToStringMap(v))
	case map[string]interface{}:
		value = CopyAndInsensitiviseMap(v)
	}

	return value
}

func CopyAndInsensitiviseMap(m map[string]interface{}) map[string]interface{} {
	nm := make(map[string]interface{})

	for key, val := range m {
		lkey := strings.ToLower(key)
		switch v := val.(type) {
		case map[interface{}]interface{}:
			nm[lkey] = CopyAndInsensitiviseMap(cast.ToStringMap(v))
		case map[string]interface{}:
			nm[lkey] = CopyAndInsensitiviseMap(v)
		default:
			nm[lkey] = v
		}
	}

	return nm
}

func DeepSearch(m map[string]interface{}, path []string) map[string]interface{} {
	for _, k := range path {
		m2, ok := m[k]
		if !ok {
			// intermediate key does not exist
			// => create it and continue from there
			m3 := make(map[string]interface{})
			m[k] = m3
			m = m3
			continue
		}
		m3, ok := m2.(map[string]interface{})
		if !ok {
			// intermediate key is a value
			// => replace with a new map
			m3 = make(map[string]interface{})
			m[k] = m3
		}
		// continue search from here
		m = m3
	}
	return m
}

func InArray(target string, strArray []string) bool {
	for _, element := range strArray {
		if target == element {
			return true
		}
	}
	return false
}

func GetFileContent(filePath string) string {
	if filePath == "" || filePath == "." || filePath == "./" || filePath == "/" {
		return ""
	}
	if strings.HasPrefix(filePath, "./") {
		filePath = strings.TrimLeft(filePath, "./")
	}
	if !strings.HasPrefix(filePath, "/") {
		filePath = GetPwd() + "/" + filePath
	}
	// 读取文件内容
	if Exists(filePath) && IsFile(filePath) {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return ""
		}
		return string(content)
	}
	return ""
}
