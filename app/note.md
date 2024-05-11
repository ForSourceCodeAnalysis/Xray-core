这个目录对应的是`core.Config`里面的App配置，不包含Inbound和Outbound
包含
- dispatcher
- proxyman.Inbound
- proxyman.Outbound
- api
- metrics
- stats
- log
- router
- dns
- policy
- reverse
- fakedns
- observatory
- burstObservatory  

其中，`dispatcher`和`proxyman`并没有和配置文件里面的配置项对应，是在初始化`App`时，直接添加到`App`里面的
```
config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		},
	}
```




