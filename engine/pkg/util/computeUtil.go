package util

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bilibili/gengine/builder"
	"github.com/bilibili/gengine/context"
	"github.com/bilibili/gengine/engine"
	"github.com/goccy/go-json"
	"github.com/mattn/go-runewidth"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	ComputeMap = map[string]func(xs ...float64) float64{
		"max":   Max,     // 最大值
		"min":   Min,     // 最小值
		"sum":   Sum,     // 求和
		"avg":   Average, // 平均值
		"round": Round,   // 4舍5入取整
		"int":   Int,     // 向下取整
		"mod":   Mod,     // 取余
		"power": Power,   // 取次方数
		"abs":   Abs,     // 取绝对值
		"log":   Logs,    // 取对数
	}
	ComputeIFMap = map[string]func(values []interface{}, expression string) (result float64){
		"countif": CountIf, // 满足给定条件的个数
		"sumif":   SumIf,   // 对满足条件的数求和
	}
	ComputeLenMap = map[string]func(val string) (result int){
		"len":  Len,  // 获取长度
		"lenb": LenB, // 获取字符长度
	}
	ComputeLRMap = map[string]func(text string, n ...int) (result string){
		"left":  Left,  // 从一个文本字符串的第一个字符开始返回指定个数的字符
		"right": Right, // 从一个文本字符串的最后一个字符开始返回指定个数的字符
	}
	IFConstant         = "IF"         // 条件
	ISNumberConstant   = "ISNUMBER"   // 检查一个数是否为数值
	MidConstant        = "MID"        // 取字符串中间的值
	FindConstant       = "FIND"       // 返回一个字符串在另一个字符串中出现的起始位置
	ConcatConstant     = "CONCAT"     // 拼接多个文本
	SubstituteConstant = "SUBSTITUTE" // 将字符串中的部分字符替换为新的字符
	ReplaceConstant    = "REPLACE"    // 将一个字符串中的部分字符用另一个字符串替换

	Number        = "number"
	String        = "string"
	Date          = "date"
	OperatorSlice = []string{">", ">=", "<", "<=", "=", "==", "!=", "<>"}
)

// MediumToLast 将计算表达式转换为后缀表达式
func MediumToLast(expression string, output string) string {
	var sta Stack
	b := CharAt(expression)
	for i := 0; i < len(b); i++ {
		char := b[i]
		char = PercentToDecimal(char)
		switch char {
		case "-":
			if i == 0 || (i > 0 && (b[i-1] == "+" || b[i-1] == "-" || b[i-1] == "/" || b[i-1] == "÷" || b[i-1] == "*" || b[i-1] == "×" || b[i-1] == "(")) {
				output += char + " "
				continue
			}
			output = doOperate(char, 1, output, &sta)
		case "+":
			output = doOperate(char, 1, output, &sta)
		case "/":
			output = doOperate(char, 2, output, &sta)
		case "÷":
			output = doOperate(char, 2, output, &sta)
		case "*":
			output = doOperate(char, 2, output, &sta)
		case "×":
			output = doOperate(char, 2, output, &sta)
		case "(":
			sta.push(char)
		case ")":
			output = gotParOperate(output, &sta)
		default:
			output += char + " "
		}
	}
	for len(sta.List) > 0 {
		o, _ := sta.pop()
		output += o + " "
	}
	return output[0 : len(output)-1]
}

// CharAt 计算表达式切割
func CharAt(expression string) []string {
	var charStr []string
	var ch = []rune(expression)
	var str string
	for i, _ := range ch {
		//for i := 0; i < len(ch); i++ {
		if ch[i] == '+' || ch[i] == '-' || ch[i] == '*' || ch[i] == '/' || ch[i] == '(' || ch[i] == ')' || ch[i] == '×' || ch[i] == '÷' || ch[i] == '=' || ch[i] == '>' || ch[i] == '<' || ch[i] == '!' || ch[i] == '|' || ch[i] == '&' || ch[i] == ',' {
			if len(str) > 0 {
				charStr = append(charStr, str)
				str = ""
			}
			// 上一个符号也是运算符或是第一个字符
			if (i > 0 && ch[i] == '-' && (ch[i-1] == '+' || ch[i-1] == '-' || ch[i-1] == '*' || ch[i-1] == '/' || ch[i-1] == '×' || ch[i-1] == '÷')) || (i == 0 && ch[i] == '-') {
				str += string(ch[i])
				continue
			}
			charStr = append(charStr, string(ch[i]))
		} else {
			str += string(ch[i])
			if i == len(ch)-1 && len(str) > 0 {
				charStr = append(charStr, str)
			}
		}
	}
	return charStr
}

