package util

import (
	"encoding/json"
	"errors"
	"gopkg.in/yaml.v2"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"zenith.engine.com/engine/pkg/sql"
)

var (
	operatorMap = map[string]string{
		"EQ": "==",
		"LE": "<=",
		"LT": "<",
		"NE": "!=",
		"GE": ">=",
		"GT": ">",
	}
	operatorDateMap = map[string]string{
		"EQ": "timeEqual",
		"LQ": "timeBeforeEqual",
		"LT": "timeBefore",
		"NE": "!timeEqual",
		"GE": "timeAfterEqual",
		"GT": "timeAfter",
	}
	logicMap = map[string]string{
		"AND": "&&",
		"OR":  "||",
		"":    "&&",
		"and": "&&",
		"or":  "||",
	}
	explainMap = map[string]func(code string, rule Rule) string{
		"start":     ExplainStart,
		"condition": ExplainCondition2,
		"action":    ExplainAction,
		"compute":   ExplainCompute,
		"end":       ExplainEnd,
	}
	ArgumentsDefiner = "$"
	NodeMap          = sync.Map{}
)

const (
	FS  = "string"
	FIN = "int"
	FMD = "method"
	FDT = "date"
	FTP = "typeof"
)

type Rules struct {
	Rules      []Rule      `yaml:"rules,omitempty"`
	Calculates []Calculate `yaml:"calculates,omitempty"`
}
type Rule struct {
	Code                string          `yaml:"code,omitempty"` // 编码
	Id                  string          `yaml:"id,omitempty"`
	Type                string          `yaml:"type,omitempty"`     // 类型
	Name                string          `yaml:"name,omitempty"`     // 名称
	Desc                string          `yaml:"desc,omitempty"`     // 描述
	Salience            int             `yaml:"salience,omitempty"` // 优先级
	Label               string          `yaml:"label,omitempty"`    // 标签
	Decision            string          `yaml:"decision,omitempty"`
	Conditions          []Condition     `yaml:"conditions,omitempty"` // 条件项对象
	Content             string          `yaml:"content,omitempty"`
	Execute             []string        `yaml:"execute,omitempty"`         // 执行默认规则
	ExecuteType         string          `yaml:"executeType,omitempty"`     // 执行类型
	ExecuteCallRule     ExecuteCallRule `yaml:"executeCallRule,omitempty"` // 执行其他规则
	ExecuteCustom       string          `yaml:"executeCustom,omitempty"`   // 执行自定义脚本
	ExecuteSQL          ExecuteSQL      `yaml:"executeSQL,omitempty"`      // 执行sql属性
	ExecuteHttp         ExecuteHttp     `yaml:"executeHttp,omitempty"`     // 执行http属性
	ExecuteJson         ExecuteJson     `yaml:"executeJson,omitempty"`     // 执行json属性
	Next                []string        `yaml:"next,omitempty"`            // 下一步节点code
	Perform             []string        `yaml:"perform,omitempty"`
	Together            bool            `yaml:"together,omitempty"`            // 是否聚合
	TogetherCount       int             `yaml:"togetherCount"`                 // 聚合数量
	TogetherWeight      int             `yaml:"togetherWeight"`                // 聚合权重
	Expression          string          `yaml:"expression,omitempty"`          // 计算表达式
	ExpressionArguments string          `yaml:"expressionArguments,omitempty"` // 计算后返回值参数名称
	ExpressionPrecision int             `yaml:"expressionPrecision,omitempty"` // 计算精度
	ExpressionResult    bool            `yaml:"expressionResult,omitempty"`    // 计算后是否作为结果
	//Value       string      `yaml:"value"`
	//Logics      []Logic     `yaml:"logics"`
	//Result      string      `yaml:"result"`
}
type Condition struct {
	Condition      string    `yaml:"condition,omitempty"`
	Features       []Feature `yaml:"feature,omitempty"`
	Next           []string  `yaml:"next,omitempty"`
	Execute        []string  `yaml:"execute,omitempty"`
	DefaultFeature bool      `yaml:"defaultFeature,omitempty"`
}
type Feature struct {
	Field    string `yaml:"field,omitempty"`
	Operator string `yaml:"operator,omitempty"`
	Value    string `yaml:"value,omitempty"`
	Type     string `yaml:"type,omitempty"`
	Logic    string `yaml:"logic,omitempty"`
}
type ExecuteCallRule struct {
	Code      string                 `yaml:"code,omitempty"`
	Version   string                 `yaml:"version,omitempty"`
	Result    []ExecuteMappingResult `yaml:"result"`
	Arguments []ExecuteCallArguments `yaml:"arguments"`
	Loop      bool                   `yaml:"loop"`
	LoopKey   string                 `yaml:"loopKey"`
	LoopJoin  bool                   `yaml:"loopJoin"`
}

