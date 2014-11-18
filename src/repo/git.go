package repo

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type GitRepository struct{}

func runGitCommandOrDie(cmd *exec.Cmd) string {
	out, err := cmd.Output()
	if err != nil {
		log.Print(cmd.Args)
		log.Print(out)
		log.Fatal(err)
	}
	return strings.Trim(string(out), " \n")
}

func (gitRepository GitRepository) ListBranches() []Alias {
	out := runGitCommandOrDie(
		exec.Command("git", "branch", "-av", "--list", "--abbrev=40", "--no-color"))
	lines := strings.Split(out, "\n")
	aliases := make([]Alias, 0)
	for _, line := range lines {
		line = strings.Trim(line, "* ")
		splitLine := strings.Split(line, " ")
		masterName := splitLine[0]
		for _, lineComponent := range splitLine[1:] {
			if len(lineComponent) == 40 {
				revisionHash := lineComponent
				aliases = append(aliases, Alias{masterName, Revision(revisionHash)})
			}
		}
	}
	return aliases
}

func (gitRepository GitRepository) ReadRevisionContents(revision Revision) *RevisionContents {
	out := runGitCommandOrDie(exec.Command("git", "ls-tree", "-r", string(revision)))
	lines := strings.Split(out, "\n")
	paths := make([]string, len(lines))
	for index, line := range lines {
		line = strings.Replace(lines[index], "\t", " ", -1)
		lineParts := strings.Split(line, " ")
		paths[index] = lineParts[len(lineParts)-1]
	}
	return &RevisionContents{revision, paths}
}

func (gitRepository GitRepository) getSubject(revision Revision) string {
	return runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=\"format:%s\"", "-s"))
}

func (gitRepository GitRepository) getAuthorName(revision Revision) string {
	return runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=\"format:%an\"", "-s"))
}

func (gitRepository GitRepository) getAuthorEmail(revision Revision) string {
	return runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=\"format:%ae\"", "-s"))
}

func (gitRepository GitRepository) getTimestamp(revision Revision) int64 {
	out := runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=\"format:%t\"", "-s"))
	timestamp, err := strconv.ParseInt(out, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return timestamp
}

func (gitRepository GitRepository) ReadRevisionMetadata(revision Revision) RevisionMetadata {
	return RevisionMetadata{
		Revision:    revision,
		Timestamp:   gitRepository.getTimestamp(revision),
		Subject:     gitRepository.getSubject(revision),
		AuthorName:  gitRepository.getAuthorName(revision),
		AuthorEmail: gitRepository.getAuthorEmail(revision),
	}
}

func (gitRepository GitRepository) ReadFileAtRevision(revision Revision, path string) []Line {
	out := runGitCommandOrDie(
		exec.Command("git", "blame", "-s", "--abbrev=40", string(revision), "--", path))
	lines := strings.Split(out, "\n")
	result := make([]Line, len(lines))
	for _, line := range lines {
		revision := Revision(line[0:40])
		lineNumberIndex := strings.Index(line, ")")
		lineNumber, err := strconv.Atoi(strings.Trim(line[41:lineNumberIndex], " "))
		if err != nil {
			log.Fatal(err)
		}
		contents := line[lineNumberIndex+1:]
		if contents != "" {
			result = append(result, Line{revision, lineNumber, contents})
		}
	}
	return result
}
