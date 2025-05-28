package yaml

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

func TestYaml(t *testing.T) {

	bytearr, err := os.ReadFile("app.yml")
	if err != nil {
		panic("测试bug")
	}
	map1, err := Read(bytearr)
	if err != nil {
		panic("测试bug")
	}

	// var nacos confnacos.NacosConfigProperties
	var nacos constant.ClientConfig

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true, // 允许弱类型转换
		Result:           &nacos,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		panic("测试bug")
	}

	if err := decoder.Decode(map1); err != nil {
		panic("测试bug")
	}

	// if err := util.MapToStruct(map1, &nacos); err != nil {
	// 	panic(err)
	// }

	fmt.Printf("-----%+v\n", nacos)

	// val := map1["nacos.config.dataidlist"]
	val, _ := GetValue(map1, "nacos", "config", "dataidlist")
	// if wapvalue, ok := val.([]map[interface{}]interface{}); ok {

	// }
	slog.Info("sha a!!! ", "值", val)
	for key, value := range map1 {
		slog.Info("**", "key", key, "value", value)
	}

}

// 基于 key-value的方式获取对应的key的值aaa.bbb.ccc
func GetValue(m map[string]interface{}, keys ...string) (interface{}, bool) {
	var current interface{} = m
	var ok bool
	for _, key := range keys {

		switch v := current.(type) {
		case map[string]interface{}:
			current, ok = v[key]
			if !ok {
				return nil, false
			}
		case map[interface{}]interface{}:
			current, ok = v[key]
			if !ok {
				return nil, false
			}
		default:
			return nil, false
		}

	}
	return current, true
}
