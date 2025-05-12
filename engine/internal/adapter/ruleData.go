package adapter

import (
	"zenith.engine.com/engine/config"
	"zenith.engine.com/engine/internal/cache"
	"zenith.engine.com/engine/internal/repository"
)

type RuleData interface {
	GetProcessRecord(executeId string) []repository.Process    // get the rule execution process...
	SetProcessRecord(executeId string, p []repository.Process) // set process record in redis or memory
	DelProcessRecord(executeId string)                         // delete execution record after rule execution

	SetDebugRunner(debugRunner repository.DebugRunner)      // if the information is stored in redis or memory under the debug call
	GetDebugRunner(executeId string) repository.DebugRunner // get the relevant information parameters in the debug mode...
	DelDebugRunner(executeId string)                        // delete debug information

	SetErrorInfo(errorInfoRunner repository.ErrorInfoRunner)  //save information after execution failure
	GetErrorInfo(executeId string) repository.ErrorInfoRunner //get information after execution failure...
	DelErrorInfo(executeId string)                            // delete error information
}

func GetData() RuleData {
	var data RuleData
	switch config.Conf.Rule.Data {
	case "memory":
		data = &repository.Process{}
	case "redis":
		data = &cache.ProcessCache{}
	default:
	}
	return data
}
