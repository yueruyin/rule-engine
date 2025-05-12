package util

import (
	"encoding/json"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v2"
	"math"
	"reflect"
	"strconv"
	"strings"
)

const (
	DEV  = "dev"
	PROD = "prod"
)

var (
	beginCode         = "begin_"
	endCode           = "end_"
	actionCode        = "action_"
	conditionCode     = "condition_"
	conditionItemCode = "conditionItem_"
	computeCode       = "compute_"

	line          = "RDLine"
	begin         = "RuleBegin"
	end           = "RuleEnd"
	action        = "RuleAction"
	condition     = "RuleCondition"
	conditionItem = "RuleConditionItem"
	compute       = "RuleCompute"
	actionSQL     = "RuleSql"
	actionHttp    = "RuleHttp"
	actionCall    = "RuleCall"
	actionJson    = "RuleJson"
	analysisMap   = map[string]func(code map[string]interface{}, designInfo *DesignJson) Rule{
		begin:      AnalysisBegin,
		end:        AnalysisEnd,
		action:     AnalysisAction,
		condition:  AnalysisCondition,
		compute:    AnalysisCompute,
		actionSQL:  AnalysisAction,
		actionHttp: AnalysisAction,
		actionCall: AnalysisAction,
		actionJson: AnalysisAction,
	}
)

type DesignJson struct {
	Id              string                            `json:"id" binding:"required"`
	ContainerId     string                            `json:"containerid"`
	Width           interface{}                       `json:"width"`
	Height          interface{}                       `json:"height"`
	Attrs           interface{}                       `json:"attrs"`
	ControlType     string                            `json:"controlType"`
	ModelType       string                            `json:"modelType"`
	CurIndex        int64                             `json:"curIndex"`
	RootModels      map[string]map[string]interface{} `json:"rootModels"`
	LinkGroups      map[string]map[string]interface{} `json:"linkGroups"`
	DesignerVersion string                            `json:"designerVersion"`
	SaveTempStyle   interface{}                       `json:"SAVE_TEMP_STYLE"`
}

type ResultDefine struct {
	Tracks          bool     `json:"tracks"`
	Arguments       bool     `json:"arguments"`
	Time            bool     `json:"time"`
	BasicInfo       bool     `json:"basicInfo"`
	Result          bool     `json:"result"`
	ResultArguments []string `json:"resultArguments"`
	Compress        bool     `json:"compress"` // 是否压缩
}

// AnalysisDesign /* 将设计图转换为规则Yaml**/
func AnalysisDesign(designJson string) (error, string) {
	var designInfo *DesignJson
	var rules Rules
	err := json.Unmarshal([]byte(designJson), &designInfo)
	if err != nil {
		// 转换错误解析失败
		return err, ""
	}
	// 节点线归类
	availableModels, invalidModels, availableLine, invalidLine := VerifyInvalidNode(designInfo.RootModels)
	maps.Copy(availableModels, availableLine)
	// 处理rootModels 只保留可用节点
	designInfo.RootModels = availableModels

	// 修改linkGroups去掉不可用节点关系
	for k, v := range designInfo.LinkGroups {
		// 去掉不需要的节点
		for i, _ := range invalidModels {
			if strings.HasPrefix(k, i) {
				delete(designInfo.LinkGroups, k)
			}
		}
		// 去掉不需要的线
		lines := v["lineIds"].([]interface{})
		var copyLines []interface{}
		for _, l := range lines {
			isInvalid := true
			for lid, _ := range invalidLine {
				if lid == l {
					//copyLines = append(lines[0:i], lines[i+1:]...)
					isInvalid = false
					break
				}
			}
			if isInvalid {
				copyLines = append(copyLines, l)
			}
		}
		v["lineIds"] = copyLines
	}
	// 解析节点
	for _, model := range designInfo.RootModels {
		//analysis, ok := analysisMap[k[:strings.Index(k, "_")+1]]

		modelType := model["modelType"].(string)
		analysis, ok := analysisMap[modelType]
		if ok {
			rules.Rules = append(rules.Rules, analysis(model, designInfo))
		}
	}

	// 处理下一步关系
	BuildRelationships(&rules, designInfo)
	marshal, err := yaml.Marshal(&rules)
	if err != nil {
		return err, ""
	}
	return nil, string(marshal)
}

