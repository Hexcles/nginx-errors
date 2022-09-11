/*
Copyright 2017 The Kubernetes Authors.
Copyright 2022 Robert Ma (https://github.com/Hexcles), Faire Wholesale Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// main binary.
package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	// FormatHeader name of the header used to extract the format
	FormatHeader = "X-Format"

	// CodeHeader name of the header used as source of the HTTP status code to return
	CodeHeader = "X-Code"

	// OriginalURI name of the header with the original URL from NGINX
	OriginalURI = "X-Original-URI"

	// Namespace name of the header that contains information about the Ingress namespace
	Namespace = "X-Namespace"

	// IngressName name of the header that contains the matched Ingress
	IngressName = "X-Ingress-Name"

	// ServiceName name of the header that contains the matched Service in the Ingress
	ServiceName = "X-Service-Name"

	// ServicePort name of the header that contains the matched Service port in the Ingress
	ServicePort = "X-Service-Port"

	// RequestID is a unique ID that identifies the request - same as for backend service
	RequestID = "X-Request-ID"

	// ContentType name of the header that defines the format of the reply
	ContentType = "Content-Type"

	// DefaultFormatVar is the name of the environment variable indicating
	// the default error MIME type that should be returned if either the
	// client does not specify an Accept header, or the Accept header provided
	// cannot be mapped to a file extension.
	DefaultFormatVar = "DEFAULT_RESPONSE_FORMAT"

	// StatusCodeMapping is the name of the environment variable specifying a mapping
	// of status codes, e.g. `494:400,529:503`.
	StatusCodeMapping = "STATUS_CODE_MAPPING"

	// DebugVar is the name of the environment variable indicating the debug mode.
	// A non-empty value turns on debug logging and response headers.
	DebugVar = "DEBUG"
)

const errFilesPath = "/www"

func main() {
	debugMode := os.Getenv(DebugVar) != ""

	defaultFormat := "text/html"
	if os.Getenv(DefaultFormatVar) != "" {
		defaultFormat = os.Getenv(DefaultFormatVar)
	}

	statusMapping := parseStatusCodeMapping(os.Getenv(StatusCodeMapping))

	http.HandleFunc("/", errorHandler(errFilesPath, defaultFormat, statusMapping, debugMode))

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.ListenAndServe(fmt.Sprintf(":8080"), nil)
}

func errorHandler(path, defaultFormat string, statusMapping map[int]int, debugMode bool) func(http.ResponseWriter, *http.Request) {
	defaultExts, err := mime.ExtensionsByType(defaultFormat)
	if err != nil || len(defaultExts) == 0 {
		panic("couldn't get file extension for default format: " + defaultFormat)
	}
	defaultExt := defaultExts[0]

	return func(w http.ResponseWriter, r *http.Request) {
		if debugMode {
			w.Header().Set(FormatHeader, r.Header.Get(FormatHeader))
			w.Header().Set(CodeHeader, r.Header.Get(CodeHeader))
			w.Header().Set(OriginalURI, r.Header.Get(OriginalURI))
			w.Header().Set(Namespace, r.Header.Get(Namespace))
			w.Header().Set(IngressName, r.Header.Get(IngressName))
			w.Header().Set(ServiceName, r.Header.Get(ServiceName))
			w.Header().Set(ServicePort, r.Header.Get(ServicePort))
			w.Header().Set(RequestID, r.Header.Get(RequestID))
		}

		format := r.Header.Get(FormatHeader)
		if format == "" {
			format = defaultFormat
			if debugMode {
				log.Printf("format not specified. Using %v", format)
			}
		}

		ext := defaultExt
		if cext, err := mime.ExtensionsByType(format); err != nil {
			if debugMode {
				log.Printf("unexpected error reading media type extension: %v. Using %v", err, defaultExt)
			}
			format = defaultFormat
		} else if len(cext) == 0 {
			if debugMode {
				log.Printf("couldn't get media type extension. Using %v", defaultExt)
			}
			format = defaultFormat
		} else {
			ext = cext[0]
		}
		// special case for compatibility
		if ext == ".htm" {
			ext = ".html"
		}
		w.Header().Set(ContentType, format)

		code, err := strconv.Atoi(r.Header.Get(CodeHeader))
		if err != nil {
			// not configurable because it should never happen when called by ingress controller
			code = 404
			log.Printf("unexpected error reading return code: %v. Using %v", err, code)
		}
		if newCode, ok := statusMapping[code]; ok {
			log.Printf("mapping status code %d to %d", code, newCode)
			code = newCode
		}

		file := fmt.Sprintf("%s/%d%s", path, code, ext)
		f, err := os.Open(file)
		if err != nil {
			if debugMode {
				log.Printf("falling back due to error opening file: %v", err)
			}
			file = fmt.Sprintf("%s/%dxx%s", path, code/100, ext)
			f, err = os.Open(file)
			if err != nil {
				log.Printf("returning 404 due to unexpected error opening file: %v", err)
				http.NotFound(w, r)
				return
			}
		}
		defer f.Close()

		log.Printf("serving custom error response for code %v and format %v from file %v", code, format, file)
		w.WriteHeader(code)
		io.Copy(w, f)
	}
}

func parseStatusCodeMapping(config string) map[int]int {
	if config == "" {
		return nil
	}
	mapping := make(map[int]int)
	for _, pair := range strings.Split(config, ",") {
		src, dst, found := strings.Cut(pair, ":")
		if !found {
			panic("invalid status mapping: " + config)
		}
		srcCode, err := strconv.Atoi(src)
		if err != nil {
			panic(err)
		}
		dstCode, err := strconv.Atoi(dst)
		if err != nil {
			panic(err)
		}
		mapping[srcCode] = dstCode
	}
	return mapping
}
