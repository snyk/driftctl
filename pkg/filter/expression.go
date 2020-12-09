package filter

import (
	"fmt"

	"github.com/jmespath/go-jmespath"
)

func BuildExpression(expressionStr string) (*jmespath.JMESPath, error) {
	expr, err := jmespath.Compile(fmt.Sprintf("[?%s]", expressionStr))
	if err != nil {
		return nil, err
	}
	return expr, nil
}
