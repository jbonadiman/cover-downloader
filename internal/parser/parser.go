package parser

import (
	"regexp"
)

func GetGameSerialSet(content string) map[string]bool {
	serialPattern := regexp.MustCompile(`\w{4}-\d{5}`)
	matches := serialPattern.FindAllString(string(content), -1)

	return uniqueStrings(matches)
}

func uniqueStrings(slice []string) map[string]bool {
	uniqueMap := make(map[string]bool)

	for _, str := range slice {
		if !uniqueMap[str] {
			uniqueMap[str] = true
		}
	}

	return uniqueMap
}
