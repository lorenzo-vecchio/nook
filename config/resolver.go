package config

import (
	"os"
	"regexp"

	"github.com/joho/godotenv"
)

var varPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

func ResolveEnvVars(input string, extra map[string]string) string {
	return varPattern.ReplaceAllStringFunc(input, func(match string) string {
		name := match[2 : len(match)-1]

		if extra != nil {
			if val, ok := extra[name]; ok {
				return val
			}
		}

		if val, ok := os.LookupEnv(name); ok {
			return val
		}

		return match
	})
}

func LoadEnvFile(path string) (map[string]string, error) {
	env, err := godotenv.Read(path)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func ResolveAllEnvVars(input string, envFile string, extra map[string]string) string {
	combined := make(map[string]string)
	for k, v := range extra {
		combined[k] = v
	}
	if envFile != "" {
		envMap, err := LoadEnvFile(envFile)
		if err == nil {
			for k, v := range envMap {
				combined[k] = v
			}
		}
	}
	return ResolveEnvVars(input, combined)
}
