package pixiv

import (
	"errors"
	"fmt"
	"log"
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

func setOriginJson(s, Title, FileName, ID string, Date []string, Favour bool) string {
	s, _ = sjson.Set(s, "picture.id="+ID+".title", Title)
	s, _ = sjson.Set(s, "picture.id="+ID+".date.-1", Date)
	s, _ = sjson.Set(s, "picture.id="+ID+".filename", FileName)
	s, _ = sjson.Set(s, "picture.id="+ID+".favour", Favour)
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
	list := loadFromTransform(getJson(p))
	CheckThumbnail(list)
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

//LoadFavour load favour page picture
func LoadFavour() []Picture {
	list := loadFromTransform(getFavour())
	CheckThumbnail(list)
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
	} else if getBySearchId(id, p) == 0 {
		log.Println(id)
		s = dataReader(DownloadPath)
		local = false
	} else {
		return []Picture{}, errors.New("Can't find picture by id")
	}
	pic = NewPicture(gjson.Get(s, "picture.id="+id+".title").String(), gjson.Get(s, "picture.id="+id+".filename").String(), gjson.Get(s, "picture.id="+id+".favour").String())
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
					favour := gjson.Get(value.String(), "favour").String()
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
					favour := gjson.Get(value.String(), "favour").String()
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
	if len(list) > 50 {
		return list[len(list)-50 : len(list)]
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
				favour := gjson.Get(value.String(), "favour").String()
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

func importFromJson(s1, s2 string) string {
	gjson.Parse(gjson.Get(s1, "picture").String()).ForEach(
		func(key, value gjson.Result) bool {
			if strings.HasPrefix(key.String(), "id=") {
				filename := gjson.Get(value.String(), "filename").String()
				title := gjson.Get(value.String(), "title").String()
				favour := gjson.Get(value.String(), "favour").Bool()
				date := gjson.Get(value.String(), "date")
				id := strings.Split(filename, ".")[0]
				list := make([]string, 0)
				date.ForEach(func(key, value gjson.Result) bool {
					list = append(list, value.String())
					return true
				})
				if !gjson.Get(s2, "picture.id="+id).Exists() {
					s2 = setOriginJson(s2, title, filename, id, list, favour)
				} else if gjson.Get(s2, "picture.id="+id+".favour").String() == "" {
					fmt.Println(id)
					s2, _ = sjson.Set(s2, "picture.id="+id+".favour", favour)
					s2, _ = sjson.Set(s2, "picture.id="+id+".date", list)
				}
			}
			return true
		})
	return s2
}
