package output

import (
	"sort"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/output"
)

type Output interface {
	Write(analysis *analyser.Analysis) error
}

var supportedOutputTypes = []string{
	ConsoleOutputType,
	JSONOutputType,
	HTMLOutputType,
	PlanOutputType,
}

var supportedOutputExample = map[string]string{
	ConsoleOutputType: ConsoleOutputExample,
	JSONOutputType:    JSONOutputExample,
	HTMLOutputType:    HTMLOutputExample,
	PlanOutputType:    PlanOutputExample,
}

func SupportedOutputsExample() []string {
	examples := make([]string, 0, len(supportedOutputExample))
	for _, ex := range supportedOutputExample {
		examples = append(examples, ex)
	}
	sort.Strings(examples)
	return examples
}

func Example(key string) string {
	return supportedOutputExample[key]
}

func IsSupported(key string) bool {
	for _, o := range supportedOutputTypes {
		if o == key {
			return true
		}
	}
	return false
}

func GetOutput(config OutputConfig) Output {
	switch config.Key {
	case JSONOutputType:
		return NewJSON(config.Path)
	case HTMLOutputType:
		return NewHTML(config.Path)
	case PlanOutputType:
		return NewPlan(config.Path)
	case ConsoleOutputType:
		fallthrough
	default:
		return NewConsole()
	}
}

// ShouldPrint indicate if we should use the global output or not (e.g. when outputting to stdout).
func ShouldPrint(outputs []OutputConfig, quiet bool) bool {
	for _, c := range outputs {
		p := GetPrinter(c, quiet)
		if _, ok := p.(*output.VoidPrinter); ok {
			return false
		}
	}
	return true
}

func GetPrinter(config OutputConfig, quiet bool) output.Printer {
	if quiet {
		return &output.VoidPrinter{}
	}

	switch config.Key {
	case JSONOutputType:
		if isStdOut(config.Path) {
			return &output.VoidPrinter{}
		}
		fallthrough
	case PlanOutputType:
		if isStdOut(config.Path) {
			return &output.VoidPrinter{}
		}
		fallthrough
	case ConsoleOutputType:
		fallthrough
	default:
		return output.NewConsolePrinter()
	}
}

func isStdOut(path string) bool {
	return path == "/dev/stdout" || path == "stdout"
}
