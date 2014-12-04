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

package dashboard

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"repo"
	"resources"
	"strconv"
)

const (
	fileContentsResource = "file_contents.html"
)

type Dashboard struct {
	Repository   repo.Repository
	TodoRegex    string
	ExcludePaths string
}

func (db Dashboard) readRevisionAndPathParams(r *http.Request) (repo.Revision, string, error) {
	revisionParam := r.URL.Query().Get("revision")
	if revisionParam == "" {
		return repo.Revision(""), "", errors.New("Missing the revision parameter")
	}
	fileName, err := url.QueryUnescape(r.URL.Query().Get("fileName"))
	if err != nil || fileName == "" {
		return repo.Revision(""), "", errors.New("Missing the fileName parameter")
	}
	revision, err := db.Repository.ValidateRevision(revisionParam)
	if err != nil {
		return repo.Revision(""), "", err
	}
	return revision, fileName, nil
}

// Serve the aliases JSON for a repo.
func (db Dashboard) ServeAliasesJson(w http.ResponseWriter, r *http.Request) {
	err := repo.WriteJson(w, db.Repository)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Server error \"%s\"", err)
	}
}

// Serve the JSON for a single revision.
// The ID of the revision is taken from the URL parameters of the request.
func (db Dashboard) ServeRevisionJson(w http.ResponseWriter, r *http.Request) {
	revisionParam := r.URL.Query().Get("id")
	if revisionParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing required parameter 'id'")
		return
	}
	revision, err := db.Repository.ValidateRevision(revisionParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid revision: %s", revisionParam)
		return
	}
	err = repo.WriteTodosJson(
		w, db.Repository, revision, db.TodoRegex, db.ExcludePaths)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Server error \"%s\"", err)
	}
}

// Serve the details JSON for a single TODO.
// The revision, path, and line number are all taken from the URL parameters of the request.
func (db Dashboard) ServeTodoJson(w http.ResponseWriter, r *http.Request) {
	revision, fileName, err := db.readRevisionAndPathParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}
	lineNumberParam := r.URL.Query().Get("lineNumber")
	if lineNumberParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Missing the lineNumber param")
		return
	}
	lineNumber, err := strconv.Atoi(lineNumberParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid format for the lineNumber parameter: %s", err)
		return
	}
	todoId := repo.TodoId{
		Revision:   revision,
		FileName:   fileName,
		LineNumber: lineNumber,
	}
	repo.WriteTodoDetailsJson(w, db.Repository, todoId)
}

// Serve the redirect for browsing a file.
// The revision, path, and line number are all taken from the URL parameters of the request.
func (db Dashboard) ServeBrowseRedirect(w http.ResponseWriter, r *http.Request) {
	revision, fileName, err := db.readRevisionAndPathParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}
	lineNumberParam := r.URL.Query().Get("lineNumber")
	if lineNumberParam == "" {
		lineNumberParam = "1"
	}
	lineNumber, err := strconv.Atoi(lineNumberParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid format for the lineNumber parameter: %s", err)
		return
	}
	http.Redirect(w, r, db.Repository.GetBrowseUrl(
		revision, fileName, lineNumber), http.StatusMovedPermanently)
}

// Serve the contents for a single file.
// The revision, path, and line number are all taken from the URL parameters of the request.
func (db Dashboard) ServeFileContents(w http.ResponseWriter, r *http.Request) {
	htmlTemplate, err := template.New("fileContentsTemplate").Parse(
		string(resources.Constants[fileContentsResource]))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Server error \"%s\"", err)
	}
	revision, fileName, err := db.readRevisionAndPathParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}
	contents := db.Repository.ReadFileSnippetAtRevision(revision, fileName, 1, -1)
	err = htmlTemplate.Execute(w, contents)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Server error \"%s\"", err)
	}
}
