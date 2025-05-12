package repository

import "sync"

var (
	//ProcessRecordMap = make(map[string][]Process)
	//DebugRunnerMap   = make(map[string]DebugRunner)
	//ErrorInfoMap     = make(map[string]ErrorInfoRunner)
	ProcessRecordMap = sync.Map{}
	DebugRunnerMap   = sync.Map{}
	ErrorInfoMap     = sync.Map{}
)

type Process struct {
	Name           string                 `json:"name"`
	Code           string                 `json:"code"`
	Type           string                 `json:"type"`
	ExecuteTime    string                 `json:"executeTime"`
	InputArguments map[string]interface{} `json:"input"`
	OutArguments   map[string]interface{} `json:"output"`
	Status         string                 `json:"status"`
	Error          interface{}            `json:"error"`
}

type DebugRunner struct {
	ExecuteId     string                   `json:"executeId"`
	Code          string                   `json:"code"`
	NextCode      string                   `json:"nextCode"`
	Arguments     map[string]interface{}   `json:"arguments"`
	RuleResultMap map[string][]interface{} `json:"ruleResultMap"`
}

type ErrorInfoRunner struct {
	ExecuteId     string                   `json:"executeId"`
	Code          string                   `json:"code"`
	NextCode      string                   `json:"nextCode"`
	Arguments     map[string]interface{}   `json:"arguments"`
	RuleResultMap map[string][]interface{} `json:"ruleResultMap"`
	ErrorMsg      string                   `json:"errorMsg"`
	Id            uint64                   `json:"id"`
	RuleCode      string                   `json:"ruleCode"`
	Version       string                   `json:"version"`
	Uri           string                   `json:"uri"`
	BusinessId    string                   `json:"businessId"`
}

func (*Process) GetProcessRecord(executeId string) []Process {
	value, b := ProcessRecordMap.Load(executeId)
	if !b {
		return []Process{}
	}
	return value.([]Process)
}

func (*Process) SetProcessRecord(executeId string, p []Process) {
	ProcessRecordMap.Store(executeId, p)
}

func (*Process) DelProcessRecord(executeId string) {
	ProcessRecordMap.Delete(executeId)
}

func (*Process) SetDebugRunner(debugRunner DebugRunner) {
	DebugRunnerMap.Store(debugRunner.ExecuteId, debugRunner)
}

func (*Process) GetDebugRunner(executeId string) DebugRunner {
	value, b := DebugRunnerMap.Load(executeId)
	if !b {
		return DebugRunner{}
	}
	return value.(DebugRunner)
}

func (*Process) DelDebugRunner(executeId string) {
	DebugRunnerMap.Delete(executeId)
}

func (*Process) SetErrorInfo(errorInfo ErrorInfoRunner) {
	ErrorInfoMap.Store(errorInfo.ExecuteId, errorInfo)
}

func (*Process) GetErrorInfo(executeId string) ErrorInfoRunner {
	value, b := ErrorInfoMap.Load(executeId)
	if !b {
		return ErrorInfoRunner{}
	}
	return value.(ErrorInfoRunner)
}

func (*Process) DelErrorInfo(executeId string) {
	ErrorInfoMap.Delete(executeId)
}
