package eval

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"

	"time"

	"github.com/jaracil/ei"
	"github.com/pkg/errors"
)

const (
	// ResNotMatch when expression don't match using env values
	ResNotMatch ResultRes = 0
	// ResOK when expression matches using env values
	ResOK ResultRes = 1
	// ResInvalidExpr when the expression is invalid
	ResInvalidExpr ResultRes = -1
	// ResInvalidOp when the operator is invalid
	ResInvalidOp ResultRes = -2
	// ResInvalidEnv when the env is invalid
	ResInvalidEnv ResultRes = -3
	// ResInvalidEnvItem when the env item is invalid
	ResInvalidEnvItem ResultRes = -4
	// ResTypeMismatch when the item type on the expression doesn't match the one from env
	ResTypeMismatch ResultRes = -5
)

// ResultRes is the numeric code result of the evaluation
type ResultRes int

// Result of the evaluation
type Result struct {
	Res  ResultRes
	Iden string
}

type evalCtx struct {
	Iden string
}

type CompiledExpr = map[string]interface{}

var MockedNow = time.Now

// Eval evaluates expression with env returning Result
func Eval(expr string, env map[string]interface{}) Result {
	setBaseEnv(env)

	exprMap, err := CompileExpr(expr)
	if err != nil {
		return Result{Res: ResInvalidExpr, Iden: ""}
	}
	ctx := evalCtx{}
	res := eval(&ctx, exprMap, env)
	return Result{Res: res, Iden: ctx.Iden}
}

// Eval evaluates expression with env returning Result
func EvalCompiled(exprMap CompiledExpr, env map[string]interface{}) Result {
	setBaseEnv(env)
	ctx := evalCtx{}
	res := eval(&ctx, exprMap, env)
	return Result{Res: res, Iden: ctx.Iden}
}

func setBaseEnv(env map[string]interface{}) {
	now := MockedNow()
	env["wday"] = int(now.Weekday())
	env["mon"] = int(now.Month()) - 1
	env["mday"] = now.Day()
	env["year"] = now.Year() - 1900
	env["hour"] = now.Hour()
	env["min"] = now.Minute()
	env["sec"] = now.Second()
	env["daymin"] = now.Hour()*60 + now.Minute()
	env["daysec"] = now.Hour()*3600 + now.Minute()*60 + now.Second()
}

func CompileExpr(expr string) (decoded CompiledExpr, err error) {
	exprMap := CompiledExpr{}

	if err := json.Unmarshal([]byte(expr), &exprMap); err != nil {
		return nil, err
	}
	return exprMap, nil
}

func eval(ctx *evalCtx, expr CompiledExpr, env map[string]interface{}) ResultRes {
	if len(expr) != 1 {
		return ResInvalidExpr
	}

	for k, v := range expr {
		switch k {
		case "$not":
			return evalNot(ctx, v, env)
		case "$exists":
			return evalExists(ctx, v, env)
		case "$or", "$and", "$nor":
			return evalListOp(ctx, v, env, k)
		case "$true":
			return ResOK
		case "$false":
			return ResNotMatch
		default:
			ctx.Iden = k

			if len(k) > 0 && k[0] == '$' {
				return ResInvalidOp
			}

			valEnv := getEnvValue(env, k)
			return evalSimple(ctx, v, valEnv)
		}
	}

	return ResInvalidExpr
}

func getEnvValue(env map[string]interface{}, key string) interface{} {
	if env == nil {
		return nil
	}

	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		nestedEnv, isMap := env[parts[0]].(map[string]interface{})
		if !isMap {
			return nil
		}

		return getEnvValue(nestedEnv, strings.Join(parts[1:], "."))
	}

	return env[key]
}

func canCastToNumber(val interface{}) bool {
	_, isBool := val.(bool)
	return isBool || isNumber(val)
}

func isNumber(val interface{}) bool {
	t := reflect.TypeOf(val)
	if t == nil {
		return false
	}
	return t.Kind() == reflect.Float32 || t.Kind() == reflect.Float64 ||
		t.Kind() == reflect.Int || t.Kind() == reflect.Int8 || t.Kind() == reflect.Int16 || t.Kind() == reflect.Int32 || t.Kind() == reflect.Int64 ||
		t.Kind() == reflect.Uint || t.Kind() == reflect.Uint8 || t.Kind() == reflect.Uint16 || t.Kind() == reflect.Uint32 || t.Kind() == reflect.Uint64
}

func toFloat64(val interface{}) (float64, error) {
	if isNumber(val) {
		return ei.N(val).Float64()
	}
	if valBool, ok := val.(bool); ok {
		if valBool {
			return 1, nil
		}
		return 0, nil
	}
	return 0, errors.Errorf("error converting to float64: %v", val)
}

func isString(val interface{}) bool {
	_, ok := val.(string)
	return ok
}

func evalSimple(ctx *evalCtx, valExpr, valEnv interface{}) ResultRes {
	if valExprMap, ok := valExpr.(map[string]interface{}); ok {
		return evalOp(valExprMap, valEnv)
	}

	return evalCmp(valExpr, valEnv, "$eq")
}

