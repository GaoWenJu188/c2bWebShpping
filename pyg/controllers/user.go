package controllers

import (
	"github.com/astaxie/beego"
	"regexp"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"encoding/json"
	"math/rand"
	"time"
	"fmt"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
	"github.com/astaxie/beego/utils"
)

type UserController struct {
	beego.Controller
}

//展示注册页面
func(this*UserController)ShowRegister(){
	this.TplName = "register.html"
}

func RespFunc(this* UserController,resp map[string]interface{}){
	//3.把容器传递给前段
	this.Data["json"] = resp
	//4.指定传递方式
	this.ServeJSON()
}

type Message struct {
	Message string
	RequestId string
	BizId string
	Code string
}

//发送短信
func(this*UserController)HandleSendMsg(){
	//接受数据
	phone := this.GetString("phone")
	resp := make(map[string]interface{})

	defer RespFunc(this,resp)
	//返回json格式数据
	//校验数据
	if phone == ""{
		beego.Error("获取电话号码失败")
		//2.给容器赋值
		resp["errno"] = 1
		resp["errmsg"] = "获取电话号码错误"
		return
	}
	//检查电话号码格式是否正确
	reg,_ :=regexp.Compile(`^1[3-9][0-9]{9}$`)
	result := reg.FindString(phone)
	if result == ""{
		beego.Error("电话号码格式错误")
		//2.给容器赋值
		resp["errno"] = 2
		resp["errmsg"] = "电话号码格式错误"
		return
	}
	//发送短信   SDK调用
	client, err := sdk.NewClientWithAccessKey("cn-hangzhou", "LTAIu4sh9mfgqjjr", "sTPSi0Ybj0oFyqDTjQyQNqdq9I9akE")
	if err != nil {
		beego.Error("电话号码格式错误")
		//2.给容器赋值
		resp["errno"] = 3
		resp["errmsg"] = "初始化短信错误"
		return
	}
	//生成6位数随机数

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode :=fmt.Sprintf("%06d",rnd.Int31n(1000000))


	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["RegionId"] = "cn-hangzhou"
	request.QueryParams["PhoneNumbers"] = phone
	request.QueryParams["SignName"] = "品优购"
	request.QueryParams["TemplateCode"] = "SMS_164275022"
	request.QueryParams["TemplateParam"] = "{\"code\":"+vcode+"}"

	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		beego.Error("电话号码格式错误")
		//2.给容器赋值
		resp["errno"] = 4
		resp["errmsg"] = "短信发送失败"
		return
	}
	//json数据解析
	var message Message
	json.Unmarshal(response.GetHttpContentBytes(),&message)
	if message.Message != "OK"{
		beego.Error("电话号码格式错误")
		//2.给容器赋值
		resp["errno"] = 6
		resp["errmsg"] = message.Message
		return
	}

	resp["errno"] = 5
	resp["errmsg"] = "发送成功"
	resp["code"]=vcode
}
//用户注册
func (this*UserController)HandleRegister(){
	phone:= this.GetString("phone")
	password:= this.GetString("password")
	repassword:=this.GetString("repassword")
	if phone==""||password==""||repassword==""{
		beego.Error("用户信息获取错误")
		this.Data["errmsg"]="用户信息获取错误"
		this.TplName="register.html"
		return
	}
	if password!=repassword{
		beego.Error("两次密码不一致")
		this.Data["errmsg"]="两次密码不一致"
		this.TplName="register.html"
		return
	}
	//处理数据
	//orm插入数据
	o:=orm.NewOrm()
	var user models.User
	user.Name=phone
	user.Pwd=password
	user.Phone=phone
	o.Insert(&user)
	//注册成功之后去到激活叶页面，但是要传过去username用cookie
	this.Ctx.SetCookie("userName",user.Name,60*10)
	this.Redirect("/register-email",302)
}
//展示用户激活email页面
func (this*UserController)ShowRegisterEmail(){
	this.TplName="register-email.html"
}
//处理用户激活页面业务
func (this*UserController)HandleRegisterEmail(){
	email:= this.GetString("email")
	password:= this.GetString("password")
	repassword:= this.GetString("repassword")
	if email==""|| password==""||repassword==""{
		beego.Error("输入数据不完整")
		this.Data["errmsg"]="输入数据不完整"
		this.TplName="register-email.html"
		return
	}
	if password!=repassword{
		beego.Error("两次密码不一致")
		this.Data["errmsg"]="两次密码不一致"
		this.TplName="register-email.html"
		return
	}
	//对email格式检验
	reg:= regexp.MustCompile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	result:= reg.FindString(email)
	if result==""{
		beego.Error("email格式不正确")
		this.Data["errmsg"]="email格式不正确"
		this.TplName="register-email.html"
		return
	}
	//配置邮件信息
	//utils
	config := `{"username":"759948611@qq.com","password":"lgrrtpdfhwzebfeg","host":"smtp.qq.com","port":587}`
	emailReg:= utils.NewEMail(config)
	//内容配置
	emailReg.Subject="品优购用户激活"
	emailReg.From="759948611@qq.com"
	emailReg.To=[]string{email}
	userName:= this.Ctx.GetCookie("userName")
	emailReg.HTML=	`<a href="http://192.168.182.111:8080/active?userName=`+userName+`">点击激活用户</a>`
	//发送邮件
	emailReg.Send()
	this.Ctx.WriteString("小伙子你很优秀哟")
}