func doOperate(ch string, per int64, output string, sta *Stack) string {
	for len(sta.List) > 0 {
		var peri int64 = 0
		if ch == "(" {
			sta.push(ch)
			break
		}
		chp, _ := sta.pop()
		if chp == "+" || chp == "-" {
			peri = 1
		}
		if chp == "*" || chp == "/" || chp == "×" || chp == "÷" {
			peri = 2
		}
		//if chp == "max" || chp == "min" || chp == "sum" || chp == "avg" {
		//	peri = 3
		//}
		if per <= peri {
			output += chp + " "
		} else {
			sta.push(chp)
			break
		}
	}
	sta.push(ch)
	return output
}

func gotParOperate(output string, sta *Stack) string {
	for len(sta.List) > 0 {
		ch, _ := sta.pop()
		if ch == "(" {
			break
		} else {
			output += ch + " "
		}
	}
	return output
}

// ComputeResult 计算不包含函数的公式
func ComputeResult(lastCompute string) string {
	lastCompute = strings.Trim(lastCompute, " ")
	b := strings.Split(lastCompute, " ")
	var sumSta Stack
	for i := 0; i < len(b); i++ {
		char := b[i]
		if len(char) == 0 {
			continue
		}
		if char == "+" || char == "-" || char == "*" || char == "/" || char == "×" || char == "÷" {
			s1, _ := sumSta.pop()
			s2, _ := sumSta.pop()
			x, _ := strconv.ParseFloat(s1, 64)
			y, _ := strconv.ParseFloat(s2, 64)
			if char == "+" {
				y = y + x
			}
			if char == "-" {
				y -= x
			}
			if char == "*" || char == "×" {
				y = y * x
			}
			if char == "/" || char == "÷" {
				if y == 0 {
					y = 0
				} else {
					y = y / x
				}
			}
			sumSta.push(strconv.FormatFloat(y, 'f', 16, 64))
		} else {
			sumSta.push(string(b[i]))
		}
	}
	sum, _ := sumSta.pop()
	return sum
}

// ComputeFuncResult 计算包含函数的公式
func ComputeFuncResult(expression string) string {
	var resultFloat64 float64
	// 验证表达式中是否包含函数关键字 (不区分大小写)
	if ComputeMapContains(expression) {
		// 有函数表达式
		var xs []float64
		// 获取此函数名称
		lk := strings.Index(expression, "(")
		fn := expression[0:lk]
		// 分割参数
		idx := strings.Split(expression, ",")
		// 反拼接
		for i, v := range idx {
			if i > 0 && ComputeMapContains(v) {
				for j := 0; j < len(idx)-i; j++ {
					idx[i] = idx[i] + "," + idx[i+1]
					idx = append(idx[:i+1], idx[i+2:]...)
					if strings.Contains(idx[i], ")") {
						break
					}
				}
			}
		}
		for i, s := range idx {
			if i == 0 {
				// 处理第一个参数
				xs = AppendComputeResult(xs, expression[len(fn)+1:len(s)], 2)
			} else if i == len(idx)-1 {
				// 处理最后一个参数
				xs = AppendComputeResult(xs, s[:len(s)-1], 2)
			} else {
				// 处理中间参数
				xs = AppendComputeResult(xs, s, 2)
			}
		}
		// 执行基础函数
		if ComputeMap[fn] != nil {
			// 执行对应计算函数
			resultFloat64 = ComputeMap[strings.ToLower(fn)](xs...)
		}
	} else {
		// 无函数表达式
		var output string
		lastExp := MediumToLast(expression, output)
		result := ComputeResult(lastExp)
		resultFloat64, _ = strconv.ParseFloat(result, 64)
	}
	return strconv.FormatFloat(resultFloat64, 'f', 2, 64)
}

// SplitAtCommas split s at commas, ignoring commas in strings.
func SplitAtCommas(s string) []string {
	res := []string{}
	var beg int
	var inString bool

	for i := 0; i < len(s); i++ {
		if s[i] == ',' && !inString {
			res = append(res, s[beg:i])
			beg = i + 1
		} else if s[i] == '"' {
			if !inString {
				inString = true
			} else if i > 0 && s[i-1] != '\\' {
				inString = false
			}
		}
	}
	return append(res, s[beg:])
}

