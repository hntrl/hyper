package stdlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hntrl/hyper/internal/symbols"
)

func sendRequest(method string, path string, config *HTTPRequestConfigValue, body symbols.ValueObject) (*HTTPResponseValue, error) {
	var err error

	base := &url.URL{}
	if config != nil && config.BaseURL != nil {
		base, err = url.Parse(string(*config.BaseURL))
		if err != nil {
			return nil, err
		}
	}
	requestUrl, err := base.Parse(string(path))
	if err != nil {
		return nil, err
	}

	contentType := "application/json"
	client := http.Client{}
	req, err := http.NewRequest(method, requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	if config != nil {
		if config.Headers != nil {
			for k, v := range config.Headers {
				req.Header.Set(k, v)
			}
		}
		if config.Params != nil {
			params := req.URL.Query()
			for k, v := range config.Params {
				params.Set(k, v)
			}
			req.URL.RawQuery = params.Encode()
		}
		if config.Auth != nil {
			if config.Auth.Username != nil {
				if config.Auth.Password != nil {
					req.SetBasicAuth(*config.Auth.Username, *config.Auth.Password)
				} else {
					req.SetBasicAuth(*config.Auth.Username, "")
				}
			}
		}
		if config.ContentType != nil {
			contentType = string(*config.ContentType)
		}
		if config.Timeout != nil {
			client.Timeout = time.Millisecond * time.Duration(int64(*config.Timeout))
		}
	}

	if body != nil {
		if config != nil && config.ContentType != nil {
			contentType = string(*config.ContentType)
		}

		var bodyBytes []byte
		switch contentType {
		case "application/x-www-form-urlencoded":
			data := url.Values{}
			if mapValue, ok := body.(*symbols.MapValue); ok {
				for k, v := range mapValue.Map() {
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
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", contentType)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resHeaders := make(map[string]string)
	for k, v := range res.Header {
		resHeaders[k] = strings.Join(v, ",")
		if err != nil {
			return nil, err
		}
	}
	return &HTTPResponseValue{
		Status:  res.StatusCode,
		Headers: resHeaders,
		body:    res.Body,
	}, nil
}

type RequestPackage struct{}

func (rp RequestPackage) Get(key string) (symbols.ScopeValue, error) {
	functions := map[string]symbols.Callable{
		"get": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String,
				symbols.NewNilableClass(HTTPRequestConfig),
			},
			Returns: HTTPResponse,
			Handler: func(url symbols.StringValue, config *HTTPRequestConfigValue) (symbols.ValueObject, error) {
				return sendRequest(http.MethodGet, string(url), config, nil)
			},
		}),
		"post": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String,
				nil,
				symbols.NewNilableClass(HTTPRequestConfig),
			},
			Returns: HTTPResponse,
			Handler: func(url symbols.StringValue, body symbols.ValueObject, config *HTTPRequestConfigValue) (symbols.ValueObject, error) {
				return sendRequest(http.MethodPost, string(url), config, body)
			},
		}),
		"put": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String,
				nil,
				symbols.NewNilableClass(HTTPRequestConfig),
			},
			Returns: HTTPResponse,
			Handler: func(url symbols.StringValue, body symbols.ValueObject, config *HTTPRequestConfigValue) (symbols.ValueObject, error) {
				return sendRequest(http.MethodPut, string(url), config, body)
			},
		}),
		"delete": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String,
				symbols.NewNilableClass(HTTPRequestConfig),
			},
			Returns: HTTPResponse,
			Handler: func(url symbols.StringValue, config *HTTPRequestConfigValue) (symbols.ValueObject, error) {
				return sendRequest(http.MethodDelete, string(url), config, nil)
			},
		}),
		"options": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String,
				symbols.NewNilableClass(HTTPRequestConfig),
			},
			Returns: HTTPResponse,
			Handler: func(url symbols.StringValue, config *HTTPRequestConfigValue) (symbols.ValueObject, error) {
				return sendRequest(http.MethodOptions, string(url), config, nil)
			},
		}),
		"head": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String,
				symbols.NewNilableClass(HTTPRequestConfig),
			},
			Returns: HTTPResponse,
			Handler: func(url symbols.StringValue, config *HTTPRequestConfigValue) (symbols.ValueObject, error) {
				return sendRequest(http.MethodHead, string(url), config, nil)
			},
		}),
		"patch": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.String,
				nil,
				symbols.NewNilableClass(HTTPRequestConfig),
			},
			Returns: HTTPResponse,
			Handler: func(url symbols.StringValue, body symbols.ValueObject, config *HTTPRequestConfigValue) (symbols.ValueObject, error) {
				return sendRequest(http.MethodPatch, string(url), config, body)
			},
		}),
	}
	if fn, ok := functions[key]; ok {
		return fn, nil
	}
	if status, ok := statusCodes[key]; ok {
		return status, nil
	}
	return nil, nil
}

