package pod

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func IsLowerThanAny(checkVersion string, targetVersions []string) (bool, error) {
	if strings.Contains(checkVersion, "latest") {
		return false, nil
	}

	checkVer, err := parseVersion(checkVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse check version %s: %w", checkVersion, err)
	}

	for _, targetVersion := range targetVersions {
		targetVer, err := parseVersion(targetVersion)
		if err != nil {
			return false, fmt.Errorf("failed to parse target version %s: %w", targetVersion, err)
		}

		// check whether the major and minor versions are the same
		if checkVer.Major == targetVer.Major && checkVer.Minor == targetVer.Minor {
			// In the same major and minor version, compare the patch version
			if checkVer.Patch < targetVer.Patch {
				return true, nil
			}
			return false, nil
		}
	}

	return false, fmt.Errorf("version %s not in target versions", checkVersion)
}

// for version formats like "3.3-latest", it will return invalid format error
func parseVersion(versionStr string) (Version, error) {
	// handle non-standard formats, remove the part after the hyphen
	parts := strings.Split(versionStr, "-")
	versionCore := parts[0]

	// parse the core version part
	versionParts := strings.Split(versionCore, ".")
	if len(versionParts) != 3 {
		return Version{}, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", versionParts[0])
	}

	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", versionParts[1])
	}

	patch, err := strconv.Atoi(versionParts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %s", versionParts[2])
	}

	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

func GetImageVersion(image string) string {
	parts := strings.LastIndex(image, ":")
	return image[parts+1:]
}
