package handler

import (
	"encoding/json"
	"github.com/bilibili/gengine/engine"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
	"reflect"
	"strconv"
	"sync"
	"time"
	"zenith.engine.com/engine/config"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/internal/adapter"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

type ExecuteRuleParam struct {
	Id            uint64                      `json:"id"`
	Code          string                      `json:"code"`
	Version       string                      `json:"version"`
	Arguments     map[string]interface{}      `json:"arguments"`
	BusinessId    string                      `json:"businessId"`
	Variable      int64                       `json:"variable"`
	Result        []util.ExecuteMappingResult `json:"result"`
	CallArguments []util.ExecuteCallArguments `json:"callArguments"`
	Env           string                      `json:"env"`
}

type ExecuteRuleResult struct {
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
}

// ExecuteRuleBatchParallel 批量执行规则 并行
func ExecuteRuleBatchParallel(ctx *gin.Context) {
	var eps []ExecuteRuleParam
	err := ctx.BindJSON(&eps)
	util.Log.Println("通过CODE批量并发执行规则")
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	var executeRuleResultList []ExecuteRuleResult
	var responseChannel = make(chan sync.Map, len(eps))
	defer close(responseChannel)
	w := sync.WaitGroup{}
	for _, ep := range eps {
		w.Add(1)

		go func(ep ExecuteRuleParam) {
			responseChannel <- execute(ep)
			w.Done()
		}(ep)
	}
	w.Wait()
	for i := 0; i < len(eps); i++ {
		select {
		case val := <-responseChannel:
			// 处理结果
			result, _ := val.Load("executeRuleResult")
			executeRuleResult := result.(ExecuteRuleResult)
			executeRuleResultList = append(executeRuleResultList, executeRuleResult)

		}
	}
	internal.Success(ctx, executeRuleResultList, e.GetMsg(e.SUCCESS))
}

// ExecuteRuleBatch 批量执行规则 串行
func ExecuteRuleBatch(ctx *gin.Context) {
	var eps []ExecuteRuleParam
	err := ctx.BindJSON(&eps)
	util.Log.Println("通过CODE批量顺序执行规则")
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	var executeRuleResultList []ExecuteRuleResult
	for _, ep := range eps {
		resultMap := execute(ep)
		// 处理结果
		result, _ := resultMap.Load("executeRuleResult")
		executeRuleResult := result.(ExecuteRuleResult)
		executeRuleResultList = append(executeRuleResultList, executeRuleResult)
	}
	internal.Success(ctx, executeRuleResultList, e.GetMsg(e.SUCCESS))
}

// ExecuteRuleId 通过id执行规则
func ExecuteRuleId(ctx *gin.Context) {
	var ep ExecuteRuleParam
	err := ctx.BindJSON(&ep)
	util.Log.Println("通过ID执行规则  ID:{} ", ep.Id)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	design, err := adapter.GetStorage().GetById(ep.Id)
	ep.Code = design.Code
	ep.Version = design.Version
	// 从数据库查询
	ruleDesign, err := adapter.GetStorage().GetPublishDesignByCode(ep.Code, ep.Version, ep.Env)
	err, yamlCode, data, executeId, startExecuteTime, arguments := executeRuleDesign(ruleDesign, ep)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	// 重新加载最新规则在内存中保证每次都是最新规则
	err = loadDesign(nil, yamlCode, ruleDesign.Design)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	// 执行开始规则
	_, err = ExecuteEnginePool("", "", "", util.ParseRuleCode(yamlCode, "start"), MapToSyncMap(data), executeId)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	// 结束规则执行计时
	endExecuteTime := time.Now()
	value, _ := RuleResultMap.Load(executeId)
	// 获取返回值定义
	err, df := util.AnalysisResultJson(ruleDesign.DesignJson)
	// 清除此次执行全局变量
	defer ExecuteEngineEnd(executeId, false)
	setErrorInfo(executeId, ep.BusinessId, ruleDesign)
	internal.Success(ctx, ParseResult(df, ruleDesign, ep, executeId, ParseResultMap(value.(*sync.Map)), startExecuteTime, endExecuteTime, arguments), e.GetMsg(e.SUCCESS))
}

// ExecuteRule 执行单个规则
func ExecuteRule(ctx *gin.Context) {
	var ep ExecuteRuleParam
	err := ctx.BindJSON(&ep)
	util.Log.Println("通过CODE执行规则  code:{} version:{}", ep.Code, ep.Version)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	resultMap := execute(ep)
	resultErr, _ := resultMap.Load("error")
	if resultErr != nil {
		internal.Fail(ctx, nil, resultErr.(error).Error())
		return
	}
	result, _ := resultMap.Load("executeRuleResult")
	internal.Success(ctx, result, e.GetMsg(e.SUCCESS))
}

// execute 执行规则
func execute(ep ExecuteRuleParam) sync.Map {
	var result = sync.Map{}
	var ruleDesign repository.RuleDesign
	var err error
	// 先从内存查询
	v, b := repository.RuleDesignMap.Load(util.ParseYamlCode(ep.Code, ep.Version))
	if b && v != nil {
		ruleDesign = v.(repository.RuleDesign)
	} else {
		// 从数据库查询
		ruleDesign, err = adapter.GetStorage().GetPublishDesignByCode(ep.Code, ep.Version, ep.Env)
		if err != nil {
			result.Store("executeRuleResult", ExecuteRuleResult{})
			result.Store("error", err)
			return result
		}
	}
	err, yamlCode, data, executeId, startExecuteTime, arguments := executeRuleDesign(ruleDesign, ep)
	if err != nil {
		result.Store("executeRuleResult", ExecuteRuleResult{})
		result.Store("error", err)
		return result
	}
	// 执行开始规则
	_, err = ExecuteEnginePool("", "", "", util.ParseRuleCode(yamlCode, "start"), MapToSyncMap(data), executeId)
	// 结束规则执行计时
	endExecuteTime := time.Now()
	if err != nil {
		result.Store("executeRuleResult", ExecuteRuleResult{})
		result.Store("error", err)
		return result
	}
	value, _ := RuleResultMap.Load(executeId)
	// 获取返回值定义
	err, df := util.AnalysisResultJson(ruleDesign.DesignJson)
	// 清除此次执行全局变量
	defer ExecuteEngineEnd(executeId, false)
	setErrorInfo(executeId, ep.BusinessId, ruleDesign)
	result.Store("executeRuleResult", ParseResult(df, ruleDesign, ep, executeId, ParseResultMap(value.(*sync.Map)), startExecuteTime, endExecuteTime, arguments))
	result.Store("error", nil)
	return result
}

// RouterExecuteRule 使用路径uri执行规则
func RouterExecuteRule(ctx *gin.Context) {
	var ep ExecuteRuleParam
	err := ctx.ShouldBind(&ep)
	uri := ctx.FullPath()

	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	ruleDesign, _ := adapter.GetStorage().GetByUri(uri, ep.Version)
	// 解析yaml处理规则参数
	err, yamlCode, data, executeId, startExecuteTime, arguments := executeRuleDesign(ruleDesign, ep)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.RuleVariableError))
		return
	}
	// 执行开始规则
	_, err = ExecuteEnginePool("", "", "", util.ParseRuleCode(yamlCode, "start"), MapToSyncMap(data), executeId)
	// 结束规则执行计时
	endExecuteTime := time.Now()
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.RuleExecuteError))
		return
	}
	value, _ := RuleResultMap.Load(executeId)
	// 拼装返回
	executeRuleResult := ExecuteRuleResult{
		Id:               ruleDesign.Id,
		Code:             ruleDesign.Code,
		Uri:              ruleDesign.Uri,
		Version:          ruleDesign.Version,
		BusinessId:       ep.BusinessId,
		ExecuteId:        executeId,
		Arguments:        SyncMapToMap(arguments),
		Reference:        reference(ep.Arguments),
		Tracks:           adapter.GetData().GetProcessRecord(executeId),
		Result:           value.(map[string][]interface{}),
		StartExecuteTime: startExecuteTime.UnixMicro(),
		EndExecuteTime:   endExecuteTime.UnixMicro(),
		ExecuteTime:      FormatSubtle(startExecuteTime),
	}
	defer ExecuteEngineEnd(executeId, false)
	setErrorInfo(executeId, ep.BusinessId, ruleDesign)
	internal.Success(ctx, executeRuleResult, e.GetMsg(e.SUCCESS))

}

