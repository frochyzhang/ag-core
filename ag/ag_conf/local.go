package ag_conf

import (
	"embed"
	"fmt"
	"github.com/frochyzhang/ag-core/ag/ag_conf/reader/prop"
	"github.com/frochyzhang/ag-core/ag/ag_conf/reader/yaml"
	"github.com/frochyzhang/ag-core/ag/ag_ext"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const (
	// CustomerYamlSource 客户配置的yaml资源
	CustomerYamlSource string = "CustomerYamlSource"
	// CustomerPropertiesSource 客户配置的properties资源
	CustomerPropertiesSource string = "CustomerPropertiesSource"
	// LocalDefaultYamlSource 二进制可执行文件中默认配置的yaml文件
	LocalDefaultYamlSource string = "LocalDefaultYamlSource"
	// LocalDefaultPropertiesSource 二进制可执行中默认配置的properties文件
	LocalDefaultPropertiesSource string = "LocalDefaultPropertiesSource"
	// LocalProfileYamlSource 基于用户配置的profile加载对应的yaml文件
	LocalProfileYamlSource string = "LocalProfileYamlSource"
	// LocalProfilePropertiesSource 基于用户配置的profile加载对应的properties文件
	LocalProfilePropertiesSource string = "LocalProfilePropertiesSource"
)

type LocalConfigLoded string

var localConfLoadOnce sync.Once

// LoadLocalConfig 加载本地配置 本地支持yaml|yml|properties后缀的三个文件
// 1. 先判断环境变量或者进程变量中是否配置app.conf对应的key的值,未配置取执行二进制文件中的app.suffix文件
// 2. 获取app.profile的环境的设置,例如dev,sit,uat等
// 3. 如果1的场景未配置，则按照app.suffix app_profile.suffix的顺序加载,后续的内容会覆盖前者
func LoadLocalConfig(env IConfigurableEnvironment, localEmbed embed.FS) {
	localConfLoadOnce.Do(
		func() {
			doLoadLocalConfig(env, localEmbed)
		},
	)
}

func LoadLocalConfigToState(env IConfigurableEnvironment, localEmbed embed.FS) LocalConfigLoded {
	LoadLocalConfig(env, localEmbed)
	return "localConfigLoaded"
}

func doLoadLocalConfig(env IConfigurableEnvironment, localEmbed embed.FS) {
	slog.Info("--- LoadLocalConfig ---")
	// 1. 判断 os 或者env中是否设置文件路径
	appConf := env.GetProperty("app.conf")

	if appConf == "" {
		slog.Info("os or env no config app.conf key")
		// 加载二进制执行文件中的默认配置
		err := loadLocalConfigFile(env, localEmbed)
		if err != nil {
			panic(fmt.Errorf("load local config file err:%s", err))
		}
	} else {
		// 执行客户配置的文件
		err := LoadCustomerConfigFile(env, appConf)
		if err != nil {
			panic(fmt.Errorf("load customer config file err:%s", err))
		}
	}
}

func loadLocalConfigFile(env IConfigurableEnvironment, localEmbed embed.FS) error {
	// 先去加载默认的 app.yaml|app.yml|app.properties
	err := loadLocalDefaultFile(env, localEmbed)
	if err != nil {
		panic(err)
	}

	// 设置的环境类型
	profile := env.GetProperty("app.profile")
	if profile == "" {
		return nil
	}
	return loadLocalProfileFile(env, localEmbed, profile)
}

// loadLocalProfileFile 基于profile加载本地文件
func loadLocalProfileFile(env IConfigurableEnvironment, localEmbed embed.FS, profile string) error {

	// 加载yaml后缀的本地文件
	err := loadLocalProfileYamlFile(env, profile, "yaml", localEmbed)
	if err != nil {
		return err
	}
	// 加载yml后缀的本地文件
	err = loadLocalProfileYamlFile(env, profile, "yml", localEmbed)
	if err != nil {
		return err
	}
	// 加载properties后缀的文件
	err = loadLocalProfilePropertyFile(env, profile, localEmbed)
	if err != nil {
		return err
	}
	return nil

}

// loadLocalProfileYamlFile 加载yaml格式的环境变量
func loadLocalProfileYamlFile(env IConfigurableEnvironment, profile string, suffix string, localEmbed embed.FS) error {
	yamlContext, err := localEmbed.ReadFile("conf/app_" + profile + "." + suffix)
	if err != nil {
		slog.Error("process local profile yaml config file err", "errormsg", err)
	} else {
		mapcontext, err := yaml.Read(yamlContext)
		if err != nil {
			return err
		}
		flatmapcontext, err := ag_ext.GetFlattenedMap(mapcontext)
		if err != nil {
			return err
		}
		if env.GetPropertySources().Contains(LocalDefaultYamlSource) {
			env.GetPropertySources().AddBefore(LocalDefaultYamlSource, &MapPropertySource{
				NamedPropertySource: NamedPropertySource{
					Name: LocalProfileYamlSource,
				},
				Source: flatmapcontext,
			})
			return nil
		}

		// yaml|yml的优先级要比properties的优先级高
		if env.GetPropertySources().Contains(LocalProfilePropertiesSource) {
			env.GetPropertySources().AddBefore(LocalProfilePropertiesSource, &MapPropertySource{
				NamedPropertySource: NamedPropertySource{
					Name: LocalProfileYamlSource,
				},
				Source: flatmapcontext,
			})
			return nil
		}

		env.GetPropertySources().AddLast(&MapPropertySource{
			NamedPropertySource: NamedPropertySource{
				Name: LocalProfileYamlSource,
			},
			Source: flatmapcontext,
		})
	}
	return nil
}

// loadLocalProfilePropertyFile 加载本地环境变量
func loadLocalProfilePropertyFile(env IConfigurableEnvironment, profile string, localEmbed embed.FS) error {
	yamlContext, err := localEmbed.ReadFile("conf/app_" + profile + ".properties")
	if err != nil {
		slog.Error("process local profile properties config file err", "errormsg", err)
	} else {
		mapcontext, err := prop.Read(yamlContext)
		if err != nil {
			return err
		}
		flatmapcontext, err := ag_ext.GetFlattenedMap(mapcontext)
		if err != nil {
			return err
		}
		// yaml|yml的优先级要比properties的优先级高
		if env.GetPropertySources().Contains(LocalDefaultPropertiesSource) {
			env.GetPropertySources().AddBefore(LocalDefaultPropertiesSource, &MapPropertySource{
				NamedPropertySource: NamedPropertySource{
					Name: LocalProfilePropertiesSource,
				},
				Source: flatmapcontext,
			})
			return nil
		}

		env.GetPropertySources().AddLast(&MapPropertySource{
			NamedPropertySource: NamedPropertySource{
				Name: LocalProfilePropertiesSource,
			},
			Source: flatmapcontext,
		})

	}
	return nil
}

// loadLocalDefaultFile 按照顺序加载本地配置文件
// 先加载 yaml|yml 再加载properties文件,和springboot对于配置文件的优先级保持一致
func loadLocalDefaultFile(env IConfigurableEnvironment, localEmbed embed.FS) error {

	defaultYamlContext, err := localEmbed.ReadFile("app.yaml")
	if err != nil {
		slog.Info("process local default yaml config file err", "errormsg", err)
	} else {
		mapcontext, err := yaml.Read(defaultYamlContext)
		if err != nil {
			return err
		}
		flatmapcontext, err := ag_ext.GetFlattenedMap(mapcontext)
		if err != nil {
			return err
		}
		env.GetPropertySources().AddLast(&MapPropertySource{
			NamedPropertySource: NamedPropertySource{
				Name: LocalDefaultYamlSource,
			},
			Source: flatmapcontext,
		})
	}
	defaultYmlContext, err := localEmbed.ReadFile("app.yml")
	if err != nil {
		slog.Info("process local default yml config file err", "errormsg", err)
	} else {
		mapcontext, err := yaml.Read(defaultYmlContext)
		if err != nil {
			return err
		}
		flatmapcontext, err := ag_ext.GetFlattenedMap(mapcontext)
		if err != nil {
			return err
		}
		env.GetPropertySources().AddLast(&MapPropertySource{
			NamedPropertySource: NamedPropertySource{
				Name: LocalDefaultYamlSource,
			},
			Source: flatmapcontext,
		})
	}
	defaultPropertyContext, err := localEmbed.ReadFile("app.properties")
	if err != nil {
		slog.Info("process local default properties config file err", "errormsg", err)
	} else {
		mapcontext, err := prop.Read(defaultPropertyContext)
		if err != nil {
			return err
		}
		flatmapcontext, err := ag_ext.GetFlattenedMap(mapcontext)
		if err != nil {
			return err
		}
		env.GetPropertySources().AddLast(&MapPropertySource{
			NamedPropertySource: NamedPropertySource{
				Name: LocalDefaultPropertiesSource,
			},
			Source: flatmapcontext,
		})
	}
	return nil
}

// LoadCustomerConfigFile 加载客户环境变量中指定的应用配置文件
func LoadCustomerConfigFile(env IConfigurableEnvironment, customerConfigFile string) error {

	suffix := filepath.Ext(customerConfigFile)
	switch suffix {
	case ".yaml":
		return loadYamlFile(env, customerConfigFile)
	case ".yml":
		return loadYamlFile(env, customerConfigFile)
	case ".property":
		return loadPropertiesFile(env, customerConfigFile)
	default:
		return fmt.Errorf("file suffix:%s is not supported", suffix)
	}
}

// loadYamlFile 加载yaml文件
func loadYamlFile(env IConfigurableEnvironment, yamlfile string) error {

	context, err := os.ReadFile(yamlfile)
	if err != nil {
		return err
	}

	contextMap, err := yaml.Read(context)
	if err != nil {
		return err
	}
	flatmapcontext, err := ag_ext.GetFlattenedMap(contextMap)
	if err != nil {
		return err
	}
	// 如果localProperties已经存在,此时需要将yaml添加到它之前
	env.GetPropertySources().AddLast(&MapPropertySource{
		NamedPropertySource: NamedPropertySource{
			Name: CustomerYamlSource,
		},
		Source: flatmapcontext,
	})
	return nil
}

// loadPropertiesFile 加载property后缀的文件
func loadPropertiesFile(env IConfigurableEnvironment, propertiesfile string) error {
	context, err := os.ReadFile(propertiesfile)
	if err != nil {
		return err
	}

	contextMap, err := prop.Read(context)
	if err != nil {
		return err
	}
	flatmapcontext, err := ag_ext.GetFlattenedMap(contextMap)
	if err != nil {
		return err
	}
	// 字节数组转换为map
	env.GetPropertySources().AddLast(&MapPropertySource{
		NamedPropertySource: NamedPropertySource{
			Name: CustomerPropertiesSource,
		},
		Source: flatmapcontext,
	})
	return nil
}

// GetLocalMap 获取本地配置或者客户配置文件中的所有map并合并
func GetLocalMap(env IConfigurableEnvironment) map[string]interface{} {
	sournamelist := []string{LocalDefaultPropertiesSource,
		LocalDefaultYamlSource,
		LocalProfilePropertiesSource,
		LocalProfileYamlSource,
		CustomerPropertiesSource,
		CustomerYamlSource,
	}

	target := make(map[string]interface{}, 20)
	for _, sourename := range sournamelist {
		localmap := env.GetPropertySources().Get(sourename)
		if localmap != nil {
			for key, value := range localmap.GetSource() {
				target[key] = value
			}
		}
	}

	return target
}
