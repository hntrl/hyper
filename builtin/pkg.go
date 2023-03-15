package builtin

import "github.com/hntrl/lang/symbols"

var Packages = map[string]symbols.Object{
	"math":    MathPackage{},
	"errors":  ErrorsPackage{},
	"units":   UnitsPackage{},
	"request": RequestPackage{},
	"mime":    MimeTypesPackage{},
}

var Classes = map[string]symbols.Class{
	"MimeType":      MimeType{},
	"RequestConfig": RequestConfig{},
	"HTTPResponse":  HTTPResponse{},
	"Dimension":     Dimension{},
}
