package e

const (
	SUCCESS       = 0
	ERROR         = 500
	InvalidParams = 400

	// 成员错误
	ErrorExistUser      = 10002
	ErrorNotExistUser   = 10003
	ErrorFailEncryption = 10006
	ErrorNotCompare     = 10007

	HaveSignUp           = 20001
	ErrorActivityTimeout = 20002

	ErrorAuthCheckTokenFail    = 30001 //token 错误
	ErrorAuthCheckTokenTimeout = 30002 //token 过期
	ErrorAuthToken             = 30003
	ErrorAuth                  = 30004
	ErrorAuthNotFound          = 30005
	ErrorDatabase              = 40001

	RuleCodeExist       = 10008
	RuleDesignSaveError = 10009
	RuleCodeNotExist    = 10010
	RuleInfoSaveError   = 10011
	RuleExecuteError    = 10012
	RuleNotExist        = 10013
	RuleTableError      = 10014
	RuleUriError        = 10015
	RuleParseYamlError  = 10016
	RuleVariableError   = 10017

	RuleGroupSaveError     = 10018
	RuleNotExistAndPublish = 10019

	GenPasswordError          = 10020
	PasswordError             = 10021
	PasswordInconsistentError = 10022
	JsonFormatError           = 10023
)
