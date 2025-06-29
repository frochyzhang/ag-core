package ag_nacos

import "github.com/frochyzhang/ag-core/ag/ag_conf"

// NacosPropertySource nacos配置实体
type NacosPropertySource struct {
	ag_conf.MapPropertySource
}

// NewNacosPropertySource naocs远程配置相关内容 当前是main方法主动放入env
func NewNacosPropertySource(name string, source map[string]any) *NacosPropertySource {
	return &NacosPropertySource{
		MapPropertySource: ag_conf.MapPropertySource{
			NamedPropertySource: ag_conf.NamedPropertySource{
				Name: name,
			},
			Source: source,
		},
	}

}
