package routers

import (
	"pyg/pyg/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/register", &controllers.UserController{}, "get:ShowRegister;post:HandleRegister")
	beego.Router("/sendMsg",&controllers.UserController{},"post:HandleSendMsg")
	//用户激活页面
	beego.Router("/register-email",&controllers.UserController{},"get:ShowRegisterEmail;post:HandleRegisterEmail")


}
