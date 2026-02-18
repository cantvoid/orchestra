package fetcher

import (
	"testing"
	"time"
)

func TestGetLinks(t *testing.T) {
	const testingLink string = "https://gist.githubusercontent.com/cantvoid/1300f1cea3ca8f004ccba6ea7753561a/raw/41fb2532a8f23bd4556ad000b4de4107818cf36e/fakelinks.txt"
	//^ gist with 5 fake links for testing
	const linksExpected int = 5
	links, err := GetLinks(testingLink, 60*time.Second)
	if err != nil {
		t.Fatalf("failed to fetch links: %s", err)
	}
	if len(links) != linksExpected {
		t.Errorf("expected to get %d links, got %d instead", linksExpected, len(links))
	}

}
