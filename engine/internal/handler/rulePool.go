package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bilibili/gengine/engine"
	"github.com/gin-gonic/gin"
	dameng "github.com/godoes/gorm-dameng"
	oracle "github.com/godoes/gorm-oracle"
	jsoniter "github.com/json-iterator/go"
	"github.com/panjf2000/ants/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"zenith.engine.com/engine/config"
	"zenith.engine.com/engine/internal/adapter"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/pkg/util"
)

type ExecuteEngineNext struct {
	name      string
	t         string
	code      string
	nextCode  string
	arguments *sync.Map
	executeId string
	debug     DebugExecuteParam
}

var (
	EnginePoolMap  = sync.Map{}
	RuleResultMap  = sync.Map{}
	TogetherMap    = sync.Map{}
	max            = int64(runtime.NumCPU()) * 10
	min            = max / 2
	executePool, _ = ants.NewPool(runtime.NumCPU()*100, ants.WithNonblocking(false)) //执行调度池子
)

// SetPool 设置规则池/*
func SetPool(code string, pool *engine.GenginePool) {
	EnginePoolMap.Store(code, pool)
}

// GetPool 获取规则池/*
func GetPool(code string) *engine.GenginePool {
	val, ok := EnginePoolMap.Load(code)
	if ok {
		return val.(*engine.GenginePool)
	} else {
		return nil
	}
}

// DeleteDefaultPool 删除 默认规则池/*
func DeleteDefaultPool(code string) {
	EnginePoolMap.Delete(code)
}

// DeletePools 删除 engine规则池/*
func DeletePools(code string, design string) error {
	var rs = new(util.Rules)
	if err := yaml.Unmarshal([]byte(design), &rs); err != nil {
		return errors.New("规则Yaml解析失败")
	}
	for _, rule := range rs.Rules {
		EnginePoolMap.Delete(util.ParseRuleCode(code, rule.Code))
	}
	for _, calculate := range rs.Calculates {
		EnginePoolMap.Delete(util.ParseRuleCode(code, calculate.Code))
	}
	return nil
}

func GetProcess(executeId string) []*repository.Process {
	val, ok := EnginePoolMap.Load(executeId)
	if ok {
		return val.([]*repository.Process)
	} else {
		return nil
	}
}

// BuildEnginePool 构造G engine规则池/*
func BuildEnginePool(ruleStr string) *engine.GenginePool {
	apis := make(map[string]interface{})
	apis["print"] = fmt.Println
	apis["panic"] = Panic
	apis["isNil"] = isNil
	// 执行函数
	apis["executeEnginePool"] = ExecuteEnginePool
	apis["executeEngineRule"] = ExecuteEngineRule
	apis["executeEngineCallRule"] = ExecuteEngineCallRule
	apis["executeEngineCallRuleFor"] = ExecuteEngineCallRuleFor
	apis["buildExecuteEngineNext"] = BuildExecuteEngineNext

	apis["processRecord"] = ProcessRecord
	apis["togetherCheck"] = TogetherCheck
	apis["togetherCheck2"] = TogetherCheck2
	apis["togetherCondition"] = TogetherCondition
	// 返回值函数
	apis["setRuleResult"] = SetRuleResult
	apis["appendRuleResult"] = AppendRuleResult
	apis["removeRuleResult"] = RemoveRuleResult
	// 时间函数
	apis["dateAfterNow"] = DateAfterNow
	apis["dateNow"] = DateNow
	apis["timeAfter"] = TimeAfter
	apis["timeAfterEqual"] = TimeAfterEqual
	apis["timeBefore"] = TimeBefore
	apis["timeBeforeEqual"] = TimeBeforeEqual
	apis["timeEqual"] = TimeEqual
	// 数据库相关函数
	apis["executeSql"] = ExecuteSql
	apis["executeSqlResult"] = ExecuteSqlResult
	apis["connectDataBase"] = ConnectDataBase
	apis["connectDataExecuteSql"] = ConnectDataExecuteSql
	// 计算函数
	apis["computeExpression"] = ComputeExpression
	// 取参函数
	apis["loadNumberArguments"] = LoadNumberArguments
	apis["loadStringArguments"] = LoadStringArguments
	apis["loadDateArguments"] = LoadDateArguments
	apis["loadArgumentsAndTypeOf"] = loadArgumentsAndTypeOf
	// http请求函数
	apis["XGet"] = XGet
	apis["XPost"] = XPost
	apis["XSend"] = XSend
	// json转换函数
	apis["handleJson"] = HandleJson
	// 转换函数
	apis["toNumber"] = util.ToNumber
	apis["ToMap"] = ToMap
	apis["toDate"] = util.ToDate
	apis["toBool"] = util.ToBool
	apis["stringToArray"] = StringToArray
	apis["lenArray"] = LenArray
	// 计算函数
	apis["sum"] = util.Sum
	apis["max"] = util.Max
	apis["min"] = util.Min
	apis["avg"] = util.Average
	apis["mod"] = util.Mod
	apis["round"] = util.Round
	apis["int"] = util.Int
	apis["power"] = util.Power
	apis["abs"] = util.Abs
	apis["log"] = util.Logs
	apis["countif"] = util.CountIf
	apis["sumif"] = util.SumIf
	apis["isnumber"] = util.IsDigit
	apis["find"] = util.Find
	apis["mid"] = util.Mid
	apis["concat"] = util.Concat
	apis["len"] = util.Len
	apis["lenB"] = util.LenB
	apis["substitute"] = util.Substitute
	apis["replace"] = util.Replace
	apis["left"] = util.Left
	apis["right"] = util.Right
	// debug
	//apis["breakPointIn"] = BreakPointIn
	apis["argumentsGet"] = util.ArgumentsGet
	apis["argumentsSet"] = util.ArgumentsSet

	if config.Conf.Rule.PoolMax > 0 {
		max = config.Conf.Rule.PoolMax
	}
	if config.Conf.Rule.PoolMin > 0 {
		min = config.Conf.Rule.PoolMin
	}
	pool, _ := engine.NewGenginePool(min, max, 1, ruleStr, apis)
	return pool
}