type ExecuteCallArguments struct {
	Selected bool   `yaml:"selected,omitempty"`
	Key      string `yaml:"key,omitempty"`
	Value    string `yaml:"value,omitempty"`
	Def      string `yaml:"def,omitempty"`
}

type ExecuteSQL struct {
	Sql      string                 `yaml:"sql,omitempty"`
	DataBase string                 `yaml:"dataBase,omitempty"`
	Result   []ExecuteMappingResult `yaml:"result,omitempty"`
}

type ExecuteHttp struct {
	System string                 `yaml:"system,omitempty"`
	Method string                 `yaml:"method,omitempty"`
	Url    string                 `yaml:"url,omitempty"`
	Query  []ExecuteHttpParams    `yaml:"query,omitempty"`
	Header []ExecuteHttpParams    `yaml:"header,omitempty"`
	Body   string                 `yaml:"body,omitempty""`
	Result []ExecuteMappingResult `yaml:"result,omitempty"`
}

type ExecuteHttpParams struct {
	Selected bool        `yaml:"selected,omitempty"`
	Key      string      `yaml:"key,omitempty"`
	Value    interface{} `yaml:"value,omitempty"`
	Desc     string      `yaml:"desc,omitempty"`
}

type ExecuteJson struct {
	JsonArguments string                 `json:"jsonArguments,omitempty"`
	JsonModel     string                 `json:"jsonModel,omitempty"`
	Result        []ExecuteMappingResult `json:"result,omitempty"`
	SetResult     []ExecuteSetResult     `json:"setResult,omitempty"`
}

type ExecuteMappingResult struct {
	Path     string      `yaml:"path,omitempty"`
	Title    string      `yaml:"title,omitempty"`
	Def      interface{} `yaml:"def,omitempty"`
	Selected bool        `yaml:"selected,omitempty"`
}

type ExecuteSetResult struct {
	Key      string      `yaml:"key,omitempty"`
	Value    interface{} `yaml:"value,omitempty"`
	Selected bool        `yaml:"selected,omitempty"`
}

type Node struct {
	Code           string  `yaml:"code"`
	Parent         []*Node `yaml:"parent"`
	Children       []*Node `yaml:"children"`
	Status         bool    `yaml:"status"`
	Type           string  `yaml:"type"`
	Together       bool    `yaml:"together"`
	TogetherWeight int     `yaml:"togetherWeight"`
}

// BuildNode 创建全部节点图
func BuildNode(code string, rules *Rules) *Node {
	// 处理rule条件节点
	for i, rule := range rules.Rules {
		if rule.Type == "condition" {
			for _, cd := range rule.Conditions {
				rules.Rules[i].Next = append(rules.Rules[i].Next, cd.Next...)
			}
		}
	}
	// 创建根节点
	var node = new(Node)
	for _, rule := range rules.Rules {
		var cNode = new(Node)
		cNode.Code = ParseRuleCode(code, rule.Code)
		cNode.Type = rule.Type
		cNode.Together = rule.Together
		cNode.TogetherWeight = rule.TogetherWeight
		node.Children = append(node.Children, cNode)
	}
	// 处理子节点
	for _, child := range node.Children {
		for _, rule := range rules.Rules {
			if child.Code == ParseRuleCode(code, rule.Code) {
				for _, s := range rule.Next {
					{
						for _, c := range node.Children {
							if ParseRuleCode(code, s) == c.Code {
								// 判断是否已经存在这个子节点了
								cm := ConvertStrSlice2Map(child.Children)
								if InMap(cm, c.Code) {
									continue
								}
								child.Children = append(child.Children, c)
							}
						}
					}
				}
			}
		}
	}
	return node
}

