package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/martian/log"
	"github.com/segmentio/ksuid"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/internal/adapter"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

var (
	// DebugNodeQueue executeId: map[int]util.queue
	DebugNodeQueue = sync.Map{}
)

type DebugExecuteParam struct {
	Code                string                 `json:"code" binding:"required"`
	Version             string                 `json:"version"`
	Arguments           map[string]interface{} `json:"arguments"`
	ExecuteArguments    map[string]interface{} `json:"executeArguments"`
	BusinessId          string                 `json:"businessId"`
	BreakPoint          []string               `json:"breakPoint"`
	BreakPointArguments map[string]string      `json:"breakPointArguments"`
	ExecuteId           string                 `json:"executeId"`
	Mode                int                    `json:"mode"` //0.执行到下一个断点 1.执行到下一步 2.恢复程序运行,执行到结束节点 3.终止
}

type ExecuteDebugRuleResult struct {
	Id               uint64                   `json:"id,omitempty"`
	Uri              string                   `json:"uri,omitempty"`
	Code             string                   `json:"code,omitempty"`
	Version          string                   `json:"version,omitempty"`
	BusinessId       string                   `json:"businessId,omitempty"`
	ExecuteId        string                   `json:"executeId,omitempty"`
	ExecuteTime      string                   `json:"executeTime,omitempty"`
	StartExecuteTime int64                    `json:"startExecuteTime,omitempty"`
	EndExecuteTime   int64                    `json:"endExecuteTime,omitempty"`
	Arguments        map[string]interface{}   `json:"arguments,omitempty"`
	Reference        map[string]string        `json:"reference,omitempty"`
	Tracks           []repository.Process     `json:"tracks,omitempty"`
	Result           map[string][]interface{} `json:"result,omitempty"`
	Done             bool                     `json:"done"`
}

// DeBugExecuteRule debug执行规则
func DeBugExecuteRule(ctx *gin.Context) {
	var ep DebugExecuteParam

	err := ctx.BindJSON(&ep)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	epJson, _ := json.Marshal(ep)
	log.Infof("通过CODE执行规则Debug  param:{}", string(epJson))
	// 获取开始节点
	ruleDesign, _ := adapter.GetStorage().GetPublishDesignByCode(ep.Code, ep.Version, util.DEV)
	yamlCode := util.ParseYamlCode(ruleDesign.Code, ruleDesign.Version)
	var executeId string
	startExecuteTime := time.Now()
	// 处理断点code
	for i, s := range ep.BreakPoint {
		ep.BreakPoint[i] = util.ParseRuleCode(yamlCode, s)
	}
	var isDone bool
	// 验证是否有执行id 没有执行id代表第一次debug
	if len(ep.ExecuteId) == 0 {
		pool := GetPool(util.ParseRuleCode(yamlCode, "start"))
		err = loadDesign(pool, yamlCode, ruleDesign.Design)
		if err != nil {
			internal.Fail(ctx, nil, err.Error())
		}
		data := make(map[string]interface{})
		// 生成本地执行规则唯一id
		executeId = ksuid.New().String()
		data["executeId"] = executeId
		data["debugParam"] = ep
		data["arguments"] = MapToSyncMap(ep.Arguments)
		data["inputArguments"] = MapToSyncMap(ep.Arguments)

		data["code"] = ep.Code
		data["result"] = make(map[string]interface{})
		repository.ProcessRecordMap.Store(executeId, []repository.Process{})
		RuleResultMap.Store(executeId, &sync.Map{})
		var n int
		// 执行开始规则
		queueMap := make(map[int]util.Queue)
		m, _ := ExecuteEnginePoolDebug("", "", "", util.ParseRuleCode(yamlCode, "start"), MapToSyncMap(data), executeId, ep)
		DebugNodeQueue.Store(executeId, queueMap)
		queueMap[n] = DebugNextQueuePush(m)
		isDone = DebugNext2(n, queueMap, ep)
	} else {
		// 继续执行本次执行id规则
		executeId = ep.ExecuteId
		// 获取下一步规则code
		//debugRunner := adapter.GetData().GetDebugRunner(executeId)
		//if reflect.DeepEqual(debugRunner, repository.DebugRunner{}) {
		//	internal.Fail(ctx, nil, "没有找到此executeId debug记录")
		//	return
		//}
		//pool := GetPool(debugRunner.NextCode)
		//err = loadDesign(pool, yamlCode, ruleDesign.Design)
		if err != nil {
			internal.Fail(ctx, nil, err.Error())
		}
		//RuleResultMap.Store(executeId, debugRunner.RuleResultMap)
		// 继续执行上一次断点到的规则
		queueMap, err := DebugQueueLoad(executeId)
		if err != nil {
			internal.Fail(ctx, nil, err.Error())
			return
		}
		isDone = DebugNext2(len(queueMap)-1, queueMap, ep)
	}

	// 结束规则执行计时
	endExecuteTime := time.Now()
	// 如果已经走到结束节点删除debug相关信息
	processList := adapter.GetData().GetProcessRecord(executeId)
	value, _ := RuleResultMap.Load(executeId)
	// 拼装返回
	executeDebugRuleResult := ExecuteDebugRuleResult{
		Id:               ruleDesign.Id,
		Code:             ep.Code,
		Uri:              ruleDesign.Uri,
		Version:          ruleDesign.Version,
		BusinessId:       ep.BusinessId,
		ExecuteId:        executeId,
		Arguments:        ep.Arguments,
		Reference:        reference(ep.Arguments),
		Tracks:           processList,
		Result:           ParseResultMap(value.(*sync.Map)),
		StartExecuteTime: startExecuteTime.UnixMicro(),
		EndExecuteTime:   endExecuteTime.UnixMicro(),
		ExecuteTime:      FormatSubtle(startExecuteTime),
		Done:             isDone,
	}
	if isDone || ep.Mode == 3 {
		// 执行完成后或终止执行后删掉内存中数据
		defer ExecuteEngineEnd(executeId, true)
	}
	internal.Success(ctx, executeDebugRuleResult, e.GetMsg(e.SUCCESS))
}

