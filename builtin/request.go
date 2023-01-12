package builtin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hntrl/lang/build"
)

type RequestPackage struct{}

func sendRequest(method string, path string, config RequestConfig, body build.ValueObject) (*HTTPResponse, error) {
	var err error

	base := &url.URL{}
	if config.BaseURL != nil {
		base, err = url.Parse(string(*config.BaseURL))
		if err != nil {
			return nil, err
		}
	}
	requestUrl, err := base.Parse(string(path))
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	if config.Timeout != nil {
		client.Timeout = time.Millisecond * time.Duration(int64(*config.Timeout))
	}

	bodyBytes, err := json.Marshal(body.Value())
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, requestUrl.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	if config.Headers != nil {
		for k := range config.Headers.Fields() {
			val := config.Headers.Get(k)
			if valueObj, ok := val.(build.ValueObject); ok {
				req.Header.Set(k, fmt.Sprintf("%v", valueObj.Value()))
			}
		}
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resHeaders := build.NewGenericObject()
	for k, v := range res.Header {
		err := resHeaders.Set(k, build.StringLiteral(strings.Join(v, ",")))
		if err != nil {
			return nil, err
		}
	}
	return &HTTPResponse{
		Status:  build.IntegerLiteral(res.StatusCode),
		Headers: resHeaders,
		body:    res.Body,
	}, nil
}

func (rp RequestPackage) Get(key string) build.Object {
	methods := map[string]build.Object{
		"get": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
				build.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				url := args[0].(build.StringLiteral)
				config := args[1].(*build.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodGet, string(url), config, build.NilLiteral{})
			},
		}),
		"post": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
				build.NewOptionalClass(build.GenericObject{}),
				build.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				url := args[0].(build.StringLiteral)
				config := args[2].(*build.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodPost, string(url), config, args[1])
			},
		}),
		"put": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
				build.NewOptionalClass(build.GenericObject{}),
				build.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				url := args[0].(build.StringLiteral)
				config := args[2].(*build.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodPut, string(url), config, args[1])
			},
		}),
		"delete": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
				build.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				url := args[0].(build.StringLiteral)
				config := args[2].(*build.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodDelete, string(url), config, build.NilLiteral{})
			},
		}),
		"options": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
				build.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				url := args[0].(build.StringLiteral)
				config := args[2].(*build.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodOptions, string(url), config, build.NilLiteral{})
			},
		}),
		"head": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
				build.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				url := args[0].(build.StringLiteral)
				config := args[2].(*build.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodHead, string(url), config, build.NilLiteral{})
			},
		}),
		"patch": build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{
				build.String{},
				build.NewOptionalClass(build.GenericObject{}),
				build.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				url := args[0].(build.StringLiteral)
				config := args[2].(*build.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodPatch, string(url), config, args[1])
			},
		}),
	}
	if fn, ok := methods[key]; ok {
		return fn
	}
	if status, ok := statusCodes[key]; ok {
		return status
	}
	return nil
}

type RequestConfig struct {
	Headers *build.GenericObject  `hash:"ignore"`
	Params  *build.GenericObject  `hash:"ignore"`
	Timeout *build.IntegerLiteral `hash:"ignore"`
	BaseURL *build.StringLiteral  `hash:"ignore"`
}

func (rc RequestConfig) ClassName() string {
	return "RequestConfig"
}
func (rc RequestConfig) Fields() map[string]build.Class {
	return map[string]build.Class{
		"headers": build.NewOptionalClass(build.GenericObject{}),
		"params":  build.NewOptionalClass(build.GenericObject{}),
		"timeout": build.NewOptionalClass(build.Integer{}),
		"baseURL": build.NewOptionalClass(build.String{}),
	}
}
func (rc RequestConfig) Constructors() build.ConstructorMap {
	csMap := build.NewConstructorMap()
	csMap.AddGenericConstructor(rc, func(fields map[string]build.ValueObject) (build.ValueObject, error) {
		config := RequestConfig{}
		if headers, ok := fields["headers"]; ok {
			lit := headers.(*build.NilableObject).Object.(*build.GenericObject)
			config.Headers = lit
		}
		if params, ok := fields["params"]; ok {
			lit := params.(*build.NilableObject).Object.(*build.GenericObject)
			config.Headers = lit
		}
		if timeout, ok := fields["timeout"]; ok {
			lit := timeout.(*build.NilableObject).Object.(build.IntegerLiteral)
			config.Timeout = &lit
		}
		if baseURL, ok := fields["baseURL"]; ok {
			lit := baseURL.(*build.NilableObject).Object.(build.StringLiteral)
			config.BaseURL = &lit
		}
		return config, nil
	})
	return csMap
}
func (rc RequestConfig) Get(key string) build.Object {
	switch key {
	case "headers":
		return rc.Headers
	case "params":
		return rc.Params
	case "timeout":
		return rc.Timeout
	case "baseURL":
		return rc.BaseURL
	}
	return nil
}

