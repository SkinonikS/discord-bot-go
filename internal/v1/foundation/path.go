package foundation

import (
	"errors"
	"fmt"
	"path"

	"github.com/gookit/goutil/fsutil"
)

const (
	StorageDirName    = "storage"
	MigrationsDirName = "migrations"
	ConfigDirName     = "config"
)

var (
	ErrNotDirectory = errors.New("not a directory")
)

type Path struct {
	rootDirectory string
}

func NewPath(rootDirectory string) (*Path, error) {
	if !fsutil.IsDir(rootDirectory) {
		return nil, fmt.Errorf("%w: %s", ErrNotDirectory, rootDirectory)
	}

	return &Path{rootDirectory: rootDirectory}, nil
}

func (p *Path) ConfigPath(elem ...string) string {
	return p.Path(append([]string{ConfigDirName}, elem...)...)
}

func (p *Path) Path(elem ...string) string {
	return path.Join(append([]string{p.rootDirectory}, elem...)...)
}

func (p *Path) StoragePath(elem ...string) string {
	return p.Path(append([]string{StorageDirName}, elem...)...)
}

func (p *Path) MigrationsPath(elem ...string) string {
	return p.Path(append([]string{MigrationsDirName}, elem...)...)
}
