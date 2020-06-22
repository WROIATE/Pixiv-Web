package pixiv

import "os"

type Pixiv struct {
	Mode        string
	Date        string
	DownloadDir string
	DataSwap    string
	Status      int
}

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

func New(mode string) *Pixiv {
	dir := "../static/picture/"
	os.Mkdir(dir, os.ModePerm)
	return &Pixiv{mode, DateFormat(mode), dir, "", 0}
}

type Picture struct {
	Id    string
	Path  string
	Title string
}
