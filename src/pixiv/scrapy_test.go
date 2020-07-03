package pixiv

import (
	"fmt"
	"testing"
)

func TestScrapy(t *testing.T) {
	fmt.Print(FindByID("0"))
	CleanSearchCache()
}
