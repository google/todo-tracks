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
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	hashFormat = "^([[:xdigit:]]){40}$"
)

var hashRegexp *regexp.Regexp

func init() {
	var err error
	hashRegexp, err = regexp.Compile(hashFormat)
	if err != nil {
		log.Fatal(err)
	}
}

type todosCacheEntry struct {
	Present bool
	Todos   []Line
}

type gitRepository struct {
	BlobTodosCache     map[string]todosCacheEntry
	RevisionTodosCache map[Revision]todosCacheEntry
}

func NewGitRepository(todoRegex, excludePaths string) Repository {
	repository := &gitRepository{
		BlobTodosCache:     make(map[string]todosCacheEntry),
		RevisionTodosCache: make(map[Revision]todosCacheEntry),
	}
	go func() {
		// Pre-load all of the TODOs for the current branches
		for _, alias := range repository.ListBranches() {
			repository.LoadRevisionTodos(alias.Revision, todoRegex, excludePaths)
		}
	}()
	return repository
}

func runGitCommand(cmd *exec.Cmd) (string, error) {
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(out), " \n"), nil
}

func runGitCommandOrDie(cmd *exec.Cmd) string {
	out, err := runGitCommand(cmd)
	if err != nil {
		log.Print(cmd.Args)
		log.Print(out)
		log.Fatal(err)
	}
	return out
}

func splitCommandOutputLine(line string) []string {
	lineParts := make([]string, 0)
	for _, part := range strings.Split(line, " ") {
		if part != "" {
			lineParts = append(lineParts, part)
		}
	}
	return lineParts
}

func (repository *gitRepository) ListBranches() []Alias {
	out := runGitCommandOrDie(
		exec.Command("git", "branch", "-av", "--list", "--abbrev=40", "--no-color"))
	lines := strings.Split(out, "\n")
	aliases := make([]Alias, 0)
	for _, line := range lines {
		line = strings.Trim(line, "* ")
		lineParts := splitCommandOutputLine(line)
		if len(lineParts) >= 2 && len(lineParts[1]) == 40 {
			branch := lineParts[0]
			revision := Revision(lineParts[1])
			aliases = append(aliases, Alias{branch, revision})
		}
	}
	return aliases
}

func (repository *gitRepository) ReadRevisionContents(revision Revision) *RevisionContents {
	out := runGitCommandOrDie(exec.Command("git", "ls-tree", "-r", string(revision)))
	lines := strings.Split(out, "\n")
	paths := make([]string, 0)
	for _, line := range lines {
		line = strings.Replace(line, "\t", " ", -1)
		lineParts := strings.SplitN(line, " ", 4)
		paths = append(paths, lineParts[len(lineParts)-1])
	}
	return &RevisionContents{revision, paths}
}

func (repository *gitRepository) getSubject(revision Revision) string {
	return runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=%s", "-s"))
}

func (repository *gitRepository) getAuthorName(revision Revision) string {
	return runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=%an", "-s"))
}

func (repository *gitRepository) getAuthorEmail(revision Revision) string {
	return runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=%ae", "-s"))
}

func (repository *gitRepository) getTimestamp(revision Revision) int64 {
	out := runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=%ct", "-s"))
	timestamp, err := strconv.ParseInt(out, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return timestamp
}

func (repository *gitRepository) ReadRevisionMetadata(revision Revision) RevisionMetadata {
	return RevisionMetadata{
		Revision:    revision,
		Timestamp:   repository.getTimestamp(revision),
		Subject:     repository.getSubject(revision),
		AuthorName:  repository.getAuthorName(revision),
		AuthorEmail: repository.getAuthorEmail(revision),
	}
}

func (repository *gitRepository) getFileBlobOrDie(revision Revision, path string) string {
	out := runGitCommandOrDie(exec.Command("git", "ls-tree", "-r", string(revision)))
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, path) {
			lineParts := strings.Split(strings.Replace(line, "\t", " ", -1), " ")
			return lineParts[2]
		}
	}
	log.Fatal("Failed to lookup blob hash for " + path)
	return ""
}

func parseBlameOutputOrDie(fileName string, out string) []Line {
	result := make([]Line, 0)
	for out != "" {
		// First split off the next blame section
		split := strings.SplitN(out, "\n\t", 2)
		blame := split[0]
		// Then split off the source line that goes with that blame section
		split = strings.SplitN(split[1], "\n", 2)
		contents := strings.TrimPrefix(split[0], "\t")
		// And update the out variable to be what is left.
		if len(split) == 2 {
			out = split[1]
		} else {
			out = ""
		}

		// Finally, parse the blame section and add to the result.
		blameParts := strings.Split(blame, "\n")
		firstLineParts := strings.Split(blameParts[0], " ")
		revision := Revision(firstLineParts[0])
		lineNumber, err := strconv.Atoi(firstLineParts[1])
		if err != nil {
			log.Fatal(err)
		}
		for _, blamePart := range blameParts[1:] {
			if strings.HasPrefix(blamePart, "filename ") {
				fileName = strings.SplitN(blamePart, " ", 2)[1]
			}
		}
		result = append(result, Line{revision, fileName, lineNumber, contents})
	}
	return result
}

func compileRegexsOrDie(commaSeparatedString string) []*regexp.Regexp {
	regexs := make([]*regexp.Regexp, 0)
	for _, regexString := range strings.Split(commaSeparatedString, ",") {
		if regexString != "" {
			regex, err := regexp.Compile(regexString)
			if err != nil {
				log.Fatal(err)
			}
			regexs = append(regexs, regex)
		}
	}
	return regexs
}

