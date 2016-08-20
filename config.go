package singular

import (
	"io/ioutil"
	"net/url"

	"gopkg.in/yaml.v2"
)

// Config define a client config
type Config struct {
	ServerAddr string            `yaml:"server_addr,omitempty"`
	AuthToken  string            `yaml:"auth_token"`
	Proxy      map[string]string `yaml:"proxy"`
}

// ParseConfig parse config file
func ParseConfig(path string) (config *Config, err error) {
	configBuf, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	config = new(Config)
	if err = yaml.Unmarshal(configBuf, &config); err != nil {
		return
	}

	for _, addr := range config.Proxy {
		if _, err := url.Parse(addr); err != nil {
			return config, err
		}
	}

	return
}