func isNil(o interface{}) bool {
	if o == nil {
		return true
	}
	return false
}

// StringToArray 字符串数组转数组
func StringToArray(arrayAny any) []any {
	var array []any
	arrayString := arrayAny.(string)
	if len(arrayString) > 0 {
		err := json.Unmarshal([]byte(arrayString), &array)
		if err != nil {
			panic(arrayString + "转换失败")
		}
	}
	return array
}

// LenArray 获取数组长度
func LenArray(array []any) int {
	return len(array)
}

// ProcessRecord 流转记录存储/*
func ProcessRecord(name string, code string, t string, startTime time.Time, inputArguments *sync.Map, outArguments *sync.Map, status *sync.Map, executeId string) {
	process := adapter.GetData().GetProcessRecord(executeId)
	v, b := status.Load("err")
	if b {
		process = append(process, repository.Process{
			Code:           code,
			Name:           name,
			Type:           t,
			ExecuteTime:    FormatSubtle(startTime),
			OutArguments:   SyncMapToMap(outArguments),
			InputArguments: SyncMapToMap(inputArguments),
			Status:         "1",
			Error:          v.(error).Error(),
		})
	} else {
		process = append(process, repository.Process{
			Code:           code,
			Name:           name,
			Type:           t,
			ExecuteTime:    FormatSubtle(startTime),
			OutArguments:   SyncMapToMap(outArguments),
			InputArguments: SyncMapToMap(inputArguments),
			Status:         "0",
		})
	}

	adapter.GetData().SetProcessRecord(executeId, process)
}

// TogetherCondition 条件对聚合节点的处理/*
func TogetherCondition(yaml string, code string, executeId string, nextStr string, isConditionNode bool) {
	next := strings.Split(nextStr, "|")
	// 取出关系
	v, _ := util.NodeMap.Load(yaml)
	node := v.(*util.Node)
	for _, child := range node.Children {
		if child.Code == util.ParseRuleCode(yaml, code) {
			TogetherFor2(yaml, child, executeId, next, isConditionNode, true)
		}
	}
}

// TogetherFor 逐级遍历到根节点 适用于if else 模式分支
func TogetherFor(yaml string, child *util.Node, executeId string, next []string) {
	for _, c := range child.Children {
		// 排除掉是此次节点
		if SplitContents(yaml, c.Code, next) {
			continue
		}
		if c.Together {
			TogetherCheck(c.Code, -1, executeId)
		}
		if len(c.Children) > 0 {
			TogetherFor(yaml, c, executeId, next)
		}
	}
}

// TogetherFor2 逐级遍历到根节点 适用于多if 分支 预执行
func TogetherFor2(yaml string, child *util.Node, executeId string, next []string, isConditionNode bool, b bool) {
	for _, c := range child.Children {
		// 排除掉不是此次节点
		if !SplitContents(yaml, c.Code, next) && b {
			continue
		}
		//if c.Together {
		TogetherNext(c.Code, child.TogetherWeight, executeId)
		//}
		if len(c.Children) > 0 && isConditionNode {
			if TogetherCheck2(c.Code, c.TogetherWeight, executeId) {
				TogetherFor2(yaml, c, executeId, next, isConditionNode, false)
			}
		}
	}
}

// SplitContents 是否包含本次条件
func SplitContents(yamlCode string, s string, n []string) bool {
	for _, v := range n {
		if util.ParseRuleCode(yamlCode, v) == s {
			return true
		}
	}
	return false
}

//TogetherCheck2  验证是否满足权重 聚合节点继续执行
func TogetherCheck2(code string, togetherWeight int, executeId string) bool {
	value, ok := TogetherMap.Load(executeId + code)
	if ok {
		if value.(int) >= togetherWeight {
			TogetherMap.Delete(executeId + code)
			return true
		}
	}
	return false
}

// TogetherNext 增加下一个节点运行时权重
func TogetherNext(code string, togetherWeight int, executeId string) {
	value, ok := TogetherMap.Load(executeId + code)
	if ok {
		// 当前节点已经有权重 追加
		TogetherMap.Store(executeId+code, value.(int)+togetherWeight)
		return
	}
	TogetherMap.Store(executeId+code, togetherWeight)
}

// TogetherCheck 验证是否满足聚合继续执行/*
func TogetherCheck(code string, togetherCount int, executeId string) bool {
	value, ok := TogetherMap.Load(executeId + code)
	if ok {
		TogetherMap.Store(executeId+code, value.(int)+1)
		if togetherCount == -1 {
			return false
		}
		if value.(int)+1 >= togetherCount {
			TogetherMap.Delete(executeId + code)
			return true
		}
		return false
	}
	TogetherMap.Store(executeId+code, 1)
	return false
}

