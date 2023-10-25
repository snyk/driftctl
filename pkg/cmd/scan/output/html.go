package output

import (
	"embed"
	"encoding/base64"
	"html/template"
	"math"
	"os"
	"time"

	"github.com/snyk/driftctl/enumeration/alerter"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/analyser"
)

const HTMLOutputType = "html"
const HTMLOutputExample = "html://PATH/TO/FILE.html"

// assets holds our static web content.
//
//go:embed assets/*
var assets embed.FS

type HTML struct {
	path string
}

type HTMLTemplateParams struct {
	IsSync          bool
	ScanDate        string
	Coverage        int
	Summary         analyser.Summary
	Unmanaged       []*resource.Resource
	Deleted         []*resource.Resource
	Alerts          alerter.Alerts
	Stylesheet      template.CSS
	ScanDuration    string
	ProviderName    string
	ProviderVersion string
	LogoSvg         template.HTML
	FaviconBase64   string
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

	logoSvgFile, err := assets.ReadFile("assets/driftctl_light.svg")
	if err != nil {
		return err
	}

	faviconFile, err := assets.ReadFile("assets/favicon.ico")
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"getResourceTypes": func() []string {
			resources := make([]*resource.Resource, 0)
			resources = append(resources, analysis.Unmanaged()...)
			resources = append(resources, analysis.Deleted()...)

			return distinctResourceTypes(resources)
		},
		"getIaCSources": func() []string {
			resources := make([]*resource.Resource, 0)
			resources = append(resources, analysis.Deleted()...)
			resources = append(resources, analysis.Managed()...)

			return distinctIaCSources(resources)
		},
		"rate": func(count int) float64 {
			if analysis.Summary().TotalResources == 0 {
				return 0
			}
			rate := 100 * float64(count) / float64(analysis.Summary().TotalResources)
			return math.Floor(rate*100) / 100
		},
	}

	tmpl, err := template.New("main").Funcs(funcMap).Parse(string(tmplFile))
	if err != nil {
		return err
	}

	data := &HTMLTemplateParams{
		IsSync:          analysis.IsSync(),
		ScanDate:        analysis.Date.Format("Jan 02, 2006"),
		Coverage:        analysis.Coverage(),
		Summary:         analysis.Summary(),
		Unmanaged:       analysis.Unmanaged(),
		Deleted:         analysis.Deleted(),
		Alerts:          analysis.Alerts(),
		Stylesheet:      template.CSS(styleFile),
		ScanDuration:    analysis.Duration.Round(time.Second).String(),
		ProviderName:    analysis.ProviderName,
		ProviderVersion: analysis.ProviderVersion,
		LogoSvg:         template.HTML(logoSvgFile),
		FaviconBase64:   base64.StdEncoding.EncodeToString(faviconFile),
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}

func distinctResourceTypes(resources []*resource.Resource) []string {
	types := make([]string, 0)

	for _, res := range resources {
		found := false
		for _, v := range types {
			if v == res.ResourceType() {
				found = true
				break
			}
		}
		if !found {
			types = append(types, res.ResourceType())
		}
	}

	return types
}

func distinctIaCSources(resources []*resource.Resource) []string {
	types := make([]string, 0)

	for _, res := range resources {
		if res.Src() == nil {
			continue
		}

		found := false
		for _, v := range types {
			if v == res.Src().Source() {
				found = true
				break
			}
		}
		if !found {
			types = append(types, res.Src().Source())
		}
	}

	return types
}