// ComputeFuncResult1 函数表达式计算,计算精度
func ComputeFuncResult1(expressions string, precision int, isNextCompute bool) (string, error) {
	if len(expressions) == 0 {
		return expressions, nil
	}
	var resultFloat64 float64
	// 验证表达式中是否包含函数关键字 (不区分大小写)
	if ComputeMapContains(expressions) {
		//r, _ := regexp.Compile("([a-z]*)(\\([^(^)]*?\\))")
		r, _ := regexp.Compile("([a-z]*)(\\([^()]*?\\))")
		expressionList := r.FindAllString(expressions, 100)
		// 替换后字符串
		var expressionLast string
		for _, expressionStr := range expressionList {
			// 获取此函数名称
			lk := strings.Index(expressionStr, "(")
			fn := expressionStr[0:lk]
			// 分割参数
			idx := strings.Split(expressionStr[lk+1:len(expressionStr)-1], ",")
			// 计算函数
			if ComputeMap[strings.ToLower(fn)] != nil {
				var xs []float64
				for _, s := range idx {
					xs = AppendComputeResult(xs, s, precision)
				}
				resultFloat64 = ComputeMap[strings.ToLower(fn)](xs...)
				expressionLast = strconv.FormatFloat(resultFloat64, 'f', precision, 64)
			} else if strings.ToLower(fn) == strings.ToLower(IFConstant) {
				// IF函数
				var p1 string
				var p2 string
				if IsFormula(idx[1]) {
					p1, _ = ComputeFuncResult1(idx[1], precision, isNextCompute)
				} else {
					p1 = idx[1]
				}
				if IsFormula(idx[2]) {
					p2, _ = ComputeFuncResult1(idx[2], precision, isNextCompute)
				} else {
					p2 = idx[2]
				}
				expressionLast = IF2(idx[0], p1, p2).(string)
			} else if ComputeIFMap[strings.ToLower(fn)] != nil {
				idxN := ConvertSlice(idx)
				var v1 []any
				json.Unmarshal([]byte(idxN[0]), &v1)
				resultFloat64 = ComputeIFMap[strings.ToLower(fn)](v1, idxN[1])
				expressionLast = strconv.FormatFloat(resultFloat64, 'f', precision, 64)
			} else if ComputeLenMap[strings.ToLower(fn)] != nil {
				resultInt := ComputeLenMap[strings.ToLower(fn)](idx[0])
				expressionLast = strconv.Itoa(resultInt)
			} else if ComputeLRMap[strings.ToLower(fn)] != nil {
				if len(idx) == 1 {
					expressionLast = ComputeLRMap[strings.ToLower(fn)](idx[0])
				} else {
					if strings.Contains(idx[1], ".") {
						idx[1] = Split(idx[1], 0, strings.Index(idx[1], "."))
					}
					n, _ := strconv.Atoi(idx[1])
					expressionLast = ComputeLRMap[strings.ToLower(fn)](idx[0], n)
				}
			} else if strings.ToLower(fn) == strings.ToLower(ISNumberConstant) {
				b := IsDigit(idx[0])
				expressionLast = strconv.FormatBool(b)
			} else if strings.ToLower(fn) == strings.ToLower(ConcatConstant) {
				var xs []string
				for _, s := range idx {
					xs = append(xs, s)
				}
				expressionLast = Concat(xs...)
			} else if strings.ToLower(fn) == strings.ToLower(MidConstant) {
				//if !IsDigit(idx[1]) || !IsDigit(idx[2]) {
				//	return "", errors.New(expressions + "参数传入有误")
				//}
				si, _ := strconv.ParseFloat(idx[1], 64)
				chartsNum, _ := strconv.ParseFloat(idx[2], 64)
				s, err := Mid(idx[0], si, chartsNum)
				if err != nil {
					return "", errors.New(expressions + err.Error())
				}
				expressionLast = s
			} else if strings.ToLower(fn) == strings.ToLower(FindConstant) {
				var f1 []float64
				for i, v := range idx {
					if i < 2 {
						continue
					}
					v1, _ := strconv.ParseFloat(v, 64)
					f1 = append(f1, v1)
				}
				i, err := Find(idx[0], idx[1], f1...)
				if err != nil {
					return "", errors.New(expressions + err.Error())
				}
				expressionLast = strconv.Itoa(i)
			} else if strings.ToLower(fn) == strings.ToLower(SubstituteConstant) {
				if len(idx) > 3 {
					n, _ := strconv.Atoi(idx[3])
					expressionLast = Substitute(idx[0], idx[1], idx[2], n)
				} else {
					expressionLast = Substitute(idx[0], idx[1], idx[2])
				}
			} else if strings.ToLower(fn) == strings.ToLower(ReplaceConstant) {
				idx1, _ := strconv.Atoi(idx[1])
				idx2, _ := strconv.Atoi(idx[2])
				result, err := Replace(idx[0], idx1, idx2, idx[3])
				if err != nil {
					return "", errors.New(expressions + err.Error())
				}
				expressionLast = result
			} else {
				expressionLast, _ = ComputeFuncResult1(idx[0], precision, isNextCompute)
			}
			// 替换表达式
			expressions = strings.ReplaceAll(expressions, expressionStr, expressionLast)
			// 验证是否继续做下一步计算
			if !IsDigit(expressionLast) {
				isNextCompute = false
			}
			return ComputeFuncResult1(expressions, precision, isNextCompute)
		}
	} else {
		if !isNextCompute {
			return expressions, nil
		}
		// 无函数表达式
		var output string
		lastExp := MediumToLast(expressions, output)
		result := ComputeResult(lastExp)
		if !IsDigit(result) {
			return result, nil
		}
		resultFloat64, _ = strconv.ParseFloat(result, 64)
		if resultFloat64 > 9.999999999999998e+15 {
			return result, nil
		}
	}
	return strconv.FormatFloat(resultFloat64, 'f', precision, 64), nil
}

