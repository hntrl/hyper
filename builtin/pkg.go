package builtin

import (
	"github.com/hntrl/lang/build"
)

func RegisterDefaults(ctx *build.BuildContext) {
	// Register the default packages
	ctx.RegisterPackage("math", MathPackage{})
	ctx.RegisterPackage("errors", ErrorsPackage{})
	ctx.RegisterPackage("units", UnitsPackage{})
	ctx.RegisterPackage("request", RequestPackage{})
	ctx.RegisterPackage("mime", MimeTypesPackage{})
}
