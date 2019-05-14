package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
)

type GoodsController struct {
	beego.Controller
}
//展示首页
func(this*GoodsController)ShowIndex(){
	userName := this.GetSession("userName")
	if userName != nil {
		this.Data["userName"] = userName.(string)
	}else {
		this.Data["userName"] = ""
	}

	//获取类型信息并传递给前段
	//获取一级菜单
	o := orm.NewOrm()
	//接受对象
	var oneClass []models.TpshopCategory
	//查询
	o.QueryTable("TpshopCategory").Filter("Pid",0).All(&oneClass)


	//获取第二级
	var types []map[string]interface{}
	for _,v := range oneClass{
		//行容器
		t := make(map[string]interface{})

		var temp []models.TpshopCategory
		o.QueryTable("TpshopCategory").Filter("Pid",v.Id).All(&temp)
		t["t1"] = v
		t["t2"] = temp
		types = append(types,t)
	}
	//获取第三级菜单
	for _,v1:=range types{
		var erji []map[string]interface{}
		//因为v1里面有两个值t1是一级菜单，t2才是二级菜单
		for _,v2:=range v1["t2"].([]models.TpshopCategory){
			t:=make(map[string]interface{})
			var thirdClass []models.TpshopCategory
			//获取三级菜单
			o.QueryTable("TpshopCategory").Filter("Pid",v2.Id).All(&thirdClass)
			t["t22"]=v2
			t["t23"]=thirdClass
			erji=append(erji,t)
		}
		//把二级容器放到宗容器中
		v1["t3"]=erji
	}
	this.Data["types"] = types
	this.TplName = "index.html"
}

