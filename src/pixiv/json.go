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
				//println(key.String())
				if strings.Contains(value.String(), p.Mode+p.Date) {
					filename := gjson.Get(value.String(), "filename").String()
					title := gjson.Get(value.String(), "title").String()
					list = append(list, transform{Title: title, Name: filename})
				}
				//println(value.String())
			}
			return true
		})
	return list
}

func setJson(s string, pic picture) string {
	s, _ = sjson.Set(s, "picture.id="+pic.id+".title", pic.title)
	s, _ = sjson.Set(s, "picture.id="+pic.id+".date.-1", pic.date)
	s, _ = sjson.Set(s, "picture.id="+pic.id+".filename", pic.filename)
	return s
}

func NewPicture(title, id string) Picture {
	return Picture{
		Id:    strings.Split(id, ".")[0],
		Path:  fmt.Sprintf("/static/picture/%s", id),
		Title: title,
	}
}

func LoadPictures(p Pixiv) []Picture {
	files := getJson(p)
	list := make([]Picture, 0, len(files))
	for _, f := range files {
		check := strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".png")
		if check {
			list = append(list, NewPicture(f.Title, f.Name))
		}
	}
	return list
}
