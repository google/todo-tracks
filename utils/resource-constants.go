/*
Copyright 2014 Google Inc. All rights reserved.

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

// Program resource-constants generates Go source files that embed resource files read at compile time.
//
// Usage:
//   bin/resource-constants --base_dir <directory-with-static-files>/ > src/resources/constants.go
//
// Using the generated code:
//   import "resources"
//
//   var fileContents = resources.Constants["fileName"]
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var baseDir = flag.String("base_dir", "./", "Directory under which to look for resource files")
var supportedExtensions = flag.String("supported_extensions", "html,js,css", "Comma-separated list of supported file extensions")

func isSupported(fileName string) bool {
	for _, extension := range strings.Split(*supportedExtensions, ",") {
		if strings.HasSuffix(fileName, "."+extension) {
			return true
		}
	}
	return false
}

func main() {
	flag.Parse()
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "package resources")
	fmt.Fprintln(&buf, "var Constants = map[string][]byte{")
	filepath.Walk(*baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isSupported(path) {
			fileName := path[len(*baseDir):]
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			fmt.Fprintf(&buf, "%q: {", fileName)
			for _, b := range bytes {
				fmt.Fprintf(&buf, " %d,", b)
			}
			fmt.Fprintln(&buf, "},")
		}
		return nil
	})
	fmt.Fprintln(&buf, "}")
	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("Invalid generated source: %v", err)
	}
	fmt.Print(string(src))
}
