package stdlib

import "github.com/hntrl/hyper/src/hyper/symbols"

var Packages = map[string]symbols.Object{
	"errors":  ErrorsPackage{},
	"math":    MathPackage{},
	"mime":    MimeTypesPackage{},
	"request": RequestPackage{},
	"time":    TimePackage{},
	"units":   UnitsPackage{},
}
