package filter

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cloudskiff/driftctl/pkg/stringutils"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/r3labs/diff/v2"

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
		typeVal := stringutils.EscapableSplit(line)
		nbArgs := len(typeVal)
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

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (r *DriftIgnore) FilterResources(resources []resource.Resource) []resource.Resource {

	results := make([]resource.Resource, 0, len(resources))

	for _, res := range resources {
		_, isExclusionRule := r.resExclusionList[fmt.Sprintf("%s.%s", res.TerraformType(), res.TerraformId())]
		_, isExclusionWildcardRule := r.resExclusionList[fmt.Sprintf("%s.*", res.TerraformType())]
		if !isExclusionRule && !isExclusionWildcardRule {
			results = append(results, res)
		}
	}

	return results
}

func (r *DriftIgnore) FilterDrift(diffs []analyser.Difference) []analyser.Difference {

	results := make([]analyser.Difference, 0, len(diffs))

	for _, dif := range diffs {
		exclusionRules, isExclusionRule := r.driftExclusionList[fmt.Sprintf("%s.%s", dif.Res.TerraformType(), dif.Res.TerraformId())]
		exclusionWildcardRules, isExclusionWildcardRule := r.driftExclusionList[fmt.Sprintf("%s.*", dif.Res.TerraformType())]

		if !isExclusionRule && !isExclusionWildcardRule {
			results = append(results, dif) // we don't have rules to ignore drift on this resource
			continue
		}

		if !isExclusionRule {
			exclusionRules = exclusionWildcardRules
		}

		changelog := make([]diff.Change, 0, len(dif.Changelog))
		for _, change := range dif.Changelog {
			if r.isExcluded(exclusionRules, change) {
				continue // Change is excluded we don't append it to the changelog
			}
			changelog = append(changelog, change)
		}

		if len(changelog) <= 0 {
			continue // All changes where ignored we don't keep this difference
		}

		dif.Changelog = changelog      // Update changelog
		results = append(results, dif) // Keep this diff
	}

	return results
}

func (r *DriftIgnore) isExcluded(rules []string, change diff.Change) bool {
RuleCheck:
	for _, rule := range rules {
		path := stringutils.EscapableSplit(rule)
		if len(path) > len(change.Path) {
			continue // path size does not match
		}

		for i := range path {
			if path[i] != strings.ToLower(change.Path[i]) && path[i] != "*" {
				continue RuleCheck // found a diff in path that was not a wildcard
			}
		}
		return true
	}
	return false
}
