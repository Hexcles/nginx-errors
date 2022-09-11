package main

import (
	"bufio"
	"mime"
	"os"
	"strings"
)

const mimeFilePath = "etc/mime.types"

var mimeMap = make(map[string]string)

func init() {
	f, err := os.Open(mimeFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			panic("Invalid line in mime.types: " + line)
		}
		mimeMap[fields[0]] = fields[1]
	}
}

// ExtensionByType returns an extension associated with the given MIME type.
// This function uses the standard `mime.ParseMediaType` to parse the MIME
// type, but unlike `mime.ExtensionsByType`:
// 1. It does not have a built-in list.
// 2. It does not add a leading dot to the returned value.
// If the MIME type is invalid, an error will be returned; if an extension
// cannot be found, an empty string will be returned.
func ExtensionByType(fullType string) (string, error) {
	justType, _, err := mime.ParseMediaType(fullType)
	if err != nil {
		return "", err
	}
	return mimeMap[justType], nil
}
