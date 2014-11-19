package main

import (
	"flag"
	"fmt"
	"html"
	"net/http"
	"repo"
	"strconv"
)

var port int

func init() {
	flag.IntVar(&port, "port", 8080, "Port on which to start the server")
}

func serveRepoDetails(repository repo.Repository) {
	http.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "src"+r.URL.Path)
	})
	http.HandleFunc("/aliases",
		func(w http.ResponseWriter, r *http.Request) {
			err := repo.WriteJson(w, repository)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Server error \"%s\"", err)
			}
		})
	http.HandleFunc("/revision",
		func(w http.ResponseWriter, r *http.Request) {
			revisionParam := r.URL.Query().Get("id")
			if revisionParam == "" {
				w.WriteHeader(400)
				fmt.Fprint(w, "Missing required parameter 'id'")
				return
			}
			revision := repo.Revision(revisionParam)
			err := repo.WriteTodosJson(w, repository, revision)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Server error \"%s\"", err)
			}
		})
	http.HandleFunc("/todo",
		func(w http.ResponseWriter, r *http.Request) {
			revisionParam := r.URL.Query().Get("revision")
			lineNumberParam := r.URL.Query().Get("lineNumber")
			fileName := r.URL.Query().Get("fileName")
			if revisionParam == "" || fileName == "" || lineNumberParam == "" {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Missing at least one required parameter")
				return
			}
			revision := repo.Revision(revisionParam)
			lineNumber, err := strconv.Atoi(lineNumberParam)
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Invalid format for the lineNumber parameter: %s", err)
				return
			}
			todoId := repo.TodoId{
				Revision:   revision,
				FileName:   fileName,
				LineNumber: lineNumber,
			}
			repo.WriteTodoDetailsJson(w, repository, todoId)
		})
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "<body>")
			for _, alias := range repository.ListBranches() {
				fmt.Fprintf(w, "<p>Branch: \"%s\",\tRevision: \"%s\"\n",
					alias.Branch, string(alias.Revision))
				fmt.Fprintf(w, "<ul>\n")
				for _, todoLine := range repo.LoadTodos(repository, alias.Revision) {
					fmt.Fprintf(w,
						"<li>%s[%d]: \"%s\"</li>\n",
						todoLine.FileName,
						todoLine.LineNumber,
						html.EscapeString(todoLine.Contents))
				}
				fmt.Fprintf(w, "</ul>\n")
				fmt.Fprintf(w, "</body>")
			}
		})
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func main() {
	flag.Parse()
	var gitRepository repo.GitRepository
	serveRepoDetails(gitRepository)
}
