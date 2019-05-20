package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
	"strconv"
	"github.com/gomodule/redigo/redis"
	"time"
	"strings"
	"github.com/smartwalle/alipay"
)

type OrderController struct {
	beego.Controller
}

func(this*OrderController)ShowOrder(){
	//获取数据
	goodsIds := this.GetStrings("checkGoods")

	//校验数据
	if len(goodsIds) == 0 {
		this.Redirect("/user/showCart",302)
		return
	}
	//处理数据
	//获取当前用户的所有收货地址
	name := this.GetSession("userName")

	o := orm.NewOrm()
	var addrs []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name",name.(string)).All(&addrs)
	this.Data["addrs"] = addrs

	conn,_ := redis.Dial("tcp","127.0.0.1:6379")

	//获取商品,获取总价和总件数
	var goods []map[string]interface{}
	var totalPrice ,totalCount int

	for _,v := range goodsIds{
		temp := make(map[string]interface{})
		id,_ := strconv.Atoi(v)
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)
		//获取商品数量
		count,_ := redis.Int(conn.Do("hget","cart_"+name.(string),id))
		//计算小计
		littlePrice := count * goodsSku.Price
		//把商品信息放到行容器
		temp["goodsSku"] = goodsSku
		temp["count"] = count
		temp["littlePrice"] = littlePrice

		totalPrice += littlePrice
		totalCount += 1

		goods = append(goods,temp)

	}

	//返回数据
	this.Data["totalPrice"] = totalPrice
	this.Data["totalCount"] = totalCount
	this.Data["truePrice"] = totalPrice + 10
	this.Data["goods"] = goods
	this.Data["goodsIds"]=goodsIds
	this.TplName = "place_order.html"
}
//上传数据
func (this*OrderController)HandlePushOrder(){
	//addrId":addrId,"payId":payId,"goodsIds":goodsIds,"totalCount":totalCount,"totalPrice":totalPrice
	addrId,err1:= this.GetInt("addrId")
	payId,err2:=this.GetInt("payId")
	goodsIds:=this.GetString("goodsIds")
	totalCount,err3:=this.GetInt("totalCount")
	totalPrice,err4:=this.GetInt("totalPrice")

	resp:=make(map[string]interface{})
	defer RespFunc(&this.Controller,resp)

	name:= this.GetSession("userName")
	if name==nil{
		resp["errno"]=2
		resp["errmsg"]="当前用户未登陆"
		return
	}
	if err1!=nil||err2!=nil||err3!=nil||err4!=nil||goodsIds==""{
		resp["errno"]=1
		resp["errmsg"]="数据传输不完整"
		return
	}
	//处理数据
	//把数据插入到mysql数据库中
	//获取用户对象和地址对象
	o:=orm.NewOrm()
	var user models.User
	user.Name=name.(string)
	o.Read(&user,"Name")

	var address models.Address
	address.Id=addrId
	o.Read(&address)

	var orderInfo models.OrderInfo
	orderInfo.User=&user
	orderInfo.Address=&address
	orderInfo.PayMethod=payId
	orderInfo.TotalCount=totalCount
	orderInfo.TotalPrice=totalPrice
	orderInfo.TransitPrice=10
	orderInfo.OrderId=time.Now().Format("20060102150405"+strconv.Itoa(user.Id))
	//开始事物
	o.Begin()
	o.Insert(&orderInfo)
	conn,_:= redis.Dial("tcp","127.0.0.1:6379")
	defer conn.Close()

	goodsSlices:=strings.Split(goodsIds[1:len(goodsIds)-1]," ")
	for _,v:= range goodsSlices{
		//插入订单商品列表

		//获取商品信息
		id,_:=strconv.Atoi(v)
		var goodsSku models.GoodsSKU
		goodsSku.Id=id
		o.Read(&goodsSku)

		//获取商品数量
		oldStork:=goodsSku.Stock
		beego.Info("原始库存等于:",oldStork)
	 	count,_:= redis.Int(conn.Do("hget","cart_"+name.(string),id))
		//获取小计
		littlePrice := count* goodsSku.Price
		//插入
		var orderGoods models.OrderGoods
		orderGoods.OrderInfo=&orderInfo
		orderGoods.GoodsSKU=&goodsSku
		orderGoods.Count=count
		orderGoods.Price=littlePrice
		//插入之前要先更新库存和销量
		if goodsSku.Stock<count{
			resp["errno"]=4
			resp["errmsg"]="库存不足"
			o.Rollback()
			return
		}
		//goodsSku.Stock-=count
		//goodsSku.Sales+=count
		time.Sleep(time.Second*5)
		o.Read(&goodsSku)

		qs:=o.QueryTable("GoodsSKU").Filter("Id",id).Filter("Stock",oldStork)
		_,err:= qs.Update(orm.Params{"Stock":goodsSku.Stock-count,"Sales":goodsSku.Sales+count})
		if err!=nil{
			resp["errno"]=7
			resp["errmsg"]="购买失败，清重新派对"
			o.Rollback()
			return
		}


		_,err=o.Insert(&orderGoods)
		if err!=nil{
			resp["errno"]=3
			resp["errmsg"]="服务器异常"
			o.Rollback()
			return
		}
		_,err=conn.Do("hdel","cart_"+name.(string),id)
		beego.Info(err)//缓存   项目未编译成功  未知错误
		//返回数据
		//提交事物
		o.Commit()
		resp["errno"]=5
		resp["errmsg"]="OK"

	}
}

