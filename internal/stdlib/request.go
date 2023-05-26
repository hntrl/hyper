package stdlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hntrl/hyper/internal/symbols"
)

type RequestPackage struct{}

func sendRequest(method string, path string, config RequestConfig, body symbols.ValueObject) (*HTTPResponse, error) {
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

	req, err := http.NewRequest(method, requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	if body != nil {
		contentType := "application/json"
		if config.ContentType != nil {
			contentType = string(*config.ContentType)
		}

		var bodyBytes []byte
		switch contentType {
		case "application/x-www-form-urlencoded":
			data := url.Values{}
			if generic, ok := body.(*symbols.MapObject); ok {
				for k, v := range generic.Data {
					data.Set(k, fmt.Sprintf("%v", v.Value()))
				}
			}
			bodyBytes = []byte(data.Encode())
		default:
			bodyBytes, err = json.Marshal(body.Value())
			if err != nil {
				return nil, err
			}
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", contentType)
	}

	if config.Auth != nil {
		if config.Auth.Username != nil {
			if config.Auth.Password != nil {
				req.SetBasicAuth(string(*config.Auth.Username), string(*config.Auth.Password))
			} else {
				req.SetBasicAuth(string(*config.Auth.Username), "")
			}
		}
	}

	if config.Headers != nil {
		for k := range config.Headers.Fields() {
			val, err := config.Headers.Get(k)
			if err != nil {
				return nil, err
			}
			if valueObj, ok := val.(symbols.ValueObject); ok {
				req.Header.Set(k, fmt.Sprintf("%v", valueObj.Value()))
			}
		}
	}

	if config.Params != nil {
		params := req.URL.Query()
		for k := range config.Params.Fields() {
			val, err := config.Params.Get(k)
			if err != nil {
				return nil, err
			}
			if valueObj, ok := val.(symbols.ValueObject); ok {
				params.Set(k, fmt.Sprintf("%v", valueObj.Value()))
			}
		}
		req.URL.RawQuery = params.Encode()
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resHeaders := symbols.NewMapObject()
	for k, v := range res.Header {
		err := resHeaders.Set(k, symbols.StringLiteral(strings.Join(v, ",")))
		if err != nil {
			return nil, err
		}
	}
	return &HTTPResponse{
		Status:  symbols.IntegerLiteral(res.StatusCode),
		Headers: resHeaders,
		body:    res.Body,
	}, nil
}

func (rp RequestPackage) Get(key string) (symbols.Object, error) {
	methods := map[string]symbols.Object{
		"get": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
				symbols.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				url := args[0].(symbols.StringLiteral)
				config := args[1].(*symbols.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodGet, string(url), config, nil)
			},
		}),
		"post": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
				symbols.MapObject{},
				symbols.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				url := args[0].(symbols.StringLiteral)
				config := args[2].(*symbols.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodPost, string(url), config, args[1])
			},
		}),
		"put": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
				symbols.MapObject{},
				symbols.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				url := args[0].(symbols.StringLiteral)
				config := args[2].(*symbols.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodPut, string(url), config, args[1])
			},
		}),
		"delete": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
				symbols.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				url := args[0].(symbols.StringLiteral)
				config := args[2].(*symbols.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodDelete, string(url), config, nil)
			},
		}),
		"options": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
				symbols.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				url := args[0].(symbols.StringLiteral)
				config := args[2].(*symbols.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodOptions, string(url), config, nil)
			},
		}),
		"head": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
				symbols.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				url := args[0].(symbols.StringLiteral)
				config := args[2].(*symbols.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodHead, string(url), config, nil)
			},
		}),
		"patch": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String{},
				symbols.MapObject{},
				symbols.NewOptionalClass(RequestConfig{}),
			},
			Returns: HTTPResponse{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				url := args[0].(symbols.StringLiteral)
				config := args[2].(*symbols.NilableObject).Object.(RequestConfig)
				return sendRequest(http.MethodPatch, string(url), config, args[1])
			},
		}),
	}
	if fn, ok := methods[key]; ok {
		return fn, nil
	}
	if status, ok := statusCodes[key]; ok {
		return status, nil
	}
	return nil, nil
}

type RequestConfig struct {
	Headers     *symbols.MapObject      `hash:"ignore"`
	Auth        *AuthConfig             `hash:"ignore"`
	Params      *symbols.MapObject      `hash:"ignore"`
	ContentType *symbols.StringLiteral  `hash:"ignore"`
	Timeout     *symbols.IntegerLiteral `hash:"ignore"`
	BaseURL     *symbols.StringLiteral  `hash:"ignore"`
}

