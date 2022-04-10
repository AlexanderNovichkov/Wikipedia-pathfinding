package pathfinding

import (
	"errors"
	"log"
	"net/url"
)

type node struct {
	pageUrl *url.URL
	parent  *node

	wikipediaLinks          []*url.URL
	wikipediaLinksExtracted chan struct{}
}

func FindPath(start url.URL, finish url.URL) ([]*url.URL, error) {
	if !isWikipediaUrl(&start) || !isWikipediaUrl(&finish) {
		return nil, errors.New("start and finish must be wikipedia urls")
	}

	start = *getBaseOfWikipediaUrl(&start)
	finish = *getBaseOfWikipediaUrl(&finish)

	if start.String() == finish.String() {
		return []*url.URL{&start}, nil
	}

	queue := make([]*node, 0)
	queue = append(queue, &node{&start, nil, nil, make(chan struct{})})
	visited := make(map[string]bool)
	visited[start.String()] = true

	pool := newWorkersPool(512, 64)
	pool.start()
	defer pool.stop()

	ExtractNodeLinks(queue[0])

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		<-current.wikipediaLinksExtracted
		for _, link := range current.wikipediaLinks {
			if !visited[link.String()] {
				visited[link.String()] = true
				newNode := &node{link, current, nil, make(chan struct{})}
				queue = append(queue, newNode)

				if link.String() == finish.String() {
					return buildPath(newNode), nil
				}

				pool.addTask(func() { ExtractNodeLinks(newNode) })
			}
		}
	}

	return nil, nil
}

func buildPath(finishNode *node) []*url.URL {
	path := make([]*url.URL, 0)
	for current := finishNode; current != nil; current = current.parent {
		path = append(path, current.pageUrl)
	}
	reverse(path)
	return path
}

func reverse(path []*url.URL) {
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
}

func ExtractNodeLinks(node *node) {
	defer close(node.wikipediaLinksExtracted)

	var links []*url.URL
	var err error

	for try := 0; try < 3; try++ {
		links, err = extractLinks(node.pageUrl)
		if err == nil {
			break
		}
	}
	if links == nil {
		log.Println("error extracting links from page ", node.pageUrl, err)
		return
	}

	for _, link := range links {
		if isWikipediaUrl(link) {
			node.wikipediaLinks = append(node.wikipediaLinks, getBaseOfWikipediaUrl(link))
		}
	}
}
