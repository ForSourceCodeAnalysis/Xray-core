# 系统架构
![架构图](framework-fWbTmHWQ.png)
内核分为三层：应用层、代理层和传输层  
*具体的可以参考[官方文档](https://xtls.github.io/development/intro/design.html)*
# 代码流程
通过命令启动各种服务来实现各种功能  
## 配置加载  
配置文档中的配置项定义在[这里](infra\conf\xray.go)，程序通过[Build](infra\conf\xray.go)将面向用户的配置结构转换为[程序中的配置](core\config.pb.go)结构。其实就是创建各配置/模块/服务的结构体实例。
## `core.Config`结构分析  
这个就是内核配置了，相对于配置文件中的配置，字段少了很多，当前版本在用的如下
```
    // Inbound handler configurations. Must have at least one item.
	Inbound []*InboundHandlerConfig `protobuf:"bytes,1,rep,name=inbound,proto3" json:"inbound,omitempty"`
	// Outbound handler configurations. Must have at least one item. The first
	// item is used as default for routing.
	Outbound []*OutboundHandlerConfig `protobuf:"bytes,2,rep,name=outbound,proto3" json:"outbound,omitempty"`
	// App is for configurations of all features in Xray. A feature must
	// implement the Feature interface, and its config type must be registered
	// through common.RegisterConfig.
	App []*serial.TypedMessage `protobuf:"bytes,4,rep,name=app,proto3" json:"app,omitempty"`
	
	// Configuration for extensions. The config may not work if corresponding
	// extension is not loaded into Xray. Xray will ignore such config during
	// initialization.
	Extension []*serial.TypedMessage `protobuf:"bytes,6,rep,name=extension,proto3" json:"extension,omitempty"`
```
本质和用户配置没有太大区别，只是把很多模块移到了`App`字段中。配置加载的时候，会通过调用各模块的`Build()`方法实例化各模块，然后加入`App`中。  
## 服务初始化及启动
配置加载完成后，根据`core.Config`实例化[`core.Instance`](core\xray.go)，在此过程中，程序会将`App`字段里面的配置转换成`core.Instance`里面的`Features`。启动时，`core.Instance`会调用[`Start()`](core\xray.go)方法，此方法会循环调用`Instance.Features.Start()`，所以启动时，本质上是启动了各模块的服务。
### 日志模块流程分析
日志配置项比较简单，配置文件配置项如下  
```
{
  "log": {
    "access": "文件地址",
    "error": "文件地址",
    "loglevel": "warning",
    "dnsLog": false
  }
}
``` 

转换到`core.Config.App`里面结构如下  
```
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ErrorLogType  LogType      `protobuf:"varint,1,opt,name=error_log_type,json=errorLogType,proto3,enum=xray.app.log.LogType" json:"error_log_type,omitempty"`
	ErrorLogLevel log.Severity `protobuf:"varint,2,opt,name=error_log_level,json=errorLogLevel,proto3,enum=xray.common.log.Severity" json:"error_log_level,omitempty"`
	ErrorLogPath  string       `protobuf:"bytes,3,opt,name=error_log_path,json=errorLogPath,proto3" json:"error_log_path,omitempty"`
	AccessLogType LogType      `protobuf:"varint,4,opt,name=access_log_type,json=accessLogType,proto3,enum=xray.app.log.LogType" json:"access_log_type,omitempty"`
	AccessLogPath string       `protobuf:"bytes,5,opt,name=access_log_path,json=accessLogPath,proto3" json:"access_log_path,omitempty"`
	EnableDnsLog  bool         `protobuf:"varint,6,opt,name=enable_dns_log,json=enableDnsLog,proto3" json:"enable_dns_log,omitempty"`
}
```
本质上没有区别，只是转换成了代码数据结构。接下来会进行`features.Feature`化，首先在[初始化函数init中](app\log\log.go)注册`creator`函数，`creator`函数会返回一个`log.Instance`实例  
```
type Instance struct {
	sync.RWMutex
	config       *Config
	accessLogger log.Handler
	errorLogger  log.Handler
	active       bool
	dns          bool
}
```
底层实现使用的是[这里的](common\log\log.go)，日志本身带有读写锁。
日志初始化后，会直接启动





