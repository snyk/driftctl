package output

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/pkg/remote/alerts"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"

	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/pkg/resource"
)

const ConsoleOutputType = "console"
const ConsoleOutputExample = "console://"

type Console struct {
	summary string
}

func NewConsole() *Console {
	return &Console{
		`Total coverage is {{ analysis.Coverage }}`,
	}
}

func (c *Console) Write(analysis *analyser.Analysis) error {
	if analysis.Summary().TotalDeleted > 0 {
		var sources []string
		groupedBySource := make(map[string][]*resource.Resource)

		for _, deletedResource := range analysis.Deleted() {
			key := ""
			if deletedResource.Source != nil {
				key = deletedResource.Source.Source()
			}

			if _, exist := groupedBySource[key]; !exist {
				groupedBySource[key] = []*resource.Resource{deletedResource}
				continue
			}

			groupedBySource[key] = append(groupedBySource[key], deletedResource)
		}

		for s := range groupedBySource {
			sources = append(sources, s)
		}
		sort.Strings(sources)

		fmt.Println("Found missing resources:")

		for _, source := range sources {
			indentBase := "  "
			if source != "" {
				fmt.Print(color.BlueString("%sFrom %s\n", indentBase, source))
				indentBase += indentBase
			}
			for _, deletedResource := range groupedBySource[source] {
				humanStringSource := deletedResource.ResourceType()
				if deletedResource.SourceString() != "" {
					humanStringSource = deletedResource.SourceString()
				}
				humanString := fmt.Sprintf("%s- %s (%s)", indentBase, deletedResource.ResourceId(), humanStringSource)

				if humanAttrs := formatResourceAttributes(deletedResource); humanAttrs != "" {
					humanString += fmt.Sprintf("\n%s    %s", indentBase, humanAttrs)
				}
				fmt.Println(humanString)
			}
		}
	}

	if analysis.Summary().TotalUnmanaged > 0 {
		fmt.Println("Found resources not covered by IaC:")
		unmanagedByType, keys := groupByType(analysis.Unmanaged())
		for _, ty := range keys {
			fmt.Printf("  %s:\n", ty)
			for _, res := range unmanagedByType[ty] {
				humanString := fmt.Sprintf("    - %s", res.ResourceId())
				if humanAttrs := formatResourceAttributes(res); humanAttrs != "" {
					humanString += fmt.Sprintf("\n        %s", humanAttrs)
				}
				fmt.Println(humanString)
			}
		}
	}

	if analysis.Summary().TotalDrifted > 0 {
		var sources []string
		groupedBySource := make(map[string][]analyser.Difference)
		for _, difference := range analysis.Differences() {
			key := ""
			if difference.Res.Source != nil {
				key = difference.Res.Source.Source()
			}
			if _, exist := groupedBySource[key]; !exist {
				groupedBySource[key] = []analyser.Difference{difference}
				continue
			}
			groupedBySource[key] = append(groupedBySource[key], difference)
		}

		for s := range groupedBySource {
			sources = append(sources, s)
		}
		sort.Strings(sources)

		fmt.Println("Found changed resources:")
		for _, source := range sources {
			indentBase := "  "
			if source != "" {
				fmt.Print(color.BlueString("%sFrom %s\n", indentBase, source))
				indentBase += indentBase
			}
			for _, difference := range groupedBySource[source] {
				humanStringSource := difference.Res.ResourceType()
				if difference.Res.SourceString() != "" {
					humanStringSource = difference.Res.SourceString()
				}
				humanString := fmt.Sprintf("%s- %s (%s):", indentBase, difference.Res.ResourceId(), humanStringSource)
				whiteSpace := indentBase + "    "
				if humanAttrs := formatResourceAttributes(difference.Res); humanAttrs != "" {
					humanString += fmt.Sprintf("\n%s%s", whiteSpace, humanAttrs)
					whiteSpace += "    "
				}
				fmt.Println(humanString)
				for _, change := range difference.Changelog {
					path := strings.Join(change.Path, ".")
					pref := fmt.Sprintf("%s %s:", color.YellowString("~"), path)
					if change.Type == diff.CREATE {
						pref = fmt.Sprintf("%s %s:", color.GreenString("+"), path)
					} else if change.Type == diff.DELETE {
						pref = fmt.Sprintf("%s %s:", color.RedString("-"), path)
					}
					if change.Type == diff.UPDATE {
						if change.JsonString {
							prefix := "           "
							fmt.Printf("%s%s\n%s%s\n", whiteSpace, pref, prefix, jsonDiff(change.From, change.To, isatty.IsTerminal(os.Stdout.Fd())))
							continue
						}
					}
					fmt.Printf("%s%s %s => %s", whiteSpace, pref, prettify(change.From), prettify(change.To))
					if change.Computed {
						fmt.Printf(" %s", color.YellowString("(computed)"))
					}
					fmt.Printf("\n")
				}
			}
		}
	}

	c.writeSummary(analysis)

	enumerationErrorMessage := ""
	for _, a := range analysis.Alerts() {
		for _, alert := range a {
			fmt.Println(color.YellowString(alert.Message()))
			if alert, ok := alert.(*alerts.RemoteAccessDeniedAlert); ok && enumerationErrorMessage == "" {
				enumerationErrorMessage = alert.GetProviderMessage()
			}
		}
	}

	if enumerationErrorMessage != "" {
		_, _ = fmt.Fprintf(os.Stderr, "\n%s\n", color.YellowString(enumerationErrorMessage))
	}

	return nil
}

