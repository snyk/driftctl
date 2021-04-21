package output

import (
	"embed"
	"fmt"
	"html/template"
	"os"
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

func NewHTML(path string) *HTML {
	return &HTML{path}
}

func (c *HTML) Write(analysis *analyser.Analysis) error {
	type TemplateParams struct {
		ScanDate    string
		Coverage    int
		Summary     analyser.Summary
		Managed     []resource.Resource
		Unmanaged   []resource.Resource
		Differences []analyser.Difference
		Deleted     []resource.Resource
		Alerts      alerter.Alerts
		Stylesheet  template.CSS
	}

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
	}

	tmpl, err := template.New("main").Funcs(funcMap).Parse(string(tmplFile))
	if err != nil {
		return err
	}

	data := &TemplateParams{
		ScanDate:    time.Now().Format("Jan 02, 2006"),
		Summary:     analysis.Summary(),
		Coverage:    analysis.Coverage(),
		Managed:     analysis.Managed(),
		Unmanaged:   analysis.Unmanaged(),
		Differences: analysis.Differences(),
		Deleted:     analysis.Deleted(),
		Alerts:      analysis.Alerts(),
		Stylesheet:  template.CSS(styleFile),
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}
