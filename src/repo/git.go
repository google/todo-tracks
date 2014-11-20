package repo

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
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
		lineParts := strings.SplitN(line, " ", 4)
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

func (gitRepository GitRepository) getFileBlobOrDie(revision Revision, path string) string {
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

func (gitRepository GitRepository) readRawFileOrDie(revision Revision, path string) string {
	blob := gitRepository.getFileBlobOrDie(revision, path)
	return runGitCommandOrDie(exec.Command("git", "show", blob))
}

func parseBlameOutputOrDie(fileName string, out string, result []Line) []Line {
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

func (gitRepository GitRepository) LoadTodos(revision Revision, path string, todoRegex string, result []Line) []Line {
	raw := gitRepository.readRawFileOrDie(revision, path)
	rawLines := strings.Split(raw, "\n")
	for lineNumber, lineContents := range rawLines {
		matched, err := regexp.MatchString(todoRegex, lineContents)
		if err == nil && matched {
			// git-blame numbers lines starting from 1 rather than 0
			gitLineNumber := lineNumber + 1
			out := runGitCommandOrDie(exec.Command(
				"git", "blame", "--root", "--line-porcelain",
				"-L", fmt.Sprintf("%d,+1", gitLineNumber),
				string(revision), "--", path))
			result = parseBlameOutputOrDie(path, out, result)
		}
	}
	return result
}

func (gitRepository GitRepository) ReadFileSnippetAtRevision(revision Revision, path string, startLine, endLine int) string {
	out := gitRepository.readRawFileOrDie(revision, path)
	lines := strings.Split(out, "\n")
	if startLine < 0 {
		startLine = 0
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	lines = lines[startLine:endLine]
	var buffer bytes.Buffer
	for _, line := range lines {
		buffer.WriteString(line)
		buffer.WriteString("\n")
	}
	return buffer.String()
}
