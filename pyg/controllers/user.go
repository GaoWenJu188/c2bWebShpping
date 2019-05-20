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
	"strings"
	"github.com/gomodule/redigo/redis"
)

type UserController struct {
	beego.Controller
}
//展示注册页面
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}
func RespFunc(this *beego.Controller, resp map[string]interface{}) {
	//3.把容器传递给前段
	this.Data["json"] = resp
	//4.指定传递方式
	this.ServeJSON()
}

type Message struct {
	Message   string
	RequestId string
	BizId     string
	Code      string
}
//发送短信
func (this *UserController) HandleSendMsg() {
	//接受数据
	phone := this.GetString("phone")
	resp := make(map[string]interface{})

	defer RespFunc(&this.Controller, resp)
	//返回json格式数据
	//校验数据
	if phone == "" {
		beego.Error("获取电话号码失败")
		//2.给容器赋值
		resp["errno"] = 1
		resp["errmsg"] = "获取电话号码错误"
		return
	}
	//检查电话号码格式是否正确
	reg, _ := regexp.Compile(`^1[3-9][0-9]{9}$`)
	result := reg.FindString(phone)
	if result == "" {
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
	vcode := fmt.Sprintf("%06d", rnd.Int31n(1000000))

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
	request.QueryParams["TemplateParam"] = "{\"code\":" + vcode + "}"

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
	json.Unmarshal(response.GetHttpContentBytes(), &message)
	if message.Message != "OK" {
		beego.Error("电话号码格式错误")
		//2.给容器赋值
		resp["errno"] = 6
		resp["errmsg"] = message.Message
		return
	}

	resp["errno"] = 5
	resp["errmsg"] = "发送成功"
	resp["code"] = vcode
}
//用户注册
func (this *UserController) HandleRegister() {
	phone := this.GetString("phone")
	password := this.GetString("password")
	repassword := this.GetString("repassword")
	if phone == "" || password == "" || repassword == "" {
		beego.Error("用户信息获取错误")
		this.Data["errmsg"] = "用户信息获取错误"
		this.TplName = "register.html"
		return
	}
	if password != repassword {
		beego.Error("两次密码不一致")
		this.Data["errmsg"] = "两次密码不一致"
		this.TplName = "register.html"
		return
	}
	//处理数据
	//orm插入数据
	o := orm.NewOrm()
	var user models.User
	user.Name = phone
	user.Pwd = password
	user.Phone = phone
	o.Insert(&user)
	//注册成功之后去到激活叶页面，但是要传过去username用cookie
	this.Ctx.SetCookie("userName", user.Name, 60*10)
	this.Redirect("/register-email", 302)
}
//展示用户激活email页面
func (this *UserController) ShowRegisterEmail() {
	this.TplName = "register-email.html"
}
//处理用户激活页面业务
func (this *UserController) HandleRegisterEmail() {
	email := this.GetString("email")
	password := this.GetString("password")
	repassword := this.GetString("repassword")
	if email == "" || password == "" || repassword == "" {
		beego.Error("输入数据不完整")
		this.Data["errmsg"] = "输入数据不完整"
		this.TplName = "register-email.html"
		return
	}
	if password != repassword {
		beego.Error("两次密码不一致")
		this.Data["errmsg"] = "两次密码不一致"
		this.TplName = "register-email.html"
		return
	}
	//对email格式检验
	reg := regexp.MustCompile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	result := reg.FindString(email)
	if result == "" {
		beego.Error("email格式不正确")
		this.Data["errmsg"] = "email格式不正确"
		this.TplName = "register-email.html"
		return
	}
	//配置邮件信息
	//utils
	config := `{"username":"759948611@qq.com","password":"lgrrtpdfhwzebfeg","host":"smtp.qq.com","port":587}`
	emailReg := utils.NewEMail(config)
	//内容配置
	emailReg.Subject = "品优购用户激活"
	emailReg.From = "759948611@qq.com"
	emailReg.To = []string{email}
	userName := this.Ctx.GetCookie("userName")
	emailReg.HTML = `<a href="http://127.0.0.1:8080/active?userName=` + userName + `">点击激活用户</a>`
	//发送邮件
	emailReg.Send()
	o := orm.NewOrm()
	var user models.User
	user.Name = userName

	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error("未查询到用户")
		return
	}
	user.Email = email
	o.Update(&user)
	this.Ctx.WriteString("小伙子你很优秀哟")
}
//激活用户
func (this *UserController) ActiveUser() {
	userName := this.GetString("userName")
	if userName == "" {
		beego.Error("邮件未或取到用户名")
		this.Redirect("/register-email", 302)
		return
	}
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error("没有此用户", err)
		this.Redirect("/register-email", 302)
		return
	}
	user.Active = true
	o.Update(&user, "Active")
	this.TplName = "login.html"
}
//展示的登陆界面
func (this *UserController) ShowLogin() {
	userName := this.Ctx.GetCookie("loginName")
	beego.Info("yonghuming:",userName)
	if userName != "" {
		this.Data["checked"] = "checked"
	} else {
		this.Data["checked"] = ""
	}
	this.Data["userName"] = userName
	this.TplName = "login.html"
}
//用户登陆
func (this *UserController) Login() {
	userName := this.GetString("userName")
	pwd := this.GetString("pwd")
	if userName == "" || pwd == "" {
		beego.Error("用户信息不完整")
		this.TplName = "login.html"
		return
	}
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error("用户名不正确")
		this.TplName = "login.html"
		return
	}
	user.Pwd = pwd
	err = o.Read(&user, "Pwd")
	if err != nil {
		beego.Error("密码不正确")
		this.TplName = "login.html"
		return
	}
	if user.Active != true{
		this.Data["errmsg"] = "该用户没有激活，请县激活！"
		this.TplName = "login.html"
		return
	}
	check:= this.GetString("m1")
	if check == "2"{
		this.Ctx.SetCookie("loginName", user.Name, 60*10)
	}else {
		this.Ctx.SetCookie("loginName",user.Name,-1)
	}

	this.SetSession("userName",userName)
	this.Redirect("/index",302)
}
//推出登陆
func (this*UserController)Logout(){
	this.DelSession("userName")
	this.Redirect("/index",302)
}
//展示用户中心项
func (this*UserController)ShowUserCenterInfo(){
	userName:= this.GetSession("userName")
	var user models.User
	o:=orm.NewOrm()
	beego.Info(userName)
	user.Name=userName.(string)
	err:= o.Read(&user,"Name")
	phone:=user.Name
	sphone:= strings.Split(phone,"")
	for i:=3;i<7;i++{
		//phone= strings.Replace(phone,sphone[i],"*",1)
		sphone[i]="*"
	}
	phone= strings.Join(sphone,"")

	beego.Info(sphone)
	//err := o.QueryTable("User").Filter("Name", userName.(string)).One(&user)
	beego.Info(user)
	if err!=nil{
		beego.Error("未找到用户信息")
		this.Redirect("/index",302)
		return
	}
	var address models.Address
	qs:=o.QueryTable("Address").RelatedSel("User").Filter("User__Name",userName)
	err= qs.Filter("IsDefault",true).One(&address)

	//获取用户登陆时候的浏览记录、
	conn,err:= redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		beego.Info("连接redis失败")
		return
	}
	defer conn.Close()
	goodsIds,_:= redis.Ints(conn.Do("lrange","history_"+userName.(string),0,4))
	var goods []models.GoodsSKU
	for _,v:=range goodsIds{
		var goodsSku models.GoodsSKU
		goodsSku.Id=v
		o.Read(&goodsSku)
		goods=append(goods,goodsSku)
	}

	this.Data["goods"]=goods


	this.Data["userName"]=userName.(string)
	this.Data["address"]=address
	this.Data["phone"]=phone
	this.Data["user"]=user
	this.Data["tplName"]="1"
	this.Layout="layout.html"
	this.TplName="user_center_info.html"
}
//展示用户地址信息
func (this*UserController)ShowUserSite(){
	userName:= this.GetSession("userName").(string)
	o:=orm.NewOrm()
	var address models.Address
	qs:=o.QueryTable("Address").RelatedSel("User").Filter("User__Name",userName)
	qs.Filter("IsDefault",true).One(&address)
	this.Data["address"]=address
	this.Data["userName"]=userName
	this.Data["tplName"]="3"
	this.Layout="layout.html"
	this.TplName="user_center_site.html"
}
//给用户添加地址信息
func (this*UserController)HandleAddAddr(){
	receiver:= this.GetString("receiver")
	addr:= this.GetString("addr")
	postCode:= this.GetString("postCode")
	phone:= this.GetString("phone")
	if receiver==""||addr==""||postCode==""||phone==""{
		beego.Error("地址信息不完整")
		this.TplName="user_center_site.html"
		return
	}
	var user models.User
	userName:= this.GetSession("userName")
	o:=orm.NewOrm()
	user.Name=userName.(string)
	o.Read(&user,"Name")
	var addres models.Address
	addres.Receiver=receiver
	addres.Addr=addr
	addres.PostCode=postCode
	addres.Phone=phone
	addres.User=&user
	//查询有没有默认值为true的,如果有的话就改为false
	var oldAddres models.Address

	qs:=o.QueryTable("Address").RelatedSel("User").Filter("User__Name",userName.(string))
	err:= qs.Filter("IsDefault",true).One(&oldAddres)
	if err==nil{
		oldAddres.IsDefault=false
		o.Update(&oldAddres,"IsDefault")
	}
	addres.IsDefault=true
	_,err= o.Insert(&addres)
	if err!=nil{
		beego.Error("插入地址信息失败")
		this.TplName="user_center_site.html"
		return
	}
	this.Redirect("/user/userSite",302)
}
////用户中心订单的展示
func (this*UserController)ShowUserOrder(){
	userName:= this.GetSession("userName")
	if userName==""{
		beego.Error("用户未登陆")
		this.Redirect("/index",302)
		return
	}
	//获取订单信息
	o:=orm.NewOrm()
	var orderInfos []models.OrderInfo
	o.QueryTable("OrderInfo").RelatedSel("User").Filter("User__Name",userName.(string)).OrderBy("-Time").All(&orderInfos)
	//定义宗容器
	var orders []map[string]interface{}

	for _,v:=range orderInfos{
		temp:=make(map[string]interface{})
		//获取当前订单的所有商品
		var orderGoods []models.OrderGoods
		o.QueryTable("OrderGoods").RelatedSel("OrderInfo","GoodsSKU").Filter("OrderInfo__Id",v.Id).All(&orderGoods)

		temp["orderGoods"]=orderGoods
		temp["orderInfo"]=v
		orders=append(orders,temp)
	}

	this.Data["orders"]=orders
	this.Data["userName"]=userName.(string)
	this.Data["tplName"]="2"
	this.Layout="layout.html"
	this.TplName="user_center_order.html"
}
//展示用户购物车
func (this *UserController)ShowUserCart(){
	userName:= this.GetSession("userName")
	if userName==""{
		beego.Error("用户没有登陆")
		this.Redirect("/index",302)
		return
	}
	this.Data["userName"]=userName.(string)

	this.Layout="user_center_head.html"
	this.TplName="cart.html"
}


