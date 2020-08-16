package webx

import (
	"context"
	"github.com/bytepowered/flux"
	"io"
	"net/http"
	"net/url"
)

const (
	charsetUTF8 = "charset=UTF-8"
)

// MIME types
const (
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationForm            = "application/x-www-form-urlencoded"
)

// Headers
const (
	HeaderAccept              = "Accept"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestID          = "X-Request-ID"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXXSSProtection                  = "X-XSS-Protection"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderXCSRFToken                      = "X-CSRF-Token"
	HeaderReferrerPolicy                  = "Referrer-Policy"

	// Ext
	HeaderXRequestId = flux.XRequestId
)

// Web interfaces
type (
	// WebMiddleware defines a function to process middleware.
	WebMiddleware func(WebRouteHandler) WebRouteHandler

	// WebRouteHandler defines a function to serve HTTP requests.
	WebRouteHandler func(WebContext) error

	// WebRouteHandler defines a function to handle errors.
	WebErrorHandler func(err error, ctx WebContext)

	// WebSkipper
	WebSkipper func(ctx WebContext) bool
)

// WebContext defines a context for http server handlers/middleware
type WebContext interface {
	Method() string
	Host() string
	UserAgent() string
	Request() *http.Request
	RequestURI() string
	RequestURLPath() string
	RequestHeader() http.Header
	RequestBody() (io.ReadCloser, error)

	QueryValues() url.Values
	PathValues() url.Values
	FormValues() (url.Values, error)
	CookieValues() []*http.Cookie

	QueryValue(name string) string
	PathValue(name string) string
	FormValue(name string) string
	CookieValue(name string) (*http.Cookie, bool)

	Response() http.ResponseWriter
	ResponseHeader() http.Header
	ResponseWrite(statusCode int, bytes []byte) error

	SetValue(name string, value interface{})
	GetValue(name string) interface{}
}

// WebServerWriter 实现将错误消息和响应数据写入Response实例
type WebServerWriter interface {
	WriteError(webc WebContext, requestId string, header http.Header, error *flux.StateError) error
	WriteBody(webc WebContext, requestId string, header http.Header, status int, body interface{}) error
}

// WebServer
type WebServer interface {
	SetWebErrorHandler(h WebErrorHandler)
	SetRouteNotFoundHandler(h WebRouteHandler)
	AddWebInterceptor(m WebMiddleware)
	AddWebMiddleware(m WebMiddleware)
	AddWebRouteHandler(method, pattern string, h WebRouteHandler, m ...WebMiddleware)
	Start(addr string) error
	StartTLS(addr string, certFile, keyFile string) error
	Shutdown(ctx context.Context) error
}