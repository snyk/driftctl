package output

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/fatih/color"
	"github.com/nsf/jsondiff"
	"github.com/r3labs/diff/v2"

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
		fmt.Printf("Found deleted resources:\n")
		deletedByType := groupByType(analysis.Deleted())
		for ty, resources := range deletedByType {
			fmt.Printf("  %s:\n", ty)
			for _, res := range resources {
				humanString := res.TerraformId()
				if stringer, ok := res.(fmt.Stringer); ok {
					humanString = stringer.String()
				}
				fmt.Printf("    - %s\n", humanString)
			}
		}
	}

	if analysis.Summary().TotalUnmanaged > 0 {
		fmt.Printf("Found unmanaged resources:\n")
		unmanagedByType := groupByType(analysis.Unmanaged())
		for ty, resource := range unmanagedByType {
			fmt.Printf("  %s:\n", ty)
			for _, res := range resource {
				humanString := res.TerraformId()
				if stringer, ok := res.(fmt.Stringer); ok {
					humanString = stringer.String()
				}
				fmt.Printf("    - %s\n", humanString)
			}
		}
	}

	if analysis.Summary().TotalDrifted > 0 {
		fmt.Printf("Found drifted resources:\n")
		for _, difference := range analysis.Differences() {
			humanString := difference.Res.TerraformId()
			if stringer, ok := difference.Res.(fmt.Stringer); ok {
				humanString = stringer.String()
			}
			fmt.Printf("  - %s (%s):\n", humanString, difference.Res.TerraformType())
			for _, change := range difference.Changelog {
				path := strings.Join(change.Path, ".")
				pref := fmt.Sprintf("%s %s:", color.YellowString("~"), path)
				if change.Type == diff.CREATE {
					pref = fmt.Sprintf("%s %s:", color.GreenString("+"), path)
				} else if change.Type == diff.DELETE {
					pref = fmt.Sprintf("%s %s:", color.RedString("-"), path)
				}
				if change.Type == diff.UPDATE {
					isJsonString := isFieldJsonString(difference.Res, path)
					if isJsonString {
						prefix := "        "
						fmt.Printf("    %s\n%s%s\n", pref, prefix, jsonDiff(change.From, change.To, prefix))
						continue
					}
				}
				fmt.Printf("    %s %s => %s", pref, prettify(change.From), prettify(change.To))
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
			fmt.Printf("%s\n", color.YellowString(alert.Message()))
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
		fmt.Printf(" - %s deleted on cloud provider\n", deleted)

		drifted := successWriter.Sprintf("0")
		if analysis.Summary().TotalDrifted > 0 {
			drifted = errorWriter.Sprintf("%d", analysis.Summary().TotalDrifted)
		}
		fmt.Printf(" - %s drifted from IaC\n", boldWriter.Sprintf("%s/%d", drifted, analysis.Summary().TotalManaged))
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

func groupByType(resources []resource.Resource) map[string][]resource.Resource {
	result := map[string][]resource.Resource{}
	for _, res := range resources {
		if result[res.TerraformType()] == nil {
			result[res.TerraformType()] = []resource.Resource{res}
			continue
		}
		result[res.TerraformType()] = append(result[res.TerraformType()], res)
	}
	return result
}

func isFieldJsonString(res resource.Resource, fieldName string) bool {
	t := reflect.TypeOf(res)
	var field reflect.StructField
	var ok bool
	if t.Kind() == reflect.Ptr {
		field, ok = t.Elem().FieldByName(fieldName)
	}
	if t.Kind() != reflect.Ptr {
		field, ok = t.FieldByName(fieldName)
	}
	if !ok {
		return false
	}

	return field.Tag.Get("jsonstring") == "true"
}

func jsonDiff(a, b interface{}, prefix string) string {
	aStr := fmt.Sprintf("%s", a)
	bStr := fmt.Sprintf("%s", b)
	opts := jsondiff.DefaultConsoleOptions()
	opts.Prefix = prefix
	opts.Indent = "  "
	opts.Added = jsondiff.Tag{
		Begin: color.GreenString("+ "),
	}
	opts.Removed = jsondiff.Tag{
		Begin: color.RedString("- "),
	}
	opts.Changed = jsondiff.Tag{
		Begin: color.YellowString("~ "),
	}
	_, str := jsondiff.Compare([]byte(aStr), []byte(bStr), &opts)
	return str
}
