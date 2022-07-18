package setting

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Listen struct {
	IP   string `yaml:"IP"`
	Port int    `yaml:"Port"`
}

func LoadSettings(yamlPath string) Listen {
	settings := Listen{}
	b, _ := os.ReadFile(yamlPath)
	yaml.Unmarshal(b, &settings)
	return settings
}
