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
	exclusionList map[string]struct{}
}

func NewDriftIgnore() *DriftIgnore {
	d := DriftIgnore{
		exclusionList: map[string]struct{}{},
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
		typeVal := strings.SplitN(line, ".", 2)
		if len(typeVal) != 2 {
			logrus.WithFields(logrus.Fields{
				"line": line,
			}).Warnf("unable to parse line, invalid length, got %d expected 2", len(typeVal))
			continue
		}
		logrus.WithFields(logrus.Fields{
			"type": typeVal[0],
			"id":   typeVal[1],
		}).Debug("Found ignore rule in .driftignore")
		r.exclusionList[line] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (r *DriftIgnore) Run(resources []resource.Resource) []resource.Resource {

	results := make([]resource.Resource, 0, len(resources))

	for _, res := range resources {
		_, isExclusionRule := r.exclusionList[fmt.Sprintf("%s.%s", res.TerraformType(), res.TerraformId())]
		_, isExclusionWildcardRule := r.exclusionList[fmt.Sprintf("%s.*", res.TerraformType())]
		if !isExclusionRule && !isExclusionWildcardRule {
			results = append(results, res)
		}
	}

	return results
}
