package foundation

import (
	"errors"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/gookit/goutil/fsutil"
)

const (
	StorageDirName    = "storage"
	MigrationsDirName = "migrations"
	ConfigDirName     = "config"
	BinariesDirName   = "binaries"
	I18nDirName       = "i18n"

	PathLookupPrefix = "$PATH:"
)

var (
	ErrNotDirectory   = errors.New("not a directory")
	ErrBinaryNotFound = errors.New("binary not found")
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

func (p *Path) ResolveFrom(elem ...string) (string, error) {
	if len(elem) == 0 {
		return p.rootDirectory, nil
	}

	first := elem[0]
	if strings.HasPrefix(first, PathLookupPrefix) {
		binary := strings.TrimPrefix(first, PathLookupPrefix)

		pth, err := exec.LookPath(binary)
		if err != nil {
			return "", fmt.Errorf("%w: %s", ErrBinaryNotFound, binary)
		}
		return pth, nil
	}

	if path.IsAbs(first) {
		return path.Clean(filepath.Join(elem...)), nil
	}

	all := append([]string{p.rootDirectory}, elem...)
	return path.Join(all...), nil
}

func (p *Path) BinariesPath(elem ...string) string {
	return p.Path(append([]string{BinariesDirName}, elem...)...)
}

func (p *Path) Path(elem ...string) string {
	return path.Join(append([]string{p.rootDirectory}, elem...)...)
}

func (p *Path) I18nPath(elem ...string) string {
	return p.Path(append([]string{I18nDirName}, elem...)...)
}

func (p *Path) ConfigPath(elem ...string) string {
	return p.Path(append([]string{ConfigDirName}, elem...)...)
}

func (p *Path) StoragePath(elem ...string) string {
	return p.Path(append([]string{StorageDirName}, elem...)...)
}

func (p *Path) MigrationsPath(elem ...string) string {
	return p.Path(append([]string{MigrationsDirName}, elem...)...)
}
