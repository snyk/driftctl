package filter

import (
	"fmt"

	"github.com/jmespath/go-jmespath"
)

func BuildExpression(expressionStr string) (*jmespath.JMESPath, error) {
	sprintf := fmt.Sprintf("[?%s]", expressionStr)

	parse, _ := jmespath.NewParser().Parse(sprintf)
	fmt.Println(parse.String())

	expr, err := jmespath.Compile(sprintf)
	if err != nil {
		return nil, err
	}
	return expr, nil
}
