package util

import (
	"encoding/json"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// StrVal 获取变量的字符串值
// 浮点型 3.0将会转换成字符串3, "3"
// 非数值或字符类型的变量将会被转换成JSON格式字符串
func StrVal(value interface{}) string {
	var key string
	if value == nil {
		return key
	}
	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}
	return key
}

// HasArguments 替换变量
func HasArguments(arguments map[string]interface{}, val string) string {
	if len(arguments) == 0 {
		return val
	}
	//if strings.HasPrefix(val, ArgumentsDefiner) {
	switch arguments[val].(type) {
	default:
		if arguments[val] == nil {
			val = ""
		} else {
			val = arguments[val].(string)
		}
	case float64:
		if arguments[val].(float64) == math.Trunc(arguments[val].(float64)) {
			val = strconv.Itoa(int(math.Ceil(arguments[val].(float64))))
		} else {
			val = strconv.FormatFloat(arguments[val].(float64), 'f', -1, 64)
		}
	case []map[string]interface{}:
		b, _ := json.Marshal(arguments[val].([]map[string]interface{}))
		val = string(b)
	case bool:
		val = strconv.FormatBool(arguments[val].(bool))
	}
	//}
	return val
}

// MapJsonValToAny 验证map里面的值是否是json 如果是转为任意类型
func MapJsonValToAny(m map[string]interface{}) map[string]interface{} {
	j := jsoniter.Config{
		UseNumber: true,
	}.Froze()
	for mk, mv := range m {
		if json.Valid([]byte(StrVal(mv))) {
			var object interface{}
			j.Unmarshal([]byte(StrVal(mv)), &object)
			m[mk] = object
		}
	}
	return m
}

// Duplicate 数组去重
func Duplicate(a interface{}) (ret []interface{}) {
	va := reflect.ValueOf(a)
	for i := 0; i < va.Len(); i++ {
		if i > 0 && reflect.DeepEqual(va.Index(i-1).Interface(), va.Index(i).Interface()) {
			continue
		}
		ret = append(ret, va.Index(i).Interface())
	}
	return ret
}

// ConvertStrSlice2Map 将字符串 slice 转为 map[string]struct{}。
func ConvertStrSlice2Map(sl []*Node) map[string]struct{} {
	set := make(map[string]struct{}, len(sl))
	for _, v := range sl {
		set[v.Code] = struct{}{}
	}
	return set
}

// InMap 判断字符串是否在 map 中。
func InMap(m map[string]struct{}, s string) bool {
	_, ok := m[s]
	return ok
}

