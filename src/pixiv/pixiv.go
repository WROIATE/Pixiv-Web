package pixiv

import "os"

//Pixiv the pixiv spider class
type Pixiv struct {
	Mode        string
	Date        string
	DownloadDir string
	DataSwap    string
	Status      int
	Msg         chan int
}

//DownloadPath "./PixivDownload/"
const DownloadPath string = "./PixivDownload/"

type picture struct {
	id       string
	date     string
	title    string
	filename string
}

type transform struct {
	Title  string
	Name   string
	Favour string
}

// New return a Pixiv pointer
func New(mode string) *Pixiv {
	dir := "./PixivDownload/"
	os.Mkdir(dir, os.ModePerm)
	return &Pixiv{mode, DateFormat(mode), dir, "", 0, make(chan int)}
}

//Picture site struct
type Picture struct {
	ID     string
	Path   string
	Title  string
	Origin string
	Favour string
	Local  string
}
