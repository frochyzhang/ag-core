package reader

import (
	"github.com/frochyzhang/ag-core/ag/ag_conf/reader/json"
	"github.com/frochyzhang/ag-core/ag/ag_conf/reader/prop"
	"github.com/frochyzhang/ag-core/ag/ag_conf/reader/yaml"
)

type Reader func(b []byte) (map[string]any, error)

var Readers = map[string]Reader{
	"yaml":       yaml.Read,
	"yml":        yaml.Read,
	"json":       json.Read,
	"properties": prop.Read,
}
