package foundation

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gookit/goutil/fsutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	DefaultEnv Env = "dev"
	ModuleName     = "foundation"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(
			func() (*Path, error) {
				appRootDir, ok := os.LookupEnv("APP_ROOT_DIR")
				if ok && appRootDir != "" {
					return NewPath(appRootDir)
				}

				_, filename, _, ok := runtime.Caller(0)
				if !ok {
					rootDir, err := os.Getwd()
					if err != nil {
						return nil, fmt.Errorf("can not get current working directory: %w", err)
					}

					return NewPath(rootDir)
				}

				dirname := filepath.Dir(filename)
				appRootDir = fsutil.Realpath(dirname + "/../../..")
				return NewPath(appRootDir)
			},
			func() Env {
				appEnv, ok := os.LookupEnv("APP_ENV")
				if !ok {
					return DefaultEnv
				}
				return Env(appEnv)
			},
		),
		fx.Invoke(func(buildInfo *BuildInfo, path *Path, env Env, log *zap.Logger) {
			log.Info("app build info", zap.String("tag", buildInfo.Tag()), zap.String("commit", buildInfo.Commit()), zap.String("buildTime", buildInfo.BuildTime()))
			log.Info("detected app environment", zap.String("env", env.String()))
			log.Info("detected app working directory", zap.String("dir", path.Path()))
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
