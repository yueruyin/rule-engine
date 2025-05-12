package e

var MsgFlags = map[uint]string{
	SUCCESS:                    "ok",
	ERROR:                      "fail",
	InvalidParams:              "请求参数错误",
	HaveSignUp:                 "已经报名了",
	ErrorActivityTimeout:       "活动过期了",
	ErrorAuthCheckTokenFail:    "Token鉴权失败",
	ErrorAuthCheckTokenTimeout: "Token已超时",
	ErrorAuthToken:             "Token生成失败",
	ErrorAuth:                  "Token错误",
	ErrorNotCompare:            "不匹配",
	ErrorDatabase:              "数据库操作出错,请重试",
	ErrorAuthNotFound:          "Token不能为空",
	RuleCodeExist:              "规则编码已经存在",
	RuleDesignSaveError:        "规则设计保存失败",
	RuleCodeNotExist:           "规则编码不存在",
	RuleInfoSaveError:          "默认规则保存失败",
	RuleExecuteError:           "规则执行失败",
	RuleNotExist:               "规则不存在",
	RuleParseYamlError:         "解析规则失败",
	RuleUriError:               "规则uri已存在",
	RuleVariableError:          "全局参数获取失败",
	RuleGroupSaveError:         "创建规则组失败",
	RuleNotExistAndPublish:     "规则编码不存在或规则未发布请发布后重试!",
	GenPasswordError:           "生成密码失败",
	ErrorExistUser:             "用户不存在",
	PasswordError:              "密码错误",
	PasswordInconsistentError:  "密码不一致",
	JsonFormatError:            "请传入json格式字符串",
}

// GetMsg 获取状态码对应信息
func GetMsg(code uint) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[ERROR]
}
