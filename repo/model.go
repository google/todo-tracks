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

package repo

import (
	"encoding/json"
	"io"
)

type Revision string
type RevisionContents struct {
	Revision Revision
	Paths    []string
}

type RevisionMetadata struct {
	Revision    Revision
	Timestamp   int64
	Subject     string
	AuthorName  string
	AuthorEmail string
}

type Alias struct {
	Branch   string
	Revision Revision
	// TODO: Add LastModified and LastModifiedBy fields based on the RevisionMetadata
}

type Line struct {
	Revision   Revision
	FileName   string
	LineNumber int
	Contents   string
}

// Key that uniquely identifies a TODO.
type TodoId struct {
	Revision   Revision
	FileName   string
	LineNumber int
}

type TodoDetails struct {
	Id               TodoId
	RevisionMetadata RevisionMetadata
	Context          string
}

type TodoStatus struct {
	BranchesMissing []Alias
	BranchesPresent []Alias
	BranchesRemoved []Alias
}

type Repository interface {
	// Get an opaque ID that uniquely identifies this repo on this machine.
	GetRepoId() string
	// Get the path to this repo on this machine.
	GetRepoPath() string

	ListBranches() []Alias
	IsAncestor(ancestor, descendant Revision) bool
	ReadRevisionContents(revision Revision) *RevisionContents
	ReadRevisionMetadata(revision Revision) RevisionMetadata
	ReadFileSnippetAtRevision(revision Revision, path string, startLine, endLine int) string
	LoadRevisionTodos(revision Revision, todoRegex, excludePaths string) []Line
	LoadFileTodos(revision Revision, path string, todoRegex string) []Line
	FindClosingRevisions(todoId TodoId) []Revision

	GetBrowseUrl(revision Revision, path string, lineNumber int) string

	// Check that the given string is a valid revision.
	// This is intended for user input validation.
	ValidateRevision(revisionString string) (Revision, error)

	// Check that the given path is in the given revision.
	// This is intended for user input validation, and assumes that ValidateRevision
	// has already been called.
	ValidatePathAtRevision(revision Revision, path string) error

	// Check that the given line number exists for the given path in the given revision.
	// This is intended for user input validation, and assumes that ValidatePathAtRevision
	// has already been called.
	ValidateLineNumberInPathAtRevision(revision Revision, path string, lineNumber int) error
}

func WriteJson(w io.Writer, repository Repository) error {
	bytes, err := json.Marshal(repository.ListBranches())
	if err != nil {
		return err
	}
	w.Write(bytes)
	return nil
}

func LoadTodoDetails(repository Repository, todoId TodoId, linesBefore int, linesAfter int) *TodoDetails {
	startLine := todoId.LineNumber - linesBefore
	endLine := todoId.LineNumber + linesAfter + 1
	context := repository.ReadFileSnippetAtRevision(
		todoId.Revision, todoId.FileName, startLine, endLine)
	return &TodoDetails{
		Id:               todoId,
		RevisionMetadata: repository.ReadRevisionMetadata(todoId.Revision),
		Context:          context,
	}
}

func LoadTodoStatus(repository Repository, todoId TodoId) *TodoStatus {
	closingRevs := repository.FindClosingRevisions(todoId)
	missing := make([]Alias, 0)
	present := make([]Alias, 0)
	removed := make([]Alias, 0)
Branches:
	for _, alias := range repository.ListBranches() {
		if alias.Revision == todoId.Revision {
			present = append(present, alias)
		} else if repository.IsAncestor(todoId.Revision, alias.Revision) {
			for _, closingRev := range closingRevs {
				if repository.IsAncestor(closingRev, alias.Revision) {
					removed = append(removed, alias)
					continue Branches
				}
			}
			present = append(present, alias)
		} else {
			missing = append(missing, alias)
		}
	}
	return &TodoStatus{
		BranchesMissing: missing,
		BranchesPresent: present,
		BranchesRemoved: removed,
	}
}

func WriteTodosJson(w io.Writer, repository Repository, revision Revision, todoRegex, excludePaths string) error {
	bytes, err := json.Marshal(repository.LoadRevisionTodos(revision, todoRegex, excludePaths))
	if err != nil {
		return err
	}
	w.Write(bytes)
	return nil
}

func WriteTodoDetailsJson(w io.Writer, repository Repository, todoId TodoId) error {
	// TODO: Make the lines before and after a parameter.
	bytes, err := json.Marshal(LoadTodoDetails(repository, todoId, 5, 5))
	if err != nil {
		return err
	}
	w.Write(bytes)
	return nil
}

func WriteTodoStatusDetailsJson(w io.Writer, repository Repository, todoId TodoId) error {
	bytes, err := json.Marshal(LoadTodoStatus(repository, todoId))
	if err != nil {
		return err
	}
	w.Write(bytes)
	return nil
}