// ExecuteEnginePoolDebug Debug模式执行下一步编码规则
func ExecuteEnginePoolDebug(name string, t string, code string, nextCode string, arguments *sync.Map, executeId string, debug DebugExecuteParam) (map[string]interface{}, error) {
	pool := GetPool(nextCode)
	if pool == nil {
		return nil, errors.New("规则执行失败")
	}
	data := make(map[string]interface{})
	data["startTime"] = time.Now()
	data["arguments"] = arguments
	data["inputArguments"] = CopySyncMapVal(arguments)
	data["executeId"] = executeId
	data["debugParam"] = debug
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
			RuleResultMap: value.(map[string][]interface{}),
			ErrorMsg:      err.Error(),
		})
		SetRuleResult(executeId, "errorMsg", err.Error())
		return nil, nil
	}
	return m, nil
}

// DebugNext debug下一步执行处理
//func DebugNext(m map[string]interface{}) (map[string]interface{}, bool) {
//	var nextCodes = make([]string, 0)
//	for _, i := range m {
//		me := i.(map[string]ExecuteEngineNext)
//		for _, m2 := range me {
//			nextCodes = append(nextCodes, m2.nextCode)
//			DebugQueueStore(m2.executeId, nextCodes)
//			if BreakPointIn(m2.code, m2.nextCode, m2.executeId, m2.arguments, m2.debug.Mode, m2.debug.BreakPoint, m2.debug.BreakPointArguments) {
//				return nil, false
//			}
//			m, _ := ExecuteEnginePoolDebug(m2.name, m2.t, m2.code, m2.nextCode, m2.arguments, m2.executeId, m2.debug)
//			return m, true
//		}
//	}
//	return nil, false
//}

//DebugNextQueuePush 转为debug信息存储到队列中
func DebugNextQueuePush(m map[string]interface{}) util.Queue {
	var queue util.Queue
	for _, i := range m {
		me := i.(map[string]ExecuteEngineNext)
		keys := NextMapKeyToSlicesSort(me)
		for _, next := range keys {
			queue.Push(me[next])
		}
	}
	return queue
}

// NextMapKeyToSlicesSort map key转切片并排序
func NextMapKeyToSlicesSort(m map[string]ExecuteEngineNext) (keys []string) {
	for s := range m {
		keys = append(keys, s)
	}
	sort.Strings(keys)
	return keys
}

//DebugNext2 debug执行下一步
func DebugNext2(n int, m map[int]util.Queue, ep DebugExecuteParam) (isEnd bool) {
	if len(m) == 0 {
		isEnd = true
	}
	var b = false
	for len(m) > 0 {
		queue := m[n]
		if queue.QueueIsEmpty() {
			delete(m, n)
			n--
			// 退队列后自动进入下一步
			b = false
			continue
		} else {
			v := queue.First()
			een := v.(ExecuteEngineNext)
			// 替换执行中变量
			if len(ep.ExecuteArguments) > 0 {
				for key, val := range ep.ExecuteArguments {
					een.arguments.Store(key, val)
				}
			}
			// 验证是否满足停止 断点验证
			if b && BreakPointIn(een.code, een.arguments, ep) {
				DebugNodeQueue.Store(een.executeId, m)
				break
			}
			er, _ := ExecuteEnginePoolDebug("", "", een.code, een.nextCode, een.arguments, een.executeId, ep)
			queue.Pop()
			// 更改map中的队列
			m[n] = queue
			n++
			m[n] = DebugNextQueuePush(er)
		}
		b = true
	}
	return isEnd
}