//MultistageExpression 多段表达式处理
func MultistageExpression(str string) string {
	if ComputeMapContains(str) {
		var ch = []rune(str)
		var ir = 0
		var expression string
		var expressions []string
		var b bool
		for i, r := range ch {
			expression += string(ch[i])
			if r == '(' {
				b = true
				ir = ir + 1
			}
			if r == ')' {
				ir = ir - 1
			}
			if r == '+' || r == '-' || r == '*' || r == '/' || r == '×' || r == '÷' && IsDigit(expression) {
				expressions = append(expressions, expression[:len(expression)-1])
				expression = expression[len(expression)-1:]
			}
			if ir == 0 && b {
				expressions = append(expressions, expression)
				expression = ""
				b = false
			}
		}
		if len(expression) > 0 {
			expressions = append(expressions, expression)
		}
		var resultExpression string
		for i, s := range expressions {
			if IsDigit(s) {
				resultExpression += s
			} else {
				if i == 0 {
					resultExpression += ComputeFuncResult(s)
				} else {
					resultExpression += s[:1]
					if ComputeMapContains(s[1:]) {
						resultExpression += ComputeFuncResult(s[1:])
					} else {
						resultExpression += s[1:]
					}
				}
			}
		}
		return ComputeFuncResult(resultExpression)
	} else {
		return ComputeFuncResult(str)
	}
}

// ComputeMapContains 验证是否包含内置函数
func ComputeMapContains(expression string) bool {
	for k, _ := range ComputeMap {
		if strings.Contains(strings.ToLower(expression), strings.ToLower(k)) {
			return true
		}
	}
	for k, _ := range ComputeIFMap {
		if strings.Contains(strings.ToLower(expression), strings.ToLower(k)) {
			return true
		}
	}
	for k, _ := range ComputeLenMap {
		if strings.Contains(strings.ToLower(expression), strings.ToLower(k)) {
			return true
		}
	}
	for k, _ := range ComputeLRMap {
		if strings.Contains(strings.ToLower(expression), strings.ToLower(k)) {
			return true
		}
	}
	if strings.Contains(strings.ToLower(expression), strings.ToLower(IFConstant)) ||
		strings.Contains(strings.ToLower(expression), strings.ToLower(FindConstant)) ||
		strings.Contains(strings.ToLower(expression), strings.ToLower(ISNumberConstant)) ||
		strings.Contains(strings.ToLower(expression), strings.ToLower(ConcatConstant)) ||
		strings.Contains(strings.ToLower(expression), strings.ToLower(MidConstant)) ||
		strings.Contains(strings.ToLower(expression), strings.ToLower(SubstituteConstant)) ||
		strings.Contains(strings.ToLower(expression), strings.ToLower(ReplaceConstant)) {
		return true
	}
	return false
}

// AppendComputeResult 添加计算结果
func AppendComputeResult(xs []float64, s string, p int) []float64 {
	var r float64
	if IsDigit(s) {
		// 是数字直接添加
		r, _ = strconv.ParseFloat(s, 64)
	} else {
		// 继续解析函数表达式
		e, _ := ComputeFuncResult1(s, p, true)
		r, _ = strconv.ParseFloat(e, 64)
	}
	return append(xs, r)
}

