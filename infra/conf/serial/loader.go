package serial

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/ghodss/yaml"
	"github.com/pelletier/go-toml"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
	json_reader "github.com/xtls/xray-core/infra/conf/json"
)

type offset struct {
	line int
	char int
}

func findOffset(b []byte, o int) *offset {
	if o >= len(b) || o < 0 {
		return nil
	}

	line := 1
	char := 0
	for i, x := range b {
		if i == o {
			break
		}
		if x == '\n' {
			line++
			char = 0
		} else {
			char++
		}
	}

	return &offset{line: line, char: char}
}

// DecodeJSONConfig reads from reader and decode the config into *conf.Config
// syntax error could be detected.
func DecodeJSONConfig(reader io.Reader) (*conf.Config, error) {
	jsonConfig := &conf.Config{}

	//最大支持1m左右的配置
	jsonContent := bytes.NewBuffer(make([]byte, 0, 10240))

	//TeeReader根据输入的io.Reader创建一个新的Reader，并将内容同时写入到指定的io.Writer中
	//这有什么用呢，直接把内容给io.Writer不可以吗？
	//不可以。仔细阅读其功能描述就可以发现，旧的io.Reader内容其实复制了两次，一次是io.Writer，
	//另一次是新的io.Reader，如果直接复制给io.Writer，旧的io.Reader内容其实已经被清空了，无法再次使用
	//所以根本的原因就是，想读取io.Reader后，还想继续使用io.Reader
	jsonReader := io.TeeReader(&json_reader.Reader{
		Reader: reader,
	}, jsonContent)
	decoder := json.NewDecoder(jsonReader)

	if err := decoder.Decode(jsonConfig); err != nil {
		var pos *offset
		cause := errors.Cause(err)
		switch tErr := cause.(type) {
		case *json.SyntaxError:
			pos = findOffset(jsonContent.Bytes(), int(tErr.Offset))
		case *json.UnmarshalTypeError:
			pos = findOffset(jsonContent.Bytes(), int(tErr.Offset))
		}
		if pos != nil {
			return nil, newError("failed to read config file at line ", pos.line, " char ", pos.char).Base(err)
		}
		return nil, newError("failed to read config file").Base(err)
	}

	return jsonConfig, nil
}

func LoadJSONConfig(reader io.Reader) (*core.Config, error) {
	jsonConfig, err := DecodeJSONConfig(reader)
	if err != nil {
		return nil, err
	}

	pbConfig, err := jsonConfig.Build()
	if err != nil {
		return nil, newError("failed to parse json config").Base(err)
	}

	return pbConfig, nil
}

// DecodeTOMLConfig reads from reader and decode the config into *conf.Config
// using github.com/pelletier/go-toml and map to convert toml to json.
func DecodeTOMLConfig(reader io.Reader) (*conf.Config, error) {
	tomlFile, err := io.ReadAll(reader)
	if err != nil {
		return nil, newError("failed to read config file").Base(err)
	}

	configMap := make(map[string]interface{})
	if err := toml.Unmarshal(tomlFile, &configMap); err != nil {
		return nil, newError("failed to convert toml to map").Base(err)
	}

	jsonFile, err := json.Marshal(&configMap)
	if err != nil {
		return nil, newError("failed to convert map to json").Base(err)
	}

	return DecodeJSONConfig(bytes.NewReader(jsonFile))
}

func LoadTOMLConfig(reader io.Reader) (*core.Config, error) {
	tomlConfig, err := DecodeTOMLConfig(reader)
	if err != nil {
		return nil, err
	}

	pbConfig, err := tomlConfig.Build()
	if err != nil {
		return nil, newError("failed to parse toml config").Base(err)
	}

	return pbConfig, nil
}

// DecodeYAMLConfig reads from reader and decode the config into *conf.Config
// using github.com/ghodss/yaml to convert yaml to json.
func DecodeYAMLConfig(reader io.Reader) (*conf.Config, error) {
	yamlFile, err := io.ReadAll(reader)
	if err != nil {
		return nil, newError("failed to read config file").Base(err)
	}

	jsonFile, err := yaml.YAMLToJSON(yamlFile)
	if err != nil {
		return nil, newError("failed to convert yaml to json").Base(err)
	}

	return DecodeJSONConfig(bytes.NewReader(jsonFile))
}

func LoadYAMLConfig(reader io.Reader) (*core.Config, error) {
	yamlConfig, err := DecodeYAMLConfig(reader)
	if err != nil {
		return nil, err
	}

	pbConfig, err := yamlConfig.Build()
	if err != nil {
		return nil, newError("failed to parse yaml config").Base(err)
	}

	return pbConfig, nil
}
