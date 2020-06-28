package pixiv

import (
	"fmt"
	"testing"
)

func TestScrapy(t *testing.T) {
	p := New("weekly")
	fmt.Println(LoadPictures(*p))
}
