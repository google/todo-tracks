package repo

import (
	"encoding/json"
	"io"
	"regexp"
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

type Repository interface {
	ListBranches() []Alias
	ReadRevisionContents(revision Revision) *RevisionContents
	ReadRevisionMetadata(revision Revision) RevisionMetadata
	ReadFileAtRevision(revision Revision, path string) []Line
}

func WriteJson(w io.Writer, repository Repository) error {
	bytes, err := json.Marshal(repository.ListBranches())
	if err != nil {
		return err
	}
	w.Write(bytes)
	return nil
}

const (
	// TODO: Make this configurable.
	TodoRegex = "[^[:alpha:]](t|T)(o|O)(d|D)(o|O)[^[:alpha:]]"
)

// TODO: Return a slice of TodoId instead of Line.
func LoadTodos(repository Repository, revision Revision) []Line {
	todos := make([]Line, 0)
	for _, path := range repository.ReadRevisionContents(revision).Paths {
		for _, line := range repository.ReadFileAtRevision(revision, path) {
			matched, err := regexp.MatchString(TodoRegex, line.Contents)
			if err == nil && matched {
				todos = append(todos, line)
			}
		}
	}
	return todos
}

func WriteTodosJson(w io.Writer, repository Repository, revision Revision) error {
	bytes, err := json.Marshal(LoadTodos(repository, revision))
	if err != nil {
		return err
	}
	w.Write(bytes)
	return nil
}

// TODO: Add a method for getting a JSON blob of the TodoDetails given a TodoId.
