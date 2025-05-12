package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/sql"
	"zenith.engine.com/engine/pkg/util"
)

type PreviewSqlParam struct {
	DataBase  string                      `json:"dataBase"`
	Sql       string                      `json:"sql"`
	Arguments map[string]string           `json:"arguments"`
	Result    []util.ExecuteMappingResult `json:"result,omitempty"`
}

// ValidateSql /engine/sql/validate 接口处理逻辑
func ValidateSql(ctx *gin.Context) {
	var params map[string]string
	err := ctx.BindJSON(&params)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	sqlStr := params["sql"]
	parameters, err := sql.ValidateSql(sqlStr)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	data := make(map[string]interface{})
	data["parameters"] = parameters
	internal.Success(ctx, data, e.GetMsg(e.SUCCESS))
}

//DataList 获取可用规则数据库列表
func DataList(ctx *gin.Context) {
	var dataKeyList []string
	repository.DataMap.Range(func(key, value any) bool {
		dataKeyList = append(dataKeyList, key.(string))
		return true
	})
	sort.Strings(dataKeyList)
	internal.Success(ctx, dataKeyList, e.GetMsg(e.SUCCESS))
}

// PreviewSql 执行sql预览结果
func PreviewSql(ctx *gin.Context) {
	var previewSqlParam PreviewSqlParam
	err := ctx.BindJSON(&previewSqlParam)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	var arguments []interface{}
	var sqlArguments string
	previewSqlParam.Sql, sqlArguments = util.ExplainSqlArgumentsPreview(previewSqlParam.Sql)
	var as []string
	// 判断sql是否以limit结尾
	// 未来需要根据不同的数据库，实现不同的处理，目前只支持mysql的处理
	r, err := regexp.Compile("(?i)\\s*limit\\s+\\d+(\\s*,\\s*\\d+)?\\s*$")
	if err != nil {
		panic("创建sql limit验证正则表达式失败")
	}
	matched := r.MatchString(previewSqlParam.Sql)
	if !matched {
		// 没有以limit结尾，自动添加，避免数据过多
		dbVal, _ := repository.DataMap.Load(previewSqlParam.DataBase)
		db := dbVal.(*gorm.DB)
		if db.Dialector.Name() == "oracle" {
			// oracle需要特殊处理
		} else {
			previewSqlParam.Sql += " LIMIT 1000"
		}
	}
	if len(sqlArguments) > 0 {
		as = strings.Split(sqlArguments[1:], ",")
	}
	for _, s := range as {
		arguments = append(arguments, previewSqlParam.Arguments[util.ArgumentsDefiner+s[11:len(s)-2]])
	}
	m, err := ConnectDataPreviewSql(previewSqlParam.DataBase, previewSqlParam.Sql, arguments...)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	data, err := json.Marshal(m)
	mapping := make(map[string]interface{})
	// 隐式转换
	m = util.MapInterfaceNumberToString(m)
	for _, rs := range previewSqlParam.Result {
		if rs.Selected {
			rc := gjson.Get(string(data), rs.Path)
			if rc.Exists() && len(rc.String()) > 0 {
				mapping[rs.Title] = rc.String()
			} else if len(rs.Path) == 0 && len(util.ToStr(rs.Def)) == 0 {
				// 没有配置路径和默认值时，表示取所有数据
				if m != nil && reflect.TypeOf(m).String() == "map[string]interface {}" {
					var c []map[string]interface{}
					m = append(c, m.(map[string]interface{}))
				}
				mapping[rs.Title] = m
			} else {
				mapping[rs.Title] = rs.Def
			}
		}
	}

	shr := SendHttpResult{Origin: m, Mapping: mapping}
	internal.Success(ctx, shr, e.GetMsg(e.SUCCESS))
}