func (repository *gitRepository) LoadRevisionTodos(
	revision Revision, todoRegex, excludePaths string) []Line {
	todosChannel := make(chan []Line, 1)
	go repository.asyncLoadRevisionTodos(revision, todoRegex, excludePaths, todosChannel)
	return <-todosChannel
}

func (repository *gitRepository) loadRevisionPaths(revision Revision, excludePaths string) []string {
	// Since this is specified by the user who started the server, we treat erros as fatal.
	excludeRegexs := compileRegexsOrDie(excludePaths)
	includePath := func(path string) bool {
		for _, regex := range excludeRegexs {
			if regex.MatchString(path) {
				return false
			}
		}
		return true
	}
	revisionPaths := make([]string, 0)
	for _, path := range repository.ReadRevisionContents(revision).Paths {
		if includePath(path) {
			revisionPaths = append(revisionPaths, path)
		}
	}
	return revisionPaths
}

func (repository *gitRepository) asyncLoadRevisionTodos(
	revision Revision, todoRegex, excludePaths string, todosChannel chan []Line) {
	if !repository.RevisionTodosCache[revision].Present {
		revisionPaths := repository.loadRevisionPaths(revision, excludePaths)
		todoChannels := make([]chan []Line, 0)
		for _, path := range revisionPaths {
			blob := repository.getFileBlobOrDie(revision, path)
			channel := make(chan []Line, 1)
			todoChannels = append(todoChannels, channel)
			go repository.asyncLoadFileTodos(revision, path, blob, todoRegex, channel)
		}
		todos := make([]Line, 0)
		for _, channel := range todoChannels {
			pathTodos := <-channel
			todos = append(todos, pathTodos...)
		}
		// TODO: Consider grouping the TODOs based on the containing file.
		cacheEntry := todosCacheEntry{
			Present: true,
			Todos:   todos,
		}
		repository.RevisionTodosCache[revision] = cacheEntry
	}
	todosChannel <- repository.RevisionTodosCache[revision].Todos
}

func (repository *gitRepository) LoadFileTodos(
	revision Revision, path string, todoRegex string) []Line {
	blob := repository.getFileBlobOrDie(revision, path)
	todosChannel := make(chan []Line, 1)
	go repository.asyncLoadFileTodos(revision, path, blob, todoRegex, todosChannel)
	return <-todosChannel
}

func (repository *gitRepository) asyncLoadFileTodos(
	revision Revision, path, blob, todoRegex string, todosChannel chan []Line) {
	if !repository.BlobTodosCache[blob].Present {
		raw := runGitCommandOrDie(exec.Command("git", "show", blob))
		rawLines := strings.Split(raw, "\n")
		blobTodos := make([]Line, 0)
		for lineNumber, lineContents := range rawLines {
			matched, err := regexp.MatchString(todoRegex, lineContents)
			if err == nil && matched {
				// git-blame numbers lines starting from 1 rather than 0
				gitLineNumber := lineNumber + 1
				out := runGitCommandOrDie(exec.Command(
					"git", "blame", "--root", "--line-porcelain",
					"-L", fmt.Sprintf("%d,+1", gitLineNumber),
					string(revision), "--", path))
				blobTodos = append(blobTodos, parseBlameOutputOrDie(path, out)...)
			}
		}
		repository.BlobTodosCache[blob] = todosCacheEntry{
			Present: true,
			Todos:   blobTodos,
		}
	}
	todosChannel <- repository.BlobTodosCache[blob].Todos
}

func (repository *gitRepository) ReadFileSnippetAtRevision(revision Revision, path string, startLine, endLine int) string {
	blob := repository.getFileBlobOrDie(revision, path)
	out := runGitCommandOrDie(exec.Command("git", "show", blob))
	lines := strings.Split(out, "\n")
	if startLine < 1 {
		startLine = 1
	}
	if endLine > len(lines) || endLine < 0 {
		endLine = len(lines) + 1
	}
	// Git treats lines as starting from 1, so we have to move our indices before slicing
	startIndex := startLine - 1
	endIndex := endLine - 1
	lines = lines[startIndex:endIndex]
	var buffer bytes.Buffer
	for _, line := range lines {
		buffer.WriteString(line)
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func isGitHubHttpsUrl(remoteUrl string) bool {
	return strings.HasPrefix(remoteUrl, "https://github.com/") &&
		strings.HasSuffix(remoteUrl, ".git")
}

func (repository *gitRepository) GetBrowseUrl(revision Revision, path string, lineNumber int) string {
	rawUrl := fmt.Sprintf("/raw?revision=%s&fileName=%s", string(revision), url.QueryEscape(path))
	out, err := exec.Command("git", "remote", "-v").Output()
	if err != nil {
		return rawUrl
	}
	remotes := strings.Split(strings.Trim(string(out), "\n"), "\n")
	for _, remote := range remotes {
		remoteParts := strings.SplitN(remote, "\t", 2)
		if len(remoteParts) == 2 {
			remoteUrl := strings.Split(remoteParts[1], " ")[0]
			if isGitHubHttpsUrl(remoteUrl) {
				browseSuffix := fmt.Sprintf("/blob/%s/%s#L%d",
					string(revision), path, lineNumber)
				return strings.TrimSuffix(remoteUrl, ".git") + browseSuffix
			}
		}
	}
	return rawUrl
}

func (repository *gitRepository) ValidateRevision(revisionString string) (Revision, error) {
	if !hashRegexp.MatchString(revisionString) {
		return Revision(""), errors.New(fmt.Sprintf("Invalid hash format: %s", revisionString))
	}
	_, err := runGitCommand(exec.Command("git", "ls-tree", "--name-only", revisionString))
	if err != nil {
		return Revision(""), err
	}
	return Revision(revisionString), nil
}
