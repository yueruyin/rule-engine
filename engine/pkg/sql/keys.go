package sql

import (
	"strings"
)

// 定义sql关键字
const (
	KeySelect = "select"

	KeyUpdate = "update"
	KeyDelete = "delete"
	KeyInsert = "insert"
	KeyInto   = "into"

	KeyCreate = "create"
	KeyAlert  = "alert"
	KeyDrop   = "drop"

	KeyGrant  = "grant"
	KeyRevoke = "revoke"

	KeyCommit   = "commit"
	KeyRollback = "rollback"

	KeyTruncate = "truncate"

	KeyTable     = "table"
	keyTables    = "tables"
	KeyPrimary   = "primary"
	KeyDatabases = "databases"
	KeyUse       = "use"

	KeyLeftBracket  = "("
	KeyRightBracket = ")"

	KeyBlank     = " "
	KeySemicolon = ";"
	KeyComma     = ","
	KeyTab       = "\t"
	KeyEnter     = "\n"
	KeyNewLine   = "\r"

	KeyGt     = ">"
	KeyLt     = "<"
	KeyGe     = ">="
	KeyLe     = "<="
	KeyEq     = "="
	KeyNotEq  = "<>"
	KeyNotEq2 = "!="

	KeyQuota        = "\""
	KeySingleQuota  = "'"
	KeySectionQuota = "`"
)

// forbidKeys 定义sql关键字, 不包含Insert, 单独处理避免跟INSERT函数冲突
var forbidKeys = []string{KeyUpdate, KeyDelete, KeyCreate, KeyAlert, KeyDrop, KeyGrant, KeyRevoke, KeyCommit, KeyRollback, KeyTruncate, KeyUse, KeyDatabases}

var splitChars = []string{KeyLeftBracket, KeyRightBracket, KeyBlank, KeySemicolon, KeyComma, KeyGt, KeyLt, KeyGe, KeyLe, KeyEq, KeyNotEq, KeyNotEq2, KeyTab, KeyEnter, KeyNewLine}

var quotaChars = []string{KeyQuota, KeySingleQuota, KeySectionQuota}

func containsFold(elems []string, elem string) bool {
	for _, temp := range elems {
		if strings.EqualFold(temp, elem) {
			return true
		}
	}
	return false
}

// IsForbidKey 判断是否被禁止使用的关键字
func IsForbidKey(key string) bool {
	return containsFold(forbidKeys, key)
}

// IsSplitChar 判断字符是否是分割符
func IsSplitChar(key string) bool {
	return containsFold(splitChars, key)
}

// IsQuotaChar 判断字符是否是分割符
func IsQuotaChar(key string) bool {
	return containsFold(quotaChars, key)
}
