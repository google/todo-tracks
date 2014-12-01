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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var resourceFileExtensions = []string{".html", ".js", ".css"}
var baseDir string

func init() {
	flag.StringVar(&baseDir, "base_dir", "./", "Directory under which to look for resource files")
}

func isResourceFileName(fileName string) bool {
	for _, extension := range resourceFileExtensions {
		if strings.HasSuffix(fileName, extension) {
			return true
		}
	}
	return false
}

type Resource struct {
	Name  string
	Bytes []byte
}

func loadResources(dir string, resources []Resource) []Resource {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fullPath := fmt.Sprintf("%s%c%s", dir, os.PathSeparator, file.Name())
		if file.IsDir() {
			resources = loadResources(fullPath, resources)
		} else {
			fileName := file.Name()
			if isResourceFileName(fileName) {
				bytes, err := ioutil.ReadFile(fullPath)
				if err != nil {
					log.Fatal(err)
				}
				resources = append(resources, Resource{fileName, bytes})
			}
		}
	}
	return resources
}

func main() {
	flag.Parse()
	fmt.Printf("package resources\n\n")
	fmt.Printf("var Constants = map[string][]byte{\n")
	resources := make([]Resource, 0)
	resources = loadResources(baseDir, resources)
	for _, resource := range resources {
		fmt.Printf("\t\t\"%s\": {", resource.Name)
		for _, b := range resource.Bytes {
			fmt.Printf(" %d,", b)
		}
		fmt.Printf("},\n")
	}
	fmt.Printf("\t}\n")
}
