/*********************************************************************
 * auto: zxw *********************************************************
 * gengine DSL语法:****************************************************
 * const rule = ` ****************************************************
 * rule "rulename" "rule-describtion" salience  10 *******************
 * begin *************************************************************
 * //规则体 ***********************************************************
 * end` **************************************************************
 * 规则体: ************************************************************
 * 支持完整数值之间的加(+)、减(-)、乘(*)、除(/)四则运算,以及字符串之间的加法 *****
 * 完整的逻辑运算(&&、 ||、!) *******************************************
 * 支持比较运算符: 等于(==)、不等于(!=)、大于(>)、小于(<)、大于等于(>=)、*******
 * 小于等于(<=) *******************************************************
 * 支持+=, -=, *=, /=   支持小括号 *************************************
 * 优先级:括号, 非, 乘除, 加减, 逻辑运算(&&,||) 依次降低 ********************
 * 支持的数据类型：string,bool,int, int8, int16, int32, int64,uint, *****
 * uint8, uint16,uint32, uint64,float32, float64 *********************
 *
 * interpretation: 解析计算类型yaml:int,string,date,自定义,赋值 **********
 * ExplainCalculatesInt **********************************************
 * ExplainCalculatesString *******************************************
 * ExplainCalculatesDate *********************************************
 *
 * Time: 2022-12-21 **************************************************
 * Version:1.0 *******************************************************
 */

package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

/*
 * 全局变量
 */
var (
	str1            string
	str2            string
	calcOperatorMap = map[string]string{
		"ADD": "+",
		"SUB": "-",
		"MUL": "*",
		"DIV": "/",
		//"SUR": "%",
		"MOD": "/",
	}
	calcOperatorMapCustom = map[string]string{
		"SET":    StrReplace(str1, str2),
		"CUSTOM": "+",
		//StrSupple(str1, str2),
	}
)

type Calculate struct {
	Code     string  `yaml:"code"`
	Name     string  `yaml:"name"`
	Desc     string  `yaml:"desc"`
	Salience int     `yaml:"salience"`
	Type     string  `yaml:"type"`
	Value    string  `yaml:"value"`
	Logics   []Logic `yaml:"logics"`
	Result   string  `yaml:"result"`
}
type Logic struct {
	Value    string `yaml:"value"`
	Operator string `yaml:"operator"`
	Result   string `yaml:"result"`
	Unit     string `yam:"uint"`
}

/*
 * func:实现int计算的自定义规则
 * param:计算结构体
 * return:规则字符串
 */

func ExplainLogicsInt(logics []Logic) string {
	var logicValue string
	for i, logic := range logics {
		/* 获取要计算的字符串*/
		if i == 0 {
			/* 如果只有计算一个，只调用一次*/
			logicValue += calcOperatorMap[logic.Operator] + logic.Value
			continue
		}
		/*多个计算合并*/
		logicValue += calcOperatorMap[logic.Operator] + logic.Value
	}
	return logicValue
}
func ExplainLogicsStr(logics []Logic) string {
	var calcLogicStr string
	for _, logic := range logics {
		calcLogicStr += logic.Value
	}
	return calcLogicStr
}

/*其他类型获取结果*/

func ParseCalcResult(Result string, calc Calculate) string {
	var (
		/*传入的参数value与获取计算的字符串求和拼接*/
		resultStr = calc.Value + Result
	)
	//calc.Result = calc.Value+Result
	return resultStr
}

/*日期获取结果*/

func ParseCalcResultDate(Result string) string {
	var (
		resultStr = Result
	)
	return resultStr
}
func AddString(str1 string, str2 string) string {
	/*字符串拼接计算*/
	return str1 + str2
}
func JoinString(str1 []string) string {
	var ResultStr = strings.Join(str1, ",")
	return ResultStr
}
func SplitString(str1 string, str2 string) []string {
	return strings.Split(str1, str2)
}
func StrReplace(str1 string, str2 string) string {
	str1 = str2
	return str1
}
func StrSupple(str1 string, str2 string) string {
	str1 = str1 + str2
	return str1
}

// ExecuteCal 执行yaml内的perform标签*/
func ExecuteCal(rule Rule, yamlCode string) string {
	var executeStr string
	for i, exeStr := range rule.Perform {
		if i > 0 {
			executeStr += `
		` /*换行*/
		}
		executeStr += `executeEnginePool("` + ExplainStr(yamlCode, exeStr) + `",arguments,executeId)`
		/*获取执行规则池语句*/
	}
	if rule.ExecuteType == "parallel" { /*并发执行*/
		executeStr = `conc {
			` + executeStr
		executeStr = executeStr + `
		}`
	}

	return executeStr
}