func DesignYamlParseRules(design string) (*Rules, error) {
	var rs = new(Rules)
	if err := yaml.Unmarshal([]byte(design), &rs); err != nil {
		return nil, errors.New("规则Yaml解析失败")
	}
	return rs, nil
}

func ExplainYaml(code string, design string) (map[string]string, error) {
	rs, err := DesignYamlParseRules(design)
	if err != nil {
		return nil, err
	}
	// 关系存入内存
	NodeMap.Store(code, BuildNode(code, rs))
	/*map类型规则存储*/
	var poolRules = make(map[string]string)
	/*解析Rules类型规则，将start,condition,action，end类型规则注入poolRule*/
	for _, rule := range rs.Rules {
		poolRules[ParseRuleCode(code, rule.Code)] = explainMap[strings.ToLower(rule.Type)](code, rule)
	}
	/*解析calculates类型规则，将int,string,date类型规则注入poolRule*/
	for _, calc := range rs.Calculates {
		switch strings.ToLower(calc.Type) {
		case "int":
			poolRules[ParseRuleCode(code, calc.Code)] = ExplainCalculatesInt(calc)
		case "string":
			poolRules[ParseRuleCode(code, calc.Code)] = ExplainCalculatesString(calc)
		case "date":
			poolRules[ParseRuleCode(code, calc.Code)] = ExplainCalculatesDate(calc)
		default:
			return nil, errors.New("规则类型无法识别")
		}
	}
	return poolRules, nil
}

// ExplainStart 解析开始节点/*
func ExplainStart(yamlCode string, rule Rule) string {
	var ruleStr = `
	rule  "` + rule.Name + `" "` + rule.Desc + `" salience ` + strconv.Itoa(rule.Salience) + `
	begin
        print("执行-开始规则[` + rule.Name + `]")
        ` + ExplainExecute(rule.Type, rule.Code, yamlCode, rule.Execute, rule.ExecuteType) + `
		` + ParseProcessRecord(rule.Code, rule.Type) + `
 		` + ExplainTogetherCondition(yamlCode, rule.Code, rule.Next, false) + `
		` + ExplainNext(rule.Code, yamlCode, rule.Next, rule.Type) + `
        return es
	end`
	return ruleStr
}

// ExplainCondition 解析条件节点  else if/*
func ExplainCondition(yamlCode string, rule Rule) string {
	var conditionStr string
	var elseStr string
	for i, condition := range rule.Conditions {
		if i == 0 { /*获取条件，解析记录，执行下一步*/
			conditionStr += `
       ` + ParseProcessRecord(rule.Code, rule.Type) + `
		if ` + ExplainFeature(condition.Features) + ` {
          ` + ExplainTogetherCondition(yamlCode, rule.Code, condition.Next, true) + `
			` + ExplainNext(rule.Code, yamlCode, condition.Next, rule.Type) + `
		}`
			continue
		}
		if condition.DefaultFeature {
			elseStr = `
       else {
		` + ExplainTogetherCondition(yamlCode, rule.Code, condition.Next, true) + `
		` + ExplainNext(rule.Code, yamlCode, condition.Next, rule.Type) + `
		}`
		} else {
			conditionStr += `
       else if ` + ExplainFeature(condition.Features) + ` {
			` + ExplainTogetherCondition(yamlCode, rule.Code, condition.Next, true) + `
			` + ExplainNext(rule.Code, yamlCode, condition.Next, rule.Type) + `
		}`
		}
	}
	if len(elseStr) > 0 {
		conditionStr += elseStr
	}

	var content = `print("执行-条件规则[` + rule.Name + `]")
		` + conditionStr + ``
	if rule.Together {
		content = ExplainTogether(ParseRuleCode(yamlCode, rule.Code), rule.TogetherWeight, content)
	}
	var ruleStr = `
	rule  "` + rule.Name + `" "` + rule.Desc + `" salience ` + strconv.Itoa(rule.Salience) + `
	begin
		` + content + `
	end`
	return ruleStr
}

