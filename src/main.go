package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// TODO: Split this off into some sort of model package
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
	LineNumber int
	Contents   string
}

func (alias Alias) PrintVerbose(w io.Writer) {
	fmt.Fprintf(w, "<p>Branch: \"%s\",\tRevision: \"%s\"\n", alias.Branch, string(alias.Revision))
	fmt.Fprintf(w, "<ul>\n")
	for _, todoLine := range alias.Revision.LoadTodos() {
		fmt.Fprintf(w, "<li>\"%s\"</li>\n", todoLine.Contents)
	}
	fmt.Fprintf(w, "</ul>\n")
}

func (revision Revision) getSubject() string {
	out, err := exec.Command("git", "show", string(revision), "--format=\"format:%s\"", "-s").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func (revision Revision) getAuthorName() string {
	out, err := exec.Command("git", "show", string(revision), "--format=\"format:%an\"", "-s").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func (revision Revision) getAuthorEmail() string {
	out, err := exec.Command("git", "show", string(revision), "--format=\"format:%ae\"", "-s").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func (revision Revision) getTimestamp() int64 {
	out, err := exec.Command("git", "show", string(revision), "--format=\"format:%ae\"", "-s").Output()
	if err != nil {
		log.Fatal(err)
	}
	timestamp, err := strconv.ParseInt(string(out), 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return timestamp
}

func (revision Revision) GetMetadata() *RevisionMetadata {
	return &RevisionMetadata{
		Revision:    revision,
		Timestamp:   revision.getTimestamp(),
		Subject:     revision.getSubject(),
		AuthorName:  revision.getAuthorName(),
		AuthorEmail: revision.getAuthorEmail(),
	}
}

// TODO: Create a package that wraps all of the calls to git commands
func (revision Revision) Load() *RevisionContents {
	out, err := exec.Command("git", "ls-tree", "-r", string(revision)).Output()
	if err != nil {
		log.Fatal(err)
	}
	listOutput := strings.Trim(string(out), "\n ")
	lines := strings.Split(listOutput, "\n")
	paths := make([]string, len(lines))
	for index, line := range lines {
		line = strings.Replace(lines[index], "\t", " ", -1)
		lineParts := strings.Split(line, " ")
		paths[index] = lineParts[len(lineParts)-1]
	}
	return &RevisionContents{revision, paths}
}

func parseBranchListLine(line string) Alias {
	line = strings.Trim(line, "* ")
	splitLine := strings.Split(line, " ")
	masterName := splitLine[0]
	for _, lineComponent := range splitLine[1:] {
		if len(lineComponent) == 40 {
			revisionHash := lineComponent
			return Alias{masterName, Revision(revisionHash)}
		}
	}
	return Alias{Branch: masterName}
}

func ListBranches() []Alias {
	out, err := exec.Command("git", "branch", "-av", "--list", "--abbrev=40", "--no-color").Output()
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(strings.Trim(string(out), " \n"), "\n")
	aliases := make([]Alias, len(lines))
	index := 0
	for _, line := range lines {
		if line != "" {
			aliases[index] = parseBranchListLine(line)
			index += 1
		}
	}
	return aliases
}

func (revision Revision) readLines(path string) []Line {
	out, err := exec.Command("git", "blame", "-s", "--abbrev=40", string(revision), path).Output()
	if err != nil {
		log.Fatal(err)
	}
	blameOutput := strings.Trim(string(out), "\n ")
	lines := strings.Split(blameOutput, "\n")
	result := make([]Line, len(lines))
	for _, line := range lines {
		revision := Revision(line[0:40])
		lineNumberIndex := strings.Index(line, ")")
		lineNumber, err := strconv.Atoi(strings.Trim(line[41:lineNumberIndex], " "))
		if err != nil {
			log.Fatal(err)
		}
		contents := line[lineNumberIndex+2:]
		result = append(result, Line{revision, lineNumber, contents})
	}
	return result
}

const (
	TodoRegex = "[^[:alpha:]](t|T)(o|O)(d|D)(o|O)[^[:alpha:]]"
)

func (revision *Revision) LoadTodos() []Line {
	todos := make([]Line, 0)
	for _, path := range revision.Load().Paths {
		for _, line := range revision.readLines(path) {
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
	for _, alias := range ListBranches() {
		alias.PrintVerbose(w)
	}
	fmt.Fprintf(w, "</body>")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