/** executeRuleDesign 解析规则,拼装执行参数*/
func executeRuleDesign(ruleDesign repository.RuleDesign, ep ExecuteRuleParam) (error, string, map[string]interface{}, string, time.Time, *sync.Map) {
	//获取yaml接口返回值
	yamlCode := util.ParseYamlCode(ruleDesign.Code, ruleDesign.Version)
	// 通过code获取pool
	pool := GetPool(util.ParseRuleCode(yamlCode, "start"))
	err := loadDesign(pool, yamlCode, ruleDesign.Design)
	if err != nil {
		return err, "", nil, "", time.Now(), nil
	}
	data := make(map[string]interface{})
	// 生成本地执行规则唯一id
	executeId := ksuid.New().String()
	ep.Code = ruleDesign.Code
	data["executeId"] = executeId
	// 初始化 arguments
	if ep.Arguments == nil {
		ep.Arguments = make(map[string]interface{})
	}
	// 添加全局参数
	if ep.Variable > 0 {
		variables, _ := variableAppend(ep.Arguments)
		data["inputArguments"] = MapToSyncMap(variables)
	} else {
		data["inputArguments"] = MapToSyncMap(ep.Arguments)
	}
	syncArguments := MapToSyncMap(ep.Arguments)
	data["arguments"] = syncArguments
	//if err != nil {
	//	return err, "", nil, "", time.Now()
	//}

	data["code"] = ep.Code
	data["result"] = make(map[string]interface{})
	repository.ProcessRecordMap.Store(executeId, []repository.Process{})
	RuleResultMap.Store(executeId, &sync.Map{})

	//开始规则执行计时
	startExecuteTime := time.Now()
	data["startTime"] = startExecuteTime

	// 将规则信息存入内存
	repository.RuleDesignMap.Store(yamlCode, ruleDesign)
	return nil, yamlCode, data, executeId, startExecuteTime, syncArguments
}

