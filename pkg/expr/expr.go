package expr

import (
	"github.com/antonmedv/expr"
	"github.com/pkg/errors"
)

func EvalBool(expression string, env map[string]interface{}) (bool, error) {
	eval, err := expr.Eval(expression, env)
	if err != nil {
		return false, err
	}
	if result, ok := eval.(bool); !ok {
		return false, errors.Errorf("expression result is not boolean")
	} else {
		return result, nil
	}
}
