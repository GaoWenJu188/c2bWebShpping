package routers

import (
	"pyg/pyg/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)


func init() {

	beego.InsertFilter("/user/*",beego.BeforeExec,guolvfunc)
	beego.Router("/", &controllers.MainController{})
	beego.Router("/register", &controllers.UserController{}, "get:ShowRegister;post:HandleRegister")
	beego.Router("/sendMsg",&controllers.UserController{},"post:HandleSendMsg")
	//用户激活页面
	beego.Router("/register-email",&controllers.UserController{},"get:ShowRegisterEmail;post:HandleRegisterEmail")
	//激活用户
	beego.Router("/active",&controllers.UserController{},"get:ActiveUser")
	//用户登陆
	beego.Router("login",&controllers.UserController{},"get:ShowLogin;post:Login")
	//主页面
	beego.Router("/index",&controllers.GoodsController{},"get:ShowIndex")
	//推出登陆
	beego.Router("/user/logout",&controllers.UserController{},"get:Logout")
	//展示用户中心项
	beego.Router("/user/userCenterInfo",&controllers.UserController{},"get:ShowUserCenterInfo")
	//展示用户地址信息
	beego.Router("/user/userSite",&controllers.UserController{},"get:ShowUserSite;post:HandleAddAddr")
	//展示用户订单
	beego.Router("/user/userOrder",&controllers.UserController{},"get:ShowUserOrder")
	//展示用户购物车
	beego.Router("/user/userCart",&controllers.UserController{},"get:ShowUserCart")
	//展示用户生鲜页面
	beego.Router("/index_sx",&controllers.GoodsController{},"get:ShowIndexSX")
	//展示商品详情页面
	beego.Router("/goodsDetail",&controllers.GoodsController{},"get:ShowGoodsDetail")
	//展示商品列表也
	beego.Router("/goodsType",&controllers.GoodsController{},"get:ShowList")
	//搜索商品
	beego.Router("/searchGoods",&controllers.GoodsController{},"post:HandleSearchGoods")
	//添加购物车
	beego.Router("/addCart",&controllers.CartController{},"post:HandleAddCart")
	//展示购物车
	beego.Router("/user/showCart",&controllers.CartController{},"get:ShowCart")
	//处理添加购物车数量
	beego.Router("upCart",&controllers.CartController{},"post:HandleUpCart")
	//删除购物车行数据
	beego.Router("/user/deleteCart",&controllers.CartController{},"post:DeleteCart")
	//添加商品到订单
	beego.Router("/user/addOrder",&controllers.OrderController{},"post:ShowOrder")
	//提交订单
	beego.Router("/pushOrder",&controllers.OrderController{},"post:HandlePushOrder")
	//用户中心订单展示及处理
	beego.Router("/user/userOrder",&controllers.UserController{},"get:ShowUserOrder")
	//支付
	beego.Router("/pay",&controllers.OrderController{},"get:Pay")

}
func guolvfunc(ctx *context.Context){
	userName:= ctx.Input.Session("userName")
	if userName==nil{
		ctx.Redirect(302,"/index")
		return
	}
}