// AnalysisCanvas 修改画布基础属性
func AnalysisCanvas(designJson string) (string, string, string, string, uint64) {
	var designInfo *DesignJson
	json.Unmarshal([]byte(designJson), &designInfo)
	// 解析画布基础信息
	attrs := designInfo.Attrs.(map[string]interface{})
	name := attrs["ruleName"].(string)
	var code string
	if attrs["code"] != nil {
		code = attrs["ruleCode"].(string)
	}
	var desc string
	if attrs["ruleDesc"] != nil {
		desc = attrs["ruleDesc"].(string)
	}
	var uri string
	if attrs["ruleURI"] != nil {
		uri = attrs["ruleURI"].(string)
	}
	var groupId uint64
	if attrs["ruleGroup"] != nil {
		if reflect.ValueOf(attrs["ruleGroup"]).Kind() == reflect.String {
			groupId, _ = strconv.ParseUint(attrs["ruleGroup"].(string), 10, 64)
		} else {
			groupId = uint64(attrs["ruleGroup"].(float64))
		}
	}
	return name, code, desc, uri, groupId
}

// AnalysisResultJson 取出返回值配置
func AnalysisResultJson(designJson string) (error, ResultDefine) {
	var designInfo *DesignJson
	var rd ResultDefine
	err := json.Unmarshal([]byte(designJson), &designInfo)
	if err != nil {
		// 转换错误解析失败
		return err, ResultDefine{}
	}
	// 解析画布基础信息
	attrs := designInfo.Attrs.(map[string]interface{})
	if attrs["resultDefine"] == nil {
		return nil, ResultDefine{}
	}
	resultDefine := attrs["resultDefine"].(string)
	json.Unmarshal([]byte(resultDefine), &rd)
	//err = mapstructure.Decode(resultDefine, &rd)
	if err != nil {
		return nil, ResultDefine{}
	}
	return nil, rd
}

// AnalysisBegin /* 解析开始节点**/
func AnalysisBegin(v map[string]interface{}, designInfo *DesignJson) Rule {
	var rule Rule
	if strings.HasPrefix(v["id"].(string), "begin_") {
		// 开始节点
		rule.Type = "start"
		attrs := v["attrs"].(map[string]interface{})
		AnalysisBasicProperties(v, attrs, &rule)
	} else {
		//结束节点
		return AnalysisEnd(v, designInfo)
	}
	return rule
}

// AnalysisAction /* 解析执行节点,执行sql节点,执行http节点**/
func AnalysisAction(v map[string]interface{}, designInfo *DesignJson) Rule {
	var actionRule Rule
	actionRule.Type = "action"
	attrs := v["attrs"].(map[string]interface{})
	AnalysisBasicProperties(v, attrs, &actionRule)
	AnalysisExecute(attrs, &actionRule)
	AnalysisExecuteSQL(attrs, &actionRule)
	AnalysisExecuteHttp(attrs, &actionRule)
	AnalysisExecuteJson(attrs, &actionRule)
	AnalysisExecuteCallRule(attrs, &actionRule)
	AnalysisTogether(attrs, &actionRule, designInfo)
	AnalysisIsResult(attrs, &actionRule)
	return actionRule
}

// AnalysisCondition /* 解析条件节点**/
func AnalysisCondition(v map[string]interface{}, designInfo *DesignJson) Rule {
	var conditionRule Rule
	conditionRule.Type = "condition"
	attrs := v["attrs"].(map[string]interface{})
	AnalysisBasicProperties(v, attrs, &conditionRule)
	AnalysisTogether(attrs, &conditionRule, designInfo)
	return conditionRule
}

// AnalysisCompute /* 解析计算节点**/
func AnalysisCompute(v map[string]interface{}, designInfo *DesignJson) Rule {
	var computeRule Rule
	computeRule.Type = "compute"
	attrs := v["attrs"].(map[string]interface{})
	AnalysisBasicProperties(v, attrs, &computeRule)
	AnalysisTogether(attrs, &computeRule, designInfo)
	AnalysisExpression(attrs, &computeRule)
	return computeRule
}

