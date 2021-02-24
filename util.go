package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"

	"github.com/caromo/rinako/collections"
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

func find(slice []string, val string) (res int, exists bool) {
	res = -1
	for i, v := range slice {
		if val == v {
			return i, true
		}
	}
	return
}

func findCaseInsensitive(slice []string, val string) (res int, exists bool) {
	res = -1
	for i, v := range slice {
		if strings.ToLower(val) == strings.ToLower(v) {
			return i, true
		}
	}
	return
}

func appendUnique(slice []string, val string) []string {
	for _, item := range slice {
		if item == val {
			return slice
		}
	}
	return append(slice, val)
}

func strip(in string) (res string, err error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9\"]+")
	if err != nil {
		return "", err
	}
	return reg.ReplaceAllString(in, ""), nil
}

func jsonToRoles(in []byte) []collections.RoleDesc {
	var result []collections.RoleDesc
	err := json.Unmarshal(in, &result)
	if err != nil {
		log.Printf("failed to unmarshall roles: %v", err)
	}
	return result
}

func rolesToJson(r []collections.RoleDesc) []byte {
	b, err := json.Marshal(r)
	if err != nil {
		log.Printf("failed to marshall roles: %v", err)
		return nil
	}
	return b
}

func removeFromSlice(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func extractQuotes(in string) []string {
	re := regexp.MustCompile(`"[^"]+"`)
	newStrs := re.FindAllString(in, -1)
	return newStrs
}
