package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/sjson"

	"github.com/gocolly/colly"
	"github.com/kennygrant/sanitize"
	"github.com/tidwall/gjson"
)

type picture struct {
	id       string
	date     string
	title    string
	filename string
}

type transform struct {
	Title string
	Name  string
}

func getKeysAsValues(json string) gjson.Result {
	var sb strings.Builder
	sb.WriteByte('[')
	var once bool
	gjson.Parse(json).ForEach(func(key, value gjson.Result) bool {
		if !key.Exists() {
			return false
		}
		if once {
			sb.WriteByte(',')
		} else {
			once = true
		}
		sb.WriteString(key.Raw)
		return true
	})
	sb.WriteByte(']')
	return gjson.Parse(sb.String())
}

func getJson(mode, date string) []transform {
	list := []transform{}
	s := ""
	s = dataReader(s)
	gjson.Parse(gjson.Get(s, "picture").String()).ForEach(
		func(key, value gjson.Result) bool {
			if strings.HasPrefix(key.String(), "id=") {
				//println(key.String())
				if strings.Contains(value.String(), mode+date) {
					filename := gjson.Get(value.String(), "filename").String()
					title := gjson.Get(value.String(), "title").String()
					list = append(list, transform{Title: title, Name: filename})
				}
				//println(value.String())
			}
			return true
		})
	return list
}

func dataReader(data string) string {
	file, err := os.OpenFile("../static/picture/Pixiv.json", os.O_RDWR|os.O_CREATE, 0755)
	defer file.Close()
	if err != nil {
		fmt.Println("加载json错误")
		return err.Error()
	}
	b, _ := ioutil.ReadAll(file)
	if len(b) != 0 {
		data = fmt.Sprintf("%s", b)
	} else {
		data, _ = sjson.Set(data, "date.monthly", "")
		data, _ = sjson.Set(data, "date.weekly", "")
		data, _ = sjson.Set(data, "date.daily", "")
	}
	return data
}

func dataWriter(s string) {
	err := ioutil.WriteFile("../static/picture/Pixiv.json", []byte(s), 0664)
	if err != nil {
		fmt.Println("json文件写入错误")
	}
}

func setJson(s string, pic picture) string {
	s, _ = sjson.Set(s, "picture.id="+pic.id+".title", pic.title)
	s, _ = sjson.Set(s, "picture.id="+pic.id+".date.-1", pic.date)
	s, _ = sjson.Set(s, "picture.id="+pic.id+".filename", pic.filename)
	return s
}

func scrapy(data, mode, dir, date string) string {
	c := colly.NewCollector(
		colly.MaxBodySize(1024*1024*1024),
		colly.Async(true),
	)
	var mutex sync.Mutex
	c.SetRequestTimeout(60 * time.Second)
	c.Limit(&colly.LimitRule{Parallelism: 8})
	num := 0
	c.OnResponse(func(r *colly.Response) {
		if strings.Contains(r.Request.URL.String(), "ranking") {
			for i := 0; i < 50; i++ {
				id := gjson.GetBytes(r.Body, fmt.Sprintf("contents.%d.illust_id", i)).String()
				if gjson.Get(data, "picture.id="+id).Exists() {
					//fmt.Println(fmt.Sprintf(`picture.id=%s.#(date.#(=="%s"))#.id`, id, mode+"-2020-4-11"))
					fmt.Println(id + " already exsited")
					if !gjson.Get(data, fmt.Sprintf(`picture.id=%s.date.#(=="%s")`, id, mode+date)).Exists() {
						data, _ = sjson.Set(data, "picture.id="+id+".date.-1", mode+date)
						fmt.Println("update date")
					}
				} else {
					c.Visit("https://www.pixiv.net/ajax/illust/" + id)
				}
			}
		} else if strings.Contains(r.Request.URL.String(), "ajax") {
			url := gjson.GetBytes(r.Body, "body.urls.original").String()
			id := gjson.GetBytes(r.Body, "body.illustId").String()
			name := gjson.GetBytes(r.Body, "body.illustTitle").String()
			fmt.Printf("get id=%s title:%s\n", id, name)
			r.Ctx.Put("id", id)
			r.Ctx.Put("name", name)
			c.Request("GET", url, nil, r.Ctx, nil)
		} else {
			ext := filepath.Ext(r.Request.URL.String())
			cleanExt := sanitize.BaseName(ext)
			fmt.Println("Downloading " + r.Ctx.Get("id"))
			fileName := dir + fmt.Sprintf("%s.%s", r.Ctx.Get("id"), cleanExt[1:])
			if r.Save(fileName) != nil {
				fmt.Println("write error")
			} else {
				mutex.Lock()
				data = setJson(data, picture{r.Ctx.Get("id"), mode + date, r.Ctx.Get("name"), fmt.Sprintf("%s.%s", r.Ctx.Get("id"), cleanExt[1:])})
				num++
				mutex.Unlock()
				fmt.Println(r.Ctx.Get("id") + fmt.Sprintf(" Download finished,num:%d", num))
			}
		}
	})
	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		if strings.Contains(r.URL.String(), "i.pximg.net") {
			r.Headers.Set("Referer", r.URL.String())
			//fmt.Println("Visiting", r.URL.String())
		}
	})
	c.Visit(fmt.Sprintf("https://www.pixiv.net/ranking.php?mode=%s&content=illust&format=json", mode))
	//c.Visit("https://i.pximg.net/img-original/img/2020/03/13/07/36/14/80074611_p0.jpg")
	c.Wait()
	return data
}

func Crawl(mode string) {
	date := dateFormat(mode)
	dir := "../static/picture/"
	data := ""
	os.Mkdir(dir, os.ModePerm)
	data = dataReader(data)
	if dateFormat(mode) != gjson.Get(data, "date."+mode).String() {
		data = scrapy(data, mode, dir, date)
		data, _ = sjson.Set(data, "date."+mode, dateFormat(mode))
		dataWriter(data)
	}
	//fmt.Println(getJson(mode, date))
}

func WeekByDate(t time.Time) string {
	yearDay := t.YearDay()
	yearFirstDay := t.AddDate(0, 0, -yearDay+1)
	firstDayInWeek := int(yearFirstDay.Weekday())

	//今年第一周有几天
	firstWeekDays := 1
	if firstDayInWeek != 0 {
		firstWeekDays = 7 - firstDayInWeek + 1
	}
	var week int
	if yearDay <= firstWeekDays {
		week = 1
	} else {
		week = (yearDay-firstWeekDays)/7 + 2
	}
	return string(fmt.Sprintf("%d", week))
}

func dateFormat(mode string) string {
	if mode == "weekly" {
		return fmt.Sprintf("-%d-%s", time.Now().Year(), WeekByDate(time.Now()))
	} else if mode == "monthly" {
		return time.Now().Format("-2006-01")
	} else {
		return time.Now().Format("-2006-01-02")
	}
}
