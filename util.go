package main

import (
	"strings"
)

func removeHead(s []string) (head string, tail []string) {
	if len(s) > 0 {
		head = s[0]
	}
	if len(s) > 1 {
		tail = s[1:len(s)]
	}
	return
}

func toCleanRole(role string) string {
	role = strings.TrimPrefix(role, "\"")
	role = strings.TrimSuffix(role, "\"")
	return role
}
