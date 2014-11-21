package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"repo"
	"resources"
	"strconv"
	"strings"
)

var port int
var todoRegex string
var excludePaths string

func init() {
	flag.IntVar(&port, "port", 8080, "Port on which to start the server.")
	flag.StringVar(
		&todoRegex,
		"todo_regex",
		"(^|[^[:alpha:]])(t|T)(o|O)(d|D)(o|O)[^[:alpha:]]",
		"Regular expression (using the re2 syntax) to use when matching TODOs.")
	flag.StringVar(
		&excludePaths,
		"exclude_paths",
		"",
		"List of file paths to exclude when matching TODOs. This is useful if your repo contains binaries")
}

func serveStaticContent(w http.ResponseWriter, resourceName string) {
	resourceContents := resources.Constants[resourceName]
	var contentType string
	if strings.HasSuffix(resourceName, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(resourceName, ".html") {
		contentType = "text/html"
	} else if strings.HasSuffix(resourceName, ".js") {
		contentType = "text/javascript"
	} else {
		contentType = http.DetectContentType(resourceContents)
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(resourceContents)
}

func readRevisionAndPathParams(r *http.Request) (repo.Revision, string, error) {
	revisionParam := r.URL.Query().Get("revision")
	if revisionParam == "" {
		return repo.Revision(""), "", errors.New("Missing the revision parameter")
	}
	fileName, err := url.QueryUnescape(r.URL.Query().Get("fileName"))
	if err != nil || fileName == "" {
		return repo.Revision(""), "", errors.New("Missing the fileName parameter")
	}
	return repo.Revision(revisionParam), fileName, nil
}

func serveRepoDetails(repository repo.Repository) {
	http.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		resourceName := r.URL.Path[4:]
		serveStaticContent(w, resourceName)
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
			err := repo.WriteTodosJson(w, repository, revision, todoRegex, excludePaths)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Server error \"%s\"", err)
			}
		})
	http.HandleFunc("/todo",
		func(w http.ResponseWriter, r *http.Request) {
			revision, fileName, err := readRevisionAndPathParams(r)
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, err.Error())
				return
			}
			lineNumberParam := r.URL.Query().Get("lineNumber")
			if lineNumberParam == "" {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Missing the lineNumber param")
				return
			}
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
	http.HandleFunc("/browse",
		func(w http.ResponseWriter, r *http.Request) {
			revision, fileName, err := readRevisionAndPathParams(r)
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, err.Error())
				return
			}
			lineNumberParam := r.URL.Query().Get("lineNumber")
			if lineNumberParam == "" {
				lineNumberParam = "1"
			}
			lineNumber, err := strconv.Atoi(lineNumberParam)
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Invalid format for the lineNumber parameter: %s", err)
				return
			}
			http.Redirect(w, r, repository.GetBrowseUrl(
				revision, fileName, lineNumber), http.StatusMovedPermanently)
		})
	http.HandleFunc("/raw",
		func(w http.ResponseWriter, r *http.Request) {
			revision, fileName, err := readRevisionAndPathParams(r)
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, err.Error())
				return
			}
			contents := repository.ReadFileSnippetAtRevision(revision, fileName, 1, -1)
			w.Write([]byte(contents))
		})
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/ui/list_branches.html", http.StatusMovedPermanently)
		})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func main() {
	flag.Parse()
	// TODO: Add some sanity checking that the binary was started inside of a git repo directory.
	gitRepository := repo.NewGitRepository(todoRegex, excludePaths)
	serveRepoDetails(gitRepository)
}
