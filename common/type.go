package common

import (
	"context"
	"reflect"
)

// ConfigCreator is a function to create an object by a config.
type ConfigCreator func(ctx context.Context, config interface{}) (interface{}, error)

var typeCreatorRegistry = make(map[reflect.Type]ConfigCreator)

// RegisterConfig registers a global config creator. The config can be nil but must have a type.
func RegisterConfig(config interface{}, configCreator ConfigCreator) error {
	configType := reflect.TypeOf(config)
	if _, found := typeCreatorRegistry[configType]; found {
		return newError(configType.Name() + " is already registered").AtError()
	}
	typeCreatorRegistry[configType] = configCreator
	return nil
}

// typeCreatorRegistry会在每个模块对应的init函数中
// 通过调用RegisterConfig初始化对于的模块映射
// 比如log模块，就是在app\log\log.go里面初始化的
// 每个模块对应的creator不尽相同，需要查看对应的模块代码
// CreateObject creates an object by its config. The config type must be registered through RegisterConfig().
func CreateObject(ctx context.Context, config interface{}) (interface{}, error) {
	configType := reflect.TypeOf(config)
	creator, found := typeCreatorRegistry[configType]
	if !found {
		return nil, newError(configType.String() + " is not registered").AtError()
	}
	return creator(ctx, config)
}
