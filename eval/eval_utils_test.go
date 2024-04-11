package eval_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/nayarsystems/idefix/libraries/eval"
	"github.com/stretchr/testify/assert"
)

func TestMergeExpressions(t *testing.T) {
	cases := []struct {
		Operator      string
		Exprs         []string
		Expected      string
		ExpectedError bool
	}{
		{
			Operator: "$or",
			Exprs:    []string{`{"a":"b"}`},
			Expected: `{"$or":[{"a":"b"}]}`,
		},
		{
			Operator: "$or",
			Exprs:    []string{`{"a":"b"}`, `{"c":9}`},
			Expected: `{"$or":[{"a":"b"},{"c":9}]}`,
		},
		{
			Operator: "$and",
			Exprs:    []string{`{"a":"b"}`, `{"c":9}`, `{"$or":[{"x":2},{"y":"z"}]}`},
			Expected: `{"$and":[{"a":"b"},{"c":9},{"$or":[{"x":2},{"y":"z"}]}]}`,
		},
		{
			Operator: "$and",
			Exprs:    []string{},
			Expected: `{"$and":[]}`,
		},
		{
			Operator:      "$or",
			Exprs:         []string{`{"a":"b"`, `{"c":9}`},
			ExpectedError: true,
		},
	}

	for _, tt := range cases {
		errMsg := fmt.Sprintf("Operator: %q, Exprs: %v, Expected=%v, ExpectedError: %v", tt.Operator, tt.Exprs, tt.Expected, tt.ExpectedError)
		got, err := eval.MergeExpressions(context.Background(), tt.Operator, tt.Exprs...)

		if tt.ExpectedError {
			assert.Error(t, err, errMsg)
		} else {
			assert.Equal(t, tt.Expected, got, errMsg)
		}
	}
}
