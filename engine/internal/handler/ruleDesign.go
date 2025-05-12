package handler

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
	"zenith.engine.com/engine/config"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/internal/adapter"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

func GetDesign(ctx *gin.Context) {
	code := ctx.Query("code")
	version := ctx.Query("version")
	rudDesign, err := adapter.GetStorage().GetByCode(code, version)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	internal.Success(ctx, rudDesign, e.GetMsg(e.SUCCESS))
}

func GetDesignById(ctx *gin.Context) {
	id := ctx.Param("id")
	uid, _ := strconv.ParseUint(id, 10, 64)
	design, err := adapter.GetStorage().GetById(uid)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.RuleCodeNotExist))
		return
	}
	internal.Success(ctx, design, e.GetMsg(e.SUCCESS))
}

func SaveDesign(ctx *gin.Context) {
	var rud repository.RuleDesign
	err := ctx.BindJSON(&rud)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	// 验证code是否为空,如果是空的生成code(当前时间搓)
	if len(rud.Code) == 0 {
		rud.Code = strconv.FormatInt(time.Now().Unix(), 10)
	}
	rud, err = saveAndAnalysis(rud)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	internal.Success(ctx, rud, e.GetMsg(e.SUCCESS))
}

func saveAndAnalysis(rud repository.RuleDesign) (repository.RuleDesign, error) {
	var ruleMap map[string]string
	var err error
	if rud.Id == 0 {
		// 新增
		rud.Version = "1.0"
		// 将图转换为规则yaml
		ruleMap, err = analysisDesign(&rud, ruleMap)
		if err != nil {
			return rud, errors.New(err.Error())
		}
		// 解析完成入库
		createError := adapter.GetStorage().Create(&rud)
		if createError != nil {
			return rud, errors.New(err.Error())
		}
		registerUri(&rud)
	} else {
		// 编辑
		// 删除内存中规则
		err = DeletePools(util.ParseYamlCode(rud.Code, rud.Version), rud.Design)
		repository.RuleDesignMap.Delete(util.ParseYamlCode(rud.Code, rud.Version))
		if err != nil {
			return rud, errors.New(err.Error())
		}
		// 将图转换为规则yaml
		ruleMap, err = analysisDesign(&rud, ruleMap)
		if err != nil {
			return rud, errors.New(err.Error())
		}
		var updateError error
		if rud.Publish == 0 {
			// 修改画布上基础属性
			repository.UpdateCanvas(&rud)
			// 如果状态没有发布修改主表 修改子表数据
			updateError = adapter.GetStorage().Update(&rud)
			poolStore(&rud, ruleMap)
		} else {
			// 自增版本,得到版本
			rud.Version = graterVersion(rud)
			rud.Publish = 0
			// 修改画布上基础属性
			repository.UpdateCanvas(&rud)
			// 如果状态发布了 新增子表 修改主表
			updateError = adapter.GetStorage().UpdatePublish(&rud)
		}
		if updateError != nil {
			return rud, errors.New(err.Error())
		}
	}
	return rud, nil
}

// analysisDesign 将图转换为规则yaml
func analysisDesign(rud *repository.RuleDesign, ruleMap map[string]string) (map[string]string, error) {
	if len(rud.DesignJson) > 0 && rud.StorageType == 0 {
		// 将图转换为规则yaml
		err, designYaml := util.AnalysisDesign(rud.DesignJson)
		name, code, desc, uri, groupId := util.AnalysisCanvas(rud.DesignJson)
		rud.Name = name
		if len(code) > 0 {
			rud.Code = code
		}
		rud.Desc = desc
		rud.Uri = uri
		rud.GroupId = groupId
		if err != nil {
			return nil, err
		}
		rud.Design = designYaml
		// 解析规则实现..
		yamlCode := util.ParseYamlCode(rud.Code, rud.Version)
		ruleMap, err = util.ExplainYaml(yamlCode, rud.Design)
		if err != nil {
			return nil, err
		}
		// 验证是否存在,如果
		ruleDes, _ := adapter.GetStorage().GetByCode(code, rud.Version)
		if ruleDes.Id != rud.Id && ruleDes.Id != 0 {
			return nil, errors.New(e.GetMsg(e.RuleCodeExist))
		}
	}
	return ruleMap, nil
}

// poolStore 将规则放入规则池
func poolStore(rud *repository.RuleDesign, ruleMap map[string]string) {
	if len(rud.DesignJson) > 0 && rud.StorageType == 0 {
		// 规则引擎池放入内存
		for ruleCode, ruleStr := range ruleMap {
			EnginePoolMap.Store(ruleCode, BuildEnginePool(ruleStr))
		}
	}
}

func registerUri(rud *repository.RuleDesign) {
	// 规则放入gin uri
	if len(rud.Uri) > 0 {
		CallBackRouter(rud.Uri)
	} else {
		CallBackRouter("/" + rud.Version + "/" + rud.Code)
	}
}

