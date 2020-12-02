package filter

import (
	"errors"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"net/http"
)

const (
	TypeIdPermissionV2Filter = "PermissionFilter"
)

const (
	ErrorCodePermissionDenied = "PERMISSION:ACCESS_DENIED"
)

type (
	// PermissionVerifyReport 权限验证结果报告
	PermissionVerifyReport struct {
		StatusCode int    `json:"statusCode"`
		Success    bool   `json:"success"`
		ErrorCode  string `json:"errorCode"`
		Message    string `json:"message"`
	}
	// PermissionVerifyFunc 权限验证
	// @return pass 对当前请求的权限验证是否通过；
	// @return err 如果验证过程发生错误，返回error；
	PermissionVerifyFunc func(services []flux.BackendService, ctx flux.Context) (report PermissionVerifyReport, err error)
)

// PermissionConfig 权限配置
type PermissionConfig struct {
	SkipFunc   flux.FilterSkipper
	VerifyFunc PermissionVerifyFunc
}

func NewPermissionVerifyReport(success bool, errorCode, message string) PermissionVerifyReport {
	return PermissionVerifyReport{
		StatusCode: flux.StatusUnauthorized,
		Success:    success,
		ErrorCode:  errorCode,
		Message:    message,
	}
}

func NewPermissionFilter(c PermissionConfig) *PermissionFilter {
	return &PermissionFilter{
		Configs: c,
	}
}

// PermissionFilter 提供基于Endpoint.Permission元数据的权限验证
type PermissionFilter struct {
	Disabled bool
	Configs  PermissionConfig
}

func (p *PermissionFilter) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		ConfigKeyDisabled: false,
	})
	p.Disabled = config.GetBool(ConfigKeyDisabled)
	if p.Disabled {
		logger.Info("Endpoint PermissionFilter was DISABLED!!")
		return nil
	}
	if pkg.IsNil(p.Configs.SkipFunc) {
		p.Configs.SkipFunc = func(_ flux.Context) bool {
			return false
		}
	}
	if pkg.IsNil(p.Configs.VerifyFunc) {
		return fmt.Errorf("PermissionFilter.VerifyFunc is nil")
	}
	return nil
}

func (*PermissionFilter) TypeId() string {
	return TypeIdPermissionV2Filter
}

func (p *PermissionFilter) DoFilter(next flux.FilterHandler) flux.FilterHandler {
	if p.Disabled {
		return next
	}
	return func(ctx flux.Context) *flux.StateError {
		if p.Configs.SkipFunc(ctx) {
			return next(ctx)
		}
		// 必须开启Authorize才进行权限校验
		endpoint := ctx.Endpoint()
		if false == endpoint.Authorize {
			return next(ctx)
		}
		// 没有任何权限校验定义
		size := len(endpoint.Permissions)
		if size == 0 && !endpoint.Permission.IsValid() {
			return next(ctx)
		}
		services := make([]flux.BackendService, 0, 1+size)
		// Define permission first
		if endpoint.Permission.IsValid() {
			services = append(services, endpoint.Permission)
		}
		for _, id := range endpoint.Permissions {
			if srv, ok := ext.LoadBackendService(id); ok {
				services = append(services, srv)
			} else {
				return &flux.StateError{
					StatusCode: flux.StatusServerError,
					ErrorCode:  flux.ErrorCodeGatewayInternal,
					Message:    "PERMISSION:SERVICE:NOT_FOUND",
					Internal:   errors.New("service not found, id: " + id),
				}
			}
		}
		report, err := p.Configs.VerifyFunc(services, ctx)
		if nil != err {
			if serr, ok := err.(*flux.StateError); ok {
				return serr
			}
			return &flux.StateError{
				StatusCode: http.StatusForbidden,
				ErrorCode:  flux.ErrorCodeGatewayInternal,
				Message:    "PERMISSION:VERIFY:ERROR",
				Internal:   err,
			}
		}
		if !report.Success {
			return &flux.StateError{
				StatusCode: EnsurePermissionStatusCode(report.StatusCode),
				ErrorCode:  EnsurePermissionErrorCode(report.ErrorCode),
				Message:    EnsurePermissionMessage(report.Message),
			}
		}
		return next(ctx)
	}
}

// Invoke 执行权限验证的后端服务，获取响应结果；
func (p *PermissionFilter) Invoke(service flux.BackendService, ctx flux.Context) (interface{}, *flux.StateError) {
	return backend.DoInvoke(service, ctx)
}

func EnsurePermissionStatusCode(status int) int {
	if status < 100 {
		return http.StatusForbidden
	}
	return status
}

func EnsurePermissionErrorCode(code string) string {
	if "" == code {
		return ErrorCodePermissionDenied
	}
	return code
}

func EnsurePermissionMessage(message string) string {
	if "" == message {
		return ErrorCodePermissionDenied
	}
	return message
}