func (rc RequestConfig) ClassName() string {
	return "RequestConfig"
}
func (rc RequestConfig) Fields() map[string]symbols.Class {
	return map[string]symbols.Class{
		"headers":     symbols.NewOptionalClass(symbols.MapObject{}),
		"auth":        symbols.NewOptionalClass(AuthConfig{}),
		"params":      symbols.NewOptionalClass(symbols.MapObject{}),
		"contentType": symbols.NewOptionalClass(symbols.String{}),
		"data":        symbols.NewOptionalClass(symbols.MapObject{}),
		"timeout":     symbols.NewOptionalClass(symbols.Integer{}),
		"baseURL":     symbols.NewOptionalClass(symbols.String{}),
	}
}
func (rc RequestConfig) Constructors() symbols.ConstructorMap {
	csMap := symbols.NewConstructorMap()
	csMap.AddGenericConstructor(rc, func(fields map[string]symbols.ValueObject) (symbols.ValueObject, error) {
		config := RequestConfig{}
		if headers, ok := fields["headers"]; ok {
			obj := headers.(*symbols.NilableObject).Object
			if obj != nil {
				lit := obj.(*symbols.MapObject)
				config.Headers = lit
			}
		}
		if auth, ok := fields["auth"]; ok {
			obj := auth.(*symbols.NilableObject).Object
			if obj != nil {
				lit := obj.(AuthConfig)
				config.Auth = &lit
			}
		}
		if params, ok := fields["params"]; ok {
			obj := params.(*symbols.NilableObject).Object
			if obj != nil {
				lit := obj.(*symbols.MapObject)
				config.Headers = lit
			}
		}
		if contentType, ok := fields["contentType"]; ok {
			obj := contentType.(*symbols.NilableObject).Object
			if obj != nil {
				lit := obj.(symbols.StringLiteral)
				config.ContentType = &lit
			}
		}
		if timeout, ok := fields["timeout"]; ok {
			obj := timeout.(*symbols.NilableObject).Object
			if obj != nil {
				lit := obj.(symbols.IntegerLiteral)
				config.Timeout = &lit
			}
		}
		if baseURL, ok := fields["baseURL"]; ok {
			obj := baseURL.(*symbols.NilableObject).Object
			if obj != nil {
				lit := obj.(symbols.StringLiteral)
				config.BaseURL = &lit
			}
		}
		return config, nil
	})
	return csMap
}
func (rc RequestConfig) Get(key string) (symbols.Object, error) {
	switch key {
	case "headers":
		return rc.Headers, nil
	case "auth":
		return rc.Auth, nil
	case "params":
		return rc.Params, nil
	case "contentType":
		return rc.ContentType, nil
	case "timeout":
		return rc.Timeout, nil
	case "baseURL":
		return rc.BaseURL, nil
	}
	return nil, nil
}

func (rc RequestConfig) Class() symbols.Class {
	return rc
}
func (rc RequestConfig) Value() interface{} {
	return nil
}
func (rc RequestConfig) Set(key string, obj symbols.ValueObject) error {
	return nil
}

type AuthConfig struct {
	Username *symbols.StringLiteral
	Password *symbols.StringLiteral
}

func (ac AuthConfig) ClassName() string {
	return "RequestAuthConfig"
}
func (ac AuthConfig) Fields() map[string]symbols.Class {
	return map[string]symbols.Class{
		"username": symbols.NewOptionalClass(symbols.String{}),
		"password": symbols.NewOptionalClass(symbols.String{}),
	}
}
func (ac AuthConfig) Constructors() symbols.ConstructorMap {
	csMap := symbols.NewConstructorMap()
	csMap.AddGenericConstructor(ac, func(fields map[string]symbols.ValueObject) (symbols.ValueObject, error) {
		config := AuthConfig{}
		if username, ok := fields["username"]; ok {
			obj := username.(*symbols.NilableObject).Object
			if obj != nil {
				lit := obj.(symbols.StringLiteral)
				config.Username = &lit
			}
		}
		if params, ok := fields["password"]; ok {
			obj := params.(*symbols.NilableObject).Object
			if obj != nil {
				lit := obj.(symbols.StringLiteral)
				config.Password = &lit
			}
		}
		return config, nil
	})
	return csMap
}
func (ac AuthConfig) Get(key string) (symbols.Object, error) {
	switch key {
	case "username":
		return ac.Username, nil
	case "password":
		return ac.Password, nil
	}
	return nil, nil
}