var (
	HTTPRequestConfig            = HTTPRequestConfigClass{}
	HTTPRequestConfigDescriptors = &symbols.ClassDescriptors{
		Properties: symbols.ClassPropertyMap{
			"headers": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.NewNilableClass(symbols.Map),
				Getter: func(val *HTTPRequestConfigValue) (*symbols.NilableValue, error) {
					if val.Headers != nil {
						castedMapValue := symbols.NewMapValue()
						for k, v := range val.Headers {
							castedMapValue.Set(k, symbols.StringValue(v))
						}
						return symbols.NewNilableValue(castedMapValue.Class(), castedMapValue), nil
					}
					return symbols.NewNilableValue(symbols.Map, nil), nil
				},
			}),
			"params": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.NewNilableClass(symbols.Map),
				Getter: func(val *HTTPRequestConfigValue) (*symbols.NilableValue, error) {
					if val.Params != nil {
						castedMapValue := symbols.NewMapValue()
						for k, v := range val.Params {
							castedMapValue.Set(k, symbols.StringValue(v))
						}
						return symbols.NewNilableValue(castedMapValue.Class(), castedMapValue), nil
					}
					return symbols.NewNilableValue(symbols.Map, nil), nil
				},
			}),
			"auth": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.NewNilableClass(HTTPAuthConfig),
				Getter: func(val *HTTPRequestConfigValue) (*symbols.NilableValue, error) {
					if val.Auth != nil {
						return symbols.NewNilableValue(HTTPAuthConfig, val.Auth), nil
					}
					return symbols.NewNilableValue(HTTPAuthConfig, nil), nil
				},
			}),
			"contentType": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.NewNilableClass(symbols.String),
				Getter: func(val *HTTPRequestConfigValue) (*symbols.NilableValue, error) {
					if val.ContentType != nil {
						return symbols.NewNilableValue(symbols.String, symbols.StringValue(*val.ContentType)), nil
					}
					return symbols.NewNilableValue(symbols.String, nil), nil
				},
			}),
			"timeout": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.NewNilableClass(symbols.Integer),
				Getter: func(val *HTTPRequestConfigValue) (*symbols.NilableValue, error) {
					if val.Timeout != nil {
						return symbols.NewNilableValue(symbols.Integer, symbols.IntegerValue(*val.Timeout)), nil
					}
					return symbols.NewNilableValue(symbols.Integer, nil), nil
				},
			}),
			"baseURL": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.NewNilableClass(symbols.String),
				Getter: func(val *HTTPRequestConfigValue) (*symbols.NilableValue, error) {
					if val.BaseURL != nil {
						return symbols.NewNilableValue(symbols.String, symbols.StringValue(*val.BaseURL)), nil
					}
					return symbols.NewNilableValue(symbols.String, nil), nil
				},
			}),
		},
	}
)

type HTTPRequestConfigClass struct{}

func (HTTPRequestConfigClass) Name() string {
	return "HTTPRequestConfig"
}
func (HTTPRequestConfigClass) Descriptors() *symbols.ClassDescriptors {
	return HTTPRequestConfigDescriptors
}

type HTTPRequestConfigValue struct {
	Headers     map[string]string
	Params      map[string]string
	Auth        *HTTPAuthConfigValue
	ContentType *string
	Timeout     *int
	BaseURL     *string
}

func (*HTTPRequestConfigValue) Class() symbols.Class {
	return HTTPRequestConfig
}
func (*HTTPRequestConfigValue) Value() interface{} {
	return nil
}

var (
	HTTPAuthConfig            = HTTPAuthConfigClass{}
	HTTPAuthConfigDescriptors = &symbols.ClassDescriptors{
		Properties: symbols.ClassPropertyMap{
			"username": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.NewNilableClass(symbols.String),
				Getter: func(cfg *HTTPAuthConfigValue) (*symbols.NilableValue, error) {
					if cfg.Username != nil {
						return symbols.NewNilableValue(symbols.String, symbols.StringValue(*cfg.Username)), nil
					}
					return symbols.NewNilableValue(symbols.String, nil), nil
				},
			}),
			"password": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.NewNilableClass(symbols.String),
				Getter: func(cfg *HTTPAuthConfigValue) (*symbols.NilableValue, error) {
					if cfg.Password != nil {
						return symbols.NewNilableValue(symbols.String, symbols.StringValue(*cfg.Password)), nil
					}
					return symbols.NewNilableValue(symbols.String, nil), nil
				},
			}),
		},
	}
)

