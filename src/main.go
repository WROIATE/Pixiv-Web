package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

type Picture struct {
	Id    string
	Path  string
	Title string
}

type Site struct {
	Daily   string
	Weekly  string
	Monthly string
}

func NewPicture(title, id string) Picture {
	return Picture{
		Id:    strings.Split(id, ".")[0],
		Path:  fmt.Sprintf("/static/picture/%s", id),
		Title: title,
	}
}

func LoadPictures(mode string) []Picture {
	files := getJson(mode, dateFormat(mode))
	list := make([]Picture, 0, len(files))
	for _, f := range files {
		check := strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".png")
		if check {
			list = append(list, NewPicture(f.Title, f.Name))
		}
	}
	return list
}

func init() {
	Crawl("daily")
	Crawl("weekly")
	Crawl("monthly")
}

func restore() {
	dirs := []string{"static"} // 设置需要释放的目录
	isSuccess := true
	for _, dir := range dirs {
		// 解压dir目录到当前目录
		if err := RestoreAssets("../", dir); err != nil {
			isSuccess = false
			break
		}
	}
	if !isSuccess {
		for _, dir := range dirs {
			os.RemoveAll(filepath.Join("../", dir))
		}
	}
}

func main() {
	r := gin.Default()
	c := cron.New()
	restore()
	r.Static("/static", "../static")
	r.LoadHTMLGlob("../static/index.html")
	c.AddFunc("@daily", func() { Crawl("daily") })
	c.AddFunc("@weekly", func() { Crawl("weekly") })
	c.AddFunc("@monthly", func() { Crawl("monthly") })
	c.Start()
	defer c.Stop()
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Pictures": LoadPictures("daily"),
			"Site":     &Site{Daily: "true"},
		})
	})

	r.GET("/daily", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Pictures": LoadPictures("daily"),
			"Site":     Site{Daily: "true"},
		})
	})

	r.GET("/monthly", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Pictures": LoadPictures("monthly"),
			"Site":     Site{Monthly: "true"},
		})
	})

	r.GET("/weekly", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Pictures": LoadPictures("weekly"),
			"Site":     Site{Weekly: "true"},
		})
	})
	r.Run("0.0.0.0:8080") // listen and serve on 0.0.0.0:8080

}