// IsDigit 验证字符是否为数字
func IsDigit(str string) bool {
	//r, _ := regexp.Compile("^-?[0-9]+(.[0-9]+)?$")
	//return r.MatchString(str)
	if strings.HasPrefix(str, "-") {
		str = str[1:]
	}
	if strings.HasPrefix(str, ".") || strings.HasSuffix(str, ".") {
		return false
	}
	str = strings.ReplaceAll(str, ".", "")
	for _, x := range []rune(str) {
		if !unicode.IsDigit(x) {
			return false
		}
	}
	return true
}

// IsDate 验证字符是否为日期
func IsDate(str string) bool {
	str = strings.Trim(str, " ")
	r, _ := regexp.Compile("((^((1[8-9]\\d{2})|([2-9]\\d{3}))([-\\/\\._])(10|12|0?[13578])([-\\/\\._])(3[01]|[12][0-9]|0?[1-9])$)|(^((1[8-9]\\d{2})|([2-9]\\d{3}))([-\\/\\._])(11|0?[469])([-\\/\\._])(30|[12][0-9]|0?[1-9])$)|(^((1[8-9]\\d{2})|([2-9]\\d{3}))([-\\/\\._])(0?2)([-\\/\\._])(2[0-8]|1[0-9]|0?[1-9])$)|(^([2468][048]00)([-\\/\\._])(0?2)([-\\/\\._])(29)$)|(^([3579][26]00)([-\\/\\._])(0?2)([-\\/\\._])(29)$)|(^([1][89][0][48])([-\\/\\._])(0?2)([-\\/\\._])(29)$)|(^([2-9][0-9][0][48])([-\\/\\._])(0?2)([-\\/\\._])(29)$)|(^([1][89][2468][048])([-\\/\\._])(0?2)([-\\/\\._])(29)$)|(^([2-9][0-9][2468][048])([-\\/\\._])(0?2)([-\\/\\._])(29)$)|(^([1][89][13579][26])([-\\/\\._])(0?2)([-\\/\\._])(29)$)|(^([2-9][0-9][13579][26])([-\\/\\._])(0?2)([-\\/\\._])(29)$))")
	t, _ := regexp.Compile("^((20|21|22|23|[0-1]?\\d):[0-5]?\\d:[0-5]?\\d)$")
	if strings.Contains(str, " ") {
		return r.MatchString(str[0:strings.Index(str, " ")]) && t.MatchString(str[strings.Index(str, " ")+1:])
	} else {
		return r.MatchString(str)
	}
}

// Sum 求和函数
func Sum(float642 ...float64) float64 {
	var fa float64
	for _, f := range float642 {
		fa += f
	}
	return fa
}

// Max 最大值函数
func Max(float642 ...float64) float64 {
	var ma float64
	for i, v := range float642 {
		if i == 0 {
			ma = v
			continue
		}
		if v > ma {
			ma = v
		}
	}
	return ma
}

// Min 最小值函数
func Min(float642 ...float64) float64 {
	var mi float64
	for i, v := range float642 {
		if i == 0 {
			mi = v
			continue
		}
		if v < mi {
			mi = v
		}
	}
	return mi
}

//Average 平均值函数
func Average(float642 ...float64) float64 {
	sum := Sum(float642...)
	avg := sum / float64(len(float642))
	return avg
}

// Round 4舍5入取整函数
func Round(float642 ...float64) float64 {
	var decimal int
	decimal = 0
	if len(float642) > 1 {
		// 获取小数位
		decimal = int(math.Floor(float642[1]))
	}
	return math.Round(float642[0]*math.Pow10(decimal)) / math.Pow10(decimal)
}

// Int 向下取整
func Int(float642 ...float64) float64 {
	return math.Floor(float642[0])
}

// Mod 取余
func Mod(float642 ...float64) float64 {
	return math.Mod(float642[0], float642[1])
}

// Power 取次方数
func Power(float642 ...float64) float64 {
	return math.Pow(float642[0], float642[1])
}

// Abs 取绝对值
func Abs(float642 ...float64) float64 {
	return math.Abs(float642[0])
}

// Logs 取对数
func Logs(float642 ...float64) float64 {
	if len(float642) > 1 {
		return math.Log(float642[0]) / math.Log(float642[1])
	}
	return math.Log10(float642[0])
}

