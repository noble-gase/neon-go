package cors

import "slices"

type Option func(c *Cors)

// ACAO = Access-Control-Allow-Origin
func ACAO(origins ...string) Option {
	return func(c *Cors) {
		c.allowOrigins = slices.Clone(origins)
	}
}

// ACAM = Access-Control-Allow-Methods
func ACAM(methods ...string) Option {
	return func(c *Cors) {
		c.allowMethods = slices.Clone(methods)
	}
}

// ACAH = Access-Control-Allow-Headers
func ACAH(headers ...string) Option {
	return func(c *Cors) {
		c.allowHeaders = slices.Clone(headers)
	}
}

// ACEH = Access-Control-Expose-Headers 服务器暴露一些自定义的头信息，允许客户端访问
func ACEH(headers ...string) Option {
	return func(c *Cors) {
		c.exposeHeaders = slices.Clone(headers)
	}
}

// ACAC = Access-Control-Allow-Credentials
func ACAC(allow bool) Option {
	return func(c *Cors) {
		c.allowCredentials = allow
	}
}

// ACMA = Access-Control-Max-Age
func ACMA(age int) Option {
	return func(c *Cors) {
		c.maxAge = age
	}
}