// variableAppend 验证是否有全局参数 添加
func variableAppend(arguments map[string]interface{}) (map[string]interface{}, error) {
	if len(config.Conf.Variable.ApiAll) > 0 {
		// 从配置的api里面查找
		body, _ := json.Marshal(arguments)
		data, err := XPost(config.Conf.Variable.ApiAll, nil, string(body))
		if err != nil {
			return nil, err
		}
		variableMap := make(map[string]map[string]interface{})
		err = json.Unmarshal(data, &variableMap)
		if err != nil {
			return nil, err
		}
		for k, v := range variableMap {
			arguments[k] = v
		}
	}
	return arguments, nil
}

/** loadDesign 如果规则池没有获取到,解析yaml并将规则存入规则池*/
func loadDesign(pool *engine.GenginePool, yamlCode string, design string) error {
	if pool == nil {
		// 解析yaml-Rule
		ruleStrSlice, err := util.ExplainYaml(yamlCode, design)
		if err != nil {
			return err
		}
		for k, v := range ruleStrSlice {
			SetPool(k, BuildEnginePool(v))
		}
	}
	return nil
}

type ErrorExecuteParam struct {
	ExecuteId string `json:"executeId" binding:"required"`
}

func RetryExecuteRule(ctx *gin.Context) {
	var ep ErrorExecuteParam
	err := ctx.BindJSON(&ep)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	errorInfoRunner := adapter.GetData().GetErrorInfo(ep.ExecuteId)
	if reflect.DeepEqual(errorInfoRunner, repository.ErrorInfoRunner{}) { //验证此执行id是否获取到错误信息
		internal.Fail(ctx, nil, "此执行id没有任何错误信息")
		return
	}
	// 将返回值重新放入resultMap
	RuleResultMap.Store(errorInfoRunner.ExecuteId, errorInfoRunner.RuleResultMap)
	startExecuteTime := time.Now()
	pool := GetPool(errorInfoRunner.NextCode)
	ruleDesign, err := adapter.GetStorage().GetPublishDesignByCode(errorInfoRunner.RuleCode, errorInfoRunner.Version, util.DEV)
	err = loadDesign(pool, util.ParseYamlCode(errorInfoRunner.RuleCode, errorInfoRunner.Version), ruleDesign.Design)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.RuleParseYamlError))
		return
	}
	// 重试,继续执行错误后的规则
	ExecuteEnginePool("", "", "", errorInfoRunner.NextCode, MapToSyncMap(errorInfoRunner.Arguments), errorInfoRunner.ExecuteId)
	endExecuteTime := time.Now()
	value, _ := RuleResultMap.Load(errorInfoRunner.ExecuteId)
	// 拼装返回
	executeRuleResult := ExecuteRuleResult{
		Id:               errorInfoRunner.Id,
		Code:             errorInfoRunner.RuleCode,
		Uri:              errorInfoRunner.Uri,
		Version:          errorInfoRunner.Version,
		BusinessId:       errorInfoRunner.BusinessId,
		ExecuteId:        errorInfoRunner.ExecuteId,
		Arguments:        errorInfoRunner.Arguments,
		Reference:        reference(errorInfoRunner.Arguments),
		Tracks:           adapter.GetData().GetProcessRecord(errorInfoRunner.ExecuteId),
		Result:           value.(map[string][]interface{}),
		StartExecuteTime: startExecuteTime.UnixMicro(),
		EndExecuteTime:   endExecuteTime.UnixMicro(),
		ExecuteTime:      FormatSubtle(startExecuteTime),
	}
	defer ExecuteEngineEnd(errorInfoRunner.ExecuteId, false)
	errorInfo := adapter.GetData().GetErrorInfo(errorInfoRunner.ExecuteId)
	if !reflect.DeepEqual(errorInfo, repository.ErrorInfoRunner{}) && len(errorInfo.ErrorMsg) > 0 { // 验证是否运行规则时出错过
		errorInfo.Id = errorInfoRunner.Id
		errorInfo.RuleCode = errorInfoRunner.RuleCode
		errorInfo.Version = errorInfoRunner.Version
		errorInfo.BusinessId = errorInfoRunner.BusinessId
		errorInfo.Uri = errorInfoRunner.Uri
		adapter.GetData().SetErrorInfo(errorInfo)
	} else {
		// 运行成功后将错误记录删除
		adapter.GetData().DelErrorInfo(errorInfoRunner.ExecuteId)
	}
	internal.Success(ctx, executeRuleResult, e.GetMsg(e.SUCCESS))
}

