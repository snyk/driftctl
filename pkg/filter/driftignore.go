package filter

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
)

const separator = "_-_"

type DriftIgnore struct {
	driftignorePath string
	ignorePatterns  []string
	matcher         gitignore.Matcher
}

func NewDriftIgnore(path string, ignorePatterns ...string) *DriftIgnore {
	d := DriftIgnore{
		driftignorePath: path,
		ignorePatterns:  ignorePatterns,
		matcher:         gitignore.NewMatcher(nil),
	}
	var err error
	if len(ignorePatterns) > 0 {
		err = d.parseIgnorePatterns()
	} else {
		err = d.readIgnoreFile()
	}

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

	var lines []gitignore.Pattern
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		r.parseIgnorePattern(line, &lines)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	r.matcher = gitignore.NewMatcher(lines)

	return nil
}

func (r *DriftIgnore) parseIgnorePatterns() error {
	var lines []gitignore.Pattern
	for _, p := range r.ignorePatterns {
		r.parseIgnorePattern(p, &lines)
	}
	r.matcher = gitignore.NewMatcher(lines)
	return nil
}

func (r *DriftIgnore) parseIgnorePattern(line string, patterns *[]gitignore.Pattern) {
	if len(strings.ReplaceAll(line, " ", "")) <= 0 {
		return // empty
	}

	if strings.HasPrefix(line, "#") {
		return // this is a comment
	}
	line = strings.ReplaceAll(line, "/", separator)

	*patterns = append(*patterns, gitignore.ParsePattern(line, nil))
	if !strings.HasSuffix(line, "*") {
		line := fmt.Sprintf("%s.*", line)
		*patterns = append(*patterns, gitignore.ParsePattern(line, nil))
	}
}

func (r *DriftIgnore) isAnyOfChildrenTypesNotIgnored(ty resource.ResourceType) bool {
	childrenTypes := resource.GetMeta(ty).GetChildrenTypes()
	for _, childrenType := range childrenTypes {
		if !r.shouldIgnoreType(childrenType) {
			return true
		}
		if r.isAnyOfChildrenTypesNotIgnored(childrenType) {
			return true
		}
	}
	return false
}

func (r *DriftIgnore) IsTypeIgnored(ty resource.ResourceType) bool {
	// Iterate over children types, and do not ignore parent resource
	// if at least one of children type is not ignored.
	if r.isAnyOfChildrenTypesNotIgnored(ty) {
		return false
	}

	return r.shouldIgnoreType(ty)
}

func (r *DriftIgnore) shouldIgnoreType(ty resource.ResourceType) bool {
	for _, pattern := range r.ignorePatterns {
		// If a line start with a `!` and if the type match, we should not ignore it
		if strings.HasPrefix(pattern, fmt.Sprintf("!%s.", ty)) {
			return false
		}
	}

	return r.match(fmt.Sprintf("%s.*", ty))
}

func (r *DriftIgnore) IsResourceIgnored(res *resource.Resource) bool {
	return r.match(fmt.Sprintf("%s.%s", res.ResourceType(), res.ResourceId()))
}

func (r *DriftIgnore) match(strRes string) bool {
	return r.matcher.Match([]string{strings.ReplaceAll(strRes, "/", separator)}, false)
}
