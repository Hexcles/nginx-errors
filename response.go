package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const errFilesPath = "www"

var responseMap = make(map[string][]byte)

func init() {
	files, err := os.ReadDir(errFilesPath)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(errFilesPath, file.Name()))
		if err != nil {
			panic(err)
		}
		responseMap[file.Name()] = content
	}
}

// GetResponseReader returns a Reader and the file name of the response for a given status code and extension.
// Using 404 and json as an example, it first tries "404.json", and falls back to "4xx.json".
// When neither can be found, it returns a nil Reader and an empty string.
func GetResponseReader(code int, ext string, debugMode bool) (io.Reader, string) {
	file := fmt.Sprintf("%d.%s", code, ext)
	if content, ok := responseMap[file]; ok {
		return bytes.NewReader(content), file
	}
	if debugMode {
		log.Printf("%s not found. Falling back", file)
	}
	file = fmt.Sprintf("%dxx.%s", code/100, ext)
	if content, ok := responseMap[file]; ok {
		return bytes.NewReader(content), file
	}
	log.Printf("%s still not found. Returning 404", file)
	return nil, ""
}
