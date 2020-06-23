package server

import (
	"Pixiv/src/pixiv"
	"Pixiv/src/static"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type ginServer struct {
	g *gin.Engine
	c *cron.Cron
}

type Site struct {
	Daily   string
	Weekly  string
	Monthly string
}

func firstLoad(d, w, m *pixiv.Pixiv) {
	d.Crawl()
	w.Crawl()
	m.Crawl()

}

func New() *ginServer {
	return &ginServer{}
}

func exportStatic() {
	dirs := []string{"view"} // 设置需要释放的目录
	isSuccess := true
	for _, dir := range dirs {
		if err := static.RestoreAssets("./", dir); err != nil {
			isSuccess = false
			break
		}
	}
	if !isSuccess {
		for _, dir := range dirs {
			os.RemoveAll(filepath.Join("./", dir))
		}
	}
}

func (s *ginServer) InitServer() {
	daily := pixiv.New("daily")
	weekly := pixiv.New("weekly")
	monthly := pixiv.New("monthly")
	exportStatic()
	firstLoad(daily, weekly, monthly)
	s.c = cron.New()
	s.c.AddFunc("@daily", func() {
		daily.Crawl()
		weekly.Crawl()
		monthly.Crawl()
		fmt.Println("Daily Pre Download: " + s.c.Entries()[0].Prev.String())
		fmt.Println("Daily Next Download: " + s.c.Entries()[0].Next.String())
	})

	s.g = gin.Default()
	s.g.Static("/static", "./view/static")
	s.g.Static("/Pixiv", "./PixivDownload")
	s.g.LoadHTMLGlob("./view/html/index.html")
	s.g.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Pictures": pixiv.LoadPictures(*daily),
			"Site":     &Site{Daily: "true"},
		})
	})

	s.g.GET("/daily", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Pictures": pixiv.LoadPictures(*daily),
			"Site":     Site{Daily: "true"},
		})
	})

	s.g.GET("/monthly", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Pictures": pixiv.LoadPictures(*monthly),
			"Site":     Site{Monthly: "true"},
		})
	})

	s.g.GET("/weekly", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Pictures": pixiv.LoadPictures(*weekly),
			"Site":     Site{Weekly: "true"},
		})
	})
}

func (s *ginServer) Start() {
	s.c.Start()
	defer s.c.Stop()
	endless.ListenAndServe("0.0.0.0:8081", s.g)
}