// AnalysisEnd /* 解析结束节点**/
func AnalysisEnd(v map[string]interface{}, designInfo *DesignJson) Rule {
	var endRule Rule
	endRule.Type = "end"
	attrs := v["attrs"].(map[string]interface{})
	AnalysisBasicProperties(v, attrs, &endRule)
	AnalysisTogether(attrs, &endRule, designInfo)
	return endRule
}

// AnalysisBasicProperties / * 解析基础属性**/
func AnalysisBasicProperties(model map[string]interface{}, attrs map[string]interface{}, rule *Rule) {
	rule.Id = model["id"].(string)
	rule.Code = attrs["code"].(string)
	if attrs["name"] == nil || len(attrs["name"].(string)) == 0 {
		rule.Name = rule.Code
	} else {
		rule.Name = attrs["name"].(string)
	}
	if attrs["desc"] != nil {
		rule.Desc = attrs["desc"].(string)
	}
	rule.Salience = 0
	if model["togetherWeight"] != nil {
		rule.TogetherWeight = model["togetherWeight"].(int)
	}
}

// AnalysisTogether /* 解析聚合属性**/
func AnalysisTogether(attrs map[string]interface{}, rule *Rule, designInfo *DesignJson) {
	// 判断是否是聚合节点
	if attrs["together"] == nil || attrs["together"].(string) == "y" {
		rule.Together = true
		for modelId, model := range designInfo.RootModels {
			if strings.HasPrefix(modelId, "rd_line") && FormatModelsCode(model["endLinkGroupId"].(string)) == rule.Id {
				rule.TogetherCount++
			}
		}
	}
}

// AnalysisExpression /* 解析计算表达式属性*/
func AnalysisExpression(attrs map[string]interface{}, rule *Rule) {
	if attrs["expression"] == nil {
		return
	}
	expressionStr := strings.ReplaceAll(attrs["expression"].(string), " ", "")
	expressionStr = strings.ReplaceAll(expressionStr, "\n", "")
	expressionStr = strings.ReplaceAll(expressionStr, "\r", "")
	var expressionArguments string
	if attrs["expressionArguments"] != nil {
		expressionArguments = attrs["expressionArguments"].(string)
	}
	if len(expressionArguments) == 0 {
		expressionArguments = attrs["name"].(string)
	}
	// 计算精度处理
	if attrs["expressionPrecision"] == nil {
		attrs["expressionPrecision"] = 2.0
	}
	expressionPrecision := attrs["expressionPrecision"].(float64)

	rule.Expression = expressionStr
	rule.ExpressionArguments = expressionArguments
	rule.ExpressionPrecision = int(math.Floor(expressionPrecision))
	//expressionStr = strings.Replace(expressionStr, " ", "", -1)
	//if strings.Contains(expressionStr, "=") {
	//	index := strings.Index(expressionStr, "=")
	//	rule.Expression = expressionStr[:index]
	//	rule.ExpressionArguments = expressionStr[index+1:]
	//} else {
	//	rule.Expression = expressionStr
	//	rule.ExpressionArguments = attrs["name"].(string)
	//}
	// 是否作为返回结果
	AnalysisIsResult(attrs, rule)
}