// IF 条件函数  1 > sum(1,2)
func IF(expression string, s1 interface{}, s2 interface{}) interface{} {
	dataContext := context.NewDataContext()
	ruleName := time.Now().String()
	ruleStr := BuildIFExpression(ruleName, expression)
	for k, v := range ComputeMap {
		dataContext.Add(k, v)
	}
	ruleBuilder := builder.NewRuleBuilder(dataContext)
	err := ruleBuilder.BuildRuleFromString(ruleStr)
	eng := engine.NewGengine()
	//执行规则
	err = eng.Execute(ruleBuilder, true)
	// 获取结果
	resultMap, _ := eng.GetRulesResultMap()
	// 删除规则
	var ruleNames = make([]string, 1)
	ruleNames[0] = ruleName
	err = ruleBuilder.RemoveRules(ruleNames)

	if err != nil {
		panic("IF 函数计算失败")
	}
	for _, v := range resultMap {
		if v == true {
			return s1
		} else {
			return s2
		}
	}
	return s2
}

// IF2 条件函数
func IF2(expression string, s1 interface{}, s2 interface{}) interface{} {
	if expression == "true" {
		return s1
	}
	if expression == "false" {
		return s2
	}
	// 如果是数字 >0 返回真值
	if IsDigit(expression) {
		number, _ := strconv.ParseFloat(expression, 64)
		if number > 0 {
			return s1
		} else {
			return s2
		}
	}
	if ComputeResultBool(expression) {
		return s1
	}
	return s2
}

// CountIf 条件计数
func CountIf(values []interface{}, expression string) (result float64) {
	t, o, val2 := IsExpressionType(expression)
	for _, value := range values {
		b, _ := ExpressionIf(value, val2, o, t)
		if b {
			result = result + 1
		}
	}
	return result
}

// SumIf 条件求和
func SumIf(values []interface{}, expression string) (result float64) {
	t, o, val2 := IsExpressionType(expression)
	if t != Number && val2 != "*" {
		return result
	}
	for _, value := range values {
		b, r := ExpressionIf(value, val2, o, t)
		if b {
			result = result + r
		}
	}
	return result
}

// ExpressionIf 函数条件 被比值,比较值,操作符,类型
func ExpressionIf(val1 interface{}, val2 interface{}, o string, t string) (b bool, result float64) {
	if val1 == nil {
		return false, 0
	}
	if reflect.TypeOf(val1).String() != String {
		val1 = strconv.FormatFloat(val1.(float64), 'f', -1, 64)
	}
	switch t {
	case Number:
		v1, _ := strconv.ParseFloat(val1.(string), 64)
		v2, _ := strconv.ParseFloat(val2.(string), 64)
		b = CompareNumber(o, v1, v2)
		if b {
			result = v1
		}
	case Date:
		b = CompareDate(o, ToDate(val1.(string)), ToDate(val2.(string)))
	case String:
		fallthrough
	default:
		b = CompareString(val1.(string), val2.(string))
		if b {
			v1, _ := strconv.ParseFloat(val1.(string), 64)
			result = v1
		}
	}
	return b, result
}

// CompareString 字符串比较
func CompareString(s string, e string) bool {
	if strings.Contains(e, "*") {
		if strings.HasPrefix(e, "*") {
			return strings.HasSuffix(s, e[1:])
		} else if strings.HasSuffix(e, "*") {
			return strings.HasPrefix(s, e[:len(e)-1])
		} else {
			return strings.EqualFold(s, e)
		}
	} else {
		return strings.EqualFold(s, e)
	}
}

// CompareDate 日期比较
func CompareDate(o string, t1 time.Time, t2 time.Time) (b bool) {
	switch o {
	case "<":
		b = t1.Before(t2)
	case "<=":
		b = t1.Before(t2) || t1.Equal(t2)
	case ">":
		b = t1.After(t2)
	case ">=":
		b = t1.After(t2) || t1.Equal(t2)
	case "=", "==":
		b = t1.Equal(t2)
	case "!=", "<>":
		b = !t1.Equal(t2)
	default:
		b = false
	}
	return b
}

// CompareNumber 数字比较
func CompareNumber(o string, t1 float64, t2 float64) (b bool) {
	switch o {
	case "<":
		b = t1 < t2
	case "<=":
		b = t1 <= t2
	case ">":
		b = t1 > t2
	case ">=":
		b = t1 >= t2
	case "=", "==":
		b = t1 == t2
	case "!=", "<>":
		b = t1 != t2
	default:
		b = false
	}
	return b
}

// IsExpressionType 验证表达式类型 比较类型,操作符号,比较值
func IsExpressionType(expression string) (string, string, string) {
	expression = strings.TrimSpace(expression)
	for _, o := range OperatorSlice {
		if strings.HasPrefix(expression, o) {
			val2 := strings.ReplaceAll(expression, o, "")
			if IsDigit(val2) {
				return Number, o, val2
			}
			if IsDate(val2) {
				return Date, o, val2
			}
		}
	}
	return String, "", expression
}

