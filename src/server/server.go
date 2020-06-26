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
	"github.com/gorilla/websocket"
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

func LoadStatic() {
	daily.GetImageWithStrict()
	weekly.GetImageWithStrict()
	monthly.GetImageWithStrict()
	daily.DecodeTar()
	weekly.DecodeTar()
	monthly.DecodeTar()
}

func New() *ginServer {
	return &ginServer{}
}

var daily, weekly, monthly *pixiv.Pixiv

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
	daily = pixiv.New("daily")
	weekly = pixiv.New("weekly")
	monthly = pixiv.New("monthly")
	exportStatic()
	LoadStatic()
	s.c = cron.New()

	s.c.AddFunc("@daily", func() {
		LoadStatic()
		log.Println("Daily Pre Download: " + s.c.Entries()[0].Prev.String())
		log.Println("Daily Next Download: " + s.c.Entries()[0].Next.String())
	})
	//gin.SetMode(gin.ReleaseMode)
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

	s.g.GET("/reload/:mode", reload)

	s.g.GET("/download/:mode", func(c *gin.Context) {
		mode := c.Param("mode")
		var p pixiv.Pixiv
		switch mode {
		case "daily":
			p = *daily
		case "weekly":
			p = *weekly
		case "monthly":
			p = *monthly
		default:
			log.Println(c.Request.Header)
			c.JSON(403, gin.H{"error": "forbidden"})
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Header("Content-Disposition", "attachment; filename="+p.Mode+p.Date+".tar")
		c.File("./tmp/" + p.Mode + p.Date + ".tar")
	})
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func reload(c *gin.Context) {
	mode := c.Param("mode")
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	var p *pixiv.Pixiv
	switch mode {
	case "daily":
		p = daily
	case "weekly":
		p = weekly
	case "monthly":
		p = monthly
	default:
		log.Println(c.Request.Header)
		return
	}
	log.Println("reload")
	go p.Crawl()
	num := <-p.Msg
	total := num
	for num > 0 {
		err = ws.WriteJSON(gin.H{
			"num":   num,
			"total": total,
		})
		if err != nil {
			break
		}
		num = <-p.Msg
	}
	p.DecodeTar()
	err = ws.WriteJSON(gin.H{"num": 0, "total": total})
	if err != nil {
		return
	}
}

func (s *ginServer) Start() {
	s.c.Start()
	defer s.c.Stop()
	endless.ListenAndServe("0.0.0.0:8081", s.g)
}
