package output

import (
	"fmt"
	"os"
	"sort"

	"github.com/snyk/driftctl/enumeration/remote/alerts"

	"github.com/fatih/color"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/analyser"
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

		unmanaged := successWriter.Sprintf("0")
		if analysis.Summary().TotalUnmanaged > 0 {
			unmanaged = warningWriter.Sprintf("%d", analysis.Summary().TotalUnmanaged)
		}
		deleted := successWriter.Sprintf("0")
		if analysis.Summary().TotalDeleted > 0 {
			deleted = errorWriter.Sprintf("%d", analysis.Summary().TotalDeleted)
		}
		fmt.Printf(" - %s resource(s) not managed by Terraform\n", unmanaged)
		fmt.Printf(" - %s resource(s) found in a Terraform state but missing on the cloud provider\n", deleted)
	}
	if analysis.IsSync() {
		fmt.Println(color.GreenString("Congrats! Your infrastructure is fully in sync."))
	}
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
