package log

//go:generate go run github.com/xtls/xray-core/common/errors/errorgen

import (
	"context"
	"sync"

	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/log"
)

// Instance is a log.Handler that handles logs.
type Instance struct {
	//这里的读写锁不是针对日志文件的，而是针对Instance本身的
	//也就是说控制的是Instance的读写，而不是日志，这一点要搞清楚
	sync.RWMutex
	config       *Config     //日志等级类型配置
	accessLogger log.Handler //访问日志处理实现
	errorLogger  log.Handler //错误日志处理实现
	active       bool
	dns          bool
}

// New creates a new log.Instance based on the given config.
func New(ctx context.Context, config *Config) (*Instance, error) {
	g := &Instance{
		config: config,
		active: false,
		dns:    config.EnableDnsLog,
	}
	log.RegisterHandler(g)

	// start logger instantly on inited
	// other modules would log during init
	if err := g.startInternal(); err != nil {
		return nil, err
	}

	newError("Logger started").AtDebug().WriteToLog()
	return g, nil
}

// 初始化访问日志处理
func (g *Instance) initAccessLogger() error {
	handler, err := createHandler(g.config.AccessLogType, HandlerCreatorOptions{
		Path: g.config.AccessLogPath,
	})
	if err != nil {
		return err
	}
	//这个handler就是common\log\logger.go里面的generalLogger的实例
	g.accessLogger = handler
	return nil
}

// 初始化错误日志处理
func (g *Instance) initErrorLogger() error {
	handler, err := createHandler(g.config.ErrorLogType, HandlerCreatorOptions{
		Path: g.config.ErrorLogPath,
	})
	if err != nil {
		return err
	}
	g.errorLogger = handler
	return nil
}

// Type implements common.HasType.
func (*Instance) Type() interface{} {
	return (*Instance)(nil)
}

func (g *Instance) startInternal() error {
	g.Lock() //这里用的互斥锁，因为会改变g（元素）的值
	defer g.Unlock()

	if g.active {
		return nil
	}

	g.active = true

	if err := g.initAccessLogger(); err != nil {
		return newError("failed to initialize access logger").Base(err).AtWarning()
	}
	if err := g.initErrorLogger(); err != nil {
		return newError("failed to initialize error logger").Base(err).AtWarning()
	}

	return nil
}

// Start implements common.Runnable.Start().
func (g *Instance) Start() error {
	return g.startInternal()
}

// Handle implements log.Handler.
func (g *Instance) Handle(msg log.Message) {
	g.RLock() //这里是读锁，因为这里并没有改变g的值，所以读锁就可以了
	defer g.RUnlock()

	if !g.active {
		return
	}
	//具体日志类型的处理，最后都是通过调用common\log\logger.go里面的generalLogger.Handle()实现的
	switch msg := msg.(type) {
	case *log.AccessMessage:
		if g.accessLogger != nil {
			g.accessLogger.Handle(msg)
		}
	case *log.DNSLog:
		if g.dns && g.accessLogger != nil {
			g.accessLogger.Handle(msg)
		}
	case *log.GeneralMessage:
		if g.errorLogger != nil && msg.Severity <= g.config.ErrorLogLevel {
			g.errorLogger.Handle(msg)
		}
	default:
		// Swallow
	}
}

// Close implements common.Closable.Close().
func (g *Instance) Close() error {
	newError("Logger closing").AtDebug().WriteToLog()

	g.Lock()
	defer g.Unlock()

	if !g.active {
		return nil
	}

	g.active = false

	common.Close(g.accessLogger)
	g.accessLogger = nil

	common.Close(g.errorLogger)
	g.errorLogger = nil

	return nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}
