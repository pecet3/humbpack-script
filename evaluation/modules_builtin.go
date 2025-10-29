package evaluation

import (
	"github.com/pecet3/hmbk-script/object"
)

var modules map[string]*object.Module

func initModules() {
	modules = map[string]*object.Module{
		"http": {
			Name: "http",
			Env:  initHttpEnvModule(),
		},
	}
}

func initHttpEnvModule() *object.Environment {
	env := object.NewEnvironment()

	env.SetConst("hello", &object.String{
		Value: "world",
	})
	return env
}
