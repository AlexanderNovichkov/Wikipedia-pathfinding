package pathfinding

import (
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"strings"
)

func isWikipediaUrl(url *url.URL) bool {
	return strings.HasSuffix(url.Host, "wikipedia.org")
}

func getBaseOfWikipediaUrl(wikipediaUrl *url.URL) *url.URL {
	return &url.URL{
		Scheme: wikipediaUrl.Scheme,
		Host:   wikipediaUrl.Host,
		Path:   wikipediaUrl.Path,
	}
}

func extractLinks(pageUrl *url.URL) ([]*url.URL, error) {
	resp, err := http.Get(pageUrl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	links := make([]*url.URL, 0)

	tokenizer := html.NewTokenizer(resp.Body)
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			break
		}
		token := tokenizer.Token()
		if tokenType == html.StartTagToken && token.Data == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					currentUrl, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}
					currentUrlAbsolute := pageUrl.ResolveReference(currentUrl)
					links = append(links, currentUrlAbsolute)
				}
			}
		}
	}

	return links, nil
}
