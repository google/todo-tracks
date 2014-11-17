package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type Alias struct {
	Branch string
	Revision string
}

func (alias *Alias) print() {
	fmt.Printf("Branch: \"%s\", Revision: \"%s\"\n", alias.Branch, alias.Revision)
}

func parseBranchListLine(line string) Alias {
	line = strings.Trim(line, "* ")
	splitLine := strings.Split(line, " ")
	masterName := splitLine[0]
	for _, lineComponent := range splitLine[1:] {
		if len(lineComponent) == 40 {
			return Alias{masterName, lineComponent}
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

func main() {
	for _, alias := range ListBranches() {
		alias.print()
	}
}