// AnalysisExecuteSQL /*解析执行SQL属性**/
func AnalysisExecuteSQL(attrs map[string]interface{}, rule *Rule) {
	if attrs["executeSql"] == nil || reflect.TypeOf(attrs["executeSql"]).String() == "string" {
		return
	}
	executeSqlMap := attrs["executeSql"].(map[string]interface{})
	//json.Unmarshal([]byte(executeSql), &executeSqlMap)
	if executeSqlMap["dataBase"] != nil {
		rule.ExecuteSQL.DataBase = executeSqlMap["dataBase"].(string)
	}
	if executeSqlMap["sql"] != nil {
		// 将双引号替换成单引号
		rule.ExecuteSQL.Sql = strings.ReplaceAll(executeSqlMap["sql"].(string), "\"", "'")
	}
	if executeSqlMap["result"] != nil {
		executeSQLResults := executeSqlMap["result"].([]interface{})
		for _, result := range executeSQLResults {
			var executeSqlResult ExecuteMappingResult
			resultMap := result.(map[string]interface{})
			if resultMap["path"] != nil {
				executeSqlResult.Path = resultMap["path"].(string)
			}
			if resultMap["title"] != nil {
				executeSqlResult.Title = resultMap["title"].(string)
			}
			executeSqlResult.Selected = resultMap["selected"].(bool)
			if resultMap["def"] != nil {
				switch resultMap["def"].(type) {
				default:
					executeSqlResult.Def = resultMap["def"].(string)
				case float64:
					def := resultMap["def"].(float64)
					executeSqlResult.Def = strconv.FormatFloat(def, 'f', 8, 64)
				}
			}
			rule.ExecuteSQL.Result = append(rule.ExecuteSQL.Result, executeSqlResult)
		}
	}
}

// AnalysisExecuteHttp /*解析执行Http属性**/
func AnalysisExecuteHttp(attrs map[string]interface{}, rule *Rule) {
	if attrs["executeHttp"] == nil {
		return
	}
	executeHttpMap := attrs["executeHttp"].(map[string]interface{})
	rule.ExecuteHttp.System = executeHttpMap["system"].(string)
	if executeHttpMap["method"] == nil {
		rule.ExecuteHttp.Method = "GET"
	} else {
		rule.ExecuteHttp.Method = executeHttpMap["method"].(string)
	}
	if executeHttpMap["url"] != nil {
		rule.ExecuteHttp.Url = executeHttpMap["url"].(string)
	}
	rule.ExecuteHttp.Query = parseParam(executeHttpMap["query"].([]interface{}))
	if executeHttpMap["body"] != nil {
		rule.ExecuteHttp.Body = executeHttpMap["body"].(string)
	}
	rule.ExecuteHttp.Header = parseParam(executeHttpMap["header"].([]interface{}))

	var result []ExecuteMappingResult
	resultList := executeHttpMap["result"].([]interface{})
	for _, r := range resultList {
		var er ExecuteMappingResult
		resultMap := r.(map[string]interface{})
		er.Selected = resultMap["selected"].(bool)
		if resultMap["title"] != nil {
			er.Title = resultMap["title"].(string)
		}
		if resultMap["path"] != nil {
			er.Path = resultMap["path"].(string)
		}
		if resultMap["def"] != nil {
			er.Def = resultMap["def"].(string)
		}
		result = append(result, er)
	}
	rule.ExecuteHttp.Result = result
}

// AnalysisExecuteJson 解析执行Json属性
func AnalysisExecuteJson(attrs map[string]interface{}, rule *Rule) {
	if attrs["executeJson"] == nil {
		return
	}
	executeJsonRule := attrs["executeJson"].(map[string]interface{})
	if executeJsonRule["jsonArguments"] != nil {
		rule.ExecuteJson.JsonArguments = executeJsonRule["jsonArguments"].(string)
	}
	if executeJsonRule["jsonModel"] != nil {
		rule.ExecuteJson.JsonModel = executeJsonRule["jsonModel"].(string)
	}
	if executeJsonRule["result"] != nil {
		resultList := executeJsonRule["result"].([]interface{})
		var result []ExecuteMappingResult
		for _, v := range resultList {
			resultMap := v.(map[string]interface{})
			var er ExecuteMappingResult
			er.Selected = resultMap["selected"].(bool)
			er.Path = resultMap["path"].(string)
			er.Title = resultMap["title"].(string)
			if resultMap["def"] != nil {
				er.Def = resultMap["def"].(string)
			}
			result = append(result, er)
		}
		rule.ExecuteJson.Result = result
	}
	if executeJsonRule["setResult"] != nil {
		setResultList := executeJsonRule["setResult"].([]interface{})
		var setResult []ExecuteSetResult
		for _, v := range setResultList {
			var sr ExecuteSetResult
			setResultMap := v.(map[string]interface{})
			sr.Selected = setResultMap["selected"].(bool)
			sr.Key = setResultMap["key"].(string)
			sr.Value = setResultMap["value"].(string)
			setResult = append(setResult, sr)
		}
		rule.ExecuteJson.SetResult = setResult
	}
}