// DebugQueueStore 存储后续执行节点至队列
func DebugQueueStore(executeId string, nextCodes []string) {
	var queue util.Queue
	q, b := DebugNodeQueue.Load(executeId)
	if b {
		queue = q.(util.Queue)
	}
	for _, nextCode := range nextCodes {
		queue.Push(nextCode)
	}
	DebugNodeQueue.Store(executeId, queue)
}

// DebugQueueLoad 获取此executeId debug队列
func DebugQueueLoad(executeId string) (map[int]util.Queue, error) {
	v, err := DebugNodeQueue.Load(executeId)
	if !err {
		return nil, errors.New("debug失败 没有获取到执行Id:" + executeId)
	}
	return v.(map[int]util.Queue), nil
}

// BreakPointIn 断点内是否包含此节点 断点参数是否满足/*
func BreakPointIn(code string, arguments *sync.Map, ep DebugExecuteParam) bool {
	var b bool
	// 验证断点条件
	if BreakPointArguments(code, arguments, ep) {
		// 不满足断点条件
		return false
	}
	//if len(ep.BreakPointArguments) > 0 { // 验证是否有断点参数,如果当前参数不满足断点参数则不中断规则
	//	b = true
	//	wg := sync.WaitGroup{}
	//	wg.Add(len(ep.BreakPointArguments))
	//	go func() {
	//		for k, v := range ep.BreakPointArguments {
	//			if ep.Arguments[k] == nil || ep.Arguments[k] != v {
	//				b = false
	//			}
	//			wg.Done()
	//		}
	//	}()
	//	wg.Wait()
	//	if !b {
	//		return false // debug参数不满足 直接不进入debug
	//	}
	//}
	switch ep.Mode { // 判断debug模式
	case 0: // 执行到下一个断点
		fallthrough
	default:
		for _, element := range ep.BreakPoint {
			if strings.EqualFold(code, element) {
				return true
			}
		}
		b = false
	case 1: // 执行到下一步
		//// 取出关系此规则执行图
		//v, _ := util.NodeMap.Load(util.ParseYamlCode(ep.Code, ep.Version))
		//node := v.(*util.Node)
		//// 获取当前node
		//for _, child := range node.Children {
		//	if child.Code == code {
		//		node = child
		//		break
		//	}
		//}
		//// 验证下一步是否是聚合节点 并且是否满足聚合节点,如果不满足直接跳过继续执行下一步
		//if node.Together {
		//	if !TogetherCheck2(code, node.TogetherWeight, ep.ExecuteId) {
		//		b = false
		//	}
		//} else {
		//	b = true
		//}
		b = true
	case 2: // 恢复程序运行,执行到结束节点
		b = false
	case 3: // 终止执行
		b = true
	}
	return b
}

// BreakPointArguments 断点参数验证
func BreakPointArguments(code string, arguments *sync.Map, ep DebugExecuteParam) bool {
	expression := ep.BreakPointArguments[strings.ReplaceAll(code, util.ParseYamlCode(ep.Code, ep.Version)+":", "")]
	if len(expression) == 0 {
		// 没有设置断点条件
		return false
	}
	expression = strings.ReplaceAll(expression, " ", "")
	if strings.Contains(expression, util.ArgumentsDefiner) {
		char := util.CharAt(expression)
		for i, _ := range char {
			if strings.HasPrefix(char[i], util.ArgumentsDefiner) {
				argument := strings.ReplaceAll(char[i], util.ArgumentsDefiner, "")
				val, _ := arguments.Load(argument)
				// 在变量池中没有找到此变量
				if val == nil {
					continue
				}
				switch val.(type) {
				case string:
					char[i] = val.(string)
				case map[string]interface{}:
					valMap := val.(map[string]interface{})
					char[i] = valMap["value"].(string)
				case int:
					char[i] = strconv.Itoa(val.(int))
				case int64:
					char[i] = strconv.FormatInt(val.(int64), 10)
				case float64:
					char[i] = strconv.FormatFloat(val.(float64), 'f', 10, 64)
				default:
					char[i] = val.(string)
				}
			}
		}
		var str string
		for _, s := range char {
			str += s
		}
		expression = str
	}
	return !util.IF2(expression, true, false).(bool)
}
