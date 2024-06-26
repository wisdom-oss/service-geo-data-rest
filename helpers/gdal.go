package helpers

import (
	"os/exec"
	"strconv"
	"strings"
)

func SpatialReferenceInformation(input string, outputFormat string) (interface{}, error) {
	bytes, err := exec.Command("gdalsrsinfo", input, "-o", outputFormat).Output()
	if err != nil {
		return "", err
	}
	var output interface{}
	switch outputFormat {
	case "epsg":
		codeOnly := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(string(bytes))), "epsg:")
		i, err := strconv.ParseInt(codeOnly, 10, 64)
		if err != nil {
			return "", err
		}
		output = int(i)
	default:
		output = strings.TrimSpace(string(bytes))
	}
	return output, nil
}
