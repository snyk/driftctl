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
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/remote"
	"github.com/cloudskiff/driftctl/pkg/resource"
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
		fmt.Println("Found missing resources:")
		deletedByType, keys := groupByType(analysis.Deleted())
		for _, ty := range keys {
			fmt.Printf("  %s:\n", ty)
			for _, res := range deletedByType[ty] {
				humanString := fmt.Sprintf("    - %s", res.TerraformId())
				if humanAttrs := formatResourceAttributes(res); humanAttrs != "" {
					humanString += fmt.Sprintf("\n        %s", humanAttrs)
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
				humanString := fmt.Sprintf("    - %s", res.TerraformId())
				if humanAttrs := formatResourceAttributes(res); humanAttrs != "" {
					humanString += fmt.Sprintf("\n        %s", humanAttrs)
				}
				fmt.Println(humanString)
			}
		}
	}

	if analysis.Summary().TotalDrifted > 0 {
		fmt.Println("Found changed resources:")
		for _, difference := range analysis.Differences() {
			humanString := fmt.Sprintf("    - %s (%s):", difference.Res.TerraformId(), difference.Res.TerraformType())
			whiteSpace := "        "
			if humanAttrs := formatResourceAttributes(difference.Res); humanAttrs != "" {
				humanString += fmt.Sprintf("\n        %s", humanAttrs)
				whiteSpace = "            "
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
						fmt.Printf("%s%s\n%s%s\n", whiteSpace, pref, prefix, jsonDiff(change.From, change.To, prefix))
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

	c.writeSummary(analysis)

	enumerationErrorMessage := ""
	for _, alerts := range analysis.Alerts() {
		for _, alert := range alerts {
			fmt.Println(color.YellowString(alert.Message()))
			if alert, ok := alert.(*remote.EnumerationAccessDeniedAlert); ok && enumerationErrorMessage == "" {
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
		fmt.Printf(" - %s covered by IaC\n", managed)

		unmanaged := successWriter.Sprintf("0")
		if analysis.Summary().TotalUnmanaged > 0 {
			unmanaged = warningWriter.Sprintf("%d", analysis.Summary().TotalUnmanaged)
		}
		fmt.Printf(" - %s not covered by IaC\n", unmanaged)

		deleted := successWriter.Sprintf("0")
		if analysis.Summary().TotalDeleted > 0 {
			deleted = errorWriter.Sprintf("%d", analysis.Summary().TotalDeleted)
		}
		fmt.Printf(" - %s missing on cloud provider\n", deleted)

		drifted := successWriter.Sprintf("0")
		if analysis.Summary().TotalDrifted > 0 {
			drifted = errorWriter.Sprintf("%d", analysis.Summary().TotalDrifted)
		}
		fmt.Printf(" - %s changed outside of IaC\n", boldWriter.Sprintf("%s/%d", drifted, analysis.Summary().TotalManaged))
	}
	if analysis.IsSync() {
		fmt.Println(color.GreenString("Congrats! Your infrastructure is fully in sync."))
	}
}

func prettify(resource interface{}) string {
	res := reflect.ValueOf(resource)
	if resource == nil || res.Kind() == reflect.Ptr && res.IsNil() {
		return "<null>"
	}

	return awsutil.Prettify(resource)
}

func groupByType(resources []resource.Resource) (map[string][]resource.Resource, []string) {
	result := map[string][]resource.Resource{}
	for _, res := range resources {
		if result[res.TerraformType()] == nil {
			result[res.TerraformType()] = []resource.Resource{res}
			continue
		}
		result[res.TerraformType()] = append(result[res.TerraformType()], res)
	}

	keys := make([]string, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return result, keys
}

func jsonDiff(a, b interface{}, prefix string) string {
	aStr := fmt.Sprintf("%s", a)
	bStr := fmt.Sprintf("%s", b)
	d := gojsondiff.New()
	var aJson map[string]interface{}
	_ = json.Unmarshal([]byte(aStr), &aJson)
	diff, _ := d.Compare([]byte(aStr), []byte(bStr))
	f := formatter.NewAsciiFormatter(aJson, formatter.AsciiFormatterConfig{
		Coloring: isatty.IsTerminal(os.Stdout.Fd()),
	})
	// Set foreground green color for added lines and red color for deleted lines
	formatter.AsciiStyles = map[string]string{
		"+": "32",
		"-": "31",
	}
	diffStr, _ := f.Format(diff)

	return diffStr
}

func formatResourceAttributes(res resource.Resource) string {
	if res.Schema() == nil || res.Schema().HumanReadableAttributesFunc == nil {
		return ""
	}
	attributes := res.Schema().HumanReadableAttributesFunc(res.(*resource.AbstractResource))
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
