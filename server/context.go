package server

import (
	"context"
	"github.com/bytepowered/flux"
	"sync"
	"time"
)

var _ flux.Context = new(WrappedContext)

// Context接口实现
type WrappedContext struct {
	requestId      string
	webc           flux.WebContext
	endpoint       *flux.Endpoint
	attachments    *sync.Map
	values         *sync.Map
	requestReader  *WrappedRequestReader
	responseWriter *WrappedResponseWriter
	ctxLogger      flux.Logger
}

func NewContextWrapper() interface{} {
	return &WrappedContext{
		responseWriter: newResponseWrappedWriter(),
		requestReader:  newRequestWrappedReader(),
	}
}

func (c *WrappedContext) Request() flux.RequestReader {
	return c.requestReader
}

func (c *WrappedContext) Response() flux.ResponseWriter {
	return c.responseWriter
}

func (c *WrappedContext) Endpoint() flux.Endpoint {
	return *(c.endpoint)
}

func (c *WrappedContext) Upstream() (proto, host, uri, method string) {
	return c.endpoint.UpstreamProto, c.endpoint.UpstreamHost, c.endpoint.UpstreamUri, c.endpoint.UpstreamMethod
}

func (c *WrappedContext) UpstreamProto() string {
	return c.endpoint.UpstreamProto
}

func (c *WrappedContext) UpstreamService() (uri, method string) {
	return c.endpoint.UpstreamUri, c.endpoint.UpstreamMethod
}

func (c *WrappedContext) Authorize() bool {
	return c.endpoint.Authorize
}

func (c *WrappedContext) Method() string {
	return c.webc.Method()
}

func (c *WrappedContext) RequestURI() string {
	return c.webc.RequestURI()
}

func (c *WrappedContext) RequestId() string {
	return c.requestId
}

func (c *WrappedContext) Attributes() map[string]interface{} {
	copied := make(map[string]interface{})
	c.attachments.Range(func(key, value interface{}) bool {
		copied[key.(string)] = value
		return true
	})
	return copied
}

func (c *WrappedContext) SetAttribute(name string, value interface{}) {
	c.attachments.Store(name, value)
}

func (c *WrappedContext) GetAttribute(name string) (interface{}, bool) {
	v, ok := c.attachments.Load(name)
	return v, ok
}

func (c *WrappedContext) SetValue(name string, value interface{}) {
	c.values.Store(name, value)
}

func (c *WrappedContext) GetValue(name string) (interface{}, bool) {
	// first: Local values
	// then: WebContext values
	if lv, ok := c.values.Load(name); ok {
		return lv, true
	} else if cv := c.webc.GetValue(name); nil != cv {
		return cv, true
	} else {
		return nil, false
	}
}

func (c *WrappedContext) Context() context.Context {
	return c.webc.Context()
}

func (c *WrappedContext) SetContextLogger(logger flux.Logger) {
	c.ctxLogger = logger
}

func (c *WrappedContext) GetContextLogger() (flux.Logger, bool) {
	return c.ctxLogger, nil != c.ctxLogger
}

func (c *WrappedContext) Reattach(requestId string, webc flux.WebContext, endpoint *flux.Endpoint) {
	c.requestId = requestId
	c.webc = webc
	c.endpoint = endpoint
	c.attachments = new(sync.Map)
	c.values = new(sync.Map)
	c.requestReader.reattach(webc)
	// duplicated: c.responseWriter.reset()
	c.SetAttribute(flux.XRequestTime, time.Now().Unix())
	c.SetAttribute(flux.XRequestId, c.requestId)
	c.SetAttribute(flux.XRequestHost, webc.Host())
	c.SetAttribute(flux.XRequestAgent, "flux/gateway")
}

func (c *WrappedContext) Release() {
	c.requestId = ""
	c.webc = nil
	c.endpoint = nil
	c.attachments = nil
	c.values = nil
	c.requestReader.reset()
	c.responseWriter.reset()
	c.ctxLogger = nil
}