func (ac AuthConfig) Class() symbols.Class {
	return ac
}
func (ac AuthConfig) Value() interface{} {
	return nil
}
func (ac AuthConfig) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, ac)
}

type HTTPResponse struct {
	Status  symbols.IntegerLiteral `hash:"ignore"`
	Headers *symbols.MapObject     `hash:"ignore"`
	body    io.ReadCloser          `hash:"ignore"`
}

func (hr HTTPResponse) ClassName() string {
	return "HTTPResponse"
}
func (hr HTTPResponse) Fields() map[string]symbols.Class {
	return map[string]symbols.Class{
		"status":  symbols.Integer{},
		"headers": symbols.MapObject{},
	}
}
func (hr HTTPResponse) Constructors() symbols.ConstructorMap {
	return symbols.NewConstructorMap()
}
func (hr HTTPResponse) Get(key string) (symbols.Object, error) {
	switch key {
	case "status":
		return hr.Status, nil
	case "headers":
		return hr.Headers, nil
	case "object":
		return symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{},
			Returns:   &symbols.MapObject{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				resBody, err := io.ReadAll(hr.body)
				if err != nil {
					return nil, err
				}
				generic, err := symbols.FromBytes(resBody)
				if err != nil {
					return nil, err
				}
				return generic, nil
			},
		}), nil
	case "text":
		return symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{},
			Returns:   symbols.String{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				resBody, err := io.ReadAll(hr.body)
				if err != nil {
					return nil, err
				}
				return symbols.StringLiteral(resBody), nil
			},
		}), nil
	}
	return nil, nil
}

func (hr HTTPResponse) Class() symbols.Class {
	return hr
}
func (hr HTTPResponse) Value() interface{} {
	return nil
}
func (hr HTTPResponse) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, hr)
}

