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
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/google/todo-tracks/repo"
	"github.com/google/todo-tracks/resources"
)

const (
	fileContentsResource = "file_contents.html"
)

type Dashboard struct {
	Repositories map[string]*repo.Repository
	TodoRegex    string
	ExcludePaths string
}

func (db Dashboard) readRepoParam(r *http.Request) (*repo.Repository, error) {
	repoParam := r.URL.Query().Get("repo")
	if repoParam == "" {
		if len(db.Repositories) != 1 {
			return nil, errors.New("Missing the repo parameter")
		}
		for key := range db.Repositories {
			repoParam = key
			break
		}
	}
	repository := db.Repositories[repoParam]
	if repository == nil {
		return nil, errors.New(fmt.Sprintf("Unknown repo '%s'", repoParam))
	}
	return repository, nil
}

func (db Dashboard) readRepoAndRevisionParams(r *http.Request) (*repo.Repository, repo.Revision, error) {
	repository, err := db.readRepoParam(r)
	if err != nil {
		return nil, repo.Revision(""), err
	}
	revisionParam := r.URL.Query().Get("revision")
	if revisionParam == "" {
		return nil, repo.Revision(""), errors.New("Missing the revision parameter")
	}
	revision, err := (*repository).ValidateRevision(revisionParam)
	if err != nil {
		return nil, repo.Revision(""), err
	}
	return repository, revision, nil
}

func (db Dashboard) readRepoRevisionAndPathParams(r *http.Request) (*repo.Repository, repo.Revision, string, error) {
	repository, revision, err := db.readRepoAndRevisionParams(r)
	if err != nil {
		return nil, repo.Revision(""), "", err
	}
	fileName, err := url.QueryUnescape(r.URL.Query().Get("fileName"))
	if err != nil || fileName == "" {
		return nil, repo.Revision(""), "", errors.New("Missing the fileName parameter")
	}
	err = (*repository).ValidatePathAtRevision(revision, fileName)
	return repository, revision, fileName, err
}

func (db Dashboard) readRepoRevisionPathAndLineNumberParams(r *http.Request) (*repo.Repository, repo.Revision, string, int, error) {
	repository, revision, fileName, err := db.readRepoRevisionAndPathParams(r)
	if err != nil {
		return nil, repo.Revision(""), "", 0, err
	}
	lineNumberParam := r.URL.Query().Get("lineNumber")
	if lineNumberParam == "" {
		return nil, repo.Revision(""), "", 0, errors.New("Missing the lineNumber param")
	}
	lineNumber, err := strconv.Atoi(lineNumberParam)
	if err != nil {
		return nil, repo.Revision(""), "", 0, fmt.Errorf("Invalid format for the lineNumber parameter: %v", err)
	}
	err = (*repository).ValidateLineNumberInPathAtRevision(revision, fileName, lineNumber)
	return repository, revision, fileName, lineNumber, err
}

// Serve the main page.
func (db Dashboard) ServeMainPage(w http.ResponseWriter, r *http.Request) {
	if len(db.Repositories) == 1 {
		for repoId := range db.Repositories {
			http.Redirect(w, r,
				"/ui/list_branches.html#?repo="+repoId,
				http.StatusMovedPermanently)
			return
		}
	} else {
		http.Redirect(w, r, "/ui/list_repos.html", http.StatusMovedPermanently)
	}
}

// Serve the aliases JSON for a repo.
func (db Dashboard) ServeAliasesJson(w http.ResponseWriter, r *http.Request) {
	repositoryPtr, err := db.readRepoParam(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error loading repo: \"%s\"", err)
		return
	}
	repository := *repositoryPtr
	err = repo.WriteJson(w, repository)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Server error \"%s\"", err)
	}
}

// Serve the JSON for a single revision.
// The ID of the revision is taken from the URL parameters of the request.
func (db Dashboard) ServeRevisionJson(w http.ResponseWriter, r *http.Request) {
	repositoryPtr, revision, err := db.readRepoAndRevisionParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}
	repository := *repositoryPtr
	err = repo.WriteTodosJson(
		w, repository, revision, db.TodoRegex, db.ExcludePaths)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Server error \"%s\"", err)
	}
}

// Serve the details JSON for a single TODO.
// The revision, path, and line number are all taken from the URL parameters of the request.
func (db Dashboard) ServeTodoJson(w http.ResponseWriter, r *http.Request) {
	repositoryPtr, revision, fileName, lineNumber, err := db.readRepoRevisionPathAndLineNumberParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	repository := *repositoryPtr
	todoId := repo.TodoId{
		Revision:   revision,
		FileName:   fileName,
		LineNumber: lineNumber,
	}
	repo.WriteTodoDetailsJson(w, repository, todoId)
}

// Serve the status details JSON for a single TODO.
// The revision, path, and line number are all taken from the URL parameters of the request.
func (db Dashboard) ServeTodoStatusJson(w http.ResponseWriter, r *http.Request) {
	repositoryPtr, revision, fileName, lineNumber, err := db.readRepoRevisionPathAndLineNumberParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	repository := *repositoryPtr
	todoId := repo.TodoId{
		Revision:   revision,
		FileName:   fileName,
		LineNumber: lineNumber,
	}
	repo.WriteTodoStatusDetailsJson(w, repository, todoId)
}

// Serve the redirect for browsing a file.
// The revision, path, and line number are all taken from the URL parameters of the request.
func (db Dashboard) ServeBrowseRedirect(w http.ResponseWriter, r *http.Request) {
	repositoryPtr, revision, fileName, err := db.readRepoRevisionAndPathParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}
	repository := *repositoryPtr
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
	err = repository.ValidateLineNumberInPathAtRevision(revision, fileName, lineNumber)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}
	http.Redirect(w, r, repository.GetBrowseUrl(
		revision, fileName, lineNumber), http.StatusMovedPermanently)
}

type fileContents struct {
	LineNumber int
	Contents   string
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
	repositoryPtr, revision, fileName, lineNumber, err := db.readRepoRevisionPathAndLineNumberParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, err.Error())
		return
	}
	repository := *repositoryPtr
	contents := repository.ReadFileSnippetAtRevision(revision, fileName, 1, -1)
	err = htmlTemplate.Execute(w, fileContents{
		LineNumber: lineNumber,
		Contents:   contents})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Server error \"%s\"", err)
	}
}

type repoPath struct {
	Path   string
	RepoId string
}
type sortByPath []repoPath

func (rs sortByPath) Len() int           { return len(rs) }
func (rs sortByPath) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }
func (rs sortByPath) Less(i, j int) bool { return rs[i].Path < rs[j].Path }

func (db Dashboard) ServeReposJson(w http.ResponseWriter, r *http.Request) {
	repoPaths := make([]repoPath, 0)
	for repoId, repository := range db.Repositories {
		repoPaths = append(repoPaths, repoPath{(*repository).GetRepoPath(), repoId})
	}
	sort.Sort(sortByPath(repoPaths))

	reposJson, err := json.Marshal(repoPaths)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Server error \"%s\"", err)
	}
	w.Write(reposJson)
}