// LoadNumberArguments 获取数字参数
func LoadNumberArguments(key string, arguments *sync.Map) float64 {
	if strings.HasPrefix(key, util.ArgumentsDefiner) {
		key = key[1:]
	} else {
		v, _ := strconv.ParseFloat(key, 64)
		return v
	}
	if util.ArgumentsGet(arguments, key) == nil {
		return 0
	}
	switch util.ArgumentsGet(arguments, key).(type) {
	case map[string]interface{}:
		m := util.ArgumentsGet(arguments, key).(map[string]interface{})
		v, _ := strconv.ParseFloat(m["value"].(string), 64)
		return v
	case float64:
		return util.ArgumentsGet(arguments, key).(float64)
	default:
		v, _ := strconv.ParseFloat(util.ArgumentsGet(arguments, key).(string), 64)
		return v
	}
}

// LoadStringArguments 获取字符串参数
func LoadStringArguments(key string, arguments *sync.Map) string {
	if strings.HasPrefix(key, util.ArgumentsDefiner) {
		key = key[1:]
	} else {
		return key
	}
	if util.ArgumentsGet(arguments, key) == nil {
		return ""
	}
	return util.ArgumentsGet(arguments, key).(string)
}

// loadArgumentsAndTypeOf 获取参数并验证参数类型 todo 需要修改
func loadArgumentsAndTypeOf(field string, value string, operator string, arguments *sync.Map) bool {
	var b bool
	if strings.HasPrefix(field, util.ArgumentsDefiner) {
		field = util.ArgumentsGet(arguments, field[1:]).(string)
	}
	switch value {
	case "number":
		if len(field) > 0 {
			b = util.IsDigit(field)
		}
	case "string":
		b = !util.IsDigit(field)
	case "date":
		b = util.IsDate(field)
	default:
	}
	if operator == "NE" {
		return !b
	}
	return b
}

// LoadDateArguments 获取日期参数
func LoadDateArguments(key string, arguments *sync.Map) time.Time {
	if strings.HasPrefix(key, util.ArgumentsDefiner) {
		key = key[1:]
	} else {
		return util.ToDate(key)
	}
	if util.ArgumentsGet(arguments, key) == nil {
		panic("没有获取到参数:" + key)
	}
	switch util.ArgumentsGet(arguments, key).(type) {
	case string:
		return util.ToDate(util.ArgumentsGet(arguments, key).(string))
	default:
		return util.ArgumentsGet(arguments, key).(time.Time)
	}
}

// ComputeExpression 计算方法 表达式  结果,错误
func ComputeExpression(expression string, inputArguments *sync.Map, arguments *sync.Map, precision int, executeId string, status *sync.Map) string {
	expression = strings.ReplaceAll(expression, "\\n", "")
	//expression = strings.ToLower(expression)
	// 验证计算表达式里是否有变量
	if strings.Contains(expression, util.ArgumentsDefiner) {
		char := util.CharAt(expression)
		for i, _ := range char {
			if strings.HasPrefix(char[i], util.ArgumentsDefiner) {
				argument := strings.ReplaceAll(char[i], util.ArgumentsDefiner, "")
				val := util.ArgumentsGet(inputArguments, argument)
				// 在变量池中没有找到此变量
				if val != nil {
					switch val.(type) {
					case string:
						//if len(val.(string)) == 0 {
						//	char[i] = "0"
						//} else {
						char[i] = val.(string)
						//}
					case map[string]interface{}:
						valMap := val.(map[string]interface{})
						if valMap["value"] == nil {
							char[i] = util.ToStr(valMap)
						} else {
							char[i] = valMap["value"].(string)
						}
					case int:
						char[i] = strconv.Itoa(val.(int))
					case int64:
						char[i] = strconv.FormatInt(val.(int64), 10)
					case float64:
						char[i] = strconv.FormatFloat(val.(float64), 'f', 10, 64)
					case []interface{}:
						b, _ := json.Marshal(val.([]interface{}))
						char[i] = string(b)
					default:
						char[i] = val.(string)
					}
				} else {
					if len(config.Conf.Variable.ApiSingle) == 0 {
						return ""
					}
					// 从配置的api里面查找
					body, _ := json.Marshal(inputArguments)
					data, err := XPost(config.Conf.Variable.ApiSingle+"/"+argument, nil, string(body))
					if err != nil {
						panic(err)
					}
					variableMap := make(map[string]map[string]interface{})
					err = json.Unmarshal(data, &variableMap)
					if err != nil {
						return ""
					}
					// 存入全局参数arguments
					for k, v := range variableMap {
						arguments.Store(k, v)
					}
					argumentKey := variableMap[argument]
					if argumentKey != nil {
						char[i] = argumentKey["value"].(string)
					} else {
						return ""
					}
				}
			}
		}
		var str string
		for _, s := range char {
			str += s
		}
		expression = str
	}
	result, err := util.ComputeFuncResult1(strings.ToLower(expression), precision, true)
	if !util.IsDigit(result) {
		result = strings.ReplaceAll(result, "\\", "")
		result = strings.Trim(result, "\"")
	}
	if err != nil {
		status.Store("err", err)
		SetRuleResult(executeId, "errorMsg", err.Error())
	}
	return result
}

// Panic 终止规则/*
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}

// ToMap 转换为map类型
func ToMap(m interface{}) map[string]interface{} {
	switch m.(type) {
	case map[string]interface{}:
		return m.(map[string]interface{})
	}
	return nil
}

