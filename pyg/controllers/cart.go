package controllers

import (
	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
	"pyg/pyg/models"
	"github.com/astaxie/beego/orm"
)

type CartController struct {
	beego.Controller
}
//添加购物车
func(this*CartController)HandleAddCart(){
	//获取数据
	id,err:= this.GetInt("goodsId")
	num,err2:= this.GetInt("num")
	resp:=make(map[string]interface{})
	//封装，集成，多态
	defer RespFunc(&this.Controller,resp)
	if err!=nil|| err2!=nil{
		resp["errno"]=1
		resp["errmsg"]="输入数据不完整"
		return
	}
	name := this.GetSession("userName")
	if name ==nil{
		resp["errno"]=2
		resp["errmsg"]="当前用户未登陆，不能添加购物车"
		return
	}
	//处理数据
	//把数据存储在redis的hash中
 	conn,err:= redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		resp["errno"]=3
		resp["errmsg"]="服务器宜昌"
		return
	}
	defer conn.Close()
	oldNum ,err:=redis.Int(conn.Do("hget","cart_"+name.(string),id))
	_,err=conn.Do("hset","cart_"+name.(string),id,oldNum+num)
	if err!=nil{
		resp["errno"]=4
		resp["errmsg"]="添加商品到购物车失败"
		return
	}
	//返回数据
	resp["errno"]=5
	resp["errmsg"]="OK"

}
//展示购物车
func (this*CartController)ShowCart(){
	//先要连接redis
	conn,err:= redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		this.Redirect("/index_sx",302)
		return
	}
	defer conn.Close()
	name:= this.GetSession("userName")
	result,err:= redis.Ints(conn.Do("hgetall","cart_"+name.(string)))
	if err!=nil{
		this.Redirect("/index_sx",302)
		return
	}
	//定义大容器

	var goods []map[string]interface{}

	o:=orm.NewOrm()
	totalPrice :=0
	totalCount:=0
	for i:=0;i<len(result);i+=2{
		temp :=make(map[string]interface{})
		var goodsSku models.GoodsSKU
		goodsSku.Id=result[i]
		o.Read(&goodsSku)
		//给行容器赋值

		temp["goodsSku"]=goodsSku
		temp["count"]=result[i+1]

		littlePrice:=result[i+1]*goodsSku.Price
		temp["littlePrice"]=littlePrice
		totalCount++
		totalPrice+=littlePrice
		//把行容器添加到大容器里
		goods=append(goods,temp)
	}
	this.Data["totalCount"] = totalCount
	this.Data["totalPrice"] = totalPrice
	this.Data["goods"] = goods
	this.Layout="user_center_head.html"
	this.TplName = "cart.html"
}
//处理添加购物车数量
func(this*CartController)HandleUpCart(){
	id ,err:= this.GetInt("goodsId")
 	count,err2:= this.GetInt("count")

 	//定义容器
 	resp:= make(map[string]interface{})
 	defer RespFunc(&this.Controller,resp)

 	if err!=nil||err2!=nil{
 		resp["errno"]=1
 		resp["errmsg"]="传输数据不完整"
 		return
	}
	name:= this.GetSession("userName")
	if name==nil{
		resp["errno"]=3
		resp["errmsg"]="当前用户未登陆"
		return
	}

	//向redis中写入购物车数量
	conn,err:= redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		resp["errno"]=2
		resp["errmsg"]="redis的连接错误"
		return
	}
	defer  conn.Close()
	_,err= conn.Do("hset","cart_"+name.(string),id,count)
	if err!=nil{
		resp["errno"]=4
		resp["errmsg"]="redis写入数据失败"
		return
	}
	resp["errno"]=5
	resp["errmsg"]="OK"

}
//删除购物行数据
func (this*CartController)DeleteCart(){
	id,err:= this.GetInt("goodsId")
	resp :=make(map[string]interface{})
	defer RespFunc(&this.Controller,resp)
	if err!=nil{
		resp["errno"]=1
		resp["errmsg"]="无效商品id"
		return
	}
 	conn,err:= redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		resp["errno"]=2
		resp["errmsg"]="redis连接失败"
		return
	}
	userName:=this.GetSession("userName")
	o:=orm.NewOrm()
	var user models.User
	user.Name=userName.(string)
    err=o.Read(&user,"Name")
	if err!=nil{
		resp["errno"]=3
		resp["errmsg"]="用户不存在"
		return
	}
	_,err= conn.Do("hdel","cart_"+user.Name,id)
	if err!=nil{
		resp["errno"]=4
		resp["errmsg"]="redis删除失败"
		return
	}
	resp["errno"]=5
	resp["errmsg"]="OK"
}