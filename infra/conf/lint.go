package conf

type ConfigureFilePostProcessingStage interface {
	Process(conf *Config) error
}

var configureFilePostProcessingStages map[string]ConfigureFilePostProcessingStage

// 此函数只被调用了一次，注册了一个 FakeDNS
func RegisterConfigureFilePostProcessingStage(name string, stage ConfigureFilePostProcessingStage) {
	if configureFilePostProcessingStages == nil {
		configureFilePostProcessingStages = make(map[string]ConfigureFilePostProcessingStage)
	}
	configureFilePostProcessingStages[name] = stage
}

func PostProcessConfigureFile(conf *Config) error {
	for k, v := range configureFilePostProcessingStages {
		if err := v.Process(conf); err != nil {
			return newError("Rejected by Postprocessing Stage ", k).AtError().Base(err)
		}
	}
	return nil
}
