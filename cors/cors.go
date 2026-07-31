package cors

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
)

const wildcard = "*"

const (
	HeaderOrigin = "Origin"
	HeaderVary   = "Vary"
	HeaderACRM   = "Access-Control-Request-Method"
	HeaderACRH   = "Access-Control-Request-Headers"
	HeaderACAO   = "Access-Control-Allow-Origin"
	HeaderACAM   = "Access-Control-Allow-Methods"
	HeaderACAH   = "Access-Control-Allow-Headers"
	HeaderACEH   = "Access-Control-Expose-Headers"
	HeaderACAC   = "Access-Control-Allow-Credentials"
	HeaderACMA   = "Access-Control-Max-Age"
)

type Cors struct {
	allowOrigins     []string
	allowMethods     []string
	allowHeaders     []string
	exposeHeaders    []string
	allowCredentials bool
	maxAge           int
}

func (c *Cors) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 仅拦截真正的 CORS 预检，放行普通 OPTIONS
		if r.Method == http.MethodOptions && len(r.Header.Get(HeaderACRM)) != 0 {
			c.handlePreflight(w, r)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// 处理实际的跨域请求
		c.handleRequest(w, r)
		h.ServeHTTP(w, r)
	})
}

// handlePreflight 处理 CORS 预检请求
func (c *Cors) handlePreflight(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()

	headers.Add(HeaderVary, HeaderOrigin)
	headers.Add(HeaderVary, HeaderACRM)
	headers.Add(HeaderVary, HeaderACRH)

	origin := r.Header.Get(HeaderOrigin)
	if len(origin) == 0 || !c.isOriginAllowed(origin) {
		return
	}

	// 校验预检请求的方法是否被允许
	reqMethod := r.Header.Get(HeaderACRM)
	if reqMethod != http.MethodOptions && !slices.Contains(c.allowMethods, reqMethod) {
		return
	}

	// Access-Control-Allow-Origin
	if slices.Contains(c.allowOrigins, wildcard) {
		headers.Set(HeaderACAO, wildcard)
	} else {
		headers.Set(HeaderACAO, origin)
	}

	// Access-Control-Allow-Methods（回显请求方法）
	headers.Set(HeaderACAM, reqMethod)

	// Access-Control-Allow-Headers
	if reqHeaders := r.Header.Get(HeaderACRH); len(reqHeaders) != 0 {
		if slices.Contains(c.allowHeaders, wildcard) {
			headers.Set(HeaderACAH, reqHeaders)
		} else {
			headers.Set(HeaderACAH, strings.Join(c.allowHeaders, ", "))
		}
	}

	// Access-Control-Allow-Credentials
	if c.allowCredentials {
		headers.Set(HeaderACAC, "true")
	}

	// Access-Control-Max-Age
	if c.maxAge > 0 {
		headers.Set(HeaderACMA, strconv.Itoa(c.maxAge))
	}
}

// handleRequest 处理实际的跨域请求
func (c *Cors) handleRequest(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()

	headers.Add(HeaderVary, HeaderOrigin)

	origin := r.Header.Get(HeaderOrigin)
	if len(origin) == 0 || !c.isOriginAllowed(origin) {
		return
	}

	// Access-Control-Allow-Origin
	if slices.Contains(c.allowOrigins, wildcard) {
		headers.Set(HeaderACAO, wildcard)
	} else {
		headers.Set(HeaderACAO, origin)
	}

	// Access-Control-Expose-Headers
	if len(c.exposeHeaders) != 0 {
		headers.Set(HeaderACEH, strings.Join(c.exposeHeaders, ", "))
	}

	// Access-Control-Allow-Credentials
	if c.allowCredentials {
		headers.Set(HeaderACAC, "true")
	}
}

func (c *Cors) isOriginAllowed(origin string) bool {
	return slices.Contains(c.allowOrigins, wildcard) || slices.Contains(c.allowOrigins, origin)
}

// New 创建一个 CORS 中间件，默认允许所有跨域请求
func New(opts ...Option) *Cors {
	c := &Cors{
		allowOrigins: []string{wildcard},
		allowMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		allowHeaders: []string{wildcard},
	}
	for _, f := range opts {
		f(c)
	}
	return c
}
