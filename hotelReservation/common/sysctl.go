package common

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

const sysctlHome = "/proc/sys/"

func sysctlKey(key string) string {
	return filepath.Join(sysctlHome, strings.Replace(key, ".", "/", -1))
}

func readSysctl(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func setSysctl(path, value string) error {
	return ioutil.WriteFile(path, []byte(value), 0644)
}

func GetSysctl(key string) (string, error) {
	return readSysctl(sysctlKey(key))
}

func SetSysctl(key, value string) error {
	return setSysctl(sysctlKey(key), value)
}
