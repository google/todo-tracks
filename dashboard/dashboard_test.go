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

package dashboard_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/google/todo-tracks/dashboard"
	"github.com/google/todo-tracks/repo"
	"github.com/google/todo-tracks/repo/repotest"
)

const (
	TestRevision     = "testRevision"
	TestFileName     = "testFile"
	TestLineNumber   = 42
	TestTodoContents = "TODO: test this"
)

var mockAlias repo.Alias
var mockTodo repo.Line
var mockRepo repo.Repository
var mockRepos map[string]*repo.Repository

func init() {
	mockAlias = repo.Alias{"branch", repo.Revision("revision")}
	mockTodo = repo.Line{
		Revision:   repo.Revision(TestRevision),
		FileName:   TestFileName,
		LineNumber: TestLineNumber,
		Contents:   TestTodoContents,
	}

	aliases := make([]repo.Alias, 0)
	aliases = append(aliases, mockAlias)
	revisionTodos := make(map[string][]repo.Line)
	revisionTodos[TestRevision] =
		append(revisionTodos[TestRevision], mockTodo)
	mockRepo = repotest.MockRepository{
		Aliases:       aliases,
		RevisionTodos: revisionTodos,
	}
	mockRepos = make(map[string]*repo.Repository)
	mockRepos[mockRepo.GetRepoId()] = &mockRepo
}

