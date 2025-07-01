package dao

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type Parse interface {
	// 文件路径  渠道装载解析的数据
	ParseFile(conf *SchemaConfig, ch chan *TableData, wait *sync.WaitGroup)
}
type YamlParse struct {
}

func (yamlParse *YamlParse) ParseFile(conf *SchemaConfig, ch chan *TableData, wait *sync.WaitGroup) {
	// 遍历yaml文件
	yamlPath := conf.ConfigPath
	fileArr, error := os.ReadDir(yamlPath)
	if error != nil {
		log.Panic("read dir ", yamlPath, " file fail")
	}

	templateWait := &sync.WaitGroup{}
	templateWait.Add(len(fileArr))
	// 构建模版数据并行进行
	// templateWait:=&sync.WaitGroup{}
	for _, file := range fileArr {

		go func(entry os.DirEntry) {
			defer templateWait.Done()
			fileName := entry.Name()
			// 只处理yaml后缀的文件
			suffix := filepath.Ext(fileName)
			// log.Println("当前处理的文件名后缀为:",suffix)
			if suffix != ".yaml" {
				return
			}
			// 如何设置了目标文件就要过滤掉非目标文件
			if conf.TargetFileName != "" {
				if !CheckOrNotContains(strings.Split(conf.TargetFileName, ","), fileName) {
					return
				}
			}
			yamlPath := filepath.Join(yamlPath, fileName)
			data, err := os.ReadFile(yamlPath)
			if err != nil {
				log.Println("file name:", fileName, ",error:", err)
				return
			}
			yamlData := &YamlData{}
			error := yaml.Unmarshal(data, yamlData)

			if error != nil {
				log.Println(fileName, "data ref to model error", error)
				return
			}
			if yamlData.ModuleName == "" {
				yamlData.ModuleName = conf.ModuleName
			}
			// 将yaml文件解析为模板数据
			tableData := YamlDataToTemplate(yamlData)
			// 将数据放入信道通知消费
			ch <- tableData
			wait.Add(1)
		}(file)
	}
	// 关闭信道
	templateWait.Wait()
	close(ch)
}
