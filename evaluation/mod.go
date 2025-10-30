package evaluation

import (
	"github.com/pecet3/hmbk-script/object"
)

var builtInModules map[string]*object.Module

func initModules() {
	builtInModules = map[string]*object.Module{
		"http": {
			Name: "http",
			Env:  ModHttp(),
		},
	}
}