// AnalysisExecuteCallRule 执行调用其他规则(子规则)属性解析
func AnalysisExecuteCallRule(attrs map[string]interface{}, rule *Rule) {
	if attrs["executeCallRule"] == nil {
		return
	}
	executeCallRule := attrs["executeCallRule"].(map[string]interface{})
	if executeCallRule["ruleCode"] != nil {
		rule.ExecuteCallRule.Code = executeCallRule["ruleCode"].(string)
	}
	if executeCallRule["ruleVersion"] != nil {
		rule.ExecuteCallRule.Version = executeCallRule["ruleVersion"].(string)
	}
	// 子规则参数解析
	var arguments []ExecuteCallArguments
	argumentsList := executeCallRule["arguments"].([]interface{})
	for _, arg := range argumentsList {
		var eca ExecuteCallArguments
		argMap := arg.(map[string]interface{})
		eca.Selected = argMap["selected"].(bool)
		if argMap["key"] != nil {
			eca.Key = argMap["key"].(string)
		}
		if argMap["value"] != nil {
			eca.Value = argMap["value"].(string)
		}
		if argMap["def"] != nil {
			eca.Def = argMap["def"].(string)
		}
		arguments = append(arguments, eca)
	}
	rule.ExecuteCallRule.Arguments = arguments
	// 返回值映射解析
	var result []ExecuteMappingResult
	resultList := executeCallRule["result"].([]interface{})
	for _, r := range resultList {
		var er ExecuteMappingResult
		resultMap := r.(map[string]interface{})
		er.Selected = resultMap["selected"].(bool)
		if resultMap["title"] != nil {
			er.Title = resultMap["title"].(string)
		}
		if resultMap["path"] != nil {
			er.Path = resultMap["path"].(string)
		}
		if resultMap["def"] != nil {
			er.Def = resultMap["def"].(string)
		}
		result = append(result, er)
	}
	rule.ExecuteCallRule.Result = result

	if executeCallRule["loopKey"] != nil {
		rule.ExecuteCallRule.LoopKey = executeCallRule["loopKey"].(string)
	}
	if executeCallRule["loopJoin"] != nil {
		rule.ExecuteCallRule.LoopJoin = executeCallRule["loopJoin"].(bool)
	}
	//rule.ExecuteCallRule.LoopKey = "测试数据集"
	//rule.ExecuteCallRule.LoopJoin = true
}

// AnalysisIsResult 解析是否是返回值
func AnalysisIsResult(attrs map[string]interface{}, rule *Rule) {
	if attrs["expressionResult"] == nil {
		return
	}
	// 是否作为返回结果
	if attrs["expressionResult"].(string) == "y" {
		rule.ExpressionResult = true
	}
}

// parseParam 格式化参数
func parseParam(m []interface{}) []ExecuteHttpParams {
	var ep []ExecuteHttpParams
	for _, mv := range m {
		var e ExecuteHttpParams
		mp := mv.(map[string]interface{})
		e.Selected = mp["selected"].(bool)
		if mp["key"] != nil {
			e.Key = mp["key"].(string)
		}
		if mp["value"] != nil {
			e.Value = mp["value"].(string)
		}
		if mp["desc"] != nil {
			e.Desc = mp["desc"].(string)
		}
		ep = append(ep, e)
	}
	return ep
}

