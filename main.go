package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robfig/cron"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

func main() {

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
		timeUnix := time.Now().Unix()
		shorturl := RandStringBytesMaskImprSrc(6)

		stmt, err := db.Prepare("INSERT INTO urlinfo(url, cms, ele, timeUnix, longurl, shorturl) values(?,?,?,?,?,?)")
		if err != nil {
			log.Fatal(err)
		}
		stmt.Exec(url, cms, ele, timeUnix, longurl, shorturl)
		c.JSON(200, gin.H{
			"result": shorturl,
		})
	})
	//解析短连接渲染及解析短链
	r.GET("/u/:shorturl", func(c *gin.Context) {
		shorturl := c.Param("shorturl")

		if shorturl == "false" {
			c.String(200, "网络错误，刷新一下页面")
		}

		var url string
		var cms string
		var longurl string
		var ele string
		//查询该时间戳对应的数据，如果是个短链则返回短链，否则则是是中间页 就渲染
		err := db.QueryRow("select url,cms,ele,longurl, shorturl from urlinfo where shorturl = ?  limit  1", shorturl).Scan(&url, &cms, &ele, &longurl, &shorturl)
		if err != nil {
			log.Print(err)
			return
		}
		if longurl != "" { //解析短连接
			c.Redirect(http.StatusMovedPermanently, longurl)
		} else {

			r.LoadHTMLFiles("html/index.tmpl")
			c.HTML(200, "index.tmpl", gin.H{
				"url": url,
				"ele": ele,
				"cms": cms,
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
