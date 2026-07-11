package workspace

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const ideasDirName = "ideas"

type Idea struct {
	Number   int
	Filename string
	Title    string
	Rejected bool
	Created  time.Time
}

func CollectIdeas(workspaceRoot string, includeRejected bool) ([]Idea, error) {
	ideasDir := filepath.Join(workspaceRoot, "grind", ideasDirName)

	entries, err := os.ReadDir(ideasDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading ideas dir: %w", err)
	}

	var ideas []Idea
	number := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		rejected := strings.HasPrefix(name, "rejected-")
		if rejected && !includeRejected {
			continue
		}

		created, err := parseIdeaTimestamp(name)
		if err != nil {
			continue
		}

		title, err := readIdeaTitle(filepath.Join(ideasDir, name))
		if err != nil {
			title = "(untitled)"
		}

		ideas = append(ideas, Idea{
			Number:   number,
			Filename: name,
			Title:    title,
			Rejected: rejected,
			Created:  created,
		})
		number++
	}

	sort.Slice(ideas, func(i, j int) bool {
		return ideas[i].Created.Before(ideas[j].Created)
	})

	for i := range ideas {
		ideas[i].Number = i
	}

	return ideas, nil
}

func parseIdeaTimestamp(filename string) (time.Time, error) {
	base := strings.TrimPrefix(filename, "rejected-")
	base = strings.TrimSuffix(base, ".md")

	if len(base) != 14 {
		return time.Time{}, fmt.Errorf("invalid timestamp length: %s", base)
	}

	year, err := strconv.Atoi(base[0:4])
	if err != nil {
		return time.Time{}, err
	}
	month, err := strconv.Atoi(base[4:6])
	if err != nil {
		return time.Time{}, err
	}
	day, err := strconv.Atoi(base[6:8])
	if err != nil {
		return time.Time{}, err
	}
	hour, err := strconv.Atoi(base[8:10])
	if err != nil {
		return time.Time{}, err
	}
	minute, err := strconv.Atoi(base[10:12])
	if err != nil {
		return time.Time{}, err
	}
	second, err := strconv.Atoi(base[12:14])
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), nil
}

func readIdeaTitle(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			title := strings.TrimLeft(line, "# ")
			title = strings.TrimSpace(title)
			if title != "" {
				return title, nil
			}
		}
	}

	return "", fmt.Errorf("no heading found")
}
