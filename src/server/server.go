package server

import (
	"Pixiv/src/pixiv"
	"Pixiv/src/static"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

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

//Site information
type Site struct {
	Daily   string
	Weekly  string
	Monthly string
	Favour  string
	Search  string
}

// LoadStatic Run the normal crawl
func LoadStatic() {
	daily.GetImage()
	weekly.GetImage()
	monthly.GetImage()
	pixiv.CompressImgByMode(*daily)
	pixiv.CompressImgByMode(*weekly)
	pixiv.CompressImgByMode(*monthly)
	pixiv.DeleteTmp()
	daily.EncodeTar()
	weekly.EncodeTar()
	monthly.EncodeTar()
}

// New return a ginServer pointer
func New() *ginServer {
	return &ginServer{}
}

var daily, weekly, monthly *pixiv.Pixiv

func exportStatic() {
	dirs := []string{"view"}
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

//InitServer return a init server route group
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
	s.g.Static("/thumbnail", "./thumbnail")
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			if i%5 == 0 {
				return 1
			}
			return 0
		},
	}
	s.g.SetFuncMap(funcMap)
	s.g.LoadHTMLGlob("./view/html/*")

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

	s.g.GET("/favour", favourPage)

	s.g.GET("/search/:keywords", search)

	s.g.GET("/search", func(c *gin.Context) {
		c.HTML(http.StatusOK, "search.html", gin.H{
			"Pictures": pixiv.LoadSearchData(),
			"Site":     Site{Search: "true"},
		})
	})

	s.g.GET("/download/:mode", download)

	s.g.POST("/", setfavour)

	s.g.POST("/save", savePicture)

	s.g.POST("/clean", cleanCache)
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func download(c *gin.Context) {
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
	p.EncodeTar()
	err = ws.WriteJSON(gin.H{"num": 0, "total": total})
	if err != nil {
		return
	}
}

func setfavour(c *gin.Context) {

	var id, favour string
	fmt.Sscanf(c.PostForm("id"), "%s", &id)
	fmt.Sscanf(c.PostForm("favour"), "%s", &favour)
	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
	})
	if favour == "true" {
		pixiv.SetFavour(id)
	} else if favour == "false" {
		pixiv.RemoveFavour(id)
	}
}

func search(c *gin.Context) {
	var keyWords string
	keyWords = c.Param("keywords")
	response := func(list []pixiv.Picture) {
		c.HTML(http.StatusOK, "search.html", gin.H{
			"Pictures": pixiv.LoadSearchData(),
			"Result":   list,
			"keyWords": keyWords,
			"Site":     Site{Search: "true"},
		})
	}
	responseError := func() {
		c.HTML(http.StatusOK, "search.html", gin.H{
			"Pictures": pixiv.LoadSearchData(),
			"Site":     Site{Search: "true"},
		})
	}
	if match, _ := regexp.MatchString(`^(\d+)$`, keyWords); match {
		pic, err1 := pixiv.FindByID(keyWords)
		list, err2 := pixiv.FindByFileName(keyWords)
		resultList := make([]pixiv.Picture, 0)
		switch {
		case err1 == nil:
			resultList = append(resultList, pic...)
			log.Println("search by id success")
			fallthrough
		case err2 == nil:
			resultList = append(resultList, list...)
			log.Println("search by name success")
			response(resultList)
		default:
			{
				responseError()
			}
		}
	} else {
		if list, err := pixiv.FindByFileName(keyWords); err != nil {
			log.Println(err)
			responseError()
		} else {
			response(list)
		}
	}
}

func favourPage(c *gin.Context) {
	c.HTML(http.StatusOK, "other.html", gin.H{
		"Pictures": pixiv.LoadFavour(),
		"Site":     Site{Favour: "true"},
	})
}

func cleanCache(c *gin.Context) {
	pixiv.CleanSearchCache()
	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
	})
}

func savePicture(c *gin.Context) {
	var id string
	fmt.Sscanf(c.PostForm("id"), "%s", &id)
	pixiv.SaveByID(id)
	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
	})
}

// Start to listen server
func (s *ginServer) Start() {
	s.c.Start()
	defer s.c.Stop()
	endless.ListenAndServe("0.0.0.0:8081", s.g)
}