// PublishDesign 发布规则设计
func PublishDesign(ctx *gin.Context) {
	var rud repository.RuleDesign
	err := ctx.BindJSON(&rud)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	var ruleMap map[string]string
	if rud.Publish == 0 && rud.Id != 0 {
		// 修改数据库中发布数据
		rud.Publish = 1
		// 当前规则是未发布状态时
		err = DeletePools(util.ParseYamlCode(rud.Code, rud.Version), rud.Design)
		repository.RuleDesignMap.Delete(util.ParseYamlCode(rud.Code, rud.Version))
		if err != nil {
			internal.Fail(ctx, nil, err.Error())
			return
		}
		// 将图转换为规则yaml
		ruleMap, err = analysisDesign(&rud, ruleMap)
		if err != nil {
			internal.Fail(ctx, nil, err.Error())
		}
		updateError := adapter.GetStorage().Update(&rud)
		if updateError != nil {
			internal.Fail(ctx, nil, updateError.Error())
			return
		}
	} else {
		// 修改数据库中发布数据
		rud.Publish = 1
		// 当前规则已发布时
		// 自增版本,得到版本
		rud.Version = graterVersion(rud)
		// 修改画布上基础属性
		repository.UpdateCanvas(&rud)
		// 将图转换为规则yaml
		ruleMap, err = analysisDesign(&rud, ruleMap)
		if err != nil {
			internal.Fail(ctx, nil, err.Error())
		}
		// 解析完成入库
		createError := adapter.GetStorage().UpdatePublish(&rud)
		if createError != nil {
			internal.Fail(ctx, nil, createError.Error())
			return
		}
		// 注册路由
		registerUri(&rud)
	}
	// 存储规则进入规则池
	poolStore(&rud, ruleMap)
	internal.Success(ctx, rud, e.GetMsg(e.SUCCESS))
}

// DeleteDesign 删除规则设计
func DeleteDesign(ctx *gin.Context) {
	id := ctx.Param("id")
	uid, _ := strconv.ParseUint(id, 10, 64)
	design, err := adapter.GetStorage().GetById(uid)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	// 删除内存中规则
	err = DeletePools(util.ParseYamlCode(design.Code, design.Version), design.Design)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	// 删除数据库数据
	deleteErr := adapter.GetStorage().Delete(uid)

	if deleteErr != nil {
		internal.Fail(ctx, nil, deleteErr.Error())
		return
	}
	// 删除内存中规则
	repository.RuleDesignMap.Delete(util.ParseYamlCode(design.Code, design.Version))
	internal.Success(ctx, nil, e.GetMsg(e.SUCCESS))
}

// DeleteDesigns 批量删除规则设计
func DeleteDesigns(ctx *gin.Context) {
	var ids []uint64
	err := ctx.BindJSON(&ids)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	design, err := adapter.GetStorage().GetByIds(ids)
	for _, ruleDesign := range design {
		// 删除内存中规则
		err = DeletePools(util.ParseYamlCode(ruleDesign.Code, ruleDesign.Version), ruleDesign.Design)
		repository.RuleDesignMap.Delete(util.ParseYamlCode(ruleDesign.Code, ruleDesign.Version))
	}
	// 删除数据库数据
	deleteErr := adapter.GetStorage().DeleteIds(ids)
	if deleteErr != nil {
		internal.Fail(ctx, nil, deleteErr.Error())
		return
	}
	internal.Success(ctx, nil, e.GetMsg(e.SUCCESS))
}

// UpdateDesign 修改规则设计
func UpdateDesign(ctx *gin.Context) {
	var rud repository.RuleDesign
	err := ctx.BindJSON(&rud)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	// 删除内存中规则
	err = DeletePools(util.ParseYamlCode(rud.Code, rud.Version), rud.Design)
	repository.RuleDesignMap.Delete(util.ParseYamlCode(rud.Code, rud.Version))
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	// 从新解析规则
	yamlCode := util.ParseYamlCode(rud.Code, rud.Version)
	ruleMap, err := util.ExplainYaml(yamlCode, rud.Design)
	if err != nil {
		return
	}
	// 规则引擎池放入内存
	for ruleCode, ruleStr := range ruleMap {
		EnginePoolMap.Store(ruleCode, BuildEnginePool(ruleStr))
	}
	// 修改数据库中数据
	updateError := adapter.GetStorage().Update(&rud)
	if updateError != nil {
		internal.Fail(ctx, nil, updateError.Error())
		return
	}
	internal.Success(ctx, nil, e.GetMsg(e.SUCCESS))
}