func (rc RequestConfig) Class() build.Class {
	return rc
}
func (rc RequestConfig) Value() interface{} {
	return nil
}
func (rc RequestConfig) Set(key string, obj build.ValueObject) error {
	return nil
}

type HTTPResponse struct {
	Status  build.IntegerLiteral `hash:"ignore"`
	Headers *build.GenericObject `hash:"ignore"`
	body    io.ReadCloser        `hash:"ignore"`
}

func (hr HTTPResponse) ClassName() string {
	return "HTTPResponse"
}
func (hr HTTPResponse) Fields() map[string]build.Class {
	return map[string]build.Class{
		"status":  build.Integer{},
		"headers": build.GenericObject{},
	}
}
func (hr HTTPResponse) Constructors() build.ConstructorMap {
	return build.NewConstructorMap()
}
func (hr HTTPResponse) Get(key string) build.Object {
	switch key {
	case "status":
		return hr.Status
	case "headers":
		return hr.Headers
	case "object":
		return build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{},
			Returns:   build.GenericObject{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				resBody, err := io.ReadAll(hr.body)
				if err != nil {
					return nil, err
				}
				generic, err := build.FromBytes(resBody)
				if err != nil {
					return nil, err
				}
				return generic, nil
			},
		})
	case "text":
		return build.NewFunction(build.FunctionOptions{
			Arguments: []build.Class{},
			Returns:   build.String{},
			Handler: func(args []build.ValueObject, proto build.ValueObject) (build.ValueObject, error) {
				resBody, err := io.ReadAll(hr.body)
				if err != nil {
					return nil, err
				}
				return build.StringLiteral(resBody), nil
			},
		})
	}
	return nil
}

func (hr HTTPResponse) Class() build.Class {
	return hr
}
func (hr HTTPResponse) Value() interface{} {
	return nil
}
func (hr HTTPResponse) Set(key string, obj build.ValueObject) error {
	return build.CannotSetPropertyError(key, hr)
}

