package filter

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/lakego"
	"github.com/spf13/cast"
	"strings"
	"time"
)

const (
	TypeIdPermissionVerification = "PermissionVerificationFilter"
)

var (
	ErrPermissionSubjectNotFound = &flux.StateError{
		StatusCode: flux.StatusBadRequest,
		ErrorCode:  flux.ErrorCodeRequestInvalid,
		Message:    "PERMISSION:SUBJECT_NOT_FOUND",
	}
	ErrPermissionSubjectAccessDenied = &flux.StateError{
		StatusCode: flux.StatusAccessDenied,
		ErrorCode:  flux.ErrorCodeRequestInvalid,
		Message:    "PERMISSION:SUBJECT_ACCESS_DENIED",
	}
)

// 权限验证函数 检查SubjectId,Method,Patter是否具有权限
type PermissionVerificationFunc func(subjectId, method, pattern string) (bool, error)

func PermissionVerificationFactory() interface{} {
	return NewPermissionVerificationWith(nil)
}

func NewPermissionVerificationWith(provider PermissionVerificationFunc) flux.Filter {
	return &PermissionVerificationFilter{
		provider: provider,
	}
}

// Permission Filter，负责读取JWT的Subject字段，调用指定Dubbo接口判断权限
type PermissionVerificationFilter struct {
	disabled  bool
	provider  PermissionVerificationFunc
	permCache lakego.Cache
}

func (p *PermissionVerificationFilter) Invoke(next flux.FilterInvoker) flux.FilterInvoker {
	if p.disabled {
		return next
	}
	return func(ctx flux.Context) *flux.StateError {
		if false == ctx.Endpoint().Authorize {
			return next(ctx)
		}
		if err := p.doVerification(ctx); nil != err {
			return err
		} else {
			return next(ctx)
		}
	}
}

func (p *PermissionVerificationFilter) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		ConfigKeyDisabled:        false,
		ConfigKeyCacheExpiration: defValueCacheExpiration,
	})
	p.disabled = config.GetBool(ConfigKeyDisabled)
	if p.disabled {
		logger.Infof("Permission filter was DISABLED!!")
		return nil
	}
	logger.Infof("Permission filter initializing")
	if config.IsSet(UpstreamConfigKeyProtocol, UpstreamConfigKeyUri, UpstreamConfigKeyMethod) && p.provider == nil {
		p.provider = func() PermissionVerificationFunc {
			proto := config.GetString(UpstreamConfigKeyProtocol)
			host := config.GetString(UpstreamConfigKeyHost)
			uri := config.GetString(UpstreamConfigKeyUri)
			method := config.GetString(UpstreamConfigKeyMethod)
			logger.Infof("Permission filter config provider, proto:%s, method: %s, uri: %s%s", proto, method, host, host)
			return func(subjectId, method, pattern string) (bool, error) {
				switch strings.ToUpper(proto) {
				case flux.ProtoDubbo:
					return _loadPermByExchange(flux.ProtoDubbo, host, method, uri, subjectId, method, pattern)
				case flux.ProtoHttp:
					return _loadPermByExchange(flux.ProtoHttp, host, method, uri, subjectId, method, pattern)
				default:
					return false, fmt.Errorf("unknown verification protocol: %s", proto)
				}
			}
		}()
	}
	permCacheExpiration := config.GetInt64(ConfigKeyCacheExpiration)
	p.permCache = lakego.NewSimple(lakego.WithExpiration(time.Minute * time.Duration(permCacheExpiration)))
	return nil
}

func (p *PermissionVerificationFilter) Order() int {
	return OrderFilterPermissionVerification
}

func (*PermissionVerificationFilter) TypeId() string {
	return TypeIdPermissionVerification
}

func (p *PermissionVerificationFilter) doVerification(ctx flux.Context) *flux.StateError {
	jwtSubjectId, ok := ctx.GetAttribute(flux.XJwtSubject)
	if !ok {
		return ErrPermissionSubjectNotFound
	}
	endpoint := ctx.Endpoint()
	// 验证用户是否有权限访问API：(userSubId, method, uri-pattern)
	permKey := fmt.Sprintf("%s@%s#%s", jwtSubjectId, endpoint.HttpMethod, endpoint.HttpPattern)
	allowed, err := p.permCache.GetOrLoad(permKey, func(_ lakego.Key) (lakego.Value, error) {
		strSubId := cast.ToString(jwtSubjectId)
		return p.provider(strSubId, endpoint.HttpMethod, endpoint.HttpPattern)
	})
	if err == nil {
		if v, ok := allowed.(bool); ok && v {
			return nil
		} else {
			return ErrPermissionSubjectAccessDenied
		}
	} else {
		return &flux.StateError{
			StatusCode: flux.StatusServerError,
			Message:    "PERMISSION:LOAD_ACCESS",
			Internal:   err,
		}
	}
}

func _loadPermByExchange(proto string, host, method, uri string, reqSubjectId, reqMethod, reqPattern string) (bool, error) {
	exchange, _ := ext.GetExchange(proto)
	if ret, err := exchange.Invoke(&flux.Endpoint{
		UpstreamHost:   host,
		UpstreamMethod: method,
		UpstreamUri:    uri,
		Arguments: []flux.Argument{
			ext.NewStringArgument("subjectId", reqSubjectId),
			ext.NewStringArgument("method", reqMethod),
			ext.NewStringArgument("pattern", reqPattern),
		},
	}, nil); nil != err {
		return false, err
	} else {
		return strings.Contains(cast.ToString(ret), "success"), nil
	}
}
