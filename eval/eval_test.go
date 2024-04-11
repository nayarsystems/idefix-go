package eval

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type testCase struct {
	MockTime     string
	Expr         string
	Env          msi
	ExpectedRes  ResultRes
	ExpectedIden string
}

func TestExpressions(t *testing.T) {
	cases := readCasesFromFile(t, "testdata/test.tsv")
	cases = append(cases, numberCases()...)
	cases = append(cases, dateCases()...)
	cases = append(cases, malformedExprCases()...)

	for _, tt := range cases {
		if tt.MockTime != "" {
			MockTime(t, tt.MockTime)
		}

		errMsg := fmt.Sprintf("Expr: %q, Env: %+v -> ExpectedRes=%v, ExpectedIden=%v", tt.Expr, tt.Env, tt.ExpectedRes, tt.ExpectedIden)
		got := Eval(tt.Expr, tt.Env)

		assert.Equal(t, tt.ExpectedRes, got.Res, errMsg)
		assert.Equal(t, tt.ExpectedIden, got.Iden, errMsg)
	}
}

func readCasesFromFile(t *testing.T, fileName string) []*testCase {
	cases := []*testCase{}

	file, err := os.Open(fileName)
	if err != nil {
		t.Fatalf("Error opening test cases file: %v", err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, "\t")

		numParts := len(parts)
		if numParts < 4 || numParts > 5 {
			t.Fatalf("Test case line should have 4 (or 5) elements separated by tabs: %q", line)
		}

		tCase := testCase{Expr: parts[0]}

		if err := json.Unmarshal([]byte(parts[1]), &tCase.Env); err != nil {
			t.Fatalf("Error parsing element %q as JSON in line %q", parts[1], line)
		}

		res, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil {
			t.Fatalf("Expected int result code on %q in line %q", parts[2], line)
		}

		tCase.ExpectedRes = ResultRes(res)
		tCase.ExpectedIden = strings.TrimSpace(parts[3])

		cases = append(cases, &tCase)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading test cases: %v", err)
	}

	return cases
}

func malformedExprCases() []*testCase {
	return []*testCase{
		{
			Expr:        `"a"`,
			Env:         msi{},
			ExpectedRes: ResInvalidExpr,
		},
		{
			Expr:        `{`,
			Env:         msi{},
			ExpectedRes: ResInvalidExpr,
		},
		{
			Expr:        `"a"}`,
			Env:         msi{},
			ExpectedRes: ResInvalidExpr,
		},
	}
}

func numberCases() []*testCase {
	return []*testCase{
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": int(1)},
			ExpectedRes:  ResOK,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": int(2)},
			ExpectedRes:  ResNotMatch,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": int8(1)},
			ExpectedRes:  ResOK,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": int8(2)},
			ExpectedRes:  ResNotMatch,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": int16(1)},
			ExpectedRes:  ResOK,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": int16(2)},
			ExpectedRes:  ResNotMatch,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": int32(1)},
			ExpectedRes:  ResOK,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": int32(2)},
			ExpectedRes:  ResNotMatch,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": int64(1)},
			ExpectedRes:  ResOK,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": int64(2)},
			ExpectedRes:  ResNotMatch,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": uint(1)},
			ExpectedRes:  ResOK,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": uint(2)},
			ExpectedRes:  ResNotMatch,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": uint8(1)},
			ExpectedRes:  ResOK,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": uint8(2)},
			ExpectedRes:  ResNotMatch,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": uint16(1)},
			ExpectedRes:  ResOK,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": uint16(2)},
			ExpectedRes:  ResNotMatch,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": uint32(1)},
			ExpectedRes:  ResOK,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": uint32(2)},
			ExpectedRes:  ResNotMatch,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": uint64(1)},
			ExpectedRes:  ResOK,
			ExpectedIden: "a",
		},
		{
			Expr:         `{"a":1}`,
			Env:          msi{"a": uint64(2)},
			ExpectedRes:  ResNotMatch,
			ExpectedIden: "a",
		},
	}
}

func dateCases() []*testCase {
	return []*testCase{
		{
			MockTime:     "2020-06-12T10:07:43Z",
			Expr:         `{"$and":[{"wday": 5}, {"mon": 5}, {"mday": 12}, {"year": 120}, {"hour": 10}, {"min": 7}, {"sec": 43}, {"daymin": 607}, {"daysec": 36463}]}`,
			Env:          msi{},
			ExpectedRes:  ResOK,
			ExpectedIden: "daysec",
		},
	}
}

// MockTime changes clock.MockedNow() so it returns the passed time
func MockTime(t *testing.T, rfc3339 string) {
	MockedNow = func() time.Time {
		return Time(t, rfc3339)
	}
}

// Time returns time specified by rfc3339 string
func Time(t *testing.T, rfc3339 string) time.Time {
	now, err := time.Parse(time.RFC3339Nano, rfc3339)
	assert.NoError(t, err)
	return now
}