var statusCodes = map[string]build.IntegerLiteral{
	"StatusContinue":           build.IntegerLiteral(100), // RFC 9110, 15.2.1
	"StatusSwitchingProtocols": build.IntegerLiteral(101), // RFC 9110, 15.2.2
	"StatusProcessing":         build.IntegerLiteral(102), // RFC 2518, 10.1
	"StatusEarlyHints":         build.IntegerLiteral(103), // RFC 8297

	"StatusOK":                   build.IntegerLiteral(200), // RFC 9110, 15.3.1
	"StatusCreated":              build.IntegerLiteral(201), // RFC 9110, 15.3.2
	"StatusAccepted":             build.IntegerLiteral(202), // RFC 9110, 15.3.3
	"StatusNonAuthoritativeInfo": build.IntegerLiteral(203), // RFC 9110, 15.3.4
	"StatusNoContent":            build.IntegerLiteral(204), // RFC 9110, 15.3.5
	"StatusResetContent":         build.IntegerLiteral(205), // RFC 9110, 15.3.6
	"StatusPartialContent":       build.IntegerLiteral(206), // RFC 9110, 15.3.7
	"StatusMultiStatus":          build.IntegerLiteral(207), // RFC 4918, 11.1
	"StatusAlreadyReported":      build.IntegerLiteral(208), // RFC 5842, 7.1
	"StatusIMUsed":               build.IntegerLiteral(226), // RFC 3229, 10.4.1

	"StatusMultipleChoices":  build.IntegerLiteral(300), // RFC 9110, 15.4.1
	"StatusMovedPermanently": build.IntegerLiteral(301), // RFC 9110, 15.4.2
	"StatusFound":            build.IntegerLiteral(302), // RFC 9110, 15.4.3
	"StatusSeeOther":         build.IntegerLiteral(303), // RFC 9110, 15.4.4
	"StatusNotModified":      build.IntegerLiteral(304), // RFC 9110, 15.4.5
	"StatusUseProxy":         build.IntegerLiteral(305), // RFC 9110, 15.4.6

	"StatusTemporaryRedirect": build.IntegerLiteral(307), // RFC 9110, 15.4.8
	"StatusPermanentRedirect": build.IntegerLiteral(308), // RFC 9110, 15.4.9

	"StatusBadRequest":                   build.IntegerLiteral(400), // RFC 9110, 15.5.1
	"StatusUnauthorized":                 build.IntegerLiteral(401), // RFC 9110, 15.5.2
	"StatusPaymentRequired":              build.IntegerLiteral(402), // RFC 9110, 15.5.3
	"StatusForbidden":                    build.IntegerLiteral(403), // RFC 9110, 15.5.4
	"StatusNotFound":                     build.IntegerLiteral(404), // RFC 9110, 15.5.5
	"StatusMethodNotAllowed":             build.IntegerLiteral(405), // RFC 9110, 15.5.6
	"StatusNotAcceptable":                build.IntegerLiteral(406), // RFC 9110, 15.5.7
	"StatusProxyAuthRequired":            build.IntegerLiteral(407), // RFC 9110, 15.5.8
	"StatusRequestTimeout":               build.IntegerLiteral(408), // RFC 9110, 15.5.9
	"StatusConflict":                     build.IntegerLiteral(409), // RFC 9110, 15.5.10
	"StatusGone":                         build.IntegerLiteral(410), // RFC 9110, 15.5.11
	"StatusLengthRequired":               build.IntegerLiteral(411), // RFC 9110, 15.5.12
	"StatusPreconditionFailed":           build.IntegerLiteral(412), // RFC 9110, 15.5.13
	"StatusRequestEntityTooLarge":        build.IntegerLiteral(413), // RFC 9110, 15.5.14
	"StatusRequestURITooLong":            build.IntegerLiteral(414), // RFC 9110, 15.5.15
	"StatusUnsupportedMediaType":         build.IntegerLiteral(415), // RFC 9110, 15.5.16
	"StatusRequestedRangeNotSatisfiable": build.IntegerLiteral(416), // RFC 9110, 15.5.17
	"StatusExpectationFailed":            build.IntegerLiteral(417), // RFC 9110, 15.5.18
	"StatusTeapot":                       build.IntegerLiteral(418), // RFC 9110, 15.5.19 (Unused)
	"StatusMisdirectedRequest":           build.IntegerLiteral(421), // RFC 9110, 15.5.20
	"StatusUnprocessableEntity":          build.IntegerLiteral(422), // RFC 9110, 15.5.21
	"StatusLocked":                       build.IntegerLiteral(423), // RFC 4918, 11.3
	"StatusFailedDependency":             build.IntegerLiteral(424), // RFC 4918, 11.4
	"StatusTooEarly":                     build.IntegerLiteral(425), // RFC 8470, 5.2.
	"StatusUpgradeRequired":              build.IntegerLiteral(426), // RFC 9110, 15.5.22
	"StatusPreconditionRequired":         build.IntegerLiteral(428), // RFC 6585, 3
	"StatusTooManyRequests":              build.IntegerLiteral(429), // RFC 6585, 4
	"StatusRequestHeaderFieldsTooLarge":  build.IntegerLiteral(431), // RFC 6585, 5
	"StatusUnavailableForLegalReasons":   build.IntegerLiteral(451), // RFC 7725, 3

	"StatusInternalServerError":           build.IntegerLiteral(500), // RFC 9110, 15.6.1
	"StatusNotImplemented":                build.IntegerLiteral(501), // RFC 9110, 15.6.2
	"StatusBadGateway":                    build.IntegerLiteral(502), // RFC 9110, 15.6.3
	"StatusServiceUnavailable":            build.IntegerLiteral(503), // RFC 9110, 15.6.4
	"StatusGatewayTimeout":                build.IntegerLiteral(504), // RFC 9110, 15.6.5
	"StatusHTTPVersionNotSupported":       build.IntegerLiteral(505), // RFC 9110, 15.6.6
	"StatusVariantAlsoNegotiates":         build.IntegerLiteral(506), // RFC 2295, 8.1
	"StatusInsufficientStorage":           build.IntegerLiteral(507), // RFC 4918, 11.5
	"StatusLoopDetected":                  build.IntegerLiteral(508), // RFC 5842, 7.2
	"StatusNotExtended":                   build.IntegerLiteral(510), // RFC 2774, 7
	"StatusNetworkAuthenticationRequired": build.IntegerLiteral(511), // RFC 6585, 6
}
