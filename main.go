package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Rss struct {
    Channel struct {
        Title string `xml:"title"`
        Items []struct {
            Title string `xml:"title"`
            Link string `xml:"link"`
            Date string `xml:"pubDate"`
            Id string
            Image struct {
                Url string `xml:"href,attr"`
            } `xml:"image"`
            Enclosure struct {
                Url string `xml:"url,attr"`
            } `xml:"enclosure"`
        } `xml:"item"`
    } `xml:"channel"`
}

func main() {
    http.HandleFunc("/index", viewHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
    tpl, err := template.ParseFiles("index.html")
    if err != nil {
        log.Fatal("Error parsing html file", err)
    }

    feedsUrl := []string{"https://radioescafandro.com/feed/", "https://anchor.fm/s/1969eccc/podcast/rss", "https://www.spreaker.com/show/3258232/episodes/feed"}

    var parsedXmls []Rss
    for _, url := range feedsUrl {
        parsedXmls = append(parsedXmls, *parseRssFeed(url))
    }

    parsedXml := parsedXmls[1]
    parsedXml.Channel.Items = append(parsedXml.Channel.Items, parsedXmls[0].Channel.Items...)
    parsedXml.Channel.Items = append(parsedXml.Channel.Items, parsedXmls[2].Channel.Items...)
    sort.Slice(parsedXml.Channel.Items, func(i,j int) bool {
        t1, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 MST",parsedXml.Channel.Items[i].Date)
        t2, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 MST",parsedXml.Channel.Items[j].Date)
        return t2.Before(t1)
    })

    for i := range parsedXml.Channel.Items {
        segments := strings.Split(parsedXml.Channel.Items[i].Link, "/")
        parsedXml.Channel.Items[i].Id = segments[len(segments) - 1]
    }

    check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

    err = tpl.Execute(w, parsedXml)
	check(err)
}

func parseRssFeed(feedUrl string) *Rss {
    resp, err := http.Get(feedUrl)
    if err != nil {
        fmt.Println("Error retrieving feed")
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error parsing body")
    }

    var parsedXml Rss
    if err = xml.Unmarshal(body, &parsedXml); err != nil {
        fmt.Println("Error parsing xml")
    }

    return &parsedXml
}

