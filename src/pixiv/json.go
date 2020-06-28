package pixiv

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

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

func getJson(p Pixiv) []transform {
	list := []transform{}
	s := dataReader(p.DownloadDir)
	gjson.Parse(gjson.Get(s, "picture").String()).ForEach(
		func(key, value gjson.Result) bool {
			if strings.HasPrefix(key.String(), "id=") {
				if strings.Contains(gjson.Get(value.String(), "date").String(), p.Mode+p.Date) {
					filename := gjson.Get(value.String(), "filename").String()
					title := gjson.Get(value.String(), "title").String()
					favour := gjson.Get(value.String(), "favour").String()
					list = append(list, transform{Title: title, Name: filename, Favour: favour})
					// println(key.String())
					// println(strings.Split(value.String(), ",")[1])
				}
			}
			return true
		})
	return list
}

func setJson(s string, pic picture) string {
	s, _ = sjson.Set(s, "picture.id="+pic.id+".title", pic.title)
	s, _ = sjson.Set(s, "picture.id="+pic.id+".date.-1", pic.date)
	s, _ = sjson.Set(s, "picture.id="+pic.id+".filename", pic.filename)
	s, _ = sjson.Set(s, "picture.id="+pic.id+".favour", false)
	return s
}

//NewPicture return the Picture struct
func NewPicture(title, id, favour string) Picture {
	return Picture{
		ID:     strings.Split(id, ".")[0],
		Path:   fmt.Sprintf("/Pixiv/%s", id),
		Title:  title,
		Origin: id,
		Favour: favour,
	}
}

//LoadPictures load point pixiv mode picture
func LoadPictures(p Pixiv) []Picture {
	files := getJson(p)
	list := make([]Picture, 0, len(files))
	for _, f := range files {
		check := strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".png")
		if check {
			list = append(list, NewPicture(f.Title, f.Name, f.Favour))
		}
	}
	return list
}

func getFavour() []transform {
	list := []transform{}
	s := dataReader(DownloadPath)
	gjson.Parse(gjson.Get(s, "picture").String()).ForEach(
		func(key, value gjson.Result) bool {
			if strings.HasPrefix(key.String(), "id=") {
				//println(key.String())
				if strings.Contains(gjson.Get(value.String(), "favour").String(), "true") {
					filename := gjson.Get(value.String(), "filename").String()
					title := gjson.Get(value.String(), "title").String()
					list = append(list, transform{Title: title, Name: filename, Favour: "true"})
				}
				//println(value.String())
			}
			return true
		})
	return list
}

//SetFavour set favourite image
func SetFavour(id string) {
	s := dataReader(DownloadPath)
	s, _ = sjson.Set(s, "picture."+id+".favour", true)
	dataWriter(s, DownloadPath)
}

//RemoveFavour remove favourite image
func RemoveFavour(id string) {
	s := dataReader(DownloadPath)
	s, _ = sjson.Set(s, "picture."+id+".favour", false)
	dataWriter(s, DownloadPath)
}
