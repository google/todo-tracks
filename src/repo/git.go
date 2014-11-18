package repo

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type GitRepository struct{}

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

func (gitRepository GitRepository) ListBranches() []Alias {
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

func (gitRepository GitRepository) ReadRevisionContents(revision Revision) *RevisionContents {
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

func (gitRepository GitRepository) getSubject(revision Revision) string {
	out, err := exec.Command("git", "show", string(revision), "--format=\"format:%s\"", "-s").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func (gitRepository GitRepository) getAuthorName(revision Revision) string {
	out, err := exec.Command("git", "show", string(revision), "--format=\"format:%an\"", "-s").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func (gitRepository GitRepository) getAuthorEmail(revision Revision) string {
	out, err := exec.Command("git", "show", string(revision), "--format=\"format:%ae\"", "-s").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func (gitRepository GitRepository) getTimestamp(revision Revision) int64 {
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
