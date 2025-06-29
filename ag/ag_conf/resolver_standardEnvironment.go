package ag_conf

import (
	"github.com/frochyzhang/ag-core/ag/ag_crypto"
	"strings"
)

const (
	constSystemPropertiesPropertySourceName  = "systemProperties"
	constSystemEnvironmentPropertySourceName = "systemEnvironment"
)

// StandardEnvironment is a standard implementation of Environment.
type StandardEnvironment struct {
	AbstractEnvironment
}

// NewStandardEnvironment creates a new StandardEnvironment instance.
func NewStandardEnvironment() *StandardEnvironment {
	e := &StandardEnvironment{}
	// 初始化MutablePropertySources
	e.PropertySources = NewMutablePropertySources()
	// 初始化PropertyResolver,传入PropertySources
	e.PropertyResolver = NewPropertySourcesPropertyResolver(e.PropertySources)
	// customizePropertySources 环境变量和-D属性添加到配置源中
	e.customizePropertySources(e.PropertySources)
	return e
}

// customizePropertySources 环境变量和-D属性添加到配置源中
func (e *StandardEnvironment) customizePropertySources(ps *MutablePropertySources) {
	ps.AddLast(NewPropertiesPropertySource(constSystemPropertiesPropertySourceName, e.GetSystemProperties()))
	ps.AddLast(NewSystemEnvironmentPropertySource(constSystemEnvironmentPropertySourceName, e.GetSystemEnvironment()))
	// 做本地配置文件中相关字段的解密处理 TODO 应该在加载local配置后做一遍boot阶段的解密处理
	Decrypt(ps)
}

// Decrypt 遍历指定名字的properties,然后对应内容做解密处理 TODO 待优化
func Decrypt(ps *MutablePropertySources) {
	decryptSource := make(map[string]any, 5)
	ps.RangePropertySourceHandler(func(ps IPropertySource) (bool, error) {
		source := ps.GetSource()
		for key, value := range source {
			ciphertext, ok := value.(string)
			if ok && strings.HasPrefix(ciphertext, ConstEncryptKeyWords) {
				// plaintext, err := ag_ext.GetEncrytorPrimary().Decrypt(ciphertext)
				// 此处先临时使用base64解密,后续需重构调整
				plaintext, err := ag_crypto.Base64Encryptor.Decrypt(ciphertext)
				if err != nil {
					panic(err)
				}
				decryptSource[key] = plaintext
			}
		}
		return true, nil
	})

	// 将解密好的内容添加到env中并置换为第一位
	ps.AddFirst(&MapPropertySource{NamedPropertySource: NamedPropertySource{
		Name: ConstEncrptSystemPropertiesSource},
		Source: decryptSource,
	})

}
