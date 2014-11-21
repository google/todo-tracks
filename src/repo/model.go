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
	// TODO(ojarjur): Add a list of branches from which the TODO is missing (not yet added)
	// TODO(ojarjur): Add a list of branches that have the TODO
	// TODO(ojarjur): Add a list of branches from which the TODO has been removed
}

type Repository interface {
	ListBranches() []Alias
	ReadRevisionContents(revision Revision) *RevisionContents
	ReadRevisionMetadata(revision Revision) RevisionMetadata
	ReadFileSnippetAtRevision(revision Revision, path string, startLine, endLine int) string
	LoadRevisionTodos(revision Revision, todoRegex, excludePaths string) []Line
	LoadFileTodos(revision Revision, path string, todoRegex string) []Line
	GetBrowseUrl(revision Revision, path string, lineNumber int) string
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
	context := repository.ReadFileSnippetAtRevision(todoId.Revision, todoId.FileName, startLine, endLine)
	return &TodoDetails{
		Id:               todoId,
		RevisionMetadata: repository.ReadRevisionMetadata(todoId.Revision),
		Context:          context,
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
