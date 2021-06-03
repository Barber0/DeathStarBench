package registry

import (
	"fmt"
	"strings"
)

func GetK8sServiceAddr(name string, conf map[string]string) (string, error) {
	headChar := name[:1]
	portKey := strings.ToUpper(headChar) + name[1:] + "Port"
	portStr, ok := conf[portKey]
	if !ok {
		return "", fmt.Errorf("invalid service[%s], port not found", name)
	}
	return fmt.Sprintf("%s:%s", name, portStr), nil
}