func evalCmp(valExpr, valEnv interface{}, op string) ResultRes {
	if _, ok := valExpr.(map[string]interface{}); ok {
		return ResInvalidExpr
	}

	if (canCastToNumber(valExpr) && canCastToNumber(valEnv)) ||
		(isString(valExpr) && isString(valEnv)) {

		cmpRes, err := cmp(valEnv, valExpr)
		if err != nil {
			return ResInvalidExpr
		}

		switch op {
		case "$eq":
			if cmpRes == 0 {
				return ResOK
			}
			return ResNotMatch
		case "$ne":
			if cmpRes == 0 {
				return ResNotMatch
			}
			return ResOK
		case "$gt":
			if cmpRes > 0 {
				return ResOK
			}
			return ResNotMatch
		case "$gte":
			if cmpRes >= 0 {
				return ResOK
			}
			return ResNotMatch
		case "$lt":
			if cmpRes < 0 {
				return ResOK
			}
			return ResNotMatch
		case "$lte":
			if cmpRes <= 0 {
				return ResOK
			}
			return ResNotMatch
		}
	}

	switch op {
	case "$in":
		return in(valEnv, valExpr)
	case "$nin":
		res := in(valEnv, valExpr)
		if res == ResOK {
			return ResNotMatch
		} else if res == ResNotMatch {
			return ResOK
		} else {
			return res
		}
	case "$regex":
		re, isStr := valExpr.(string)
		if !isStr {
			return ResInvalidExpr
		}

		if re == "" {
			return ResInvalidExpr
		}

		if !isString(valEnv) {
			return ResInvalidEnvItem
		}

		match, err := regexp.MatchString(re, valEnv.(string))
		if err != nil {
			return ResInvalidExpr
		}

		if match {
			return ResOK
		}
		return ResNotMatch
	}

	if valEnv == nil {
		return ResInvalidEnvItem
	}

	return ResTypeMismatch
}

func in(elem interface{}, slice interface{}) ResultRes {
	sliceVal, ok := slice.([]interface{})
	if !ok {
		return ResInvalidExpr
	}

	if elem == nil {
		return ResInvalidEnvItem
	}

	if len(sliceVal) == 0 {
		return ResNotMatch
	}

	for _, el := range sliceVal {
		if el == elem {
			return ResOK
		}
	}
	return ResNotMatch
}

func cmp(valEnv, valExpr interface{}) (int, error) {
	if canCastToNumber(valExpr) && canCastToNumber(valEnv) {
		valExprNum, err := toFloat64(valExpr)
		if err != nil {
			return 0, err
		}

		valEnvNum, err := toFloat64(valEnv)
		if err != nil {
			return 0, err
		}

		if valEnvNum == valExprNum {
			return 0, nil
		} else if valEnvNum > valExprNum {
			return 1, nil
		} else {
			return -1, nil
		}
	} else if isString(valExpr) && isString(valEnv) {
		return strings.Compare(valEnv.(string), valExpr.(string)), nil
	} else {
		return 0, errors.New("operands with different types")
	}
}

func evalOp(valExpr map[string]interface{}, valEnv interface{}) ResultRes {
	if len(valExpr) != 1 {
		return ResInvalidExpr
	}

	if valEnv == nil {
		return ResInvalidEnvItem
	}

	for k, v := range valExpr {
		return evalCmp(v, valEnv, k)
	}

	return ResInvalidExpr
}

func evalNot(ctx *evalCtx, expr interface{}, env map[string]interface{}) ResultRes {
	exprMap, ok := expr.(map[string]interface{})
	if !ok {
		return ResInvalidExpr
	}

	res := eval(ctx, exprMap, env)
	if res == ResNotMatch {
		return ResOK
	} else if res == ResOK {
		return ResNotMatch
	} else {
		return res
	}
}

func evalExists(ctx *evalCtx, expr interface{}, env map[string]interface{}) ResultRes {
	key, ok := expr.(string)
	if !ok {
		return ResInvalidExpr
	}

	val := getEnvValue(env, key)

	if val == nil {
		return ResNotMatch
	}

	return ResOK
}

func evalListOp(ctx *evalCtx, expr interface{}, env map[string]interface{}, op string) ResultRes {
	opItems, ok := expr.([]interface{})
	if !ok {
		return ResInvalidExpr
	}

	for _, opItem := range opItems {
		opItemMap, ok := opItem.(map[string]interface{})
		if !ok {
			return ResInvalidExpr
		}

		res := eval(ctx, opItemMap, env)
		if res < 0 {
			return res
		} else if res == ResOK {
			if op == "$or" {
				return ResOK
			} else if op == "$nor" {
				return ResNotMatch
			}
		} else {
			if op == "$and" {
				return ResNotMatch
			}
		}
	}

	if op == "$or" {
		return ResNotMatch
	} else if op == "$and" {
		return ResOK
	} else {
		return ResOK
	}
}