// ExecuteEnginePool 执行下一步编码规则/*
func ExecuteEnginePool(name string, t string, code string, nextCode string, arguments *sync.Map, executeId string) (map[string]interface{}, error) {
	pool := GetPool(nextCode)
	if pool == nil {
		return nil, errors.New("规则执行失败")
	}
	data := make(map[string]interface{})
	data["startTime"] = time.Now()
	data["arguments"] = arguments
	data["inputArguments"] = CopySyncMapVal(arguments)
	data["executeId"] = executeId
	data["debugParam"] = DebugExecuteParam{}
	data["es"] = make(map[string]ExecuteEngineNext)
	data["status"] = &sync.Map{}
	arguments.Range(func(k, v any) bool {
		data[k.(string)] = v
		return true
	})
	err, m := pool.Execute(data, true)
	if err != nil { // 规则执行错误处理
		fmt.Println("执行规则失败:" + nextCode + "  error:" + err.Error())
		// 执行规则错误保证报存规则信息,可重新执行规则
		value, b := RuleResultMap.Load(executeId)
		if !b {
			value = make(map[string][]interface{})
		}
		adapter.GetData().SetErrorInfo(repository.ErrorInfoRunner{
			Code:          code,
			NextCode:      nextCode,
			ExecuteId:     executeId,
			Arguments:     SyncMapToMap(arguments),
			RuleResultMap: ParseResultMap(value.(*sync.Map)),
			ErrorMsg:      err.Error(),
		})
		SetRuleResult(executeId, "errorMsg", err.Error())
		return nil, nil
	}
	//wg := sync.WaitGroup{}
	for _, i := range m {
		me := i.(map[string]ExecuteEngineNext)
		//wg.Add(len(me))
		for _, m2 := range me {
			p := m2
			//executePool.Submit(func() {
			ExecuteEnginePool(p.name, p.t, p.code, p.nextCode, p.arguments, p.executeId)
			//wg.Done()
			//})
		}
	}
	//wg.Wait()
	return m, nil
}

// BuildExecuteEngineNext 封装下一步信息
func BuildExecuteEngineNext(name string, t string, code string, nextCode string, arguments *sync.Map, executeId string, debug DebugExecuteParam) ExecuteEngineNext {
	return ExecuteEngineNext{
		name:      name,
		t:         t,
		code:      code,
		nextCode:  nextCode,
		arguments: arguments,
		executeId: executeId,
		debug:     debug,
	}
}

// ExecuteEngineRule 执行其他规则 会污染父规则
func ExecuteEngineRule(code string, version string, arguments *sync.Map, executeId string) {
	ruleDesign, _ := adapter.GetStorage().GetByCode(code, version)
	yamlCode := util.ParseYamlCode(ruleDesign.Code, ruleDesign.Version)
	// 通过code获取开始节点
	pool := GetPool(util.ParseRuleCode(yamlCode, "start"))
	if pool == nil {
		// 解析yaml-Rule
		ruleStrSlice, err := util.ExplainYaml(yamlCode, ruleDesign.Design)
		if err != nil {
			return
		}
		for k, v := range ruleStrSlice {
			SetPool(k, BuildEnginePool(v))
		}
	}
	// 执行开始规则
	ExecuteEnginePool("", "", "", util.ParseRuleCode(yamlCode, "start"), arguments, executeId)
}

// ExecuteEngineCallRuleFor 循环执行子规则
func ExecuteEngineCallRuleFor(call func(string, string, string, string, bool, *sync.Map, string) ExecuteRuleResult, code string, version string, argumentsFormatJson string, resultFormatJson string, result bool, arguments *sync.Map, executeId string, loopKey string, loopJoin bool) {
	var loopArguments []map[string]interface{}
	loopArgumentsJson := util.HasArguments(SyncMapToMap(arguments), loopKey[1:])
	j := jsoniter.Config{
		UseNumber: true,
	}.Froze()
	j.Unmarshal([]byte(loopArgumentsJson), &loopArguments)
	argumentsFormatJson = strings.ReplaceAll(argumentsFormatJson, "\\", "")
	// 格式化结果
	var callArguments []util.ExecuteCallArguments
	j.Unmarshal([]byte(argumentsFormatJson), &callArguments)
	for i, v := range loopArguments {
		// 拼装子规则请求参数
		for k, v1 := range v {
			var executeCallArguments util.ExecuteCallArguments
			executeCallArguments.Key = k
			executeCallArguments.Value = util.StrVal(v1)
			executeCallArguments.Selected = true
			callArguments = append(callArguments, executeCallArguments)
		}
		callArgument, _ := j.Marshal(callArguments)
		argumentsFormatJson = util.FormatJson(string(callArgument))
		er := call(code, version, argumentsFormatJson, resultFormatJson, result, arguments, executeId)
		if loopJoin {
			for k, rv := range er.Result {
				v[k] = rv[len(rv)-1]
			}
			loopArguments[i] = v
		}
	}
	if loopJoin {
		arguments.Store(loopKey[1:], loopArguments)
		SetRuleResult(executeId, loopKey[1:], loopArguments)
	}
}