/** 设置规则错误信息*/
func setErrorInfo(executeId string, businessId string, design repository.RuleDesign) {
	errorInfo := adapter.GetData().GetErrorInfo(executeId)
	if !reflect.DeepEqual(errorInfo, repository.ErrorInfoRunner{}) { // 验证是否运行规则时出错过
		errorInfo.Id = design.Id
		errorInfo.RuleCode = design.Code
		errorInfo.Version = design.Version
		errorInfo.BusinessId = businessId
		errorInfo.Uri = design.Uri
		adapter.GetData().SetErrorInfo(errorInfo)
	}
}

/** 提取参数引用*/
func reference(arguments map[string]interface{}) map[string]string {
	referenceMap := make(map[string]string, len(arguments))
	var i int
	for k, _ := range arguments {
		i++
		referenceMap["ref_"+strconv.Itoa(i)] = k
	}
	return referenceMap
}

// compressTracks 压缩返回json方法
func compressTracks(arguments map[string]interface{}, process []repository.Process, reference map[string]string) {
	for i := 0; i < len(process); i++ {
		// 处理入参
		for k, v := range process[i].InputArguments {
			ak := arguments[k]
			if ak == nil {
				continue
			}
			if reflect.DeepEqual(v, ak) {
				for rk, rv := range reference {
					if reflect.DeepEqual(k, rv) {
						delete(process[i].InputArguments, k)
						process[i].InputArguments[rk] = 0
					}
				}
			}
		}

		// 处理出参
		for k, v := range process[i].OutArguments {
			ak := arguments[k]
			if ak == nil {
				continue
			}
			if reflect.DeepEqual(v, ak) {
				for rk, rv := range reference {
					if reflect.DeepEqual(k, rv) {
						delete(process[i].OutArguments, k)
						process[i].OutArguments[rk] = 0
					}
				}
			}
		}

	}
}