// AnalysisExecute /*解析执行属性**/
func AnalysisExecute(attrs map[string]interface{}, rule *Rule) {
	if attrs["executeCustom"] != nil {
		rule.ExecuteCustom = attrs["executeCustom"].(string)
	}
	if attrs["execute"] == nil {
		return
	}
	executes := attrs["execute"].(map[string]interface{})
	if executes["executeDefault"] != nil {
		executeDefaults := executes["executeDefault"].([]interface{})
		var executeDefaultStr []string
		for _, v := range executeDefaults {
			executeDefaultStr = append(executeDefaultStr, v.(string))
		}
		rule.Execute = executeDefaultStr
	}
	if executes["executeCustom"] != nil {
		rule.ExecuteCustom = executes["executeCustom"].(string)
	}
	if executes["executeType"] != nil {
		rule.ExecuteType = executes["executeType"].(string)
	}
}

// BuildRelationships /* 处理建立关系**/
func BuildRelationships(rules *Rules, designInfo *DesignJson) {
	// 建立关系
	for _, group := range designInfo.LinkGroups {
		for i, rule := range rules.Rules {
			if group["modelId"].(string) == rule.Id {
				// 获取到全部线
				lineIds := group["lineIds"].([]interface{})
				for _, id := range lineIds {
					// 找到对应线
					line := designInfo.RootModels[id.(string)]
					if FormatModelsCode(line["startLinkGroupId"].(string)) == rule.Id {
						// 找到下个节点
						toModelId := line["endLinkGroupId"].(string)
						lastIndex := strings.LastIndex(toModelId, "_")
						toModelId = toModelId[:lastIndex]
						model := designInfo.RootModels[toModelId]
						// 如果线指向条件项就不处理下一步
						if strings.HasPrefix(model["modelType"].(string), conditionItem) {
							var entityCondition Condition
							if GetAttrs(model)["defaultCondition"] != nil && GetAttrs(model)["defaultCondition"].(bool) {
								// 这是一个else
								entityCondition.DefaultFeature = true
							} else {
								// 存在没有设置条件的节点
								if GetAttrs(model)["conditionJson"] == nil {
									continue
								}
								conditionItemList := GetAttrs(model)["conditionJson"].([]interface{})
								// 解析feature
								for _, conditionItemMap := range conditionItemList {
									cm := conditionItemMap.(map[string]interface{})
									var feature Feature
									feature.Field = cm["field"].(string)
									feature.Type = cm["type"].(string)
									feature.Value = strings.ReplaceAll(ToStr(cm["value"]), "\"", "")
									feature.Logic = cm["logic"].(string)
									feature.Operator = cm["operator"].(string)
									entityCondition.Features = append(entityCondition.Features, feature)
								}
							}
							attrs := model["attrs"].(map[string]interface{})
							entityCondition.Condition = attrs["code"].(string)
							// 找到条件项下的关系(线)
							for k, itemLinkGroup := range designInfo.LinkGroups {
								if FormatModelsCode(k) == model["id"].(string) {
									itemLineIds := itemLinkGroup["lineIds"].([]interface{})
									for _, lineId := range itemLineIds {
										// 获取到线
										itemLine := designInfo.RootModels[lineId.(string)]
										if FormatModelsCode(itemLine["startLinkGroupId"].(string)) == model["id"].(string) {
											endModel := designInfo.RootModels[itemLine["endLinkGroupId"].(string)[:strings.LastIndex(itemLine["endLinkGroupId"].(string), "_")]]
											entityCondition.Next = append(rule.Next, GetAttrs(endModel)["code"].(string))
										}
									}
								}
							}
							rule.Conditions = append(rule.Conditions, entityCondition)
							rules.Rules[i] = rule
						} else {
							rules.Rules[i].Next = append(rules.Rules[i].Next, GetAttrs(model)["code"].(string))
						}
					}
				}
			}
		}
	}
}

