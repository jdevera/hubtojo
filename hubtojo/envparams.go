package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetEnvBool(key string, defValue bool) (bool, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defValue, nil
	}
	switch strings.ToLower(value) {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	default:
		return false, fmt.Errorf("environment variable %s must be true, false, 1, or 0", key)
	}
}

func GetEnvInt(key string, defValue int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defValue, nil
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("environment variable %s must be an integer: %w", key, err)
	}
	return num, nil
}

func GetEnvString(key string, defValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defValue
	}
	return value
}

func GetEnvStrict(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("environment variable %s not set or empty", key)
	}
	return value, nil
}

func GetEnvOptional(key string) *string {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}
