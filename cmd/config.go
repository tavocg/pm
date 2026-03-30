package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

type configSource struct {
	path  string
	isDir bool
}

func loadConfig(v *viper.Viper, explicitPath string) ([]string, error) {
	if explicitPath != "" {
		if err := readConfigFile(v, explicitPath, false); err != nil {
			return nil, err
		}
		return []string{explicitPath}, nil
	}

	sources, err := defaultConfigSources()
	if err != nil {
		return nil, err
	}

	return loadConfigSources(v, sources)
}

func loadConfigSources(v *viper.Viper, sources []configSource) ([]string, error) {
	var (
		loadedFiles []string
		merge       bool
	)

	for _, source := range sources {
		if source.isDir {
			files, err := configDirFiles(source.path)
			if err != nil {
				return loadedFiles, err
			}

			for _, path := range files {
				if err := readConfigFile(v, path, merge); err != nil {
					return loadedFiles, err
				}
				loadedFiles = append(loadedFiles, path)
				merge = true
			}

			continue
		}

		exists, err := configFileExists(source.path)
		if err != nil {
			return loadedFiles, err
		}
		if !exists {
			continue
		}

		if err := readConfigFile(v, source.path, merge); err != nil {
			return loadedFiles, err
		}
		loadedFiles = append(loadedFiles, source.path)
		merge = true
	}

	return loadedFiles, nil
}

func readConfigFile(v *viper.Viper, path string, merge bool) error {
	v.SetConfigFile(path)
	if merge {
		return v.MergeInConfig()
	}
	return v.ReadInConfig()
}

func defaultConfigSources() ([]configSource, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var userConfigDir string
	if runtime.GOOS == "windows" {
		userConfigDir, err = os.UserConfigDir()
		if err != nil {
			return nil, err
		}
	}

	return buildDefaultConfigSources(home, os.Getenv("XDG_CONFIG_HOME"), userConfigDir, runtime.GOOS), nil
}

func buildDefaultConfigSources(home, xdgConfigHome, userConfigDir, goos string) []configSource {
	if goos == "windows" {
		return dedupeConfigSources([]configSource{
			{path: filepath.Join(userConfigDir, "pm.yaml")},
			{path: filepath.Join(userConfigDir, "pm"), isDir: true},
			{path: filepath.Join(home, ".pm.yaml")},
			{path: filepath.Join(home, ".pm.d"), isDir: true},
		}, goos)
	}

	homeConfigDir := filepath.Join(home, ".config")
	sources := []configSource{
		{path: filepath.Join("/etc", "pm.yaml")},
		{path: filepath.Join("/etc", "pm"), isDir: true},
		{path: filepath.Join("/usr/local/etc", "pm.yaml")},
		{path: filepath.Join("/usr/local/etc", "pm"), isDir: true},
		{path: filepath.Join(home, ".pm.yaml")},
		{path: filepath.Join(home, ".pm.d"), isDir: true},
		{path: filepath.Join(homeConfigDir, "pm.yaml")},
		{path: filepath.Join(homeConfigDir, "pm"), isDir: true},
	}

	if xdgConfigHome != "" {
		sources = append(sources,
			configSource{path: filepath.Join(xdgConfigHome, "pm.yaml")},
			configSource{path: filepath.Join(xdgConfigHome, "pm"), isDir: true},
		)
	}

	return dedupeConfigSources(sources, goos)
}

func dedupeConfigSources(sources []configSource, goos string) []configSource {
	seen := make(map[string]struct{}, len(sources))
	deduped := make([]configSource, 0, len(sources))

	for _, source := range sources {
		if source.path == "" {
			continue
		}

		key := filepath.Clean(source.path)
		if goos == "windows" {
			key = strings.ToLower(key)
		}
		if source.isDir {
			key += string(filepath.Separator)
		}

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		deduped = append(deduped, source)
	}

	return deduped
}

func configFileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return !info.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func configDirFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}

	sort.Strings(files)
	return files, nil
}