func (c Console) writeSummary(analysis *analyser.Analysis) {
	boldWriter := color.New(color.Bold)
	successWriter := color.New(color.Bold, color.FgGreen)
	warningWriter := color.New(color.Bold, color.FgYellow)
	errorWriter := color.New(color.Bold, color.FgRed)
	total := boldWriter.Sprintf("%d", analysis.Summary().TotalResources)

	fmt.Printf(
		"Found %s resource(s)\n",
		total,
	)
	fmt.Printf(
		" - %s%% coverage\n",
		boldWriter.Sprintf(
			"%d",
			analysis.Coverage(),
		),
	)
	if !analysis.IsSync() {
		managed := successWriter.Sprintf("0")
		if analysis.Summary().TotalManaged > 0 {
			managed = warningWriter.Sprintf("%d", analysis.Summary().TotalManaged)
		}
		fmt.Printf(" - %s resource(s) managed by Terraform\n", managed)

		drifted := successWriter.Sprintf("0")
		if analysis.Summary().TotalDrifted > 0 {
			drifted = errorWriter.Sprintf("%d", analysis.Summary().TotalDrifted)
		}
		if analysis.Options().Deep {
			fmt.Printf("     - %s resource(s) out of sync with Terraform state\n", boldWriter.Sprintf("%s/%d", drifted, analysis.Summary().TotalManaged))
		}

		unmanaged := successWriter.Sprintf("0")
		if analysis.Summary().TotalUnmanaged > 0 {
			unmanaged = warningWriter.Sprintf("%d", analysis.Summary().TotalUnmanaged)
		}
		fmt.Printf(" - %s resource(s) not managed by Terraform\n", unmanaged)

		deleted := successWriter.Sprintf("0")
		if analysis.Summary().TotalDeleted > 0 {
			deleted = errorWriter.Sprintf("%d", analysis.Summary().TotalDeleted)
		}
		fmt.Printf(" - %s resource(s) found in a Terraform state but missing on the cloud provider\n", deleted)
	}
	if analysis.IsSync() {
		fmt.Println(color.GreenString("Congrats! Your infrastructure is fully in sync."))
	}
}

func prettify(resource interface{}) string {
	res := reflect.ValueOf(resource)
	if resource == nil || res.Kind() == reflect.Ptr && res.IsNil() {
		return "<nil>"
	}

	return awsutil.Prettify(resource)
}

func groupByType(resources []*resource.Resource) (map[string][]*resource.Resource, []string) {
	result := map[string][]*resource.Resource{}
	for _, res := range resources {
		if result[res.ResourceType()] == nil {
			result[res.ResourceType()] = []*resource.Resource{res}
			continue
		}
		result[res.ResourceType()] = append(result[res.ResourceType()], res)
	}

	keys := make([]string, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return result, keys
}

func jsonDiff(a, b interface{}, coloring bool) string {
	aStr := fmt.Sprintf("%s", a)
	bStr := fmt.Sprintf("%s", b)
	d := gojsondiff.New()
	var aJson map[string]interface{}
	_ = json.Unmarshal([]byte(aStr), &aJson)
	result, _ := d.Compare([]byte(aStr), []byte(bStr))
	f := formatter.NewAsciiFormatter(aJson, formatter.AsciiFormatterConfig{
		Coloring: coloring,
	})
	// Set foreground green color for added lines and red color for deleted lines
	formatter.AsciiStyles = map[string]string{
		"+": "32",
		"-": "31",
	}
	diffStr, _ := f.Format(result)

	return diffStr
}

func formatResourceAttributes(res *resource.Resource) string {
	if res.Schema() == nil || res.Schema().HumanReadableAttributesFunc == nil {
		return ""
	}
	attributes := res.Schema().HumanReadableAttributesFunc(res)
	if len(attributes) <= 0 {
		return ""
	}
	// sort attributes
	keys := make([]string, 0, len(attributes))
	for k := range attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// retrieve stringer
	attrString := ""
	for _, k := range keys {
		if attrString != "" {
			attrString += ", "
		}
		attrString += fmt.Sprintf("%s: %s", k, attributes[k])
	}
	return attrString
}
