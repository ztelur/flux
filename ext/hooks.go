package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	hooksPrepare  = make([]flux.PrepareHookFunc, 0, 16)
	hooksStartup  = make([]flux.Startuper, 0, 16)
	hooksShutdown = make([]flux.Shutdowner, 0, 16)
)

// StoreHookFunc 添加生命周期启动与停止的钩子接口
func StoreHookFunc(hook interface{}) {
	pkg.RequireNotNil(hook, "Hook is nil")
	if startup, ok := hook.(flux.Startuper); ok {
		hooksStartup = append(hooksStartup, startup)
	}
	if shutdown, ok := hook.(flux.Shutdowner); ok {
		hooksShutdown = append(hooksShutdown, shutdown)
	}
}

// StorePrepareHook 添加预备阶段钩子函数
func StorePrepareHook(pf flux.PrepareHookFunc) {
	hooksPrepare = append(hooksPrepare, pkg.RequireNotNil(pf, "PrepareHookFunc is nil").(flux.PrepareHookFunc))
}

func LoadPrepareHooks() []flux.PrepareHookFunc {
	dst := make([]flux.PrepareHookFunc, len(hooksPrepare))
	copy(dst, hooksPrepare)
	return dst
}

func LoadStartupHooks() []flux.Startuper {
	dst := make([]flux.Startuper, len(hooksStartup))
	copy(dst, hooksStartup)
	return dst
}

func LoadShutdownHooks() []flux.Shutdowner {
	dst := make([]flux.Shutdowner, len(hooksShutdown))
	copy(dst, hooksShutdown)
	return dst
}