// CharAtBool 布尔表达式分割
func CharAtBool(boolExpression string) []string {
	var charStr []string
	var ch = []rune(boolExpression)
	var str string
	for i, _ := range ch {
		if ch[i] == '=' || ch[i] == '>' || ch[i] == '<' || ch[i] == '|' || ch[i] == '&' || ch[i] == '!' {
			var b bool
			if str == "=" || str == ">" || str == "<" || str == "|" || str == "&" || str == "!" {
				str += string(ch[i])
				b = true
			}
			if len(str) > 0 {
				charStr = append(charStr, str)
				str = ""
				if b {
					continue
				}
			}
			if ch[i+1] != '=' && ch[i+1] != '|' && ch[i+1] != '&' {
				charStr = append(charStr, string(ch[i]))
			} else {
				str += string(ch[i])
			}
		} else {
			str += string(ch[i])
			if i == len(ch)-1 && len(str) > 0 {
				charStr = append(charStr, str)
			}
		}
	}
	return charStr
}

// IsFormula 验证是否是一个数学计算公式
func IsFormula(str string) bool {
	r, _ := regexp.Compile("(\\(*\\d+[+/*-])+((\\(*(\\d+[+/*-])*\\d+\\)*)[+/*-])*\\d+\\)*")
	return r.MatchString(str)
}

// PercentToDecimal 百分数转小数
func PercentToDecimal(string2 string) string {
	// 百分转转小数
	if strings.HasSuffix(string2, "%") {
		charNumber, _ := strconv.ParseFloat(string2[0:len(string2)-1], 64)
		string2 = strconv.FormatFloat(charNumber/100, 'f', 16, 64)
	}
	return string2
}

// ComputeResultBool 计算解析bool表达式
func ComputeResultBool(compute string) bool {
	var resultBool bool
	char := CharAtBool(compute)
	// 入栈处理
	var stack Stack
	for _, s := range char {
		// 是否是一个计算表达式
		if IsFormula(s) {
			result, _ := ComputeFuncResult1(s, 16, true)
			stack.push(result)
			continue
		}
		stack.push(s)
	}
	var val1 string
	var resultBoolList []bool
	var logicList []string
	isNot := strings.HasPrefix(compute, "!")
	for {
		if stack.Top == 0 {
			break
		}
		s1, _ := stack.pop()
		if s1 == "=" || s1 == "==" {
			val2, _ := stack.pop()
			if IsDigit(val1) && IsDigit(val2) {
				v1, _ := strconv.ParseFloat(val1, 10)
				v2, _ := strconv.ParseFloat(val2, 10)
				resultBool = v1 == v2
			} else {
				resultBool = val1 == val2
			}
		} else if s1 == "!=" || s1 == "<>" {
			val2, _ := stack.pop()
			resultBool = val1 != val2
		} else if s1 == ">" {
			val2, _ := stack.pop()
			resultBool = ToNumber(val2) > ToNumber(val1)
		} else if s1 == ">=" {
			val2, _ := stack.pop()
			resultBool = ToNumber(val2) >= ToNumber(val1)
		} else if s1 == "<" {
			val2, _ := stack.pop()
			resultBool = ToNumber(val2) < ToNumber(val1)
		} else if s1 == "<=" {
			val2, _ := stack.pop()
			resultBool = ToNumber(val2) <= ToNumber(val1)
		} else if s1 == "||" || s1 == "&&" {
			logicList = append(logicList, s1)
			continue
		} else if s1 == "!" {
			continue
		} else {
			val1 = s1
			continue
		}
		resultBoolList = append(resultBoolList, resultBool)
		val1 = ""
	}
	if len(logicList) > 0 {
		for i, b := range logicList {
			if b == "||" {
				if resultBoolList[i] || resultBoolList[i+1] {
					resultBool = true
					if !strings.Contains(fmt.Sprint(logicList), "&&") {
						break
					}
				} else {
					resultBool = false
				}
			} else if b == "&&" {
				if resultBoolList[i] && resultBoolList[i+1] {
					resultBool = true
				} else {
					resultBool = false
					if allEqual(logicList, "&&") {
						break
					}
				}
			}
		}
	}
	if isNot {
		return !resultBool
	}
	return resultBool
}

func allEqual(arr []string, c string) bool {
	for _, s := range arr {
		if s != c {
			return false
		}
	}
	return true
}

