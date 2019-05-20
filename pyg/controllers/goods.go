package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pyg/pyg/models"
	"math"
	"github.com/gomodule/redigo/redis"
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
//展示生鲜页面
func (this*GoodsController)ShowIndexSX(){
	//获取商品类型
	o:=orm.NewOrm()
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsType"]=goodsTypes
	//获取论波图
	var goodsBanners []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&goodsBanners)
	this.Data["goodsBanners"]=goodsBanners
	//获取促销商品
	var  promotionBanners []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&promotionBanners)
	this.Data["promotionBanners"]=promotionBanners
	//还要获取啥阿
	//
	//获取首页商品展示
	var goods []map[string]interface{}

	for _,v:=range goodsTypes{
		var textGoods []models.IndexTypeGoodsBanner
		var imageGoods []models.IndexTypeGoodsBanner
		qs:=o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").Filter("GoodsType__Id",v.Id).OrderBy("Index")
		qs.Filter("DisplayType",0).All(&textGoods)
		qs.Filter("DisplayType",1).All(&imageGoods)

		//定义行容器
		temp:=make(map[string]interface{})
		temp["goodsType"]=v
		temp["textGoods"]=textGoods
		temp["imageGoods"]=imageGoods

		goods=append(goods,temp)

	}
	this.Data["goods"]=goods
	this.Layout="index_sx.html"
	this.TplName="search_gouwuche.html"

}
//展示商品详情页面
func (this*GoodsController)ShowGoodsDetail(){
	id,err	:= this.GetInt("Id")
	if err!=nil{
		beego.Error("没找到商品id")
		this.Redirect("/index_sx",302)
		return
	}
	o:=orm.NewOrm()
	//或取类型
	var goodsType []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsType)
	this.Data["goodsType"]=goodsType

	var goodsSku models.GoodsSKU
/*	goodsSku.Id=id
	o.Read(&goodsSku)*/
	o.QueryTable("GoodsSKU").RelatedSel("Goods","GoodsType").Filter("Id",id).One(&goodsSku)
	//获取统一类型商品推荐
	var newGoods []models.GoodsSKU
	qs:=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name",goodsSku.GoodsType.Name)
	qs.OrderBy("-Time").Limit(2,0).All(&newGoods)

	//用户登陆的时候记录点开详情页面的数据
	name := this.GetSession("userName")
	if name !=nil{
		conn,err:= redis.Dial("tcp","127.0.0.1:6379")
		if err==nil{
			defer conn.Close()
			conn.Do("lrem","history_"+name.(string),0,id)
			_,err:= conn.Do("lpush","history_"+name.(string),id)
			beego.Info(err)
		}
	}


	this.Data["newGoods"]=newGoods
	this.Data["goodsSku"]=goodsSku
	this.Layout="detail.html"
	this.TplName="search_gouwuche.html"
}
//独立于beego框架之外的，在那个框架都可以写。
func PageEdit(pageCount int,pageIndex int)[]int{
	//不足五页
	var pages []int
	if pageCount < 5{
		for i:=1;i<=pageCount;i++{
			pages = append(pages,i)
		}
	}else if pageIndex <= 3{
		for i:=1;i<=5;i++{
			pages = append(pages,i)
		}
	}else if pageIndex >= pageCount -2{
		for i:=pageCount - 4;i<=pageCount;i++{
			pages = append(pages,i)
		}
	}else {
		for i:=pageIndex - 2;i<=pageIndex + 2;i++{
			pages = append(pages,i)
		}
	}

	return pages
}
//展示商品列表页
func(this*GoodsController)ShowList(){
	id,err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Error("类型不存在")
		this.Redirect("/index_sx",302)
		return
	}
	//获取商品类型
	o:=orm.NewOrm()
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsType"]=goodsTypes

	//获取统一类型商品推荐
	var newGoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id).OrderBy("-Time").Limit(2,0).All(&newGoods)

	this.Data["newGoods"]=newGoods


	//处理数据

	var goods []models.GoodsSKU
	sort := this.GetString("sort")
	//实现分页
	qs:=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id)
	//获取总页码
	count,_ := qs.Count()
	pageSize := 1
	pageCount := int(math.Ceil(float64(count) / float64(pageSize)))
	//获取当前页码
	pageIndex,err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	pages := PageEdit(pageCount,pageIndex)
	this.Data["pages"] = pages
	//获取上一页，下一页的值
	var prePage,nextPage int
	//设置个范围
	if pageIndex -1 <= 0{
		prePage = 1
	}else {
		prePage = pageIndex - 1
	}

	if pageIndex +1 >= pageCount{
		nextPage = pageCount
	}else {
		nextPage = pageIndex + 1
	}


	this.Data["prePage"] = prePage
	this.Data["nextPage"] = nextPage

	qs = qs.Limit(pageSize,pageSize*(pageIndex - 1))

	//获取排序
	if sort == ""{
		qs.All(&goods)
	}else if sort == "price"{
		qs.OrderBy("Price").All(&goods)
	}else {
		qs.OrderBy("-Sales").All(&goods)
	}

	this.Data["sort"] = sort

	//返回数据
	this.Data["pageIndex"]=int(pageIndex)
	this.Data["id"]=id
	this.Data["goods"] = goods
	this.Layout="list.html"
	this.TplName ="search_gouwuche.html"
}
//搜索商品
func(this*GoodsController)HandleSearchGoods(){
	search:= this.GetString("searchGoods")
	if search==""{
		beego.Error("输入为空")
		this.Redirect("index_sx",302)
		return
	}
	var goods []models.GoodsSKU
	o:=orm.NewOrm()
	o.QueryTable("GoodsSKU").Filter("Name__icontains",search).All(&goods)
	this.Data["goods"]=goods
	this.Layout="search.html"
	this.TplName="search_gouwuche.html"
}