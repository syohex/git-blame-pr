package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

func main() {
	os.Exit(_main())
}

var mergeCommitRe = regexp.MustCompile(`(?i)Merge\s+(?:pull\s+request|pr)\s+#?(\d+)\s`)

func lookupPullRequest(commitID string) (string, error) {
	cmd := exec.Command("git", "show", "--oneline", commitID)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	matches := mergeCommitRe.FindStringSubmatch(string(output))
	if matches == nil {
		return fmt.Sprintf("%8s", commitID), nil
	}

	return fmt.Sprintf("PR #%-4s", matches[1]), nil
}

func _main() int {
	if len(os.Args) < 2 {
		fmt.Println("Usage: git-blame-pr file...")
		return 1
	}

	var output bytes.Buffer
	args := []string{"blame", "--first-parent"}
	args = append(args, os.Args[1:]...)
	cmd := exec.Command("git", args...)
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return 1
	}

	blameLineRe := regexp.MustCompile(`^(\S+).*?\) (.*)$`)

	cache := make(map[string]string)
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		line := scanner.Text()
		matches := blameLineRe.FindStringSubmatch(line)
		if matches == nil {
			fmt.Println("Please check 'git blame --first-parent' output")
			return 1
		}

		commitID, source := matches[1], matches[2]
		if _, ok := cache[commitID]; !ok {
			if prNumber, err := lookupPullRequest(commitID); err != nil {
				fmt.Println(err)
				return 1
			} else {
				cache[commitID] = prNumber
			}
		}

		fmt.Printf("%s %s\n", cache[commitID], source)
	}

	if scanner.Err() != nil {
		fmt.Println(scanner.Err())
		return 1
	}

	return 0
}
