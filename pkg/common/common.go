package common

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func EqualsIgnoreCase(a, b string) bool {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	return a == b
}

func GetStarRocksRootPath(envVars []corev1.EnvVar) string {
	if envVars != nil {
		for _, env := range envVars {
			if EqualsIgnoreCase(env.Name, "STARROCKS_ROOT") {
				return env.Value
			}
		}
	}
	return "/opt/starrocks"
}

func GetFEPreStopScriptPath(feEnvVars []corev1.EnvVar) string {
	return fmt.Sprintf("%s/fe_prestop.sh", GetStarRocksRootPath(feEnvVars))
}

func GetBEPreStopScriptPath(beEnvVars []corev1.EnvVar) string {
	return fmt.Sprintf("%s/be_prestop.sh", GetStarRocksRootPath(beEnvVars))
}

func GetCNPreStopScriptPath(cnEnvVars []corev1.EnvVar) string {
	return fmt.Sprintf("%s/cn_prestop.sh", GetStarRocksRootPath(cnEnvVars))
}

func GetFEConfigDir(envVars []corev1.EnvVar) string {
	return fmt.Sprintf("%s/fe/conf", GetStarRocksRootPath(envVars))
}

func GetBEConfigDir(envVars []corev1.EnvVar) string {
	return fmt.Sprintf("%s/be/conf", GetStarRocksRootPath(envVars))
}

func GetCNConfigDir(envVars []corev1.EnvVar) string {
	return fmt.Sprintf("%s/cn/conf", GetStarRocksRootPath(envVars))
}

func GetFELogDir(envVars []corev1.EnvVar) string {
	return fmt.Sprintf("%s/fe/log", GetStarRocksRootPath(envVars))
}

func GetBELogDir(envVars []corev1.EnvVar) string {
	return fmt.Sprintf("%s/be/log", GetStarRocksRootPath(envVars))
}

func GetCNLogDir(envVars []corev1.EnvVar) string {
	return fmt.Sprintf("%s/cn/log", GetStarRocksRootPath(envVars))
}

func GetFEMetaDir(envVars []corev1.EnvVar) string {
	return fmt.Sprintf("%s/fe/meta", GetStarRocksRootPath(envVars))
}

func GetBEStorageDir(envVars []corev1.EnvVar) string {
	return fmt.Sprintf("%s/be/storage", GetStarRocksRootPath(envVars))
}
