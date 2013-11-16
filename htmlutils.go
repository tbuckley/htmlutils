package htmlutils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"code.google.com/p/go.net/html"
)

type ElementHandler func(*html.Node)

type Query struct {
	set []*html.Node
}

func (q *Query) At(i int) *html.Node {
	return q.set[i]
}

func FetchPage(page *url.URL) ([]byte, error) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("Detected redirect, skipping")
		},
	}
	resp, err := client.Get(page.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func NewQueryFromPage(body []byte) (*Query, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return &Query{[]*html.Node{doc}}, nil
}

func NewQuery(page *url.URL) (*Query, error) {
	body, err := FetchPage(page)
	if err != nil {
		return nil, err
	}
	return NewQueryFromPage(body)
}

func (q *Query) Each(fn ElementHandler) {
	for _, e := range q.set {
		fn(e)
	}
}

func (q *Query) Traverse(fn ElementHandler) {
	var helper ElementHandler
	helper = func(e *html.Node) {
		fn(e)
		child := e.FirstChild
		for child != nil {
			helper(child)
			child = child.NextSibling
		}
	}
	q.Each(helper)
}

func (q *Query) ElementsByTagName(tagName string) *Query {
	resp := &Query{}
	handler := func(node *html.Node) {
		switch node.Type {
		case html.ElementNode:
			if node.Data == tagName {
				resp.set = append(resp.set, node)
			}
		default:
		}
	}
	q.Traverse(handler)
	return resp
}

func (q *Query) Attr(key string) []string {
	var attrs []string
	q.Each(func(e *html.Node) {
		value, ok := getAttr(e, key)
		if ok {
			attrs = append(attrs, value)
		}
	})
	return attrs
}

func getAttr(node *html.Node, key string) (string, bool) {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}
