package pixiv

import (
	"archive/tar"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/nfnt/resize"
	"github.com/tidwall/sjson"
)

func dataReader(path string) string {
	data := ""
	file, err := os.OpenFile(path+"Pixiv.json", os.O_RDWR|os.O_CREATE, 0755)
	defer file.Close()
	if err != nil {
		log.Println("Load json err")
		return err.Error()
	}
	b, _ := ioutil.ReadAll(file)
	if len(b) != 0 {
		data = fmt.Sprintf("%s", b)
	} else {
		data, _ = sjson.Set(data, "date.monthly", "")
		data, _ = sjson.Set(data, "date.weekly", "")
		data, _ = sjson.Set(data, "date.daily", "")
	}
	return data
}

func dataWriter(s, path string) {
	err := ioutil.WriteFile(path+"Pixiv.json", []byte(s), 0664)
	if err != nil {
		log.Println("Write json err")
	}
}

func (p Pixiv) EncodeTar() {
	dirPath := "./tmp/"
	filePath := dirPath + p.Mode + p.Date + ".tar"
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dirPath, 0755)
	} else if err := os.RemoveAll(filePath); err != nil {
		log.Fatal("delete "+p.Mode+" err", err)
	} else {
		log.Println("remove old package " + p.Mode)
	}

	log.Println("Package picture " + p.Mode)
	var files = LoadPictures(p)
	buf, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Println(err)
	}
	defer buf.Close()
	tw := tar.NewWriter(buf)
	defer tw.Close()
	for _, file := range files {
		body, err := read(p.DownloadDir + file.Origin)
		if err != nil {
			log.Println("read err", err)
		}
		hdr := &tar.Header{
			Name: file.Origin,
			Mode: 0600,
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Println(err)
		}
		if _, err := tw.Write(body); err != nil {
			log.Println(err)
		}
	}
}

func read(src string) ([]byte, error) {
	file, _ := os.Open(src)
	defer file.Close()
	buf := make([]byte, 4096)
	filebody := make([]byte, 0)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		filebody = append(filebody, buf[:n]...)
	}
	return filebody, nil
}

// DeleteTmp Clean the tmp folder tar cache
func DeleteTmp() {
	dirPath := "./tmp"
	if err := os.RemoveAll(dirPath); err != nil {
		log.Fatal("delete tmp err", err)
	} else {
		log.Println("Clean cache")
	}
}

func CompressImg(dstpath, srcpath string, name string) {
	log.Println("start compress " + name)
	file, err := os.Open(srcpath + name)
	if err != nil {
		log.Fatal(err)
	}
	var img image.Image
	if strings.HasSuffix(name, ".jpg") {
		img, err = jpeg.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(name, ".png") {
		img, err = png.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
	}

	file.Close()

	m := resize.Resize(800, 0, img, resize.NearestNeighbor)
	if _, err := os.Stat(dstpath); os.IsNotExist(err) {
		os.Mkdir(dstpath, 0755)
	}
	out, err := os.OpenFile(dstpath+name, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	png.Encode(out, m)
	log.Println("compress finshed")
}

func CompressAllImg(p Pixiv) {
	files := LoadPictures(p)
	for _, img := range files {
		CompressImg("./thumbnail/", p.DownloadDir, img.Origin)
	}
}
