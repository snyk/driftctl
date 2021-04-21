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
}

var supportedOutputExample = map[string]string{
	ConsoleOutputType: ConsoleOutputExample,
	JSONOutputType:    JSONOutputExample,
	HTMLOutputType:    HTMLOutputExample,
}

func SupportedOutputs() []string {
	return supportedOutputTypes
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

func GetOutput(config OutputConfig, quiet bool) Output {
	output.ChangePrinter(GetPrinter(config, quiet))

	switch config.Key {
	case JSONOutputType:
		return NewJSON(config.Options["path"])
	case ConsoleOutputType:
		fallthrough
	case HTMLOutputType:
		return NewHTML(config.Options["path"])
	default:
		return NewConsole()
	}
}

func GetPrinter(config OutputConfig, quiet bool) output.Printer {
	if quiet {
		return &output.VoidPrinter{}
	}

	switch config.Key {
	case JSONOutputType:
		if isStdOut(config.Options["path"]) {
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
