package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var baseDir string

func init() {
	flag.StringVar(&baseDir, "base_dir", "./", "Directory under which to look for resource files")
}

func isResourceFile(file os.FileInfo) bool {
	if file.IsDir() {
		return false
	}
	fileName := file.Name()
	return strings.HasSuffix(fileName, ".html") || strings.HasSuffix(fileName, ".js")
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
		if file.IsDir() {
			childDir := fmt.Sprintf("%s%c%s", dir, os.PathSeparator, file.Name())
			resources = loadResources(childDir, resources)
		} else {
			fileName := file.Name()
			if strings.HasSuffix(fileName, ".html") || strings.HasSuffix(fileName, ".js") {
				bytes, err := ioutil.ReadFile(file.Name())
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
	fmt.Printf("var Contents = map[string][]byte{\n")
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