func TestServeAliasesJsonNoRepo(t *testing.T) {
	request, err := http.NewRequest("GET", "/", strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	rw := httptest.NewRecorder()
	db := dashboard.Dashboard{mockRepos, "", ""}
	db.ServeAliasesJson(rw, request)
	if rw.Code != http.StatusOK {
		t.Errorf("Expected a response code of %d, but saw %d, with a body of '%s'",
			http.StatusOK, rw.Code, rw.Body.String())
		return
	}
	var returnedAliases []repo.Alias
	err = json.Unmarshal(rw.Body.Bytes(), &returnedAliases)
	if err != nil {
		t.Error(err)
	}
	if len(returnedAliases) != 1 || returnedAliases[0] != mockAlias {
		t.Errorf("Expected a singleton slice of %s, but saw %s", mockAlias, returnedAliases)
	}
}

func TestServeAliasesJson(t *testing.T) {
	request, err := http.NewRequest("GET", "/?repo="+mockRepo.GetRepoId(), strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	rw := httptest.NewRecorder()
	db := dashboard.Dashboard{mockRepos, "", ""}
	db.ServeAliasesJson(rw, request)
	if rw.Code != http.StatusOK {
		t.Errorf("Expected a response code of %d, but saw %d, with a body of '%s'",
			http.StatusOK, rw.Code, rw.Body.String())
		return
	}
	var returnedAliases []repo.Alias
	err = json.Unmarshal(rw.Body.Bytes(), &returnedAliases)
	if err != nil {
		t.Error(err)
	}
	if len(returnedAliases) != 1 || returnedAliases[0] != mockAlias {
		t.Errorf("Expected a singleton slice of %s, but saw %s", mockAlias, returnedAliases)
	}
}

func TestServeRevisionJsonNoRepo(t *testing.T) {
	request, err := http.NewRequest("GET", "/revision", strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	rw := httptest.NewRecorder()
	db := dashboard.Dashboard{mockRepos, "", ""}
	db.ServeRevisionJson(rw, request)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected a response code of %d, but saw %d", http.StatusBadRequest, rw.Code)
	}
}

func TestServeRevisionJsonNoId(t *testing.T) {
	params := url.Values{}
	params.Add("repo", mockRepo.GetRepoId())
	request, err := http.NewRequest("GET", "/revision?"+params.Encode(), strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	rw := httptest.NewRecorder()
	db := dashboard.Dashboard{mockRepos, "", ""}
	db.ServeRevisionJson(rw, request)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected a response code of %d, but saw %d", http.StatusBadRequest, rw.Code)
	}
}

func TestServeRevisionJson(t *testing.T) {
	params := url.Values{}
	params.Add("repo", mockRepo.GetRepoId())
	params.Add("revision", TestRevision)
	request, err := http.NewRequest("GET", "/revision?"+params.Encode(), strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	rw := httptest.NewRecorder()
	db := dashboard.Dashboard{mockRepos, "", ""}
	db.ServeRevisionJson(rw, request)
	if rw.Code != http.StatusOK {
		t.Errorf("Expected a response code of %d, but saw %d, with a body of '%s'",
			http.StatusOK, rw.Code, rw.Body.String())
		return
	}
	var returnedTodos []repo.Line
	err = json.Unmarshal(rw.Body.Bytes(), &returnedTodos)
	if err != nil {
		t.Error(err)
	}
	if len(returnedTodos) != 1 || returnedTodos[0] != mockTodo {
		t.Errorf("Expected a singleton slice of %v, but saw %v", mockTodo, returnedTodos)
	}
}

func TestServeTodoJsonNoRepo(t *testing.T) {
	request, err := http.NewRequest("GET", "/todo", strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	rw := httptest.NewRecorder()
	db := dashboard.Dashboard{mockRepos, "", ""}
	db.ServeTodoJson(rw, request)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected a response code of %d, but saw %d", http.StatusBadRequest, rw.Code)
	}
}

func TestServeTodoJsonNoRevision(t *testing.T) {
	params := url.Values{}
	params.Add("repo", mockRepo.GetRepoId())
	request, err := http.NewRequest("GET", "/todo?"+params.Encode(), strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	rw := httptest.NewRecorder()
	db := dashboard.Dashboard{mockRepos, "", ""}
	db.ServeTodoJson(rw, request)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected a response code of %d, but saw %d", http.StatusBadRequest, rw.Code)
	}
}

func TestServeTodoJsonNoFileName(t *testing.T) {
	params := url.Values{}
	params.Add("repo", mockRepo.GetRepoId())
	params.Add("revision", TestRevision)
	request, err := http.NewRequest("GET", "/todo?"+params.Encode(), strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	rw := httptest.NewRecorder()
	db := dashboard.Dashboard{mockRepos, "", ""}
	db.ServeTodoJson(rw, request)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected a response code of %d, but saw %d", http.StatusBadRequest, rw.Code)
	}
}

func TestServeTodoJsonNoLineNumber(t *testing.T) {
	params := url.Values{}
	params.Add("repo", mockRepo.GetRepoId())
	params.Add("revision", TestRevision)
	params.Add("fileName", TestFileName)
	request, err := http.NewRequest("GET", "/todo?"+params.Encode(), strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	rw := httptest.NewRecorder()
	db := dashboard.Dashboard{mockRepos, "", ""}
	db.ServeTodoJson(rw, request)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected a response code of %d, but saw %d", http.StatusBadRequest, rw.Code)
	}
}

func TestServeTodoJsonInvalidLineNumber(t *testing.T) {
	params := url.Values{}
	params.Add("repo", mockRepo.GetRepoId())
	params.Add("revision", TestRevision)
	params.Add("fileName", TestFileName)
	params.Add("lineNumber", "fortyTwo")
	request, err := http.NewRequest("GET", "/todo?"+params.Encode(), strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	rw := httptest.NewRecorder()
	db := dashboard.Dashboard{mockRepos, "", ""}
	db.ServeTodoJson(rw, request)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected a response code of %d, but saw %d", http.StatusBadRequest, rw.Code)
	}
}

func TestServeTodoJson(t *testing.T) {
	params := url.Values{}
	params.Add("repo", mockRepo.GetRepoId())
	params.Add("revision", TestRevision)
	params.Add("fileName", TestFileName)
	params.Add("lineNumber", strconv.Itoa(TestLineNumber))
	request, err := http.NewRequest("GET", "/todo?"+params.Encode(), strings.NewReader(""))
	if err != nil {
		t.Error(err)
	}
	rw := httptest.NewRecorder()
	db := dashboard.Dashboard{mockRepos, "", ""}
	db.ServeTodoJson(rw, request)
	if rw.Code != http.StatusOK {
		t.Errorf("Expected a response code of %d, but saw %d, with a body of '%s'",
			http.StatusOK, rw.Code, rw.Body.String())
		return
	}
	var returnedTodo repo.TodoDetails
	err = json.Unmarshal(rw.Body.Bytes(), &returnedTodo)
	if err != nil {
		t.Error(err)
	}
	if returnedTodo.Id.Revision != mockTodo.Revision ||
		returnedTodo.Id.FileName != mockTodo.FileName ||
		returnedTodo.Id.LineNumber != mockTodo.LineNumber {
		t.Errorf("Expected %v, but saw %v", mockTodo, returnedTodo)
	}
}
