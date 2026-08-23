package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
)

var (
	globalConfig *BootstrapConfig
	configMu     sync.RWMutex
	nacosClient   config_client.IConfigClient
	nacosConfig   config.Config
)

func LoadConfig(serviceName string) (*BootstrapConfig, error) {
	cfg := config.New(
		config.WithSource(
			file.NewSource("configs/"+serviceName+"/config.yaml"),
		),
	)

	if err := cfg.Load(); err != nil {
		return nil, fmt.Errorf("加载本地配置失败: %w", err)
	}

	var bc BootstrapConfig
	if err := cfg.Scan(&bc); err != nil {
		return nil, fmt.Errorf("解析本地配置失败: %w", err)
	}

	nacosAddr := os.Getenv("NACOS_SERVER_ADDR")
	if nacosAddr != "" || bc.Nacos.ServerAddr != "" {
		addr := nacosAddr
		if addr == "" {
			addr = bc.Nacos.ServerAddr
		}

		if err := loadFromNacos(addr, serviceName, &bc); err != nil {
			log.Warn("从 Nacos 加载配置失败，使用本地配置", "error", err)
		}
	}

	bc.ApplyEnvOverrides()

	if bc.ServiceName == "" {
		bc.ServiceName = serviceName
	}

	configMu.Lock()
	globalConfig = &bc
	configMu.Unlock()

	return &bc, nil
}

func loadFromNacos(serverAddr, serviceName string, bc *BootstrapConfig) error {
	ip, port := parseNacosAddrPort(serverAddr)

	namespace := bc.Nacos.Namespace
	if namespace == "" {
		namespace = "dev"
	}

	group := bc.Nacos.Group
	if group == "" {
		group = "MARKETING_GROUP"
	}

	clientConfig := constant.ClientConfig{
		LogDir:      ".nacos/logs",
		CacheDir:    ".nacos/cache",
		LogLevel:    "warn",
		NamespaceId: namespace,
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: ip,
			Port:   uint64(port),
		},
	}

	nc, err := clients.CreateConfigClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: serverConfigs,
		constant.KEY_CLIENT_CONFIG:  clientConfig,
	})
	if err != nil {
		return fmt.Errorf("创建 Nacos 客户端失败: %w", err)
	}
	nacosClient = nc

	dataID := serviceName + ".yaml"
	if bc.Nacos.DataID != "" {
		dataID = bc.Nacos.DataID
	}

	content, err := nc.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
	if err != nil {
		return fmt.Errorf("从 Nacos 获取配置失败: %w", err)
	}

	if content != "" {
		log.Infof("从 Nacos 加载配置成功: dataID=%s, group=%s, namespace=%s", dataID, group, namespace)

		var remoteCfg BootstrapConfig
		if err := yaml.Unmarshal([]byte(content), &remoteCfg); err != nil {
			return fmt.Errorf("解析 Nacos 配置内容失败: %w", err)
		}

		mergeConfig(bc, &remoteCfg)

		err := nc.ListenConfig(vo.ConfigParam{
			DataId: dataID,
			Group:  group,
			OnChange: func(ns, g, dataID, data string) {
				log.Infof("Nacos 配置变更: dataID=%s", dataID)
				var newCfg BootstrapConfig
				if err := yaml.Unmarshal([]byte(data), &newCfg); err == nil {
					configMu.Lock()
					mergeConfig(globalConfig, &newCfg)
					configMu.Unlock()
					log.Infof("Nacos 配置已热更新")
				} else {
					log.Errorf("解析变更配置失败: %v", err)
				}
			},
		})
		if err != nil {
			log.Warn("监听 Nacos 配置变更失败", "error", err)
		}

		return nil
	}

	return fmt.Errorf("Nacos 中未找到配置: dataID=%s", dataID)
}

func mergeConfig(dst, src *BootstrapConfig) {
	if src.Server.Addr != "" {
		dst.Server.Addr = src.Server.Addr
	}
	if src.Server.Name != "" {
		dst.Server.Name = src.Server.Name
	}
	if src.Server.Timeout != 0 {
		dst.Server.Timeout = src.Server.Timeout
	}
	if src.Data.MySQL.DSN != "" {
		dst.Data.MySQL.DSN = src.Data.MySQL.DSN
	}
	if src.Data.Redis.Addr != "" {
		dst.Data.Redis.Addr = src.Data.Redis.Addr
	}
	if src.Data.Redis.Password != "" {
		dst.Data.Redis.Password = src.Data.Redis.Password
	}
	if src.Data.Redis.DB != 0 {
		dst.Data.Redis.DB = src.Data.Redis.DB
	}
	if src.Data.RabbitMQ.URL != "" {
		dst.Data.RabbitMQ.URL = src.Data.RabbitMQ.URL
	}
	if src.Data.Stock.URL != "" {
		dst.Data.Stock.URL = src.Data.Stock.URL
	}
	if src.Log.Level != "" {
		dst.Log.Level = src.Log.Level
	}
	if src.Log.Format != "" {
		dst.Log.Format = src.Log.Format
	}
	if src.Nacos.ServerAddr != "" {
		dst.Nacos.ServerAddr = src.Nacos.ServerAddr
	}
	if src.Nacos.Namespace != "" {
		dst.Nacos.Namespace = src.Nacos.Namespace
	}
	if src.Nacos.Group != "" {
		dst.Nacos.Group = src.Nacos.Group
	}
	if src.Nacos.DataID != "" {
		dst.Nacos.DataID = src.Nacos.DataID
	}
	if src.Observability.Trace.Enabled {
		dst.Observability.Trace.Enabled = src.Observability.Trace.Enabled
	}
	if src.Observability.Trace.Endpoint != "" {
		dst.Observability.Trace.Endpoint = src.Observability.Trace.Endpoint
	}
	if src.Observability.Metrics.Enabled {
		dst.Observability.Metrics.Enabled = src.Observability.Metrics.Enabled
	}
	if src.Observability.Metrics.Path != "" {
		dst.Observability.Metrics.Path = src.Observability.Metrics.Path
	}
	if src.Observability.Metrics.Port != 0 {
		dst.Observability.Metrics.Port = src.Observability.Metrics.Port
	}
}

func parseNacosAddrPort(addr string) (string, int) {
	parts := strings.SplitN(addr, ":", 2)
	ip := parts[0]
	port := 8848
	if len(parts) == 2 {
		fmt.Sscanf(parts[1], "%d", &port)
		if port == 0 {
			port = 8848
		}
	}
	return ip, port
}

func GetGlobalConfig() *BootstrapConfig {
	configMu.RLock()
	defer configMu.RUnlock()
	return globalConfig
}

func WatchConfig(ctx context.Context, onChange func(*BootstrapConfig)) {
	if nacosConfig == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				configMu.RLock()
				cfg := globalConfig
				configMu.RUnlock()

				if cfg != nil && onChange != nil {
					onChange(cfg)
				}
			}
		}
	}()
}

func InitLogger(cfg *LogConfig) log.Logger {
	level := log.InfoLevel
	switch cfg.Level {
	case "debug":
		level = log.DebugLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	}
	return log.NewLog(
		log.WithLevel(level),
		log.WithFormat(cfg.Format),
	)
}