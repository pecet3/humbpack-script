package evaluation

import (
	"github.com/pecet3/hmbk-script/object"
)

var modules map[string]map[string]*object.Builtin

func initModules() {
	modules = map[string]map[string]*object.Builtin{
		"math": moduleMathInit(),
	}
}