// ExplainCondition2 解析条件节点  多if/*
func ExplainCondition2(yamlCode string, rule Rule) string {
	var conditionStr string
	var elseStr string
	conditionStr += ParseProcessRecord(rule.Code, rule.Type)
	conditionStr += `
		b = true
        y = true`
	// 预执行解析
	for _, condition := range rule.Conditions {
		if condition.DefaultFeature {
			elseStr = `
        if !y {
			` + ExplainTogetherCondition(yamlCode, rule.Code, condition.Next, true) + `
		}`
		} else {
			conditionStr += `
        if ` + ExplainFeature(condition.Features) + ` {
			y = false
		} else {
			` + ExplainTogetherCondition(yamlCode, rule.Code, condition.Next, true) + `
		}`
		}
	}
	if len(elseStr) > 0 {
		conditionStr += elseStr
	}

	// 实际执行解析
	for _, condition := range rule.Conditions {
		if condition.DefaultFeature {
			elseStr = `
        if b {
		` + ExplainTogetherCondition(yamlCode, rule.Code, condition.Next, false) + `
		` + ExplainNext(rule.Code, yamlCode, condition.Next, rule.Type) + `
		}`
		} else {
			conditionStr += `
        if ` + ExplainFeature(condition.Features) + ` {
			b = false
			` + ExplainTogetherCondition(yamlCode, rule.Code, condition.Next, false) + `
			` + ExplainNext(rule.Code, yamlCode, condition.Next, rule.Type) + `
		}`
		}
	}

	if len(elseStr) > 0 {
		conditionStr += elseStr
	}

	var content = `print("执行-条件规则[` + rule.Name + `]")
		` + conditionStr + ``
	if rule.Together {
		content = ExplainTogether(ParseRuleCode(yamlCode, rule.Code), rule.TogetherWeight, content)
	}
	var ruleStr = `
	rule  "` + rule.Name + `" "` + rule.Desc + `" salience ` + strconv.Itoa(rule.Salience) + `
	begin
		` + content + `
	return es
	end`
	return ruleStr
}

// ExplainAction 解析执行节点/*
func ExplainAction(yamlCode string, rule Rule) string {
	var content = `
		print("执行-执行规则[` + rule.Name + `]")
        ` + rule.ExecuteCustom + `
 		` + ExplainTogetherCondition(yamlCode, rule.Code, rule.Next, false) + `
		` + ExplainExecuteSql(rule.ExecuteSQL, rule.ExpressionResult) + `
		` + ExplainExecuteHttp(rule.ExecuteHttp, rule.ExpressionResult) + `
        ` + ExplainExecuteJson(rule.ExecuteJson, rule.ExpressionResult) + `
		` + ExplainExecute(rule.Type, rule.Code, yamlCode, rule.Execute, rule.ExecuteType) /*执行其他规则*/ + `
		` + ParseProcessRecord(rule.Code, rule.Type) + `
        ` + ExplainExecuteCallRule(rule.ExecuteCallRule, rule.ExpressionResult) + `
		` + ExplainNext(rule.Code, yamlCode, rule.Next, rule.Type) + ``

	if rule.Together {
		content = ExplainTogether(ParseRuleCode(yamlCode, rule.Code), rule.TogetherWeight, content)
	}

	var ruleStr = `
	rule  "` + rule.Name + `" "` + rule.Desc + `" salience ` + strconv.Itoa(rule.Salience) + `
	begin
        ` + content + `
    return es
	end`
	return ruleStr
}

// ExplainCompute 解析计算节点/*
func ExplainCompute(yamlCode string, rule Rule) string {
	var content = `
		print("执行-执行计算规则[` + rule.Name + `]")
 		` + ExplainTogetherCondition(yamlCode, rule.Code, rule.Next, false) + `
		` + ExplainExecute(rule.Type, rule.Code, yamlCode, rule.Execute, rule.ExecuteType) /*执行其他规则*/ + `
		` + ExecuteCal(rule, yamlCode) /*执行计算类型规则*/ + `
		` + ExplainComputeExpression(rule.Expression, rule.ExpressionArguments, rule.ExpressionResult, rule.ExpressionPrecision) + `
        ` + ExplainExecuteCallRule(rule.ExecuteCallRule, rule.ExpressionResult) + `
        ` + ParseProcessRecord(rule.Code, rule.Type) + `
		` + ExplainNext(rule.Code, yamlCode, rule.Next, rule.Type) + `
		`
	if rule.Together {
		content = ExplainTogether(ParseRuleCode(yamlCode, rule.Code), rule.TogetherWeight, content)
	}

	var ruleStr = `
	rule  "` + rule.Name + `" "` + rule.Desc + `" salience ` + strconv.Itoa(rule.Salience) + `
	begin
        ` + content + `
	return es
	end`
	return ruleStr
}