// ParseResult 处理返回值
func ParseResult(define util.ResultDefine, ruleDesign repository.RuleDesign, ep ExecuteRuleParam, executeId string, result map[string][]interface{}, startExecuteTime time.Time, endExecuteTime time.Time, arguments *sync.Map) ExecuteRuleResult {
	var executeRuleResult ExecuteRuleResult
	if reflect.DeepEqual(define, util.ResultDefine{}) {
		// 拼装默认返回
		executeRuleResult = ExecuteRuleResult{
			Id:               ruleDesign.Id,
			Code:             ep.Code,
			Version:          ruleDesign.Version,
			BusinessId:       ep.BusinessId,
			ExecuteId:        executeId,
			Arguments:        SyncMapToMap(arguments),
			Reference:        reference(ep.Arguments),
			Tracks:           adapter.GetData().GetProcessRecord(executeId),
			Result:           result,
			StartExecuteTime: startExecuteTime.UnixMicro(),
			EndExecuteTime:   endExecuteTime.UnixMicro(),
			ExecuteTime:      FormatSubtle(startExecuteTime),
		}
		// 压缩返回值
		compressTracks(executeRuleResult.Arguments, executeRuleResult.Tracks, executeRuleResult.Reference)
	} else {
		executeRuleResult.BusinessId = ep.BusinessId
		executeRuleResult.ExecuteId = executeId
		// 拼装自定义返回
		if define.BasicInfo {
			executeRuleResult.Id = ruleDesign.Id
			executeRuleResult.Code = ep.Code
			executeRuleResult.Version = ruleDesign.Version
		}
		if define.Time {
			executeRuleResult.ExecuteTime = FormatSubtle(startExecuteTime)
			executeRuleResult.StartExecuteTime = startExecuteTime.UnixMicro()
			executeRuleResult.EndExecuteTime = endExecuteTime.UnixMicro()
		}
		if define.Arguments {
			executeRuleResult.Arguments = SyncMapToMap(arguments)
		}
		if define.Tracks {
			executeRuleResult.Tracks = adapter.GetData().GetProcessRecord(executeId)
		}
		if define.Compress {
			// 压缩返回值
			executeRuleResult.Reference = reference(SyncMapToMap(arguments))
			compressTracks(executeRuleResult.Arguments, executeRuleResult.Tracks, executeRuleResult.Reference)
		}
		if define.Result {
			var resultMap = make(map[string][]interface{})
			resultAllMap := result
			if result["errorMsg"] != nil {
				resultMap["errorMsg"] = resultAllMap["errorMsg"]
			}
			for _, rv := range define.ResultArguments {
				// 验证map是否包含这个key值
				if _, ok := resultAllMap[rv]; ok {
					resultMap[rv] = resultAllMap[rv]
				}
			}
			executeRuleResult.Result = resultMap
		}
	}
	return executeRuleResult
}