type HTTPAuthConfigClass struct{}

func (HTTPAuthConfigClass) Name() string {
	return "HTTPAuthConfig"
}
func (HTTPAuthConfigClass) Descriptors() *symbols.ClassDescriptors {
	return HTTPAuthConfigDescriptors
}

type HTTPAuthConfigValue struct {
	Username *string
	Password *string
}

func (HTTPAuthConfigValue) Class() symbols.Class {
	return HTTPAuthConfig
}
func (HTTPAuthConfigValue) Value() interface{} {
	return nil
}

var (
	HTTPResponse            = HTTPResponseClass{}
	HTTPResponseDescriptors = &symbols.ClassDescriptors{
		Properties: symbols.ClassPropertyMap{
			"status": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.Integer,
				Getter: func(res *HTTPResponseValue) (symbols.IntegerValue, error) {
					return symbols.IntegerValue(res.Status), nil
				},
			}),
			"headers": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: symbols.Map,
				Getter: func(res *HTTPResponseValue) (*symbols.MapValue, error) {
					headerMapValue := symbols.NewMapValue()
					for k, v := range res.Headers {
						headerMapValue.Set(k, symbols.StringValue(v))
					}
					return headerMapValue, nil
				},
			}),
		},
		Prototype: symbols.ClassPrototypeMap{
			"object": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class:     HTTPResponse,
				Arguments: []symbols.Class{},
				Returns:   symbols.Map,
				Handler: func(res *HTTPResponseValue) (symbols.ValueObject, error) {
					resBody, err := io.ReadAll(res.body)
					if err != nil {
						return nil, err
					}
					value, err := symbols.ValueFromBytes(resBody)
					if err != nil {
						return nil, err
					}
					return value, nil
				},
			}),
		},
	}
)

type HTTPResponseClass struct{}

func (HTTPResponseClass) Name() string {
	return "HTTPResponse"
}

func (HTTPResponseClass) Descriptors() *symbols.ClassDescriptors {
	return HTTPResponseDescriptors
}

type HTTPResponseValue struct {
	Status  int
	Headers map[string]string
	body    io.ReadCloser
}

func (HTTPResponseValue) Class() symbols.Class {
	return HTTPResponse
}
func (HTTPResponseValue) Value() interface{} {
	return nil
}

