package filter

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/sirupsen/logrus"
)

const separator = "_-_"

type DriftIgnore struct {
	driftignorePath string
	matcher         gitignore.Matcher
	rulesCount      int
}

func NewDriftIgnore(path string) *DriftIgnore {
	d := DriftIgnore{
		driftignorePath: path,
		matcher:         gitignore.NewMatcher(nil),
	}
	err := d.readIgnoreFile()
	if err != nil {
		logrus.Debug(err)
	}
	return &d
}

func (r *DriftIgnore) readIgnoreFile() error {
	file, err := os.Open(r.driftignorePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var patterns []gitignore.Pattern
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()

		if len(strings.ReplaceAll(line, " ", "")) <= 0 {
			continue // empty
		}

		if strings.HasPrefix(line, "#") {
			continue // this is a comment
		}
		line = strings.ReplaceAll(line, "/", separator)

		patterns = append(patterns, gitignore.ParsePattern(line, nil))
		if !strings.HasSuffix(line, "*") {
			line := fmt.Sprintf("%s.*", line)
			patterns = append(patterns, gitignore.ParsePattern(line, nil))
		}
		r.rulesCount++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	r.matcher = gitignore.NewMatcher(patterns)

	return nil
}

func (r *DriftIgnore) IsResourceIgnored(res resource.Resource) bool {
	return r.match(fmt.Sprintf("%s.%s", res.TerraformType(), res.TerraformId()))
}

func (r *DriftIgnore) IsFieldIgnored(res resource.Resource, path []string) bool {
	full := fmt.Sprintf("%s.%s.%s", res.TerraformType(), res.TerraformId(), strings.Join(path, "."))
	return r.match(full)
}

func (r *DriftIgnore) match(strRes string) bool {
	return r.matcher.Match([]string{strings.ReplaceAll(strRes, "/", separator)}, false)
}

func (r *DriftIgnore) RulesCount() int {
	return r.rulesCount
}
