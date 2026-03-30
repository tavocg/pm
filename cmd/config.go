package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

type configSource struct {
	path  string
	isDir bool
}

func loadConfig(v *viper.Viper, explicitPath string) ([]string, error) {
	merged := make(map[string]any)

	if explicitPath != "" {
		if err := mergeConfigFile(merged, explicitPath); err != nil {
			return nil, err
		}
		if err := v.MergeConfigMap(merged); err != nil {
			return nil, err
		}
		return []string{explicitPath}, nil
	}

	sources, err := defaultConfigSources()
	if err != nil {
		return nil, err
	}

	return loadConfigSources(v, sources, merged)
}

func loadConfigSources(v *viper.Viper, sources []configSource, merged map[string]any) ([]string, error) {
	var loadedFiles []string

	for _, source := range sources {
		if source.isDir {
			files, err := configDirFiles(source.path)
			if err != nil {
				return loadedFiles, err
			}

			for _, path := range files {
				if err := mergeConfigFile(merged, path); err != nil {
					return loadedFiles, err
				}
				loadedFiles = append(loadedFiles, path)
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

		if err := mergeConfigFile(merged, source.path); err != nil {
			return loadedFiles, err
		}
		loadedFiles = append(loadedFiles, source.path)
	}

	if len(loadedFiles) != 0 {
		if err := v.MergeConfigMap(merged); err != nil {
			return loadedFiles, err
		}
	}

	return loadedFiles, nil
}

func mergeConfigFile(dst map[string]any, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var src map[string]any
	if err := yaml.Unmarshal(data, &src); err != nil {
		return err
	}

	mergeConfigMaps(dst, src)
	return nil
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

func mergeConfigMaps(dst, src map[string]any) {
	for key, srcValue := range src {
		dstValue, ok := dst[key]
		if !ok {
			dst[key] = srcValue
			continue
		}

		srcMap, srcIsMap := srcValue.(map[string]any)
		dstMap, dstIsMap := dstValue.(map[string]any)
		if srcIsMap && dstIsMap {
			mergeConfigMaps(dstMap, srcMap)
			continue
		}

		srcSlice, srcIsSlice := srcValue.([]any)
		dstSlice, dstIsSlice := dstValue.([]any)
		if srcIsSlice && dstIsSlice {
			if mergedSlice, merged := mergeNamedObjectSlices(dstSlice, srcSlice); merged {
				dst[key] = mergedSlice
				continue
			}
		}

		dst[key] = srcValue
	}
}

func mergeNamedObjectSlices(dst, src []any) ([]any, bool) {
	identityKey, ok := sliceIdentityKey(dst, src)
	if !ok {
		return nil, false
	}

	merged := append([]any(nil), dst...)
	indexByName := make(map[string]int, len(merged))

	for i, item := range merged {
		obj := item.(map[string]any)
		name := obj[identityKey].(string)
		indexByName[name] = i
	}

	for _, item := range src {
		obj := item.(map[string]any)
		name := obj[identityKey].(string)

		if idx, exists := indexByName[name]; exists {
			existing := merged[idx].(map[string]any)
			mergeConfigMaps(existing, obj)
			continue
		}

		indexByName[name] = len(merged)
		merged = append(merged, item)
	}

	return merged, true
}

func sliceIdentityKey(dst, src []any) (string, bool) {
	identityKeys := []string{"manager", "remote", "task", "package"}

	for _, key := range identityKeys {
		if sliceHasIdentityKey(dst, key) && sliceHasIdentityKey(src, key) {
			return key, true
		}
	}

	return "", false
}

func sliceHasIdentityKey(items []any, key string) bool {
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			return false
		}

		name, ok := obj[key]
		if !ok {
			return false
		}

		if _, ok := name.(string); !ok {
			return false
		}
	}

	return true
}
