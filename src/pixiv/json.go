package pixiv

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

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

func isFavour(s, id string) string {
	if gjson.Get(s, "favour."+id).Exists() {
		return "true"
	}
	return "false"
}

func getByRank(p Pixiv) []transform {
	list := []transform{}
	s := dataReader(p.DownloadDir)
	for _, v := range gjson.Get(s, "rank."+p.Mode).Array() {
		filename := gjson.Get(s, "picture."+v.String()+".filename").String()
		title := gjson.Get(s, "picture."+v.String()+".title").String()
		favour := isFavour(s, v.String())
		list = append(list, transform{Title: title, Name: filename, Favour: favour})
	}
	return list
}

func getByDate(p Pixiv, date string) []transform {
	list := []transform{}
	s := dataReader(p.DownloadDir)
	gjson.Parse(gjson.Get(s, "picture").String()).ForEach(
		func(key, value gjson.Result) bool {
			if strings.HasPrefix(key.String(), "id=") {
				if strings.Contains(gjson.Get(value.String(), "date").String(), date) {
					filename := gjson.Get(value.String(), "filename").String()
					title := gjson.Get(value.String(), "title").String()
					favour := isFavour(s, key.String())
					list = append(list, transform{Title: title, Name: filename, Favour: favour})
				}
			}
			return true
		})
	return list
}

func setMetaData(s string, pic picture) string {
	s, _ = sjson.Set(s, "picture.id="+pic.id+".title", pic.title)
	s, _ = sjson.Set(s, "picture.id="+pic.id+".date.-1", pic.date)
	s, _ = sjson.Set(s, "picture.id="+pic.id+".filename", pic.filename)
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
		Local:  "",
	}
}

//LoadPictures load point pixiv mode picture
func LoadPictures(p Pixiv) []Picture {
	list := loadFromTransform(getByRank(p))
	CheckThumbnail(list)
	return list
}

func getFavour() []transform {
	list := []transform{}
	s := dataReader(DownloadPath)
	gjson.Parse(gjson.Get(s, "favour").String()).ForEach(
		func(key, value gjson.Result) bool {
			filename := gjson.Get(s, "picture."+key.String()+".filename").String()
			title := gjson.Get(s, "picture."+key.String()+".title").String()
			list = append(list, transform{Title: title, Name: filename, Favour: "true"})
			return true
		})
	return list
}

//LoadFavour load favour page picture
func LoadFavour() []Picture {
	list := loadFromTransform(getFavour())
	CheckThumbnail(list)
	return list
}

//SetFavour set favourite image
func SetFavour(id string) {
	s := dataReader(DownloadPath)
	s, _ = sjson.Set(s, "favour."+id, time.Now().Format("2006-01-02 15:04:05"))
	dataWriter(s, DownloadPath)
}

//RemoveFavour remove favourite image
func RemoveFavour(id string) {
	s := dataReader(DownloadPath)
	s, _ = sjson.Delete(s, "favour."+id)
	dataWriter(s, DownloadPath)
}

//FindByID Find a picture by its id
func FindByID(id string) ([]Picture, error) {
	p := New("")
	s := dataReader(DownloadPath)
	var pic Picture
	local := true
	if gjson.Get(s, "picture.id="+id).Exists() {
		if strings.Contains(gjson.Get(s, "picture.id="+id+".date").String(), "cache") {
			local = false
		}
	} else if getBySearchID(id, p) == 0 {
		log.Println(id)
		s = dataReader(DownloadPath)
		local = false
	} else {
		return []Picture{}, errors.New("Can't find picture by id")
	}
	pic = NewPicture(gjson.Get(s, "picture.id="+id+".title").String(), gjson.Get(s, "picture.id="+id+".filename").String(), isFavour(s, "id="+id))
	if !local {
		pic.Local = "false"
	}
	CheckThumbnail([]Picture{pic})
	return []Picture{pic}, nil
}

func getSearchData() []transform {
	list := []transform{}
	s := dataReader(DownloadPath)
	gjson.Parse(gjson.Get(s, "picture").String()).ForEach(
		func(key, value gjson.Result) bool {
			if strings.HasPrefix(key.String(), "id=") {
				if strings.Contains(gjson.Get(value.String(), "date").String(), "cache") {
					filename := gjson.Get(value.String(), "filename").String()
					title := gjson.Get(value.String(), "title").String()
					favour := isFavour(s, key.String())
					list = append(list, transform{Title: title, Name: filename, Favour: favour})
				}
			}
			return true
		})
	return list
}

//LoadSearchData Load search history
func LoadSearchData() []Picture {
	list := loadFromTransform(getSearchData())
	CheckThumbnail(list)
	return list
}

func deleteByID(id string) {
	s := dataReader(DownloadPath)
	s, err := sjson.Delete(s, "picture.id="+id)
	if err != nil {
		log.Println(err)
	} else {
		dataWriter(s, DownloadPath)
	}
}

//SaveByID Save picture to "search" from "cache"
func SaveByID(id string) {
	s := dataReader(DownloadPath)
	if strings.Contains(gjson.Get(s, "picture.id="+id+".date").String(), "cache") {
		s, err := sjson.Set(s, "picture.id="+id+".date", []string{"search"})
		if err != nil {
			log.Println(err)
		} else {
			dataWriter(s, DownloadPath)
		}
	}
}

func getByFileName(keywords string) []transform {
	list := []transform{}
	s := dataReader(DownloadPath)
	gjson.Parse(gjson.Get(s, "picture").String()).ForEach(
		func(key, value gjson.Result) bool {
			if strings.HasPrefix(key.String(), "id=") {
				if strings.Contains(gjson.Get(value.String(), "title").String(), keywords) && !strings.Contains(gjson.Get(value.String(), "date").String(), "cache") {
					filename := gjson.Get(value.String(), "filename").String()
					title := gjson.Get(value.String(), "title").String()
					favour := isFavour(s, key.String())
					list = append(list, transform{Title: title, Name: filename, Favour: favour})
				}
			}
			return true
		})
	return list
}

//FindByFileName Find a picture by its keywords
func FindByFileName(keywords string) ([]Picture, error) {
	list := loadFromTransform(getByFileName(keywords))
	if len(list) != 0 {
		CheckThumbnail(list)
		return list, nil
	}
	return list, errors.New("Can't Find This File:" + keywords)
}

func loadFromTransform(files []transform) []Picture {
	list := make([]Picture, 0, len(files))
	for _, f := range files {
		check := strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".png")
		if check {
			list = append(list, NewPicture(f.Title, f.Name, f.Favour))
		}
	}
	return list
}

func getAll() []transform {
	list := []transform{}
	s := dataReader(DownloadPath)
	gjson.Parse(gjson.Get(s, "picture").String()).ForEach(
		func(key, value gjson.Result) bool {
			if strings.HasPrefix(key.String(), "id=") {
				filename := gjson.Get(value.String(), "filename").String()
				title := gjson.Get(value.String(), "title").String()
				favour := isFavour(s, key.String())
				list = append(list, transform{Title: title, Name: filename, Favour: favour})
			}
			return true
		})
	return list
}

//FindAll return all picture
func FindAll() []Picture {
	list := loadFromTransform(getAll())
	return list
}
