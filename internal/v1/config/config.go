package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	foundation2 "github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/gookit/goutil/fsutil"
	"github.com/joho/godotenv"
	"go.uber.org/config"
	"go.uber.org/fx"
)

const (
	ConfigKey = "app"

	defaultConfigFileName = "config.yaml"
	defaultEnvFileName    = ".env"
)

type Config struct {
	Name       string `yaml:"name"`
	Debug      bool   `yaml:"debug"`
	Repository string `yaml:"repository"`
}

type Result struct {
	fx.Out
	Provider config.Provider
	Config   *Config
}

type Params struct {
	fx.In
	Env  foundation2.Env
	Path *foundation2.Path
}

func New(p Params) (Result, error) {
	if err := loadEnvFile(p.Env, p.Path); err != nil {
		return Result{}, err
	}

	configFile, err := getConfigFile(p.Path, p.Env)
	if err != nil {
		return Result{}, err
	}
	defer configFile.Close()

	provider, err := config.NewYAML(
		config.Source(configFile),
		config.Expand(os.LookupEnv),
	)
	if err != nil {
		return Result{}, err
	}

	cfg := &Config{}
	if err := provider.Get(ConfigKey).Populate(cfg); err != nil {
		return Result{}, err
	}

	return Result{
		Provider: provider,
		Config:   cfg,
	}, nil
}

func getConfigFile(path *foundation2.Path, env foundation2.Env) (fs.File, error) {
	configFileName := fmt.Sprintf("config.%s.yaml", env.String())

	configFile, err := os.Open(path.ConfigPath(configFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			configFile, err = os.Open(path.ConfigPath(defaultConfigFileName))
			if err != nil {
				errMsg := strings.Join([]string{configFileName, defaultConfigFileName}, " or ")
				return nil, fmt.Errorf("failed to open config file: %s", errMsg)
			}
		} else {
			return nil, fmt.Errorf("failed to open config file: %w", err)
		}
	}

	return configFile, nil
}

func loadEnvFile(env foundation2.Env, path *foundation2.Path) error {
	envFilePath := path.Path(fmt.Sprintf(".env.%s", env.String()))

	if !fsutil.FileExists(envFilePath) {
		envFilePath = path.Path(defaultEnvFileName)
		if !fsutil.FileExists(envFilePath) {
			return nil
		}
	}

	if err := godotenv.Load(envFilePath); err != nil {
		return fmt.Errorf("failed to load env file: %w", err)
	}

	return nil
}