// Mid 取字符串中间的值，si从1开始
func Mid(text string, si float64, chartsNum float64) (string, error) {
	// strings.Count 方法返回的长度会多1
	l := strings.Count(text, "") - 1
	i := int(si) - 1
	num := int(chartsNum)
	if l == 0 {
		return text, nil
	}
	if i < 0 || i >= l {
		return "", errors.New("开始下标错误，请输入[1, 字符长度]中的值")
	}
	if num < 0 {
		return "", errors.New("截取字符数量错误，请输入大于等于0的值")
	}
	index := 0
	result := ""
	for _, v := range text {
		if index >= i+num {
			break
		}
		if index >= i {
			result += string(v)
		}
		index++
	}
	return result, nil
}

// Find 判断一个字符串在另一个字符串中出现的位置，返回的下标从1开始
func Find(p string, s string, si ...float64) (int, error) {
	lp := strings.Count(p, "") - 1
	// 需要查找的字符串为空值符串时，返回第一个字符串的下标
	if lp == 0 {
		return 1, nil
	}
	ls := strings.Count(s, "") - 1
	if ls == 0 {
		return -1, errors.New("被查找的字符串不存在")
	}
	i := 0
	if len(si) > 0 {
		i = int(si[0]) - 1
	}
	if i < 0 {
		i = 0
	}
	blp := len(p)
	byteArr := []byte(s)
	bl := len(byteArr)
	index := 0
	for bi, _ := range s {
		// 不能超过字符串总长度
		if bi+blp > bl {
			break
		}
		if index >= i && p == string(byteArr[bi:bi+blp]) {
			return index + 1, nil
		}
		index++
	}
	return -1, errors.New("没有找到需要的字符串")
}

// Concat 字符串拼接
func Concat(strArr ...string) string {
	if len(strArr) == 0 {
		return ""
	}
	result := ""
	for _, v := range strArr {
		result += v
	}
	return result
}

// Len 获取长度
func Len(val string) int {
	return len([]rune(val))
}

// LenB 获取长度
func LenB(val string) int {
	var stripAnsiEscapeRegexp = regexp.MustCompile(`(\x9B|\x1B\[)[0-?]*[ -/]*[@-~]`)
	return runewidth.StringWidth(stripAnsiEscapeRegexp.ReplaceAllString(val, ""))
}

// Substitute 将字符串中的部分字符替换为新的字符
func Substitute(text string, old string, new string, n ...int) string {
	if len(n) == 0 {
		return strings.Replace(text, old, new, -1)
	}
	return strings.Replace(text, old, new, n[0])
}

// Replace 将一个字符串中的部分字符用另一个字符串替换 start 从1开始
func Replace(text string, start int, num int, new string) (string, error) {
	if start < 1 {
		return "", errors.New("开始位置必须大于0")
	}
	if start > Len(text) {
		return Concat(text, new), nil
	}
	start = start - 1
	end := start + num
	if end > Len(text) {
		return new, nil
	}
	textRunes := []rune(text)
	//old := Split(text, start, end)
	startText := textRunes[:start]
	endText := textRunes[end:]
	return string(startText) + new + string(endText), nil
	//return strings.Replace(text, old, new, -1), nil
}

// Left 从一个文本字符串的第一个字符开始返回指定个数的字符
func Left(text string, n ...int) string {
	if len(n) == 0 {
		return Split(text, 0, 0)
	}
	return Split(text, 0, n[0])
}

// Right 从一个文本字符串的最后一个字符开始返回指定个数的字符
func Right(text string, n ...int) string {
	if len(n) == 0 {
		return Split(text, Len(text)-1, Len(text)-1)
	}
	var r = []rune(text)
	var substring bytes.Buffer
	for i, v := range r {
		if i >= Len(text)-n[0] {
			substring.WriteString(string(v))
		}
	}
	return substring.String()
}

// Split 字符串截取
func Split(text string, start int, end int) string {
	var r = []rune(text)
	length := len(r)
	subLen := end - start
	if subLen == 0 {
		return string(r[start])
	}
	for {
		if start < 0 {
			break
		}
		if start == 0 && subLen == length {
			break
		}
		if end > length {
			subLen = length - start
		}
		if end < 0 {
			subLen = length - start + end
		}
		var substring bytes.Buffer
		if end > 0 {
			subLen = end + 1
		}
		for i := start; i < end; i++ {
			substring.WriteString(string(r[i]))
		}
		text = substring.String()
		break
	}
	return text
}