// ExplainEnd 解析结束节点/*
func ExplainEnd(yamlCode string, rule Rule) string {
	var content = ` print("执行-结束规则[` + rule.Name + `]")
        ` + ExplainExecute(rule.Type, rule.Code, yamlCode, rule.Execute, rule.ExecuteType) + `
        ` + ParseProcessRecord(rule.Code, rule.Type) + ``
	if rule.Together && rule.TogetherCount > 1 {
		content = ExplainTogether(ParseRuleCode(yamlCode, rule.Code), rule.TogetherWeight, content)
	}
	var ruleStr = `
	rule  "` + rule.Name + `" "` + rule.Desc + `" salience ` + strconv.Itoa(rule.Salience) + `
	begin
        ` + content + `
    return es
	end`
	return ruleStr
}

// ExplainFeature 解析条件分支/*
func ExplainFeature(features []Feature) string {
	var featureStr string
	for i, feature := range features {
		if i > 0 {
			featureStr += ` ` + logicMap[feature.Logic] + ` `
		}
		if strings.EqualFold(feature.Type, FS) {
			filedStr := `loadStringArguments("` + feature.Field + `",arguments)`
			featureStr += filedStr + operatorMap[feature.Operator] + `"` + feature.Value + `"`
		} else if strings.EqualFold(feature.Type, FIN) {
			filedStr := `loadNumberArguments("` + feature.Field + `",arguments)`
			featureStr += filedStr + operatorMap[feature.Operator] + feature.Value
		} else if strings.EqualFold(feature.Type, FDT) {
			filedStr := `loadDateArguments("` + feature.Field + `",arguments)`
			filedVal := `toDate("` + feature.Value + `")`
			featureStr += operatorDateMap[feature.Operator] + `(` + filedStr + `,` + filedVal + `)`
		} else if strings.EqualFold(feature.Type, FMD) {
			featureStr += feature.Value
		} else if strings.EqualFold(feature.Type, FTP) {
			featureStr = `loadArgumentsAndTypeOf("` + feature.Field + `","` + feature.Value + `","` + feature.Operator + `",arguments)`
		} else {
			featureStr += feature.Field + operatorMap[feature.Operator] + feature.Value
		}
	}
	return featureStr

}

// ExplainExecute 解析执行规则/*
func ExplainExecute(ruleType string, code string, yamlCode string, execute []string, executeType string) string {
	var executeStr string
	for i, s := range execute {
		if i > 0 {
			executeStr += `
		`
		}
		executeStr += `executeEnginePool(@name,"` + ruleType + `","","` + s + `",arguments,executeId)`
	}
	if executeType == "parallel" {
		executeStr = `conc {
        ` + executeStr
		executeStr = executeStr + `
		}`
	}
	return executeStr
}

// ExplainExecuteCallRule 解析调用其他规则 子规则/*
func ExplainExecuteCallRule(executeCallRule ExecuteCallRule, result bool) string {
	var erStr string
	if len(executeCallRule.Code) > 0 {
		var argumentsFormatStr string
		if len(executeCallRule.Arguments) > 0 {
			callArgument, _ := json.Marshal(executeCallRule.Arguments)
			argumentsFormatStr = FormatJson(string(callArgument))
		}
		var resultFormatStr string
		if len(executeCallRule.Result) > 0 {
			callResult, _ := json.Marshal(executeCallRule.Result)
			resultFormatStr = FormatJson(string(callResult))
		}
		if len(executeCallRule.LoopKey) > 0 {
			erStr = `executeEngineCallRuleFor(executeEngineCallRule,"` + executeCallRule.Code + `","` + executeCallRule.Version + `","` + argumentsFormatStr + `","` + resultFormatStr + `",` + strconv.FormatBool(result) + `,arguments,executeId,"` + executeCallRule.LoopKey + `",` + strconv.FormatBool(executeCallRule.LoopJoin) + `)`
		} else {
			erStr = `executeEngineCallRule("` + executeCallRule.Code + `","` + executeCallRule.Version + `","` + argumentsFormatStr + `","` + resultFormatStr + `",` + strconv.FormatBool(result) + `,arguments,executeId)`
		}
	}
	return erStr
}

