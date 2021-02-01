package filter

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/sirupsen/logrus"
)

type DriftIgnore struct {
	resExclusionList   map[string]struct{} // map[type.id] exists to ignore
	driftExclusionList map[string][]string // map[type.id] contains path for drift to ignore
}

func NewDriftIgnore() *DriftIgnore {
	d := DriftIgnore{
		resExclusionList:   map[string]struct{}{},
		driftExclusionList: map[string][]string{},
	}
	err := d.readIgnoreFile()
	if err != nil {
		logrus.Debug(err)
	}
	return &d
}

func (r *DriftIgnore) readIgnoreFile() error {
	file, err := os.Open(".driftignore")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		typeVal := readDriftIgnoreLine(line)
		nbArgs := len(typeVal)
		if nbArgs == 0 || strings.HasPrefix(line, "#") {
			continue // skip empty or commented out lines
		} else {
			if nbArgs < 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warnf("unable to parse line, invalid length, got %d expected >= 2", nbArgs)
				continue
			}
			if nbArgs == 2 { // We want to ignore a resource (type.id)
				logrus.WithFields(logrus.Fields{
					"type": typeVal[0],
					"id":   typeVal[1],
				}).Debug("Found ignore resource rule in .driftignore")
				r.resExclusionList[strings.Join(typeVal, ".")] = struct{}{}
				continue
			}
			// Here we want to ignore a drift (type.id.path.to.field)
			res := strings.Join(typeVal[0:2], ".")
			ignoreSublist, exists := r.driftExclusionList[res]
			if !exists {
				ignoreSublist = make([]string, 0, 1)
			}
			path := strings.Join(typeVal[2:], ".")

			logrus.WithFields(logrus.Fields{
				"type": typeVal[0],
				"id":   typeVal[1],
				"path": path,
			}).Debug("Found ignore resource field rule in .driftignore")

			ignoreSublist = append(ignoreSublist, path)
			r.driftExclusionList[res] = ignoreSublist
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (r *DriftIgnore) IsResourceIgnored(res resource.Resource) bool {
	_, isExclusionRule := r.resExclusionList[fmt.Sprintf("%s.%s", res.TerraformType(), res.TerraformId())]
	_, isExclusionWildcardRule := r.resExclusionList[fmt.Sprintf("%s.*", res.TerraformType())]
	return isExclusionRule || isExclusionWildcardRule
}

func (r *DriftIgnore) IsFieldIgnored(res resource.Resource, path []string) bool {
	exclusionRules, isExclusionRule := r.driftExclusionList[fmt.Sprintf("%s.%s", res.TerraformType(), res.TerraformId())]
	exclusionWildcardRules, isExclusionWildcardRule := r.driftExclusionList[fmt.Sprintf("%s.*", res.TerraformType())]

	if !isExclusionRule && !isExclusionWildcardRule {
		return false
	}

	if !isExclusionRule {
		exclusionRules = exclusionWildcardRules
	}

	if r.isExcluded(exclusionRules, path) {
		return true
	}

	return false
}

func (r *DriftIgnore) isExcluded(rules []string, changePath []string) bool {
RuleCheck:
	for _, rule := range rules {
		path := readDriftIgnoreLine(rule)
		if len(path) > len(changePath) {
			continue // path size does not match
		}

		for i := range path {
			if !strings.EqualFold(path[i], changePath[i]) && path[i] != "*" {
				continue RuleCheck // found a diff in path that was not a wildcard
			}
		}
		return true
	}
	return false
}

/**
 * Read a line of ignore
 * Handle split on dots and escaping
 */
func readDriftIgnoreLine(line string) []string {
	var splitted []string
	lastWordEnd := 0
	for i := range line {
		if line[i] == '.' && ((i >= 1 && line[i-1] != '\\') || (i >= 2 && line[i-1] == '\\' && line[i-2] == '\\')) {
			splitted = append(splitted, unescapeDriftIgnoreLine(line[lastWordEnd:i]))
			lastWordEnd = i + 1
			continue
		}
		if i == len(line)-1 {
			splitted = append(splitted, unescapeDriftIgnoreLine(line[lastWordEnd:]))
		}
	}
	return splitted
}

func unescapeDriftIgnoreLine(line string) string {
	var res string
	lastEscapeEnd := 0
	for i := range line {
		if line[i] == '\\' {
			if i+1 < len(line) && line[i+1] == '\\' {
				continue
			}
			if i > 1 && line[i-1] == '\\' {
				res += line[lastEscapeEnd:i]
				lastEscapeEnd = i + 1
				continue
			}
			res += line[lastEscapeEnd:i]
			lastEscapeEnd = i + 1
			continue
		}
		if i == len(line)-1 {
			res += line[lastEscapeEnd:]
		}
	}

	return res
}