var statusCodes = map[string]symbols.IntegerLiteral{
	"StatusContinue":           symbols.IntegerLiteral(100), // RFC 9110, 15.2.1
	"StatusSwitchingProtocols": symbols.IntegerLiteral(101), // RFC 9110, 15.2.2
	"StatusProcessing":         symbols.IntegerLiteral(102), // RFC 2518, 10.1
	"StatusEarlyHints":         symbols.IntegerLiteral(103), // RFC 8297

	"StatusOK":                   symbols.IntegerLiteral(200), // RFC 9110, 15.3.1
	"StatusCreated":              symbols.IntegerLiteral(201), // RFC 9110, 15.3.2
	"StatusAccepted":             symbols.IntegerLiteral(202), // RFC 9110, 15.3.3
	"StatusNonAuthoritativeInfo": symbols.IntegerLiteral(203), // RFC 9110, 15.3.4
	"StatusNoContent":            symbols.IntegerLiteral(204), // RFC 9110, 15.3.5
	"StatusResetContent":         symbols.IntegerLiteral(205), // RFC 9110, 15.3.6
	"StatusPartialContent":       symbols.IntegerLiteral(206), // RFC 9110, 15.3.7
	"StatusMultiStatus":          symbols.IntegerLiteral(207), // RFC 4918, 11.1
	"StatusAlreadyReported":      symbols.IntegerLiteral(208), // RFC 5842, 7.1
	"StatusIMUsed":               symbols.IntegerLiteral(226), // RFC 3229, 10.4.1

	"StatusMultipleChoices":  symbols.IntegerLiteral(300), // RFC 9110, 15.4.1
	"StatusMovedPermanently": symbols.IntegerLiteral(301), // RFC 9110, 15.4.2
	"StatusFound":            symbols.IntegerLiteral(302), // RFC 9110, 15.4.3
	"StatusSeeOther":         symbols.IntegerLiteral(303), // RFC 9110, 15.4.4
	"StatusNotModified":      symbols.IntegerLiteral(304), // RFC 9110, 15.4.5
	"StatusUseProxy":         symbols.IntegerLiteral(305), // RFC 9110, 15.4.6

	"StatusTemporaryRedirect": symbols.IntegerLiteral(307), // RFC 9110, 15.4.8
	"StatusPermanentRedirect": symbols.IntegerLiteral(308), // RFC 9110, 15.4.9

	"StatusBadRequest":                   symbols.IntegerLiteral(400), // RFC 9110, 15.5.1
	"StatusUnauthorized":                 symbols.IntegerLiteral(401), // RFC 9110, 15.5.2
	"StatusPaymentRequired":              symbols.IntegerLiteral(402), // RFC 9110, 15.5.3
	"StatusForbidden":                    symbols.IntegerLiteral(403), // RFC 9110, 15.5.4
	"StatusNotFound":                     symbols.IntegerLiteral(404), // RFC 9110, 15.5.5
	"StatusMethodNotAllowed":             symbols.IntegerLiteral(405), // RFC 9110, 15.5.6
	"StatusNotAcceptable":                symbols.IntegerLiteral(406), // RFC 9110, 15.5.7
	"StatusProxyAuthRequired":            symbols.IntegerLiteral(407), // RFC 9110, 15.5.8
	"StatusRequestTimeout":               symbols.IntegerLiteral(408), // RFC 9110, 15.5.9
	"StatusConflict":                     symbols.IntegerLiteral(409), // RFC 9110, 15.5.10
	"StatusGone":                         symbols.IntegerLiteral(410), // RFC 9110, 15.5.11
	"StatusLengthRequired":               symbols.IntegerLiteral(411), // RFC 9110, 15.5.12
	"StatusPreconditionFailed":           symbols.IntegerLiteral(412), // RFC 9110, 15.5.13
	"StatusRequestEntityTooLarge":        symbols.IntegerLiteral(413), // RFC 9110, 15.5.14
	"StatusRequestURITooLong":            symbols.IntegerLiteral(414), // RFC 9110, 15.5.15
	"StatusUnsupportedMediaType":         symbols.IntegerLiteral(415), // RFC 9110, 15.5.16
	"StatusRequestedRangeNotSatisfiable": symbols.IntegerLiteral(416), // RFC 9110, 15.5.17
	"StatusExpectationFailed":            symbols.IntegerLiteral(417), // RFC 9110, 15.5.18
	"StatusTeapot":                       symbols.IntegerLiteral(418), // RFC 9110, 15.5.19 (Unused)
	"StatusMisdirectedRequest":           symbols.IntegerLiteral(421), // RFC 9110, 15.5.20
	"StatusUnprocessableEntity":          symbols.IntegerLiteral(422), // RFC 9110, 15.5.21
	"StatusLocked":                       symbols.IntegerLiteral(423), // RFC 4918, 11.3
	"StatusFailedDependency":             symbols.IntegerLiteral(424), // RFC 4918, 11.4
	"StatusTooEarly":                     symbols.IntegerLiteral(425), // RFC 8470, 5.2.
	"StatusUpgradeRequired":              symbols.IntegerLiteral(426), // RFC 9110, 15.5.22
	"StatusPreconditionRequired":         symbols.IntegerLiteral(428), // RFC 6585, 3
	"StatusTooManyRequests":              symbols.IntegerLiteral(429), // RFC 6585, 4
	"StatusRequestHeaderFieldsTooLarge":  symbols.IntegerLiteral(431), // RFC 6585, 5
	"StatusUnavailableForLegalReasons":   symbols.IntegerLiteral(451), // RFC 7725, 3

	"StatusInternalServerError":           symbols.IntegerLiteral(500), // RFC 9110, 15.6.1
	"StatusNotImplemented":                symbols.IntegerLiteral(501), // RFC 9110, 15.6.2
	"StatusBadGateway":                    symbols.IntegerLiteral(502), // RFC 9110, 15.6.3
	"StatusServiceUnavailable":            symbols.IntegerLiteral(503), // RFC 9110, 15.6.4
	"StatusGatewayTimeout":                symbols.IntegerLiteral(504), // RFC 9110, 15.6.5
	"StatusHTTPVersionNotSupported":       symbols.IntegerLiteral(505), // RFC 9110, 15.6.6
	"StatusVariantAlsoNegotiates":         symbols.IntegerLiteral(506), // RFC 2295, 8.1
	"StatusInsufficientStorage":           symbols.IntegerLiteral(507), // RFC 4918, 11.5
	"StatusLoopDetected":                  symbols.IntegerLiteral(508), // RFC 5842, 7.2
	"StatusNotExtended":                   symbols.IntegerLiteral(510), // RFC 2774, 7
	"StatusNetworkAuthenticationRequired": symbols.IntegerLiteral(511), // RFC 6585, 6
}
