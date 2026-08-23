package config

import (
	"os"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
)

func LoadConfig(serviceName string) (*BootstrapConfig, error) {
	cfg := config.New(
		config.WithSource(
			file.NewSource("configs/" + serviceName + "/config.yaml"),
		),
	)

	nacosAddr := os.Getenv("NACOS_SERVER_ADDR")
	if nacosAddr != "" {
		cfg = config.New(
			config.WithSource(
				file.NewSource("configs/" + serviceName + "/config.yaml"),
			),
		)
	}

	if err := cfg.Load(); err != nil {
		return nil, err
	}

	var bc BootstrapConfig
	if err := cfg.Scan(&bc); err != nil {
		return nil, err
	}

	if bc.ServiceName == "" {
		bc.ServiceName = serviceName
	}

	bc.ApplyEnvOverrides()

	return &bc, nil
}
