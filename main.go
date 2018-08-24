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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/todo-tracks/dashboard"
	"github.com/google/todo-tracks/repo"
	"github.com/google/todo-tracks/resources"
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

func serveDashboard(dashboard dashboard.Dashboard) {
	http.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		resourceName := r.URL.Path[4:]
		serveStaticContent(w, resourceName)
	})
	http.HandleFunc("/repos", dashboard.ServeReposJson)
	http.HandleFunc("/aliases", dashboard.ServeAliasesJson)
	http.HandleFunc("/revision", dashboard.ServeRevisionJson)
	http.HandleFunc("/todo", dashboard.ServeTodoJson)
	http.HandleFunc("/todoStatus", dashboard.ServeTodoStatusJson)
	http.HandleFunc("/browse", dashboard.ServeBrowseRedirect)
	http.HandleFunc("/raw", dashboard.ServeFileContents)
	http.HandleFunc("/_ah/health",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "ok")
		})
	http.HandleFunc("/", dashboard.ServeMainPage)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// Find all local repositories under the current working directory.
func getLocalRepos() (map[string]*repo.Repository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	repos := make(map[string]*repo.Repository)
	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dir, err := os.Open(path)
			if err != nil {
				return err
			}
			children, err := dir.Readdir(-1)
			if err != nil {
				return err
			}
			for _, child := range children {
				if child.IsDir() && child.Name() == ".git" {
					gitRepo := repo.NewGitRepository(path, todoRegex, excludePaths)
					repos[gitRepo.GetRepoId()] = &gitRepo
					return filepath.SkipDir
				}
			}
		}
		return nil
	})
	return repos, nil
}

func main() {
	flag.Parse()
	repos, err := getLocalRepos()
	if err != nil {
		log.Fatal(err.Error())
	}
	if repos == nil {
		log.Fatal("Unable to find any local repositories under the current directory")
	}
	serveDashboard(dashboard.Dashboard{repos, todoRegex, excludePaths})
}
