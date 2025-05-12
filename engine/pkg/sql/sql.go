package sql

import (
	"errors"
	"strings"
)

// ValidateSql 检查sql是否被运行
func ValidateSql(sqlStr string) ([]string, error) {
	// 拆分sql
	list := SplitSql(sqlStr)
	length := len(list)
	// 汇总sql中的参数定义
	var parameters []string
	// 判断是否存在关键字
	for i, temp := range list {
		if IsForbidKey(temp) {
			return nil, errors.New("不允许使用【" + strings.ToUpper(temp) + "】关键字")
		}
		// 单独处理INSERT INTO
		if strings.EqualFold(temp, KeyInsert) && length != i+1 && strings.EqualFold(list[i+1], KeyInto) {
			return nil, errors.New("不允许使用【" + strings.ToUpper(temp) + " " + strings.ToUpper(KeyInto) + "】关键字")
		}
		// 判断是否是参数
		if len(temp) > 1 && temp[0] == '$' {
			parameters = append(parameters, temp)
		}
	}
	return parameters, nil
}

// SplitSql 把sql语句分割成一个个字符串片段，移除了不在字符串中的空格
func SplitSql(sqlStr string) []string {
	var charStr []string
	var startIndex, endIndex int
	endIndex = -1
	runes := []rune(sqlStr)
	length := len(runes)
	var quotaFlag bool
	var quotaChar rune
	for i, char := range runes {
		if i < startIndex {
			continue
		}
		// 判断是否是引号，根据引号开始和结束一致来判断
		if IsQuotaChar(string(char)) {
			// 不是转义字符
			if i == 0 || sqlStr[i-1] != '\\' {
				if !quotaFlag {
					// 字符串开始
					quotaFlag = true
					quotaChar = char
				} else if quotaChar == char {
					// 字符串结束
					quotaFlag = false
					quotaChar = 0
					// 把反引号包含进去，这儿设置endIndex = i+1表示把字符串放入列表中，如果要剔除，则修改startIndex
					endIndex = i + 1
				}
			}
		}
		// 处于字符串中
		if quotaFlag {
			continue
		}
		// 判断是否是需要分割的字符，如何处理<=等多个字符组成的分割符
		splitCharFlag := false
		if endIndex == -1 && IsSplitChar(string(char)) {
			endIndex = i
			splitCharFlag = true
		}
		// 文本末尾
		if endIndex == -1 && length == i+1 {
			endIndex = length
		}
		// 存在需要截取字符串的末尾下标时，截取并放入列表
		if endIndex != -1 {
			temp := string(runes[startIndex:endIndex])
			if temp != "" && temp != " " {
				charStr = append(charStr, temp)
			}
			if splitCharFlag && char != ' ' {
				charStr = append(charStr, string(char))
			}
			startIndex = i + 1
			endIndex = -1
		}
	}
	return charStr
}
