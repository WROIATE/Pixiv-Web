package pixiv

import "testing"

func TestScrapy(t *testing.T) {
	p := New("daily")
	p.Crawl()
}
