package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robfig/cron"
	"log"
	"time"
)

func main()  {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	db, err := sql.Open("sqlite3", "./url.db")
	if err != nil {
		log.Fatal(err)
	}
	i := 0
	c := cron.New()
	spec := "0, 40, 6, *, *, *"
	c.AddFunc(spec, func() {
		//定时删除之前的数据库 一周前
		//取5天前的时间戳
		k := time.Now()
		sd, _ := time.ParseDuration("-24h")
		s_time := k.Add(sd * 5).Unix()
		stmt, err := db.Prepare("delete from urlinfo where timeUnix <= ?")
		if err != nil {
			log.Fatal(err)
		}
		res, err := stmt.Exec(s_time)
		affect, err := res.RowsAffected()
		fmt.Println(affect)
		i++
		log.Println("cron running:", i)
	})
	c.Start()
	r.POST("/api/midpage", func(c *gin.Context) {
		//获取参数
		url := c.PostForm("url")
		cms := c.PostForm("cms")
		ele := c.PostForm("ele")
		longurl := c.PostForm("longurl")
		//入库 返回url
		timeUnix:=time.Now().Unix()

		stmt, err := db.Prepare("INSERT INTO urlinfo(url, cms, ele, timeUnix, longurl) values(?,?,?,?,?)")
		if err != nil {
			log.Fatal(err)
		}
		stmt.Exec(url, cms, ele, timeUnix, longurl)
		c.JSON(200, gin.H{
			"result":timeUnix,
		})
	})
	//解析短连接渲染及解析短链
	r.GET("/tb/:timeUnix", func(c *gin.Context) {
		timeUnix := c.Param("timeUnix")
		var url string
		var cms string
		var longurl string
		var ele string
		//查询该时间戳对应的数据，如果是个短链则返回短链，否则则是是中间页 就渲染
		err := db.QueryRow("select url,cms,ele,longurl from urlinfo where timeUnix = ?  limit  1" , timeUnix).Scan(&url, &cms, &ele, &longurl)
		if err != nil {
			log.Fatal(err)
		}
		if longurl != "" {		//解析短连接
			c.JSON(200,gin.H{
				"code":200,
				"longurl":longurl,
			})
		}else {
			r.LoadHTMLFiles("html/index.tmpl")
			c.HTML(200,"index.tmpl", gin.H{
				"url":url,
				"ele":ele,
				"cms":cms,
			})

		}





	})


	//r.LoadHTMLFiles("html/index.tmpl")
	//c.HTML(200,"index.tmpl", gin.H{
	//	"url":url,
	//	"ele":ele,
	//})


	r.Run(":80")

}
