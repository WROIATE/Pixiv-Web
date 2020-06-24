package server

import (
	"Pixiv/src/pixiv"
	"Pixiv/src/static"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fvbock/endless"
	"github.com/gin-contrib/gzip"
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

func LoadStatic(d, w, m *pixiv.Pixiv) {
	d.Crawl()
	w.Crawl()
	m.Crawl()
	pixiv.DeleteTmp()
	pixiv.DecodeTar(*d)
	pixiv.DecodeTar(*w)
	pixiv.DecodeTar(*m)

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
	LoadStatic(daily, weekly, monthly)
	s.c = cron.New()

	s.c.AddFunc("@daily", func() {
		LoadStatic(daily, weekly, monthly)
		log.Println("Daily Pre Download: " + s.c.Entries()[0].Prev.String())
		log.Println("Daily Next Download: " + s.c.Entries()[0].Next.String())
	})
	gin.SetMode(gin.ReleaseMode)
	s.g = gin.Default()
	s.g.Use(gzip.Gzip(gzip.DefaultCompression))
	s.g.Static("/static", "./view/static")
	s.g.Static("/Pixiv", "./PixivDownload")
	s.g.LoadHTMLGlob("./view/html/index.html")
	s.g.GET("/", func(c *gin.Context) {
		c.Request.URL.Path = "/daily"
		s.g.HandleContext(c)
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

	s.g.GET("/download/:mode", func(c *gin.Context) {
		mode := c.Param("mode")
		var p pixiv.Pixiv
		if mode == "daily" {
			p = *daily
		} else if mode == "weekly" {
			p = *weekly
		} else if mode == "monthly" {
			p = *monthly
		} else {
			log.Println(c.Request.Header)
			return
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Header("Content-Disposition", "attachment; filename="+p.Mode+p.Date+".tar")
		c.File("./tmp/" + p.Mode + p.Date + ".tar")
	})
}

func (s *ginServer) Start() {
	s.c.Start()
	defer s.c.Stop()
	endless.ListenAndServe("0.0.0.0:8081", s.g)
}
