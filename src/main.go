package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"repo"
)

func serveRepoDetails(repository repo.Repository) {
	http.HandleFunc("/aliases",
		func(w http.ResponseWriter, r *http.Request) {
			err := repo.WriteJson(w, repository)
			if err != nil {
				log.Fatal(err)
			}
		})
	http.HandleFunc("/revision",
		func(w http.ResponseWriter, r *http.Request) {
			revisionParam := r.URL.Query().Get("id")
			if revisionParam != "" {
				revision := repo.Revision(revisionParam)
				err := repo.WriteTodosJson(w, repository, revision)
				if err != nil {
					log.Fatal(err)
				}
			}
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
	http.ListenAndServe(":8080", nil)
}

func main() {
	gitRepository := repo.GitRepository{}
	serveRepoDetails(gitRepository)
}
