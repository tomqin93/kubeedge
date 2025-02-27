package core

import (
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"

	"github.com/kubeedge/beehive/pkg/common"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
)

// StartModules starts modules that are registered
func StartModules() {
	beehiveContext.InitContext([]string{common.MsgCtxTypeChannel})

	modules := GetModules()

	for name, module := range modules {
		var m common.ModuleInfo
		switch module.contextType {
		case common.MsgCtxTypeChannel:
			m = common.ModuleInfo{
				ModuleName: name,
				ModuleType: module.contextType,
			}
		case common.MsgCtxTypeUS:
			m = common.ModuleInfo{
				ModuleName: name,
				ModuleType: module.contextType,
				// the below field ModuleSocket is only required for using socket.
				ModuleSocket: common.ModuleSocket{
					IsRemote: module.remote,
				},
			}
		default:
			klog.Fatalf("unsupported context type: %s", module.contextType)
		}

		beehiveContext.AddModule(&m)
		beehiveContext.AddModuleGroup(name, module.module.Group())

		go module.module.Start()
		klog.Infof("starting module %s", name)
	}
}

// GracefulShutdown is if it gets the special signals it does modules cleanup
func GracefulShutdown() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM,
		syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	s := <-c
	klog.Infof("Get os signal %v", s.String())

	// Cleanup each modules
	beehiveContext.Cancel()
	modules := GetModules()
	for name := range modules {
		klog.Infof("Cleanup module %v", name)
		beehiveContext.Cleanup(name)
	}
}

// Run starts the modules and in the end does module cleanup
func Run() {
	// Address the module registration and start the core
	StartModules()
	// monitor system signal and shutdown gracefully
	GracefulShutdown()
}