var statusCodes = map[string]symbols.IntegerValue{
	"StatusContinue":           symbols.IntegerValue(100), // RFC 9110, 15.2.1
	"StatusSwitchingProtocols": symbols.IntegerValue(101), // RFC 9110, 15.2.2
	"StatusProcessing":         symbols.IntegerValue(102), // RFC 2518, 10.1
	"StatusEarlyHints":         symbols.IntegerValue(103), // RFC 8297

	"StatusOK":                   symbols.IntegerValue(200), // RFC 9110, 15.3.1
	"StatusCreated":              symbols.IntegerValue(201), // RFC 9110, 15.3.2
	"StatusAccepted":             symbols.IntegerValue(202), // RFC 9110, 15.3.3
	"StatusNonAuthoritativeInfo": symbols.IntegerValue(203), // RFC 9110, 15.3.4
	"StatusNoContent":            symbols.IntegerValue(204), // RFC 9110, 15.3.5
	"StatusResetContent":         symbols.IntegerValue(205), // RFC 9110, 15.3.6
	"StatusPartialContent":       symbols.IntegerValue(206), // RFC 9110, 15.3.7
	"StatusMultiStatus":          symbols.IntegerValue(207), // RFC 4918, 11.1
	"StatusAlreadyReported":      symbols.IntegerValue(208), // RFC 5842, 7.1
	"StatusIMUsed":               symbols.IntegerValue(226), // RFC 3229, 10.4.1

	"StatusMultipleChoices":  symbols.IntegerValue(300), // RFC 9110, 15.4.1
	"StatusMovedPermanently": symbols.IntegerValue(301), // RFC 9110, 15.4.2
	"StatusFound":            symbols.IntegerValue(302), // RFC 9110, 15.4.3
	"StatusSeeOther":         symbols.IntegerValue(303), // RFC 9110, 15.4.4
	"StatusNotModified":      symbols.IntegerValue(304), // RFC 9110, 15.4.5
	"StatusUseProxy":         symbols.IntegerValue(305), // RFC 9110, 15.4.6

	"StatusTemporaryRedirect": symbols.IntegerValue(307), // RFC 9110, 15.4.8
	"StatusPermanentRedirect": symbols.IntegerValue(308), // RFC 9110, 15.4.9

	"StatusBadRequest":                   symbols.IntegerValue(400), // RFC 9110, 15.5.1
	"StatusUnauthorized":                 symbols.IntegerValue(401), // RFC 9110, 15.5.2
	"StatusPaymentRequired":              symbols.IntegerValue(402), // RFC 9110, 15.5.3
	"StatusForbidden":                    symbols.IntegerValue(403), // RFC 9110, 15.5.4
	"StatusNotFound":                     symbols.IntegerValue(404), // RFC 9110, 15.5.5
	"StatusMethodNotAllowed":             symbols.IntegerValue(405), // RFC 9110, 15.5.6
	"StatusNotAcceptable":                symbols.IntegerValue(406), // RFC 9110, 15.5.7
	"StatusProxyAuthRequired":            symbols.IntegerValue(407), // RFC 9110, 15.5.8
	"StatusRequestTimeout":               symbols.IntegerValue(408), // RFC 9110, 15.5.9
	"StatusConflict":                     symbols.IntegerValue(409), // RFC 9110, 15.5.10
	"StatusGone":                         symbols.IntegerValue(410), // RFC 9110, 15.5.11
	"StatusLengthRequired":               symbols.IntegerValue(411), // RFC 9110, 15.5.12
	"StatusPreconditionFailed":           symbols.IntegerValue(412), // RFC 9110, 15.5.13
	"StatusRequestEntityTooLarge":        symbols.IntegerValue(413), // RFC 9110, 15.5.14
	"StatusRequestURITooLong":            symbols.IntegerValue(414), // RFC 9110, 15.5.15
	"StatusUnsupportedMediaType":         symbols.IntegerValue(415), // RFC 9110, 15.5.16
	"StatusRequestedRangeNotSatisfiable": symbols.IntegerValue(416), // RFC 9110, 15.5.17
	"StatusExpectationFailed":            symbols.IntegerValue(417), // RFC 9110, 15.5.18
	"StatusTeapot":                       symbols.IntegerValue(418), // RFC 9110, 15.5.19 (Unused)
	"StatusMisdirectedRequest":           symbols.IntegerValue(421), // RFC 9110, 15.5.20
	"StatusUnprocessableEntity":          symbols.IntegerValue(422), // RFC 9110, 15.5.21
	"StatusLocked":                       symbols.IntegerValue(423), // RFC 4918, 11.3
	"StatusFailedDependency":             symbols.IntegerValue(424), // RFC 4918, 11.4
	"StatusTooEarly":                     symbols.IntegerValue(425), // RFC 8470, 5.2.
	"StatusUpgradeRequired":              symbols.IntegerValue(426), // RFC 9110, 15.5.22
	"StatusPreconditionRequired":         symbols.IntegerValue(428), // RFC 6585, 3
	"StatusTooManyRequests":              symbols.IntegerValue(429), // RFC 6585, 4
	"StatusRequestHeaderFieldsTooLarge":  symbols.IntegerValue(431), // RFC 6585, 5
	"StatusUnavailableForLegalReasons":   symbols.IntegerValue(451), // RFC 7725, 3

	"StatusInternalServerError":           symbols.IntegerValue(500), // RFC 9110, 15.6.1
	"StatusNotImplemented":                symbols.IntegerValue(501), // RFC 9110, 15.6.2
	"StatusBadGateway":                    symbols.IntegerValue(502), // RFC 9110, 15.6.3
	"StatusServiceUnavailable":            symbols.IntegerValue(503), // RFC 9110, 15.6.4
	"StatusGatewayTimeout":                symbols.IntegerValue(504), // RFC 9110, 15.6.5
	"StatusHTTPVersionNotSupported":       symbols.IntegerValue(505), // RFC 9110, 15.6.6
	"StatusVariantAlsoNegotiates":         symbols.IntegerValue(506), // RFC 2295, 8.1
	"StatusInsufficientStorage":           symbols.IntegerValue(507), // RFC 4918, 11.5
	"StatusLoopDetected":                  symbols.IntegerValue(508), // RFC 5842, 7.2
	"StatusNotExtended":                   symbols.IntegerValue(510), // RFC 2774, 7
	"StatusNetworkAuthenticationRequired": symbols.IntegerValue(511), // RFC 6585, 6
}
