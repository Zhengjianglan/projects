package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hu17889/go_spider/core/common/page"
	"github.com/hu17889/go_spider/core/pipeline"
	"github.com/hu17889/go_spider/core/spider"
	//"os"
	//"strconv"
	"strings"
)

type MyPageProcesser struct {
}

func NewMyPageProcesser() *MyPageProcesser {
	return &MyPageProcesser{}
}

var adder int = 0

// Parse html dom here and record the parse result that we want to Page.
// Package goquery (http://godoc.org/github.com/PuerkitoBio/goquery) is used to parse html.
func (this *MyPageProcesser) Process(p *page.Page) {
	if !p.IsSucc() {
		println(p.Errormsg())
		return
	}

	query := p.GetHtmlParser()

	var urls []string
	var name string

	//fileOperation
	/*
		userFile := strconv.Itoa(adder)
		adder++
		fmt.Println("userfile:" + userFile)
		fout, err := os.Create("html.txt" + userFile)
		defer fout.Close()
		if err != nil {
			fmt.Println(userFile, err)
			return
		}
		for i := 0; i < 10; i++ {
			fout.WriteString(p.GetBodyStr())
			fout.Write([]byte("Just a test!\r\n"))
		}
	*/

	//job
	query.Find("li[class='con_list_item default_list']").Each(func(i int, s *goquery.Selection) {
		salary, _ := s.Attr("data-salary")
		company, _ := s.Attr("data-company")
		positionname, _ := s.Attr("data-positionname")
		exp := s.Find("div[class='list_item_top'] div['position'] div['p_bot']").Text()
		remark := ""
		s.Find("div[class='list_item_bot']").Each(func(j int, js *goquery.Selection) {
			//fmt.Println(j)
			//fmt.Println(js.Html())
			if j > 0 {
				return
			}
			//fmt.Println("bot")
			js.Find("div[class='li_b_l']").Each(func(lib int, libs *goquery.Selection) {
				if lib > 0 {
					return
				}
				//fmt.Println("lib")
				libs.Find("span").Each(func(sp int, sps *goquery.Selection) {
					//fmt.Println("span")
					//fmt.Println("Js:" + sps.Text())
					remark = remark + sps.Text() + ","
				})
			})
		})
		fmt.Println(salary, company, positionname, exp, remark)
		InsertJob(positionname, "", salary, exp, company, remark)
	})

	//pagelist
	query.Find("div[class='pager_container'] a").Each(func(i int, s *goquery.Selection) {
		rel, _ := s.Attr("rel")
		if rel == "nofollow" {
			return
		}
		ut := "page"
		href, _ := s.Attr("href")
		href = "http:" + href
		fmt.Println("----- spider href:" + href + ",name:" + name)
		if !ExistsUrl(ut, href) {
			urls = append(urls, href)
			InsertUrl("page", href, "")
			p.AddTargetRequests(urls, "html")
		}
	})

	//menu
	fmt.Println("beforemenu")
	query.Find("div[class='menu_box']").Each(func(i int, hs *goquery.Selection) {
		if i > 0 {
			return
		}
		hs.Find("a[data-lg-tj-id]").Each(func(k int, s *goquery.Selection) {
			fmt.Println("menu")
			href, _ := s.Attr("href")
			href = "http:" + href
			name = s.Text()
			urls = append(urls, href)
			fmt.Println("----- spider href:" + href + ",name:" + name)
			InsertUrl("menu", href, name)
			p.AddTargetRequests(urls, "html")
		})
	})
	// these urls will be saved and crawed by other coroutines.
	//p.AddTargetRequests(urls, "html")

	//name := query.Find(".entry-title .author").Text()
	name = strings.Trim(name, " \t\n")
	//repository := query.Find(".entry-title .js-current-repository").Text()
	//repository = strings.Trim(repository, " \t\n")
	//readme, _ := query.Find("#readme").Html()
	if name == "" {
		p.SetSkip(true)
	}
	// the entity we want to save by Pipeline
	//p.AddField("author", name)
	//p.AddField("project", repository)
	//p.AddField("readme", readme)
}

func (this *MyPageProcesser) Finish() {
	fmt.Printf("TODO:before end spider \r\n")
}

func main() {
	// Spider input:
	//  PageProcesser ;
	//  Task name used in Pipeline for record;
	/*
		if ExistsUrl("menu", "http://www.lagou.com/zhaopin/touzi/") {
			fmt.Println("1")
		}
	*/

	spider.NewSpider(NewMyPageProcesser(), "TaskName").
		AddUrl("https://www.lagou.com/", "html").   // Start url, html is the responce type ("html" or "json" or "jsonp" or "text")
		AddPipeline(pipeline.NewPipelineConsole()). // Print result on screen
		SetThreadnum(3).                            // Crawl request by three Coroutines
		Run()
}

func insertTest() {
	db, err := sql.Open("mysql", "zheng:123456@/spider?charset=utf8&collation=utf8_general_ci")
	checkErr(err)
	//db.Exec("set names utf8")
	stmt, err := db.Prepare(`INSERT urls1 (utype,url,title) values (?,?,?)`)
	checkErr(err)
	fmt.Println("test", "链接", "标题")
	_, err = stmt.Exec("test", "链接", "标题")
	checkErr(err)
	//id, err := res.LastInsertId()
	//checkErr(err)
	db.Close()
	//fmt.Println(id)
}

func queryTest() bool {
	db, err := sql.Open("mysql", "zheng:123456@/spider")
	checkErr(err)
	rows, err := db.Query("select url,title from urls1 ")
	count := 0
	for rows.Next() {
		count = 1
		var url string
		var title string
		var remark string
		if err := rows.Scan(&url, &title); err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(url, title, remark)
	}
	checkErr(err)
	db.Close()
	return count == 1
}

//db, _ := sql.Open("mysql", "zheng:123456@/spider?charset=utf8")

func InsertUrl(t string, url string, title string) int {
	db, err := sql.Open("mysql", "zheng:123456@/spider?charset=utf8")
	checkErr(err)
	stmt, err := db.Prepare(`INSERT urls1 (utype,url,title) values (?,?,?)`)
	checkErr(err)
	fmt.Println(t, url, title)
	res, err := stmt.Exec(t, url, title)
	checkErr(err)
	id, err := res.LastInsertId()
	checkErr(err)
	db.Close()
	fmt.Println(id)
	return 0
}

func ExistsUrl(t string, u string) bool {
	db, err := sql.Open("mysql", "zheng:123456@/spider?charset=utf8")
	checkErr(err)
	rows, err := db.Query("select 1 from urls1 where utype='" + t + "' and url='" + u + "'")
	count := 0
	for rows.Next() {
		count = 1
	}
	checkErr(err)
	db.Close()
	return count == 1
}

func InsertJob(name string, jtype string, salary string, exp string, company string, remark string) int64 {
	db, err := sql.Open("mysql", "zheng:123456@/spider?charset=utf8")
	checkErr(err)
	stmt, err := db.Prepare(`INSERT jobs (jobname,jobtype,salary,company,exp,remark) values (?,?,?,?,?,?)`)
	checkErr(err)
	//fmt.Println(t, url)
	res, err := stmt.Exec(name, jtype, salary, company, exp, remark)
	checkErr(err)
	id, err := res.LastInsertId()
	checkErr(err)
	db.Close()
	fmt.Println(id)
	return id
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
