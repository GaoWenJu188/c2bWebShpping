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
}
func guolvfunc(ctx *context.Context){
	userName:= ctx.Input.Session("userName")
	if userName==nil{
		ctx.Redirect(302,"/index")
		return
	}
}