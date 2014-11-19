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

func splitCommandOutputLine(line string) []string {
	lineParts := make([]string, 0)
	for _, part := range strings.Split(line, " ") {
		if part != "" {
			lineParts = append(lineParts, part)
		}
	}
	return lineParts
}

func (gitRepository GitRepository) ListBranches() []Alias {
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

func (gitRepository GitRepository) ReadRevisionContents(revision Revision) *RevisionContents {
	out := runGitCommandOrDie(exec.Command("git", "ls-tree", "-r", string(revision)))
	lines := strings.Split(out, "\n")
	paths := make([]string, len(lines))
	for index, line := range lines {
		line = strings.Replace(lines[index], "\t", " ", -1)
		lineParts := splitCommandOutputLine(line)
		paths[index] = lineParts[len(lineParts)-1]
	}
	return &RevisionContents{revision, paths}
}

func (gitRepository GitRepository) getSubject(revision Revision) string {
	return runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=%s", "-s"))
}

func (gitRepository GitRepository) getAuthorName(revision Revision) string {
	return runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=%an", "-s"))
}

func (gitRepository GitRepository) getAuthorEmail(revision Revision) string {
	return runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=%ae", "-s"))
}

func (gitRepository GitRepository) getTimestamp(revision Revision) int64 {
	out := runGitCommandOrDie(exec.Command(
		"git", "show", string(revision), "--format=%ct", "-s"))
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
	out := runGitCommandOrDie(exec.Command(
		"git", "blame", "--root", "-sfn", "--abbrev=40", string(revision), "--", path))
	lines := strings.Split(out, "\n")
	result := make([]Line, len(lines))
	for _, line := range lines {
		lineNumberIndex := strings.Index(line, ")")
		lineParts := splitCommandOutputLine(line)
		if lineNumberIndex > 0 && len(lineParts) > 3 {
			revision := Revision(lineParts[0])
			fileName := lineParts[1]
			lineNumber, err := strconv.Atoi(lineParts[2])
			if err != nil {
				log.Fatal(err)
			}
			contents := line[lineNumberIndex+1:]
			result = append(result, Line{revision, fileName, lineNumber, contents})
		}
	}
	return result
}