// ExecuteEngineCallRule 执行子规则 不污染父规则
func ExecuteEngineCallRule(code string, version string, argumentsFormatJson string, resultFormatJson string, result bool, arguments *sync.Map, executeId string) ExecuteRuleResult {
	callArgumentsSync := CopySyncMapVal(arguments)
	// 解析请求参数
	if len(argumentsFormatJson) > 0 {
		argumentsFormatJson = strings.ReplaceAll(argumentsFormatJson, "\\", "")
		// 格式化结果
		var callArguments []util.ExecuteCallArguments
		err := json.Unmarshal([]byte(argumentsFormatJson), &callArguments)
		if err != nil {
			SetRuleResult(executeId, "errorMsg", "返回格式解析失败! resultFormatJson:"+resultFormatJson)
		}
		for _, arg := range callArguments {
			if arg.Selected {
				if strings.HasPrefix(arg.Key, util.ArgumentsDefiner) {
					kv, _ := arguments.Load(arg.Key[1:])
					arg.Key = kv.(string)
				}
				if strings.HasPrefix(arg.Value, util.ArgumentsDefiner) {
					kv, _ := arguments.Load(arg.Value[1:])
					arg.Value = kv.(string)
				}
				if len(arg.Value) == 0 {
					arg.Value = arg.Def
				}
				callArgumentsSync.Store(arg.Key, arg.Value)
			}
		}
	}

	// 调用规则
	resultMap := execute(ExecuteRuleParam{Code: code, Version: version, Arguments: SyncMapToMap(callArgumentsSync), Env: util.DEV})
	r, _ := resultMap.Load("executeRuleResult")
	er := r.(ExecuteRuleResult)
	erJson, _ := json.Marshal(er)
	// 处理结果映射
	if len(resultFormatJson) > 0 {
		resultFormatJson = strings.ReplaceAll(resultFormatJson, "\\", "")
		// 格式化结果
		var callResults []util.ExecuteMappingResult
		err := json.Unmarshal([]byte(resultFormatJson), &callResults)
		if err != nil {
			panic("返回格式解析失败! resultFormatJson:" + resultFormatJson)
		}
		for _, callResult := range callResults {
			// 增加主规则参数
			if callResult.Selected {
				rc := gjson.Get(string(erJson), callResult.Path)
				if rc.Exists() && len(rc.String()) > 0 {
					arguments.Store(callResult.Title, rc.String())
				} else {
					arguments.Store(callResult.Title, callResult.Def)
				}
			}
			// 增加主规则返回值
			if result {
				rs, _ := arguments.Load(callResult.Title)
				AppendRuleResult(executeId, callResult.Title, rs)
			}
		}
	}
	return er
}

// FormatSubtle 格式化时间微妙/*
func FormatSubtle(startTime time.Time) string {
	return strconv.FormatInt(time.Since(startTime).Microseconds(), 10) + "μs"
}

// SetRuleResult 设置规则返回值/*
func SetRuleResult(executeId string, key interface{}, val ...interface{}) {
	var vales []interface{}
	for _, s := range val {
		vales = append(vales, s)
	}
	mk := &sync.Map{}
	mk.Store(key.(string), vales)
	RuleResultMap.Store(executeId, mk)
}

// AppendRuleResult 添加规则返回值/*
func AppendRuleResult(executeId string, key string, val ...interface{}) {
	v, b := RuleResultMap.Load(executeId)
	if !b {
		SetRuleResult(executeId, key, val)
	}
	mk := v.(*sync.Map)
	var vales []interface{}
	vs, _ := mk.Load(key)
	if vs != nil {
		v1 := vs.([]interface{})
		vales = append(vales, v1...)
	}
	for _, s := range val {
		vales = append(vales, s)
	}
	mk.Store(key, vales)
}

// RemoveRuleResult 删除规则返回值/*
func RemoveRuleResult(executeId string, key string) {
	v, _ := RuleResultMap.Load(executeId)
	mk := v.(map[string][]interface{})
	delete(mk, key)
}

// ExecuteEngineEnd 执行规则结束后操作/*
func ExecuteEngineEnd(executeId string, b bool) {
	adapter.GetData().DelProcessRecord(executeId)
	RuleResultMap.Delete(executeId)
	if b {
		DebugNodeQueue.Delete(executeId)
	}
}

// CopyMapVal 复制一个map的全部值/*
func CopyMapVal(arguments map[string]interface{}) map[string]interface{} {
	inputArguments := make(map[string]interface{})
	for key, value := range arguments {
		inputArguments[key] = value
	}
	return inputArguments
}

// CopySyncMapVal 复制一个map的全部值/*
func CopySyncMapVal(arguments *sync.Map) *sync.Map {
	var inputArguments = &sync.Map{}
	arguments.Range(func(key, value any) bool {
		inputArguments.Store(key, value)
		return true
	})
	return inputArguments
}

// ParseResultMap 转换为返回值map
func ParseResultMap(arguments *sync.Map) map[string][]interface{} {
	var resultMap = make(map[string][]interface{})
	arguments.Range(func(key, value any) bool {
		resultMap[key.(string)] = value.([]interface{})
		return true
	})
	return resultMap
}

// SyncMapToMap sync.map 转 map/*
func SyncMapToMap(arguments *sync.Map) map[string]interface{} {
	inputArguments := make(map[string]interface{})
	arguments.Range(func(key, value any) bool {
		inputArguments[key.(string)] = value
		return true
	})
	return inputArguments
}

// MapToSyncMap map 转 sync/*
func MapToSyncMap(arguments map[string]interface{}) *sync.Map {
	var inputArguments = &sync.Map{}
	for k, v := range arguments {
		inputArguments.Store(k, v)
	}
	return inputArguments
}

