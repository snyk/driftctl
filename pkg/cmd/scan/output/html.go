package output

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"github.com/snyk/driftctl/enumeration/alerter"
	"html/template"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
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
	Differences     []analyser.Difference
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

			for _, d := range analysis.Differences() {
				resources = append(resources, d.Res)
			}

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
						_, _ = fmt.Fprintf(&buf, "%s%s<br>%s%s<br>", whiteSpace, prefix, whiteSpace, jsonDiffHTML(change.From, change.To))
						continue
					}
					_, _ = fmt.Fprintf(&buf, "%s%s <span class=\"code-box-line-delete\">%s</span> => <span class=\"code-box-line-create\">%s</span>", whiteSpace, prefix, htmlPrettify(change.From), htmlPrettify(change.To))
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
		IsSync:          analysis.IsSync(),
		ScanDate:        analysis.Date.Format("Jan 02, 2006"),
		Coverage:        analysis.Coverage(),
		Summary:         analysis.Summary(),
		Unmanaged:       analysis.Unmanaged(),
		Differences:     analysis.Differences(),
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

func htmlPrettify(resource interface{}) string {
	res := reflect.ValueOf(resource)
	if resource == nil || res.Kind() == reflect.Ptr && res.IsNil() {
		return "null"
	}
	return awsutil.Prettify(resource)
}

func jsonDiffHTML(a, b interface{}) string {
	diffStr := jsonDiff(a, b, false)

	re := regexp.MustCompile(`(?m)^(?P<value>(\-)(.*))$`)
	diffStr = re.ReplaceAllString(diffStr, `<span class="code-box-line-delete">$value</span>`)

	re = regexp.MustCompile(`(?m)^(?P<value>(\+)(.*))$`)
	diffStr = re.ReplaceAllString(diffStr, `<span class="code-box-line-create">$value</span>`)

	return diffStr
}
