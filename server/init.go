package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/registry"
	"github.com/bytepowered/flux/support"
)

func init() {
	// Default logger factory
	ext.StoreLoggerFactory(DefaultLoggerFactory)
	// 参数查找与解析函数
	ext.StoreArgumentLookupFunc(support.DefaultArgumentValueLookupFunc)
	// Serializer
	// Default: JSON
	serializer := flux.NewJsonSerializer()
	ext.StoreSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.StoreSerializer(ext.TypeNameSerializerJson, serializer)
	// Endpoint registry
	// Default: ZK
	ext.StoreEndpointRegistryFactory(ext.EndpointRegistryProtoDefault, registry.DefaultRegistryFactory)
	ext.StoreEndpointRegistryFactory(ext.EndpointRegistryProtoZookeeper, registry.DefaultRegistryFactory)
	// Server
	SetServerWriterSerializer(serializer)
	SetServerResponseContentType(flux.MIMEApplicationJSONCharsetUTF8)
}