// DateAfterNow 当前时间 大于 date/*
func DateAfterNow(date time.Time) bool {
	return time.Now().After(date)
}

// TimeAfter date 大于 date1/*
func TimeAfter(date time.Time, date1 time.Time) bool {
	return date.After(date1)
}

// TimeAfterEqual date 大于等于 date1/*
func TimeAfterEqual(date time.Time, date1 time.Time) bool {
	return date.After(date1) || date.Equal(date1)
}

// TimeBeforeEqual date 小于等于 date1/*
func TimeBeforeEqual(date time.Time, date1 time.Time) bool {
	return date.Before(date1) || date.Equal(date1)
}

// TimeBefore date 小于 date1/*
func TimeBefore(date time.Time, date1 time.Time) bool {
	return date.Before(date1)
}

// TimeEqual date 等于 date1/*
func TimeEqual(date time.Time, date1 time.Time) bool {
	return date.Equal(date1)
}

// DateNow 当前时间字符串/*
func DateNow() string {
	return time.Now().String()
}

// ExecuteSql 执行sql并将结果存入参数内/*
func ExecuteSql(arguments map[string]interface{}, sql string, val ...interface{}) {
	var m = make(map[string]interface{})
	repository.DB.Raw(sql, val).Scan(&m)
	for k, v := range m {
		arguments[k] = v
	}
}

// ConnectDataBase 执行sql并将结果存入参数内/*
func ConnectDataBase(arguments *sync.Map, dbType string, username string, password string, host string, port, database string, sql string, val ...interface{}) interface{} {
	var dsn string
	var db *gorm.DB
	switch dbType {
	case "mysql":
		fallthrough
	default:
		dsn = strings.Join([]string{username, ":", password, "@tcp(", host, ":", port, ")/", database, "?charset=utf8mb4&parseTime=true"}, "")
		db = getMysqlDB(dsn)
	case "oracle":
		dsn = strings.Join([]string{username, "/", password, "@ip(", host, ":", port, ")/", database}, "")
		db = getOracleDB(dsn)
	case "dameng":
		options := map[string]string{
			"schema":         database,
			"appName":        database,
			"connectTimeout": "30000",
		}
		portI, _ := strconv.Atoi(port)
		// 构建达梦连接URL
		dns := dameng.BuildUrl(username, password, host, portI, options)
		db = getDamengDB(dns)
	}
	var m []map[string]interface{}
	// 替换自定义的sql语句
	for strings.Contains(sql, "###") {
		sql = strings.ReplaceAll(sql, "###", val[len(val)-1].(string))
		val = val[:len(val)-1]
	}
	// 验证sql类型
	upperSql := strings.ToUpper(sql)
	if strings.HasPrefix(upperSql, "INSERT ") ||
		strings.HasPrefix(upperSql, "UPDATE ") ||
		strings.HasPrefix(upperSql, "DELETE ") {
		db.Debug().Exec(sql, val...).Scan(&m)
	} else {
		db.Debug().Raw(sql, val...).Scan(&m)
	}
	for _, s := range m {
		for k, v := range s {
			arguments.Store(k, v)
		}
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	return m
}

// ConnectDataExecuteSql 获取数据库并执行sql /*
func ConnectDataExecuteSql(executeId string, arguments *sync.Map, dataKey string, sql string, resultFormatJson string, result bool, val ...interface{}) {
	var db *gorm.DB
	dbVal, _ := repository.DataMap.Load(dataKey)
	if dbVal == nil {
		// 尝试重新链接数据库获取db
		for _, data := range config.Conf.Data {
			if data.Title == dataKey {
				repository.ConnectDataDB(data)
			}
		}
		dbVal, _ = repository.DataMap.Load(dataKey)
		if dbVal == nil {
			panic("没有链接到数据库[" + dataKey + "]")
		}
	}
	db = dbVal.(*gorm.DB)
	var m []map[string]interface{}
	db.Debug().Raw(sql, val...).Scan(&m)

	for i, m2 := range m {
		m[i] = util.MapJsonValToAny(m2)
	}
	if len(resultFormatJson) > 0 {
		resultFormatJson = strings.ReplaceAll(resultFormatJson, "\\", "")
		resultFormatJson = util.ChatSyhJson(resultFormatJson)
		// 格式化结果
		var executeSQLResults []util.ExecuteMappingResult
		err := json.Unmarshal([]byte(resultFormatJson), &executeSQLResults)
		if err != nil {
			panic("返回格式解析失败! resultFormatJson:" + resultFormatJson)
		}
		data, err := json.Marshal(m)
		if len(m) == 1 {
			data, err = json.Marshal(m[0])
		}
		for _, f := range executeSQLResults {
			// 是否选中此参数作为过程参数
			if f.Selected {
				if data == nil {
					continue
				}
				v := gjson.Get(string(data), f.Path)
				var value interface{}
				if v.Exists() && len(v.String()) > 0 {
					value = v.String()
				} else if len(f.Path) == 0 && len(util.ToStr(f.Def)) == 0 {
					// 如果没有配置路径和默认值，则取所有返回结果
					//if len(m) == 1 {
					//	value = m[0]
					//} else {
					value = m
					//}
				} else {
					value = f.Def
				}
				if value == nil {
					value = ""
				}
				arguments.Store(f.Title, value)
				// 是否是返回值
				if result {
					AppendRuleResult(executeId, f.Title, value)
				}

			}
		}
	}
	//else {
	//	if arguments != nil {
	//		for _, s := range m {
	//			for k, v := range s {
	//				arguments[k] = v
	//			}
	//		}
	//	}
	//}
}

// ConnectDataPreviewSql 获取数据库并预览sql执行 /*
func ConnectDataPreviewSql(dataKey string, sql string, val ...interface{}) (interface{}, error) {
	var db *gorm.DB
	dbVal, _ := repository.DataMap.Load(dataKey)
	if dbVal == nil {
		// 尝试重新链接数据库获取db
		for _, data := range config.Conf.Data {
			if data.Title == dataKey {
				repository.ConnectDataDB(data)
			}
		}
		dbVal, _ = repository.DataMap.Load(dataKey)
		if dbVal == nil {
			return nil, errors.New("没有链接到数据库[" + dataKey + "]")
		}
	}
	db = dbVal.(*gorm.DB)
	var m []map[string]interface{}

	err := db.Debug().Raw(sql, val...).Scan(&m).Error
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}
	if len(m) == 1 {
		return util.MapJsonValToAny(m[0]), nil
	}
	for i, m2 := range m {
		m[i] = util.MapJsonValToAny(m2)
	}
	return m, nil
}

