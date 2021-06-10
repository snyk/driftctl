package output

import (
	"embed"
	"fmt"
	"html/template"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/r3labs/diff/v2"
)

const HTMLOutputType = "html"
const HTMLOutputExample = "html://PATH/TO/FILE.html"

// assets holds our static web content.
//go:embed assets/*
var assets embed.FS

type HTML struct {
	path string
}

type HTMLTemplateParams struct {
	ScanDate     string
	Coverage     int
	Summary      analyser.Summary
	Unmanaged    []resource.Resource
	Differences  []analyser.Difference
	Deleted      []resource.Resource
	Alerts       alerter.Alerts
	Stylesheet   template.CSS
	ScanDuration string
}

func NewHTML(path string) *HTML {
	return &HTML{path}
}

func (c *HTML) Write(analysis *analyser.Analysis) error {
	file := os.Stdout
	if !isStdOut(c.path) {
		f, err := os.OpenFile(c.path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		file = f
	}

	tmplFile, err := assets.ReadFile("assets/index.tmpl")
	if err != nil {
		return err
	}

	styleFile, err := assets.ReadFile("assets/style.css")
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"getResourceTypes": func() []string {
			resources := make([]resource.Resource, 0)
			resources = append(resources, analysis.Unmanaged()...)
			resources = append(resources, analysis.Deleted()...)

			for _, d := range analysis.Differences() {
				resources = append(resources, d.Res)
			}

			return distinctResourceTypes(resources)
		},
		"formatChange": func(ch analyser.Change) string {
			prefix := ""
			suffix := ""

			switch ch.Type {
			case diff.CREATE:
				prefix = "+"
			case diff.UPDATE:
				prefix = "~"
			case diff.DELETE:
				prefix = "-"
			}

			if ch.Computed {
				suffix = "(computed)"
			}

			return fmt.Sprintf("%s %s: %s => %s %s", prefix, strings.Join(ch.Path, "."), prettify(ch.From), prettify(ch.To), suffix)
		},
		"rate": func(count int) float64 {
			if analysis.Summary().TotalResources == 0 {
				return 0
			}
			return math.Round(100 * float64(count) / float64(analysis.Summary().TotalResources))
		},
		"isInt": func(str string) bool {
			_, err := strconv.ParseInt(str, 10, 32)
			if err != nil {
				return false
			}
			return true
		},
		"prettify": func(res interface{}) string {
			return prettify(res)
		},
		"sum": func(n ...int) int {
			total := 0
			for _, v := range n {
				total += v
			}
			return total
		},
		"repeatString": func(s string, count int) template.HTML {
			return template.HTML(strings.Repeat(s, count))
		},
	}

	tmpl, err := template.New("main").Funcs(funcMap).Parse(string(tmplFile))
	if err != nil {
		return err
	}

	data := &HTMLTemplateParams{
		ScanDate:     analysis.Date.Format("Jan 02, 2006"),
		Summary:      analysis.Summary(),
		Coverage:     analysis.Coverage(),
		Unmanaged:    analysis.Unmanaged(),
		Differences:  analysis.Differences(),
		Deleted:      analysis.Deleted(),
		Alerts:       analysis.Alerts(),
		Stylesheet:   template.CSS(styleFile),
		ScanDuration: analysis.Duration.Round(time.Second).String(),
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}

func distinctResourceTypes(resources []resource.Resource) []string {
	types := make([]string, 0)

	for _, res := range resources {
		found := false
		for _, v := range types {
			if v == res.TerraformType() {
				found = true
				break
			}
		}
		if !found {
			types = append(types, res.TerraformType())
		}
	}

	return types
}
