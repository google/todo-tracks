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
	"dashboard"
	"flag"
	"fmt"
	"log"
	"net/http"
	"repo"
	"resources"
	"strings"
)

const (
	fileContentsResource = "file_contents.html"
)

var port int
var todoRegex string
var excludePaths string

func init() {
	flag.IntVar(&port, "port", 8080, "Port on which to start the server.")
	flag.StringVar(
		&todoRegex,
		"todo_regex",
		"(^|[^[:alpha:]])(t|T)(o|O)(d|D)(o|O)[^[:alpha:]]",
		"Regular expression (using the re2 syntax) to use when matching TODOs.")
	flag.StringVar(
		&excludePaths,
		"exclude_paths",
		"",
		"Comma-separated list of file paths to exclude when matching TODOs. Each path is specified as a regular expression using the re2 syntax.")
}

func serveStaticContent(w http.ResponseWriter, resourceName string) {
	resourceContents := resources.Constants[resourceName]
	var contentType string
	if strings.HasSuffix(resourceName, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(resourceName, ".html") {
		contentType = "text/html"
	} else if strings.HasSuffix(resourceName, ".js") {
		contentType = "text/javascript"
	} else {
		contentType = http.DetectContentType(resourceContents)
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(resourceContents)
}

func serveRepoDetails(dashboard dashboard.Dashboard) {
	http.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		resourceName := r.URL.Path[4:]
		serveStaticContent(w, resourceName)
	})
	http.HandleFunc("/aliases", dashboard.ServeAliasesJson)
	http.HandleFunc("/revision", dashboard.ServeRevisionJson)
	http.HandleFunc("/todo", dashboard.ServeTodoJson)
	http.HandleFunc("/browse", dashboard.ServeBrowseRedirect)
	http.HandleFunc("/raw", dashboard.ServeFileContents)
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/ui/list_branches.html", http.StatusMovedPermanently)
		})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func main() {
	flag.Parse()
	// TODO: Add some sanity checking that the binary was started inside of a git repo directory.
	gitRepository := repo.NewGitRepository(todoRegex, excludePaths)
	dashboard := dashboard.Dashboard{gitRepository, todoRegex, excludePaths}
	serveRepoDetails(dashboard)
}
