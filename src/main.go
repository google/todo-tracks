package main

import (
	"fmt"
	"net/http"
	"regexp"
	"repo"
)

const (
	TodoRegex = "[^[:alpha:]](t|T)(o|O)(d|D)(o|O)[^[:alpha:]]"
)

func LoadTodos(repository repo.Repository, revision repo.Revision) []repo.Line {
	todos := make([]repo.Line, 0)
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

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<body>")
	gitRepository := repo.GitRepository{}
	for _, alias := range gitRepository.ListBranches() {
		fmt.Fprintf(w, "<p>Branch: \"%s\",\tRevision: \"%s\"\n",
			alias.Branch, string(alias.Revision))
		fmt.Fprintf(w, "<ul>\n")
		for _, todoLine := range LoadTodos(gitRepository, alias.Revision) {
			fmt.Fprintf(w, "<li>\"%s\"</li>\n", todoLine.Contents)
		}
		fmt.Fprintf(w, "</ul>\n")
	}
	fmt.Fprintf(w, "</body>")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