/*解析字符串*/

func ExplainStr(yamlCode string, calcCode string) string {
	return yamlCode + ":" + calcCode
}

/*字符串转日期*/

func StrToDate(strTimeI interface{}) time.Time {
	var str string = fmt.Sprintf("%v", strTimeI)
	t1, _ := time.Parse("2006-01-02 15:04:05", str)
	return t1
}

/*获取日期字符串*/

func DateToStr(time time.Time) string {
	str := time.Format("2006-01-02 15:04:05")
	return str
}

/*解析INT类型数据*/

func ExplainCalculatesInt(calc Calculate) string {
	var logicsStr string
	for i, _ := range calc.Logics {
		if i > 0 {
			/*获取计算的字符串*/
			logicsStr += ` `
		} else {
			logicsStr += ExplainLogicsInt(calc.Logics)
		}
	}
	var calculateIntStr = `
    rule "` + calc.Name + `" "` + calc.Desc + `" salience ` + strconv.Itoa(calc.Salience) + `
	begin 
		print("执行-自定义Int计算规则[` + calc.Name + `]")
		arguments["` + calc.Result + `"] = ` + ParseCalcResult(logicsStr, calc) + `
		` + ParseProcessRecord(calc.Code, calc.Type) + `
	end
`
	return calculateIntStr
}

/*解析string类型数据*/

func ExplainCalculatesString(calc Calculate) string {
	var calcStringStr string
	for i, logics := range calc.Logics {
		if i == 0 {
			switch logics.Operator {
			case "ADD": /*字符串添加*/
				calcStringStr += ParseCalcResult(AddString(calcOperatorMap[logics.Operator],
					ExplainLogicsStr(calc.Logics)), calc)
				break
			case "SPLIT":
				//var s = SplitString(calc.Value, logics.Value)
				calcStringStr +=
					`strings.Split("` + calc.Value + `","` + logics.Value + `")`
			case "JOIN":
				var Str = []string{calc.Value, logics.Value}
				calcStringStr = JoinString(Str)
			case "SET": /*直接赋值运算 */
				calc.Value = logics.Value
				//calcStringStr += calcOperatorMapCustom[logics.Operator]
				calcStringStr += calc.Value
			case "CUSTOM":
				//calc.Value += logics.Value
				//calcStringStr += calcOperatorMapCustom[logics.Operator]
				calcStringStr += calc.Value + calcOperatorMapCustom[logics.Operator] + "'+'" + logics.Value
			default:

			}
		}
	}
	var calculateStringStr = `
	rule "` + calc.Name + `" "` + calc.Desc + `" salience ` + strconv.Itoa(calc.Salience) + `
	begin 
		print("执行-自定义字符串计算规则[` + calc.Name + `]")
		arguments["` + calc.Result + `"] =` + calcStringStr + `
		` + ParseProcessRecord(calc.Code, calc.Type) + `
	end
`

	return calculateStringStr
}

/*解析date类型数据*/

func ExplainCalculatesDate(calc Calculate) string {
	datetime := time.Now()
	var (
		year  int
		month int
		day   int
	)
	for i, logic := range calc.Logics {
		if i == 0 { //logic.Unit == "YEAR"
			if y, err := strconv.Atoi(logic.Value); err == nil {
				year = y
			}
		} else if i == 1 { //logic.Unit == "MONTH"
			if m, err1 := strconv.Atoi(logic.Value); err1 == nil {
				month = m
			}
		} else if i == 2 { //logic.Unit == "DAY"
			if d, err2 := strconv.Atoi(logic.Value); err2 == nil {
				day = d
			}
		} else {

		}
	}
	result := datetime.AddDate(year, month, day)
	calc.Value = DateToStr(result)
	/*规则执行，规则体内只包含基本运算，不支持函数、方法调用*/

	var calculateDateStr = `		
	rule "` + calc.Name + `" "` + calc.Desc + `" salience ` + strconv.Itoa(calc.Salience) + `
	begin 
		print("执行-自定义日期计算规则[` + calc.Name + `]")
		arguments["` + calc.Result + `"] ="` + ParseCalcResultDate(calc.Value) + `"
		` + ParseProcessRecord(calc.Code, calc.Type) + `
	end
`
	return calculateDateStr
}