func (this*OrderController)Pay(){
	orderId,err:=this.GetInt("orderId")
	if err != nil {
		this.Redirect("/user/userOrder",302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var orderInfo models.OrderInfo
	orderInfo.Id = orderId
	o.Read(&orderInfo)

	//支付


	//appId, aliPublicKey, privateKey string, isProduction bool
	publiKey :=`MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAya7JWL8pYP2eXQ7gHZFq 8gYKUvFQL6LDNnnS4uqsBAWPZqLAc+fmlRTUpIDV9Ve2EZq0tdHfoxNuzLsSoCh7 dahYXMPjzXU3+W4RzVR5IsNElwVWfRocU+lzEB06z9zXAz9HhqdmoLEYa7OrTib3 B4wJLmkRhWd3TmiqlbUEO4kc96cmpavQdWJyRMkPhwUK5yZc98nQFbM5WIq/o4tj QLLPGqUn72+DSZuHt96xzOrf9hIsPKG5SvWa8xJe+ppO1hi8q7/+Rhrh5UCtRGc+ fZlshif7zsznKfJwxCQ03PTTxg0pKABAXRzUVnE4KXaaEEIGKzg2jWdHgR+lSjKF qwIDAQAB`
	privatePay:=`MIIEowIBAAKCAQEAya7JWL8pYP2eXQ7gHZFq8gYKUvFQL6LDNnnS4uqsBAWPZqLA
c+fmlRTUpIDV9Ve2EZq0tdHfoxNuzLsSoCh7dahYXMPjzXU3+W4RzVR5IsNElwVW
fRocU+lzEB06z9zXAz9HhqdmoLEYa7OrTib3B4wJLmkRhWd3TmiqlbUEO4kc96cm
pavQdWJyRMkPhwUK5yZc98nQFbM5WIq/o4tjQLLPGqUn72+DSZuHt96xzOrf9hIs
PKG5SvWa8xJe+ppO1hi8q7/+Rhrh5UCtRGc+fZlshif7zsznKfJwxCQ03PTTxg0p
KABAXRzUVnE4KXaaEEIGKzg2jWdHgR+lSjKFqwIDAQABAoIBAHWx13Q/0kD0oEcE
GEuJAhUM61dt1XKHFE6lFyku84VBTWcC0TaSfSBA0LrTKIkNT4XEd3KncE+1VnKJ
NRxbo3iM7kvsUPMkVa7sydy+UBa2Rz+ucMN+yks5r9kWhayp3pvZgL8Wz7A0yOCJ
MA3idURlNzJrRDtKnvgV4n6M7VRj7Urx+7qXkvg9IOob6tz3udrGsIk8hrqLES1L
BL1P3iBkSZVqE6O6zlvMfZ35w/Se0HfE4r1rY55neBl4buyhwTAlOv6Lu8FVBm5S
LfndPdM2CGcJrAm4GMteWGmdBVNNTC7va1ixs+ofLHa31uV7LiXGZScBvAeGKVmO
kg7qiUECgYEA6HYIPQgxxnyQXDrM3mHzG24hwSdiX7WJ3mtix1UpQoCCSUeqCirh
8ujHKKfDV6WeaYTHaQ5UIXqrrSzUD9yVVvsvIWTURX/qvQUImJfjB6P7Dx9OPNF5
ZFjbiz/ZKfKmGRlMZD/9P9FhgjTw2pVNk0EVE/txVxHYsRYlEbqKf8sCgYEA3hrn
gEXT25YB/5tcFf1CPrWPCeYf6RUX3c/zRTUtchjsyh14dIaj3eMW9srRZOM5totI
AE/wIPtlQXj+pIxfp8rpbMVTf6gRZRRRkX2h5yAQXGP6fbXKNqBI64BOt8cMN62G
gxtKrkSqDcu2j2k9ZeFHQmhi16EUnACXBGFFlaECgYEAwD3MdlyufU1KPVvLTSWH
3OlpNMmTSz9gcvYvzFUbOAn6tQt1Dc+E6FOlUHPc0kD/DphmKPVWkhFWpHJsNWng
fvxfb6ho+8jbodHl1/vUHt93onvrQdSiJWBuv2vf9hbbUepgCI/6qapIj1uky7+p
Vdv+yHWqt6zknR6JLW4tV50CgYAP4KJ+A//iKbYY3LVXiRRMQVRpY78SPYTIQY5l
eyi1iFydEkBDLEDYotxIZjVT3f6JMynBg/Vpli8l4A1sG/DWoOXQ9cZDUPN1Y191
ZCLHz/37bNZCWFWNVCYCV9jIwHz6GfiMtM3A6X2yoMZ7OA3Ak7sxXx75xTUg9dXV
5VJBIQKBgH0A6AtwBPkkjXqCvWG4oRiWlVEQ3fzFq8JE1TKOH+o+KqKVjE/zwvoS
MRddrKU5MlJjGtHUIhHAamGjDPko65NR3hWuBa0cwyi4GdrKlTqErzrni9EHJrVG
/+radalxvRE0xJ5sTlcOLM9t3q+GczCBeQ+nvSaMQL8QoVFctVdv
`
	client:=alipay.New("2016093000634090",publiKey,privatePay,false)
	var p = alipay.TradePagePay{}
	p.NotifyURL="http://192.168.182.111:8080/payOk"
	p.ReturnURL="http://192.168.182.111:8080/payOk"
	p.Subject="品优购"
	p.OutTradeNo=orderInfo.OrderId
	p.TotalAmount=strconv.Itoa(orderInfo.TotalPrice)
	p.ProductCode= "FAST_INSTANT_TRADE_PAY"
	url,err:=client.TradePagePay(p)
	if err!=nil{
		beego.Error("支付失败")
	}
	payUrl:=url.String()
	this.Redirect(payUrl,302)

}