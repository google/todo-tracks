package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// TODO: Split this off into some sort of model package
type Alias struct {
	Branch string
	Revision string
	Paths []string
}

func (alias *Alias) PrintVerbose() {
	fmt.Printf("Branch: \"%s\",\tRevision: \"%s\"\n", alias.Branch, alias.Revision)
	for _, path := range alias.Paths {
		fmt.Printf("\tPath: \"%s\"\n", path)
	}
}

// TODO: Create a package that wraps all of the calls to git commands
func listAliasPaths(revision string) []string {
	out, err := exec.Command("git", "ls-tree", "-r", revision).Output()
	if err != nil {
		log.Fatal(err)
	}
	listOutput := strings.Trim(string(out), "\n ")
	lines := strings.Split(listOutput, "\n")
	paths := make([]string, len(lines))
	for index, line := range lines {
		line = strings.Replace(lines[index], "\t", " ", -1)
		lineParts := strings.Split(line, " ")
		paths[index] = lineParts[len(lineParts) - 1]
	}
	return paths
}

func parseBranchListLine(line string) Alias {
	line = strings.Trim(line, "* ")
	splitLine := strings.Split(line, " ")
	masterName := splitLine[0]
	for _, lineComponent := range splitLine[1:] {
		if len(lineComponent) == 40 {
			revision := lineComponent
			paths := listAliasPaths(revision)
			return Alias{masterName, revision, paths}
		}
	}
	return Alias{Branch: masterName}
}

func ListBranches() []Alias {
	out, err := exec.Command("git", "branch", "-av", "--list", "--abbrev=40").Output()
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

// TODO: Serve a webpage instead of printing to stdout
func main() {
	for _, alias := range ListBranches() {
		alias.PrintVerbose()
	}
}
