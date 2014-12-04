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

package repotest

import "errors"
import "fmt"
import "repo"

type MockRepository struct {
	Aliases       []repo.Alias
	RevisionTodos map[string][]repo.Line
}

func (repository MockRepository) ListBranches() []repo.Alias {
	return repository.Aliases
}

func (repository MockRepository) ReadRevisionContents(revision repo.Revision) *repo.RevisionContents {
	return &repo.RevisionContents{
		Revision: revision,
		Paths:    make([]string, 0),
	}
}

func (repository MockRepository) ReadRevisionMetadata(revision repo.Revision) repo.RevisionMetadata {
	return repo.RevisionMetadata{
		Revision: revision,
	}
}

func (repository MockRepository) ReadFileSnippetAtRevision(revision repo.Revision, path string, startLine, endLine int) string {
	return ""
}

func (repository MockRepository) LoadRevisionTodos(revision repo.Revision, todoRegex, excludePaths string) []repo.Line {
	return repository.RevisionTodos[string(revision)]
}

func (repository MockRepository) LoadFileTodos(revision repo.Revision, path string, todoRegex string) []repo.Line {
	return make([]repo.Line, 0)
}

func (repository MockRepository) GetBrowseUrl(revision repo.Revision, path string, lineNumber int) string {
	return ""
}

func (repository MockRepository) ValidateRevision(revisionString string) (repo.Revision, error) {
	if _, ok := repository.RevisionTodos[revisionString]; ok {
		return repo.Revision(revisionString), nil
	}
	return repo.Revision(""), errors.New(fmt.Sprintf("Not a valid revision: %s", revisionString))
}

func (repository MockRepository) ValidatePathAtRevision(revision repo.Revision, path string) error {
	return nil
}
