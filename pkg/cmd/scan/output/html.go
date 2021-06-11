package output

import (
	"bytes"
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
		"rate": func(count int) float64 {
			if analysis.Summary().TotalResources == 0 {
				return 0
			}
			return math.Round(100 * float64(count) / float64(analysis.Summary().TotalResources))
		},
		//"prettify": func(res interface{}) string {
		//	return prettify(res)
		//},
		//"prettifyPaths": func(paths []string) template.HTML {
		//	return template.HTML(prettifyPaths(paths))
		//},
		"jsonDiff": func(ch analyser.Changelog) template.HTML {
			var buf bytes.Buffer

			whiteSpace := "&emsp;"
			for _, change := range ch {
				for i, v := range change.Path {
					if _, err := strconv.Atoi(v); err == nil {
						change.Path[i] = fmt.Sprintf("[%s]", v)
					}
				}
				path := strings.Join(change.Path, ".")

				switch change.Type {
				case diff.CREATE:
					pref := fmt.Sprintf("%s %s:", "+", path)
					_, _ = fmt.Fprintf(&buf, "%s%s <span class=\"code-box-line-create\">%s</span>", whiteSpace, pref, prettify(change.To))
				case diff.DELETE:
					pref := fmt.Sprintf("%s %s:", "-", path)
					_, _ = fmt.Fprintf(&buf, "%s%s <span class=\"code-box-line-delete\">%s</span>", whiteSpace, pref, prettify(change.From))
				case diff.UPDATE:
					prefix := fmt.Sprintf("%s %s:", "~", path)
					if change.JsonString {
						_, _ = fmt.Fprintf(&buf, "%s%s<br>%s%s<br>", whiteSpace, prefix, whiteSpace, jsonDiff(change.From, change.To, whiteSpace))
						continue
					}
					_, _ = fmt.Fprintf(&buf, "%s%s <span class=\"code-box-line-delete\">%s</span> => <span class=\"code-box-line-create\">%s</span>", whiteSpace, prefix, prettify(change.From), prettify(change.To))
				}

				if change.Computed {
					_, _ = fmt.Fprintf(&buf, " %s", "(computed)")
				}
				_, _ = fmt.Fprintf(&buf, "<br>")
			}

			return template.HTML(buf.String())
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

func prettifyPaths(paths []string) string {
	content := ""
	for i, v := range paths {
		var isArrayKey bool

		// If the previous path is an integer, it means the current path is part of an array
		if j := i - 1; j >= 0 && len(paths) >= j {
			_, err := strconv.Atoi(paths[j])
			isArrayKey = err == nil
		}

		if i > 0 && !isArrayKey {
			content += "<br>"
			content += strings.Repeat("&emsp;", i)
		}

		if _, err := strconv.Atoi(v); err == nil {
			content += "- "
		} else {
			content += fmt.Sprintf("%s:", v)
		}
	}

	return content
}
