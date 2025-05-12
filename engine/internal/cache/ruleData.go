package cache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v9"
	"time"
	"zenith.engine.com/engine/config"
	"zenith.engine.com/engine/internal/repository"
)

var rd *redis.Client

type ProcessCache struct {
}

const (
	debugSecond = 3600
	ruleKey     = "process:"
	debugKey    = "debug:"
	errorKey    = "error:"
)

func InitRedis() {
	rd = repository.InitRedis()
}

//func (*ProcessCache) GetProcessRecord(executeId string) []repository.Process {
//	v, _ := rd.LIndex(context.Background(), GetRuleKey(executeId), -1).Result()
//	var p = new([]repository.Process)
//	_ = json.Unmarshal([]byte(v), &p)
//	return *p
//}
//
//func (*ProcessCache) SetProcessRecord(executeId string, p []repository.Process) {
//	pJson, _ := json.Marshal(p)
//	if err := rd.RPush(context.Background(), GetRuleKey(executeId), pJson).Err(); err != nil {
//		panic(err)
//	}
//}
//
//func (*ProcessCache) DelProcessRecord(executeId string) {
//	rd.Del(context.Background(), GetRuleKey(executeId))
//}

func (*ProcessCache) GetProcessRecord(executeId string) []repository.Process {
	value, b := repository.ProcessRecordMap.Load(executeId)
	if !b {
		return []repository.Process{}
	}
	return value.([]repository.Process)
}

func (*ProcessCache) SetProcessRecord(executeId string, p []repository.Process) {
	repository.ProcessRecordMap.Store(executeId, p)
}

func (*ProcessCache) DelProcessRecord(executeId string) {
	repository.ProcessRecordMap.Delete(executeId)
}

func (*ProcessCache) SetDebugRunner(debugRunner repository.DebugRunner) {
	debugRunnerJson, _ := json.Marshal(debugRunner)
	second := config.Conf.Rule.DebugSecond
	if second == 0 {
		second = debugSecond
	}
	if err := rd.Set(context.Background(), GetDebugKey(debugRunner.ExecuteId), debugRunnerJson, time.Second*time.Duration(second)).Err(); err != nil {
		panic(err)
	}
}

func (*ProcessCache) GetDebugRunner(executeId string) repository.DebugRunner {
	v := rd.Get(context.Background(), GetDebugKey(executeId))
	debugRunner := repository.DebugRunner{}
	_ = json.Unmarshal([]byte(v.Val()), &debugRunner)
	return debugRunner
}

func (*ProcessCache) DelDebugRunner(executeId string) {
	rd.Del(context.Background(), GetDebugKey(executeId))
}

func (*ProcessCache) SetErrorInfo(errorInfo repository.ErrorInfoRunner) {
	errorInfoJson, _ := json.Marshal(errorInfo)
	if err := rd.Set(context.Background(), GetErrorKey(errorInfo.ExecuteId), errorInfoJson, time.Second*3600).Err(); err != nil {
		panic(err)
	}
}

func (*ProcessCache) GetErrorInfo(executeId string) repository.ErrorInfoRunner {
	v := rd.Get(context.Background(), GetErrorKey(executeId))
	errorInfoRunner := repository.ErrorInfoRunner{}
	_ = json.Unmarshal([]byte(v.Val()), &errorInfoRunner)
	return errorInfoRunner
}

func (*ProcessCache) DelErrorInfo(executeId string) {
	rd.Del(context.Background(), GetRuleKey(executeId))
}

func SetRuleStr(code string, ruleStr string) {
	if err := rd.RPush(context.Background(), GetRuleKey(code), ruleStr).Err(); err != nil {
		panic(err)
	}
}

func lSetRuleStr(code string, ruleStr []string) {
	if err := rd.RPush(context.Background(), GetRuleKey(code), ruleStr).Err(); err != nil {
		panic(err)
	}
}

func GetRuleStr(code string) []string {
	v, _ := rd.LRange(context.Background(), GetRuleKey(code), 0, rd.LLen(context.Background(), GetRuleKey(code)).Val()).Result()
	return v
}

func GetRuleKey(code string) string {
	return ruleKey + code
}

func GetDebugKey(code string) string {
	return debugKey + code
}

func GetErrorKey(code string) string {
	return errorKey + code
}
