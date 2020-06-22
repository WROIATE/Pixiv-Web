package server

import (
	"Pixiv/src/pixiv"
	"fmt"
	"net/http"

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

func (s *ginServer) InitServer() {
	daily := pixiv.New("daily")
	weekly := pixiv.New("weekly")
	monthly := pixiv.New("monthly")
	firstLoad(daily, weekly, monthly)
	s.c = cron.New()
	s.c.AddFunc("@daily", func() {
		daily.Crawl()
		fmt.Println("daily Pre Download: " + s.c.Entries()[0].Prev.String())
		fmt.Println("daily Next Download: " + s.c.Entries()[0].Next.String())
	})
	s.c.AddFunc("@weekly", func() {
		weekly.Crawl()
		fmt.Println("weekly Pre Download: " + s.c.Entries()[1].Prev.String())
		fmt.Println("weekly Next Download: " + s.c.Entries()[1].Next.String())
	})
	s.c.AddFunc("@monthly", func() {
		monthly.Crawl()
		fmt.Println("monthly Pre Download: " + s.c.Entries()[2].Prev.String())
		fmt.Println("monthly Next Download: " + s.c.Entries()[2].Next.String())
	})
	s.g = gin.Default()
	s.g.Static("/static", "../../static")
	s.g.LoadHTMLGlob("../../static/index.html")
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
