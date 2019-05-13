package controllers

import "github.com/astaxie/beego"

type GoodsController struct {
	beego.Controller
}
//展示首页
func (this*GoodsController)ShowIndex(){
	userName:= this.GetSession("userName")
	if userName!=nil{
		this.Data["userName"]=userName.(string)
	}else{
		this.Data["userName"]=""
	}
	this.TplName="index.html"
}