// ExplainExecuteSql 解析执行sql属性 /*
func ExplainExecuteSql(executeSQL ExecuteSQL, result bool) string {
	var esStr string
	if reflect.DeepEqual(executeSQL, ExecuteSQL{}) {
		return ""
	}
	// 解析出sql参数
	var argumentsStr string
	executeSQL.Sql, argumentsStr = ExplainSqlArguments(executeSQL.Sql)
	var resultFormatStr string
	if len(executeSQL.Result) > 0 {
		sqlResult, _ := json.Marshal(executeSQL.Result)
		resultFormatStr = FormatJson(string(sqlResult))
	}
	esStr = `connectDataExecuteSql(executeId,arguments,"` + executeSQL.DataBase + `","` + executeSQL.Sql + `","` + resultFormatStr + `",` + strconv.FormatBool(result) + `` + argumentsStr + `)`
	return esStr
}

// ExplainSqlArguments 解析sql参数
func ExplainSqlArguments(sqlStr string) (string, string) {
	sqlList := sql.SplitSql(sqlStr)
	argumentsStr := ``
	for _, s := range sqlList {
		if strings.HasPrefix(s, ArgumentsDefiner) {
			argumentsStr += `,` + `argumentsGet(arguments,"` + s[1:] + `")`
			sqlStr = strings.Replace(sqlStr, s, "?", 1)
		}
	}
	return sqlStr, argumentsStr
}

// ExplainSqlArgumentsPreview 解析sql参数
func ExplainSqlArgumentsPreview(sqlStr string) (string, string) {
	sqlList := sql.SplitSql(sqlStr)
	argumentsStr := ``
	for _, s := range sqlList {
		if strings.HasPrefix(s, ArgumentsDefiner) {
			argumentsStr += `,` + `arguments["` + s[1:] + `"]`
			sqlStr = strings.Replace(sqlStr, s, "?", 1)
		}
	}
	return sqlStr, argumentsStr
}

// ExplainExecuteHttp 解析执行http属性
func ExplainExecuteHttp(explainHttp ExecuteHttp, result bool) string {
	var ehStr string
	if reflect.DeepEqual(explainHttp, ExecuteHttp{}) {
		return ""
	}
	var queryFormatStr string
	if len(explainHttp.Query) > 0 {
		queryStr, _ := json.Marshal(explainHttp.Query)
		queryFormatStr = FormatJson(string(queryStr))
	}
	var headerFormatStr string
	if len(explainHttp.Header) > 0 {
		headerStr, _ := json.Marshal(explainHttp.Header)
		headerFormatStr = FormatJson(string(headerStr))
	}
	var resultFormatStr string
	if len(explainHttp.Result) > 0 {
		resultStr, _ := json.Marshal(explainHttp.Result)
		resultFormatStr = FormatJson(string(resultStr))
	}
	ehStr = `XSend(executeId,arguments,"` + explainHttp.System + `","` + explainHttp.Method + `","` + explainHttp.Url + `","` + queryFormatStr + `","` + FormatJson(explainHttp.Body) + `",` + strconv.FormatBool(result) + `,"` + headerFormatStr + `","` + resultFormatStr + `")`
	return ehStr
}

// ExplainExecuteJson 解析执行json属性
func ExplainExecuteJson(explainJson ExecuteJson, result bool) string {
	var ejStr string
	if reflect.DeepEqual(explainJson, ExecuteJson{}) {
		return ""
	}
	if len(explainJson.JsonArguments) == 0 {
		explainJson.JsonArguments = FormatJson(explainJson.JsonModel)
	}
	var resultFormatStr string
	if len(explainJson.Result) > 0 {
		resultStr, _ := json.Marshal(explainJson.Result)
		resultFormatStr = FormatJson(string(resultStr))
	}

	var setResultFormatStr string
	if len(explainJson.SetResult) > 0 {
		resultStr, _ := json.Marshal(explainJson.SetResult)
		setResultFormatStr = FormatJson(string(resultStr))
	}
	ejStr = `handleJson("` + explainJson.JsonArguments + `","` + resultFormatStr + `","` + setResultFormatStr + `",arguments,executeId,` + strconv.FormatBool(result) + `,status)`
	return ejStr
}

