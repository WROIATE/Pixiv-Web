package pixiv

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/sjson"

	"github.com/gocolly/colly"
	"github.com/kennygrant/sanitize"
	"github.com/tidwall/gjson"
)

func (p *Pixiv) scrapy(mode, id string) {
	c := p.newScrapy(mode)
	if mode != "search" {
		c.Visit(fmt.Sprintf("https://www.pixiv.net/ranking.php?mode=%s&content=illust&format=json", p.Mode))
	} else {
		c.Visit("https://www.pixiv.net/ajax/illust/" + id)
	}
	c.Wait()
}

//Crawl image by reload api
func (p *Pixiv) Crawl() {
	p.DataSwap = dataReader(p.DownloadDir)
	p.Date = DateFormat(p.Mode)
	p.scrapy("simple", "")
	p.DataSwap, _ = sjson.Set(p.DataSwap, "date."+p.Mode, p.Date)
	dataWriter(p.DataSwap, p.DownloadDir)
	if p.Status != 0 {
		log.Println(p.Mode + p.Date + " Have some download failed")
		p.Status = 0
	}
	p.DataSwap = ""
}

//GetImageWithStrict use strict mode crawl
func (p *Pixiv) GetImageWithStrict() {
	p.DataSwap = dataReader(p.DownloadDir)
	p.Date = DateFormat(p.Mode)
	if p.Date != gjson.Get(p.DataSwap, "date."+p.Mode).String() {
		p.scrapy("strict", "")
		p.DataSwap, _ = sjson.Set(p.DataSwap, "date."+p.Mode, p.Date)
		dataWriter(p.DataSwap, p.DownloadDir)
		if p.Status != 0 {
			log.Println(p.Mode + p.Date + fmt.Sprintf(" Have some download failed:%d", p.Status))
			p.Status = 0
		}
	} else {
		log.Println("Mode:" + p.Mode + " Already crawled today")
	}
	p.DataSwap = ""
}

//GetImage directly download image
func (p *Pixiv) GetImage() {
	p.DataSwap = dataReader(p.DownloadDir)
	p.Date = DateFormat(p.Mode)
	p.scrapy("", "")
	p.DataSwap, _ = sjson.Set(p.DataSwap, "date."+p.Mode, p.Date)
	dataWriter(p.DataSwap, p.DownloadDir)
	if p.Status != 0 {
		log.Println(p.Mode + p.Date + fmt.Sprintf(" Have some download failed:%d", p.Status))
		p.Status = 0
	}
	p.DataSwap = ""
}

func (p *Pixiv) newScrapy(mode string) *colly.Collector {
	c := colly.NewCollector(
		colly.MaxBodySize(1024*1024*1024),
		colly.Async(true),
		colly.AllowURLRevisit(),
		colly.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36`),
	)
	var mutex sync.Mutex
	c.SetRequestTimeout(600 * time.Second)
	c.Limit(&colly.LimitRule{Parallelism: 8})
	c.OnResponse(func(r *colly.Response) {
		if strings.Contains(r.Request.URL.String(), "ranking") {
			p.DataSwap, _ = sjson.Delete(p.DataSwap, "rank."+p.Mode)
			for i := 0; i < 50; i++ {
				id := gjson.GetBytes(r.Body, fmt.Sprintf("contents.%d.illust_id", i)).String()
				if id != "" {
					p.DataSwap, _ = sjson.Set(p.DataSwap, "rank."+p.Mode+".-1", "id="+id)
					if gjson.Get(p.DataSwap, "picture.id="+id).Exists() {
						//fmt.Println(fmt.Sprintf(`picture.id=%s.#(date.#(=="%s"))#.id`, id, mode+"-2020-4-11"))
						log.Println(id + " already exsited")
						if !gjson.Get(p.DataSwap, fmt.Sprintf(`picture.id=%s.date.#(=="%s")`, id, p.Mode+p.Date)).Exists() {
							p.DataSwap, _ = sjson.Set(p.DataSwap, "picture.id="+id+".date.-1", p.Mode+p.Date)
							log.Println("update date")
						}
					} else {
						c.Visit("https://www.pixiv.net/ajax/illust/" + id)
						p.Status++
					}
				}
			}
			if mode == "simple" {
				if p.Status == 0 {
					p.Msg <- 0
				}
			}
		} else if strings.Contains(r.Request.URL.String(), "ajax") {
			if gjson.GetBytes(r.Body, "error").String() == "false" {
				url := gjson.GetBytes(r.Body, "body.urls.original").String()
				id := gjson.GetBytes(r.Body, "body.illustId").String()
				name := gjson.GetBytes(r.Body, "body.illustTitle").String()
				log.Printf("get id=%s title:%s\n", id, name)
				r.Ctx.Put("id", id)
				r.Ctx.Put("name", name)
				c.Request("GET", url, nil, r.Ctx, nil)
			}
		} else {
			ext := filepath.Ext(r.Request.URL.String())
			cleanExt := sanitize.BaseName(ext)
			log.Println("Downloading " + r.Ctx.Get("id"))
			fileName := p.DownloadDir + fmt.Sprintf("%s.%s", r.Ctx.Get("id"), cleanExt[1:])
			if r.Save(fileName) != nil {
				log.Println("picture write error")
			} else {
				mutex.Lock()
				if mode != "search" {
					p.DataSwap = setMetaData(p.DataSwap, picture{r.Ctx.Get("id"), p.Mode + p.Date, r.Ctx.Get("name"), fmt.Sprintf("%s.%s", r.Ctx.Get("id"), cleanExt[1:])})
					p.Status--
				} else {
					p.DataSwap = setMetaData(p.DataSwap, picture{r.Ctx.Get("id"), "cache", r.Ctx.Get("name"), fmt.Sprintf("%s.%s", r.Ctx.Get("id"), cleanExt[1:])})
				}
				CompressImg("./thumbnail/", p.DownloadDir, fmt.Sprintf("%s.%s", r.Ctx.Get("id"), cleanExt[1:]))
				log.Println(r.Ctx.Get("id") + fmt.Sprintf(" Download finished, Remaining num:%d", p.Status))
				mutex.Unlock()
				if mode == "simple" {
					p.Msg <- p.Status
				}
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
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", string(r.Body), "\nError:", err)
		if mode == "search" {
			p.Status = -1
		} else if mode == "simple" {
			p.Status--
			p.Msg <- p.Status
		}
	})
	return c
}

func getBySearchID(id string, p *Pixiv) int {
	p.DataSwap = dataReader(p.DownloadDir)
	p.scrapy("search", id)
	dataWriter(p.DataSwap, p.DownloadDir)
	p.DataSwap = ""
	return p.Status
}