// VerifyInvalidNode /* 找到全部 有效节点/无效节点 有效线/无效线
//  availableModels 有效节点
//  invalidModels 无效节点
//  availableLine 有效线
//  invalidLine 无效线
func VerifyInvalidNode(rootModels map[string]map[string]interface{}) (map[string]map[string]interface{}, map[string]map[string]interface{}, map[string]map[string]interface{}, map[string]map[string]interface{}) {
	// 有效节点
	availableModels := make(map[string]map[string]interface{})
	// 无效节点
	invalidModels := make(map[string]map[string]interface{})
	// 有限线
	availableLine := make(map[string]map[string]interface{})
	// 无效线
	invalidLine := make(map[string]map[string]interface{})
	// 获取全部的线
	lineModels := GetLineModels(rootModels)
	// 获取开始节点
	beginModel := GetBeginModel(rootModels)
	if len(rootModels) == 0 || beginModel["id"] == nil {
		// 如果只有一个节点，且没有线 单节点解析
		if len(rootModels) == 1 {
			oneModel := GetModelOne(rootModels)
			availableModels[oneModel["id"].(string)] = oneModel
		}
		return availableModels, invalidModels, availableLine, invalidLine
	}
	beginModel["togetherWeight"] = 1
	availableModels[beginModel["id"].(string)] = beginModel
	// 找到全部可用节点
	for _, model := range lineModels {
		if FormatModelsCode(model["startLinkGroupId"].(string)) == beginModel["id"].(string) {
			NextModel(rootModels, lineModels, availableModels, model)
		}
	}
	// 找到全部不可用节点
	for modelId, model := range rootModels {
		// 如果是线就验证下一个
		if model["modelType"].(string) == line {
			continue
		}
		isAvailable := true
		for k, _ := range availableModels {
			if model["id"].(string) == k {
				isAvailable = false
				break
			}
		}
		if isAvailable {
			invalidModels[modelId] = model
		}
	}
	// 找到全部有效线/无效线
	for lmId, lm := range lineModels {
		isAvailable := true
		for _, im := range invalidModels {
			if FormatModelsCode(lm["startLinkGroupId"].(string)) == im["id"].(string) {
				isAvailable = false
				break
			}
		}
		if isAvailable {
			availableLine[lmId] = lm
		} else {
			invalidLine[lmId] = lm
		}
	}
	return availableModels, invalidModels, availableLine, invalidLine
}

// NextModel 寻找找下一个有效节点
func NextModel(rootModels map[string]map[string]interface{}, lineModels map[string]map[string]interface{}, availableModels map[string]map[string]interface{}, model map[string]interface{}) {
	for mid, m := range rootModels {
		if FormatModelsCode(model["endLinkGroupId"].(string)) == mid {
			// 设置节点权重
			if m["togetherWeight"] == nil {
				m["togetherWeight"] = 1
			} else {
				m["togetherWeight"] = m["togetherWeight"].(int) + 1
			}
			availableModels[mid] = m
			for _, lineModel := range lineModels {
				if FormatModelsCode(lineModel["startLinkGroupId"].(string)) == m["id"].(string) {
					NextModel(rootModels, lineModels, availableModels, lineModel)
				}
			}
		}
	}
}

// GetLineModels 获取全部线model
func GetLineModels(rootModels map[string]map[string]interface{}) map[string]map[string]interface{} {
	lineModels := make(map[string]map[string]interface{})
	for modelId, model := range rootModels {
		if model["modelType"].(string) == line {
			lineModels[modelId] = model
		}
	}
	return lineModels
}

// GetBeginModel 获取到开始model
func GetBeginModel(rootModels map[string]map[string]interface{}) map[string]interface{} {
	beginModel := make(map[string]interface{})
	for id, model := range rootModels {
		if strings.HasPrefix(id, beginCode) {
			beginModel = model
			break
		}
	}
	return beginModel
}

// GetModelOne 获取单个model
func GetModelOne(rootModels map[string]map[string]interface{}) map[string]interface{} {
	oneModel := make(map[string]interface{})
	for _, model := range rootModels {
		oneModel = model
		break
	}
	return oneModel
}

// FormatModelsCode 去掉最后一个下划线后面的值防止出现action_1 action_11_top 这种情况
func FormatModelsCode(code string) string {
	index := strings.LastIndex(code, "_")
	return code[0:index]
}

// GraterUri 生成uri
func GraterUri(version string, code string) string {
	return "/" + version + "/" + code
}

// GetAttrs 获取model中attrs属性
func GetAttrs(model map[string]interface{}) (attrs map[string]interface{}) {
	attrs = model["attrs"].(map[string]interface{})
	return attrs
}