// getMysqlDB 获取mysql链接/*
func getMysqlDB(connString string) *gorm.DB {
	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       connString, // DSN data source name
		DefaultStringSize:         256,        // string 类型字段的默认长度
		DisableDatetimePrecision:  true,       // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,       // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,       // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,      // 根据版本自动配置
	}), &gorm.Config{
		Logger: ormLogger,
		//SkipDefaultTransaction: true,  // 关闭自动事务 性能提升30%左右
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(config.Conf.Mysql.MaxIdleConn) //设置连接池，空闲
	sqlDB.SetMaxOpenConns(config.Conf.Mysql.MaxOpenConn) //打开
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(config.Conf.Mysql.ConnMaxLifetime))
	return db
}

// getMysqlDB 获取dameng链接/*
func getDamengDB(connString string) *gorm.DB {
	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}
	db, err := gorm.Open(dameng.Open(connString), &gorm.Config{
		Logger: ormLogger, // DSN data source name
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(config.Conf.Mysql.MaxIdleConn) //设置连接池，空闲
	sqlDB.SetMaxOpenConns(config.Conf.Mysql.MaxOpenConn) //打开
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(config.Conf.Mysql.ConnMaxLifetime))
	return db
}

func getOracleDB(connString string) *gorm.DB {
	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}
	db, err := gorm.Open(oracle.Open(connString), &gorm.Config{
		Logger: ormLogger,
		//SkipDefaultTransaction: true,  // 关闭自动事务 性能提升30%左右
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(20) //设置连接池，空闲
	sqlDB.SetMaxOpenConns(50) //打开
	sqlDB.SetConnMaxLifetime(time.Second * 300)
	return db
}

// ExecuteSqlResult 执行sql将结果返回/*
func ExecuteSqlResult(sql string, val ...interface{}) interface{} {
	var m = make(map[string]interface{})
	repository.DB.Debug().Raw(sql, val).Scan(&m)
	return m
}

// XGet get请求
func XGet(url string, params map[string]string) ([]byte, error) {
	resp, err := http.Get(ConvertToQueryParams(url, params))
	return ResponseHandle(resp, err)
}

// XPost post请求
func XPost(url string, params map[string]string, body string) ([]byte, error) {
	resp, err := http.Post(ConvertToQueryParams(url, params), "application/json", strings.NewReader(body))
	return ResponseHandle(resp, err)
}

// XSend http请求
func XSend(executeId string, arguments *sync.Map, system string, method string, url string, queryJson string, body string, isResult bool, headerJson string, resultJson string) {
	var systemInfo config.Http
	h, _ := repository.SystemMap.Load(system)
	systemInfo = h.(config.Http)
	body = strings.ReplaceAll(body, "\\", "")
	argumentsMap := SyncMapToMap(arguments)
	// 处理body
	if len(body) > 0 {
		body = util.FormatHasArgumentsJson(body)
		bodyMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(body), &bodyMap)
		if err != nil {
			panic(err.Error())
		}
		replaceBody := make(map[string]interface{})
		for k, v := range bodyMap {
			if strings.HasPrefix(k, util.ArgumentsDefiner) {
				k = util.HasArguments(argumentsMap, util.StrVal(util.ArgumentsGet(arguments, k[1:])))
			}
			if strings.HasPrefix(util.StrVal(v), util.ArgumentsDefiner) {
				v = util.ArgumentsGet(arguments, util.StrVal(v)[1:])
			}
			replaceBody[k] = v
		}
		bodyByte, _ := json.Marshal(replaceBody)
		body = string(bodyByte)
	}
	url = formatPathParam(argumentsMap, url)
	req, err := http.NewRequest(method, systemInfo.Host+url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	if err != nil {
		panic(err.Error())
	}
	// 处理query参数
	if len(queryJson) > 0 {
		var query []util.ExecuteHttpParams
		queryJson = strings.ReplaceAll(queryJson, "\\", "")
		queryJson = util.ChatSyhJson(queryJson)
		err = json.Unmarshal([]byte(queryJson), &query)
		reqQuery := req.URL.Query()
		for _, q := range query {
			if q.Selected {
				if strings.HasPrefix(q.Key, util.ArgumentsDefiner) {
					q.Key = util.HasArguments(argumentsMap, q.Key[1:])
				}
				if strings.HasPrefix(util.ToStr(q.Value), util.ArgumentsDefiner) {
					q.Value = util.HasArguments(argumentsMap, util.ToStr(q.Value)[1:])
				}
				reqQuery.Add(q.Key, util.ToStr(q.Value))
			}
		}
		req.URL.RawQuery = reqQuery.Encode()
	}
	// 处理header参数
	if len(headerJson) > 0 {
		var header []util.ExecuteHttpParams
		headerJson = strings.ReplaceAll(headerJson, "\\", "")
		headerJson = util.ChatSyhJson(headerJson)
		err = json.Unmarshal([]byte(headerJson), &header)
		for _, hp := range header {
			if hp.Selected {
				if strings.HasPrefix(hp.Key, util.ArgumentsDefiner) {
					hp.Key = util.HasArguments(argumentsMap, hp.Key[1:])
				}
				if strings.HasPrefix(util.ToStr(hp.Value), util.ArgumentsDefiner) {
					hp.Value = util.HasArguments(argumentsMap, util.ToStr(hp.Value)[1:])
				}
				req.Header.Add(hp.Key, util.ToStr(hp.Value))
			}
		}
	}
	if systemInfo.Oauth.Enable {
		adapter.GetOauthType(systemInfo.Oauth.Type).GetToken(systemInfo, req)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err.Error())
	}
	data, err := ResponseHandle(resp, err)
	var variableMap interface{}
	j := jsoniter.Config{
		UseNumber: true,
	}.Froze()
	err = j.Unmarshal(data, &variableMap)
	// 处理返回结果
	if len(resultJson) > 0 {
		var result []util.ExecuteMappingResult
		resultJson = strings.ReplaceAll(resultJson, "\\", "")
		err = json.Unmarshal([]byte(resultJson), &result)
		if err != nil {
			panic(err.Error())
		}
		for _, er := range result {
			if er.Selected {
				rc := gjson.Get(string(data), er.Path)
				var value string
				if rc.Exists() && len(rc.String()) > 0 {
					value = rc.String()
				} else {
					value = util.ToStr(er.Def)
				}
				util.ArgumentsSet(arguments, er.Title, value)
				// 是否是返回值
				if isResult {
					AppendRuleResult(executeId, er.Title, value)
				}
			}
		}

	}
}

