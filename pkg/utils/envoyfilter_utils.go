package utils

import (
	"fmt"
	"strings"
)

var efUtil = &envoyFilterUtil{}

type EnvoyFilterUtilInterface interface {
	GetName(prefix string, filterName string, suffix string) string
}

type envoyFilterUtil struct{}

func EnvoyFilterUtil() EnvoyFilterUtilInterface {
	return efUtil
}

func (envoyFilterUtil) GetName(prefix string, filterName string, suffix string) string {
	return strings.ToLower(fmt.Sprintf("%s-%s-%s", prefix, filterName, suffix))
}