// InSlice 判断字符串是否在 slice 中。
func InSlice(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

// FormatHasArgumentsJson 转换包含变量的json字符串
func FormatHasArgumentsJson(json string) string {
	// json中包含变量
	if strings.Contains(json, ArgumentsDefiner) {
		char := CharAtJson(json)
		for _, s := range char {
			json = strings.ReplaceAll(json, s, "\""+s+"\"")
		}
		return json
	} else {
		return json
	}
}

const (
	Rj = "\r"
	Jd = ",\n"
	Jn = "\n"
	Jm = ":"
	Jr = "}"
	Jl = "{"
)

var JsonParse = []string{Rj, Jd, Jn, Jm, Jr}

// CharAtJson json切割
func CharAtJson(json string) []string {
	var charStr []string
	cs := strings.Split(json, ArgumentsDefiner)
	for i, c := range cs {
		if i == 0 {
			continue
		}
		if i > 0 && strings.HasSuffix(cs[i-1], "\"") {
			continue
		} else {
			index := FIndex(JsonParse, c)
			if index != -1 {
				charStr = append(charStr, ArgumentsDefiner+c[0:index])
			}
		}
	}
	return charStr
}

func ChatSyhJson(json string) string {
	json = strings.ReplaceAll(json, "\"{", "{")
	json = strings.ReplaceAll(json, "}\"", "}")
	return json
}

// FIndex 循环找到index对应下标
func FIndex(y []string, s string) int {
	for _, s2 := range y {
		if !(strings.Index(s, s2) == -1) {
			return strings.Index(s, s2)
		}
	}
	return -1
}

// ReplaceLong 正则替换将大数变为string类型
func ReplaceLong(data interface{}) interface{} {
	j, _ := json.Marshal(data)
	//reg := regexp.MustCompile(`id\":(\d{16,20}),"`)
	reg := regexp.MustCompile(`(\d{16,20})`)
	l := len(reg.FindAllString(string(j), -1)) //正则匹配16-20位的数字，如果找到了就开始正则替换并解析
	if l != 0 {
		//fmt.Printf("\n正则替换前的数据%+v", data)
		//var mapResult map[string]interface{}
		reg := regexp.MustCompile(`(\d{16,20})}`)
		str := reg.ReplaceAllString(string(j), `"${1}"}`)
		reg = regexp.MustCompile(`(\d{16,20})]`)
		str = reg.ReplaceAllString(str, `"${1}"]`) //执行替换
		//str := reg.ReplaceAllString(string(j), `id": "${1}","`)
		j := jsoniter.Config{
			UseNumber: true,
		}.Froze()
		j.Unmarshal([]byte(str), &data)
		//data = &mapResult
	}
	return data
}

// ToNumber 转换为数字
func ToNumber(s interface{}) float64 {
	if s == nil {
		return 0
	}
	var v float64
	var err error
	switch s.(type) {
	case float64:
		v = s.(float64)
	case int:
		v = s.(float64)
	default:
		v, err = strconv.ParseFloat(s.(string), 64)
		//v, err := strconv.Atoi(s.(string))
	}
	if err != nil {
		return 0
	}
	return v
}

// ToBool 转换为bool
func ToBool(s interface{}) bool {
	if s == nil {
		return false
	}
	var v bool
	var err error
	switch s.(type) {
	case bool:
		v = s.(bool)
	default:
		v, err = strconv.ParseBool(s.(string))
	}
	if err != nil {
		return false
	}
	return v
}

// ToStr 任意类型转str
func ToStr(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

// ToDate 转时间类型
func ToDate(dateStr string) time.Time {
	local, _ := time.LoadLocation("Asia/Shanghai")
	var fs string
	if len(dateStr) > 10 {
		fs = "2006-01-02 15:04:05"
	} else {
		fs = "2006-01-02"
	}
	t, _ := time.ParseInLocation(fs, dateStr, local)
	return t
}

// ConvertSlice 处理数组参数
func ConvertSlice(idx []string) []string {
	var v string
	var idxN []string
	for _, s := range idx {
		if strings.HasPrefix(s, "[") || len(v) > 0 {
			v = v + s + ","
			if strings.HasSuffix(s, "]") {
				idxN = append(idxN, v[:len(v)-1])
				v = ""
			}
		} else {
			idxN = append(idxN, s)
		}
	}
	return idxN
}

// SimpleCopyProperties 复制对象
func SimpleCopyProperties(dst, src interface{}) (err error) {
	// 防止意外panic
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintf("%v", e))
		}
	}()
	dstType, dstValue := reflect.TypeOf(dst), reflect.ValueOf(dst)
	srcType, srcValue := reflect.TypeOf(src), reflect.ValueOf(src)
	// dst必须结构体指针类型
	if dstType.Kind() != reflect.Ptr || dstType.Elem().Kind() != reflect.Struct {
		return errors.New("dst type should be a struct pointer")
	}
	// src必须为结构体或者结构体指针，.Elem()类似于*ptr的操作返回指针指向的地址反射类型
	if srcType.Kind() == reflect.Ptr {
		srcType, srcValue = srcType.Elem(), srcValue.Elem()
	}
	if srcType.Kind() != reflect.Struct {
		return errors.New("src type should be a struct or a struct pointer")
	}
	// 取具体内容
	dstType, dstValue = dstType.Elem(), dstValue.Elem()
	// 属性个数
	propertyNums := dstType.NumField()
	for i := 0; i < propertyNums; i++ {
		// 属性
		property := dstType.Field(i)
		// 待填充属性值
		propertyValue := srcValue.FieldByName(property.Name)
		// 无效，说明src没有这个属性 || 属性同名但类型不同
		if !propertyValue.IsValid() || property.Type != propertyValue.Type() {
			continue
		}
		if dstValue.Field(i).CanSet() {
			dstValue.Field(i).Set(propertyValue)
		}
	}
	return nil
}

// MapInterfaceNumberToString 转换json.number为string
func MapInterfaceNumberToString(m interface{}) interface{} {
	if m == nil {
		return m
	}
	if reflect.TypeOf(m).String() == "[]map[string]interface {}" {
		m := m.([]map[string]interface{})
		for _, v := range m {
			for k, vv := range v {
				if vv != nil {
					if reflect.TypeOf(vv).String() == "float64" {
						v[k] = strconv.FormatFloat(vv.(float64), 'f', -1, 64)
					}
					if reflect.TypeOf(vv).String() == "json.Number" {
						v[k] = vv.(json.Number).String()
					}
				}
			}
		}
	}
	if reflect.TypeOf(m).String() == "map[string]interface {}" {
		m := m.(map[string]interface{})
		for k, vv := range m {
			if vv != nil {
				if reflect.TypeOf(vv).String() == "float64" {
					m[k] = strconv.FormatFloat(vv.(float64), 'f', -1, 64)
				}
				if reflect.TypeOf(vv).String() == "json.Number" {
					m[k] = vv.(json.Number).String()
				}
			}
		}
	}

	return m

}

// SortedJsonArray 排序数组 (需要排序对象,列名)
func SortedJsonArray(data interface{}, c string) interface{} {

	return nil
}