// HandleJson 处理json数据
func HandleJson(jsonArguments string, resultJson string, setResultJson string, arguments *sync.Map, executeId string, isResult bool, status *sync.Map) {
	var argumentName string
	argumentsMap := SyncMapToMap(arguments)
	// 判断是传入的变量还是json串
	if strings.HasPrefix(jsonArguments, util.ArgumentsDefiner) {
		argumentName = jsonArguments[1:]
		jsonArguments = util.HasArguments(argumentsMap, jsonArguments[1:])
	} else {
		jsonArguments = strings.ReplaceAll(jsonArguments, "\\", "")
	}
	// 验证是否是json格式字符串
	//if !gjson.Valid(jsonArguments) {
	//	status.Store("err", errors.New("不是json格式:{"+jsonArguments+"}"))
	//	SetRuleResult(executeId, "errorMsg", "不是json格式:{"+jsonArguments+"}")
	//	return
	//}
	if len(resultJson) > 0 {
		var result []util.ExecuteMappingResult
		resultJson = strings.ReplaceAll(resultJson, "\\", "")
		resultJson = util.ChatSyhJson(resultJson)
		json.Unmarshal([]byte(resultJson), &result)
		for _, mappingResult := range result {
			if !mappingResult.Selected {
				continue
			}
			var val string
			if gjson.Get(jsonArguments, mappingResult.Path).Exists() {
				val = gjson.Get(jsonArguments, mappingResult.Path).String()
			} else {
				val = util.ToStr(mappingResult.Def)
			}
			arguments.Store(mappingResult.Title, val)
			// 是否是返回值
			if isResult {
				AppendRuleResult(executeId, mappingResult.Title, val)
			}
		}
	}
	if len(setResultJson) > 0 {
		var setResult []util.ExecuteSetResult
		setResultJson = strings.ReplaceAll(setResultJson, "\\", "")
		json.Unmarshal([]byte(setResultJson), &setResult)
		for _, result := range setResult {
			if !result.Selected {
				continue
			}
			if len(argumentName) > 0 {
				if strings.HasPrefix(result.Key, util.ArgumentsDefiner) {
					result.Key = util.HasArguments(argumentsMap, result.Key[1:])
				}
				if !gjson.Valid(result.Value.(string)) {
					if strings.HasPrefix(result.Value.(string), util.ArgumentsDefiner) {
						result.Value = util.HasArguments(argumentsMap, result.Value.(string)[1:])
					}
				}
				jsonArguments, _ = sjson.Set(jsonArguments, result.Key, result.Value)
				arguments.Store(argumentName, jsonArguments)
			}
		}
	}

}

//ResponseHandle 也可以返回成string，改写最后的return为string(b)
func ResponseHandle(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		return []byte(""), err
	}
	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return []byte(""), err
	}
	return b, nil
}

//ConvertToQueryParams 依旧单独分出来Url拼接的部分/*
func ConvertToQueryParams(requestUrl string, params map[string]string) string {
	data := url.Values{}
	for key, value := range params {
		data.Set(key, value)
	}
	u, _ := url.ParseRequestURI(requestUrl)
	u.RawQuery = data.Encode()
	return u.String()
}