func ListDesign(ctx *gin.Context) {
	var pageResult util.PageResult
	pageSize := ctx.Query("pageSize")
	pageNum := ctx.Query("pageNum")
	name := ctx.Query("name")
	code := ctx.Query("code")
	groupId := ctx.Query("groupId")
	typeParam := ctx.Query("type")
	size, _ := strconv.Atoi(pageSize)
	num, _ := strconv.Atoi(pageNum)
	types, _ := strconv.Atoi(typeParam)
	var groupIds []string
	// 查询有权限的
	userinfo := util.CurrentUser(ctx)
	var roleId uint64
	if userinfo != nil {
		roleId = userinfo.RoleId
	} else {
		roleId = 1
	}
	rgs := repository.GetRoleGroupByRoleId(roleId)
	if len(groupId) > 0 {
		groupIds = append(groupIds, groupId)
	} else {
		for _, rg := range rgs {
			groupIds = append(groupIds, strconv.FormatUint(rg.GroupId, 10))
		}
	}
	err, designs, count := adapter.GetStorage().ListDesign(num, size, name, code, groupIds, types)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
	}
	pageResult.Data = designs
	pageResult.PageSize = size
	pageResult.PageNum = num
	pageResult.Count = count
	internal.Success(ctx, pageResult, e.GetMsg(e.SUCCESS))
}

func CheckDesign(ctx *gin.Context) {
	code := ctx.Query("code")
	version := ctx.Query("version")
	b := adapter.GetStorage().CheckDesignExist(code, version)
	internal.Success(ctx, b, e.GetMsg(e.SUCCESS))
}

func CheckCode(ctx *gin.Context) {
	code := ctx.Query("code")
	id := ctx.Query("id")
	b := adapter.GetStorage().CheckCodeExist(code, id)
	internal.Success(ctx, b, e.GetMsg(e.SUCCESS))
}

func GetParamDefine(ctx *gin.Context) {
	var dataList []interface{}
	if len(config.Conf.Variable.ApiDefine) > 0 {
		data, err := XPost(config.Conf.Variable.ApiDefine, nil, "{}")
		if err != nil {
			internal.Fail(ctx, nil, err.Error())
			return
		}
		err = json.Unmarshal(data, &dataList)
	}
	internal.Success(ctx, dataList, e.GetMsg(e.SUCCESS))
}

// CopyDesign 拷贝规则
func CopyDesign(ctx *gin.Context) {
	id := ctx.Param("id")
	group := ctx.Query("groupId")
	groupId, _ := strconv.ParseUint(group, 10, 64)
	uid, _ := strconv.ParseUint(id, 10, 64)
	design, err := adapter.GetStorage().GetById(uid)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.RuleCodeNotExist))
		return
	}
	design.Version = "1.0"
	design.Id = 0
	ks := getCopyVersion(design.Code)
	design.Name = design.Name + "-" + ks
	design.Uri = design.Uri + "-" + ks
	design.Code = design.Code + "-" + ks
	design.Publish = 0
	design.GroupId = groupId
	// 修改设计图中画布基础属性
	var designInfo util.DesignJson
	err = json.Unmarshal([]byte(design.DesignJson), &designInfo)
	attrs := designInfo.Attrs.(map[string]interface{})
	attrs["ruleName"] = design.Name
	attrs["ruleCode"] = design.Code
	attrs["ruleVersion"] = design.Version
	design.Uri = util.GraterUri(design.Version, design.Code)
	attrs["ruleURI"] = design.Uri
	attrs["ruleGroup"] = groupId
	designInfo.Attrs = attrs
	designByte, _ := json.Marshal(designInfo)
	design.DesignJson = string(designByte)
	design, err = saveAndAnalysis(design)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.RuleCodeNotExist))
		return
	}
	internal.Success(ctx, design, e.GetMsg(e.SUCCESS))
}

// GetListVersion 根据code获取全部版本数据
func GetListVersion(ctx *gin.Context) {
	code := ctx.Query("code")
	rdList, err := adapter.GetStorage().ListByCode(code)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.RuleCodeNotExist))
		return
	}
	internal.Success(ctx, rdList, e.GetMsg(e.SUCCESS))
}

// GetEngineDesign 通过编码版本获取设计信息json
func GetEngineDesign(ctx *gin.Context) {
	code := ctx.Query("code")
	version := ctx.Query("version")
	// 如果version为空，则查询最新的版本号
	if len(version) == 0 {
		rdList, err := adapter.GetStorage().ListByCode(code)
		if err != nil {
			internal.Fail(ctx, nil, e.GetMsg(e.RuleCodeNotExist))
			return
		}
		version = rdList[0]
	}
	designJson := adapter.GetStorage().GetDesignByCodeAndVersion(code, version)
	internal.Success(ctx, designJson, e.GetMsg(e.SUCCESS))
}

func getCopyVersion(code string) string {
	var i = 1
	for {
		if !adapter.GetStorage().CheckDesignExist(code+"-"+strconv.Itoa(i), "1.0") {
			break
		}
		i++
	}
	return strconv.Itoa(i)
}

var CallBackRouter func(uri string)

func RegisterRouterEvent(f func(uri string)) {
	CallBackRouter = f
}

func graterVersion(rud repository.RuleDesign) string {
	if len(rud.Version) == 0 {
		return "1.0"
	} else {
		lastVersion := adapter.GetStorage().GetLastVersionByCode(rud.Code)
		f, _ := strconv.ParseFloat(lastVersion, 64)
		f += 0.1
		return strconv.FormatFloat(f, 'f', 1, 64)
	}
}
