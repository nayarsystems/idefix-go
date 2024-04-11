package eval

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type msi = map[string]interface{}

// MergeExpressions merges exprs using operation (e.g. "$and", "$or"...)
func MergeExpressions(ctx context.Context, operation string, exprs ...string) (string, error) {
	exprsSlice := []msi{}

	for _, expr := range exprs {
		var exprMap msi
		if err := json.Unmarshal([]byte(expr), &exprMap); err != nil {
			return "", fmt.Errorf("error unmarshalling access expression when merging expressions: expr=%q, err=%+v", expr, err)
		}

		exprsSlice = append(exprsSlice, exprMap)
	}

	resultExpr := msi{operation: exprsSlice}

	result, err := json.Marshal(resultExpr)
	if err != nil {
		return "", errors.Errorf("error marshalling result after merging expressions: resultExpr=%#v, err=%+v", resultExpr, err)
	}

	return string(result), err
}