// FormatJson 格式化json
func FormatJson(jsonStr string) string {
	var chars []rune
	for _, letter := range jsonStr {
		ok, letters := SpecialLetters(letter)
		if ok {
			chars = append(chars, letters...)
		} else {
			chars = append(chars, letter)
		}
	}
	return string(chars)
}

// SpecialLetters 添加转译符\\
func SpecialLetters(letter rune) (bool, []rune) {
	if unicode.IsPunct(letter) || unicode.IsSymbol(letter) || unicode.Is(unicode.Han, letter) {
		var chars []rune
		chars = append(chars, '\\', letter)
		return true, chars
	}
	return false, nil
}

// ExplainNext 解析下一步节点/*
func ExplainNext(ruleCode string, yamlCode string, next []string, t string) string {
	var nextStr string
	for i, s := range next {
		if i > 0 {
			nextStr += `
		`
		}
		nextStr += `es["` + ParseRuleCode(yamlCode, s) + `"] = buildExecuteEngineNext(@name,"` + t + `","` + ParseRuleCode(yamlCode, ruleCode) + `","` + ParseRuleCode(yamlCode, s) + `",arguments,executeId,debugParam)`
	}
	return nextStr
}

// ExplainTogether 解析聚合属性 /*
func ExplainTogether(code string, togetherWeight int, content string) string {
	content = `if togetherCheck2("` + code + `",` + strconv.Itoa(togetherWeight) + `,executeId) {
	` + content + "}"
	return content
}

// ExplainTogetherCondition 修改下一节点权重 /*
func ExplainTogetherCondition(yamlCode string, code string, next []string, isCondition bool) string {
	var nextStr string
	for i, s := range next {
		if i == 0 {
			nextStr += s
		} else {
			nextStr += "|" + s
		}
	}
	var str string
	str = `togetherCondition("` + yamlCode + `","` + code + `",executeId,"` + nextStr + `",` + strconv.FormatBool(isCondition) + `)`
	return str
}

// ExplainComputeExpression 解析计算表达式
func ExplainComputeExpression(expression string, expressionArguments string, expressionResult bool, expressionPrecision int) string {
	if len(expression) == 0 {
		return ""
	}
	var computeExpressionStr string
	// 转译
	expression = strconv.Quote(expression)[1:]
	expression = expression[0 : len(expression)-1]
	computeExpressionStr = `computeExpression("` + expression + `",inputArguments,arguments,` + strconv.Itoa(expressionPrecision) + `,executeId,status)`
	// 是否将计算结果赋值给参数
	if len(expressionArguments) > 0 {
		computeExpressionStr = `argumentsSet(arguments,"` + expressionArguments + `",` + computeExpressionStr + `)`
		// 是否作为返回值
		if expressionResult {
			computeExpressionStr = computeExpressionStr + `
		appendRuleResult(executeId,"` + expressionArguments + `",argumentsGet(arguments,"` + expressionArguments + `"))`
		}
	}
	return computeExpressionStr
}

// ParsePanic 错误验证 中断规则/*
func ParsePanic() string {
	var panicStr = `
		panic(error)`
	return panicStr
}

// ParseProcessRecord 转换流转记录/*
func ParseProcessRecord(code string, t string) string {
	var processRecordStr string
	processRecordStr = `processRecord(@name,"` + code + `","` + t + `",startTime,inputArguments,arguments,status,executeId)`
	return processRecordStr
}

func ParseRuleCode(yamlCode string, ruleCode string) string {
	return yamlCode + ":" + ruleCode
}

func ParseYamlCode(code string, version string) string {
	return code + "-" + version
}

// BuildIFExpression 构建if规则脚本
func BuildIFExpression(name string, expression string) string {
	var ruleStr string
	ruleStr = `
	rule "` + name + `" ""  salience 0
	begin
		if ` + expression + ` {
			return true
		}else{
			return false
		}
	end `
	return ruleStr
}
