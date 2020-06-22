package pixiv

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/tidwall/sjson"
)

func dataReader(path string) string {
	data := ""
	file, err := os.OpenFile(path+"Pixiv.json", os.O_RDWR|os.O_CREATE, 0755)
	defer file.Close()
	if err != nil {
		fmt.Println("加载json错误")
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
		fmt.Println("json文件写入错误")
	}
}
