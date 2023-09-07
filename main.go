package main

import (
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Rss struct {
    Channel struct {
        Title string `xml:"title"`
        Items []Item `xml:"item"`
    } `xml:"channel"`
}
type Item struct {
    PodcastTitle string
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

    searchPodcast("nerdcast")
    // feedsUrl := []string{"https://anchor.fm/s/a9f2877c/podcast/rss", "https://anchor.fm/s/1969eccc/podcast/rss", "https://www.spreaker.com/show/3258232/episodes/feed"}
    feedsUrl := []string{"this is brazil", "nerdcast", "boa noite internet", "radio escafandro"}

    var wg sync.WaitGroup
    var items []Item
    c := make(chan []Item)
    for _, url := range feedsUrl {
        wg.Add(1)
        go func(url string) {
            defer wg.Done()
            url1 := searchPodcast(url)
            res := parseRssFeed(url1)
            for i := range res.Channel.Items {
                i = i
                res.Channel.Items[i].PodcastTitle = res.Channel.Title
            }
            c <- res.Channel.Items
        }(url)
        items = append(items, <- c...)
    }
    wg.Wait()

    sort.Slice(items, func(i,j int) bool {
        t1, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 MST",items[i].Date)
        t2, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 MST",items[j].Date)
        return t2.Before(t1)
    })

    for i := range items {
        segments := strings.Split(items[i].Link, "/")
        segment := segments[len(segments) - 1]
        if segment == "" {
            segment = segments[len(segments) - 2]
        }
        items[i].Id = segment
    }

    pages := paginate(items, 20)
    items = pages[0]

    check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

    err = tpl.Execute(w, items)
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

func searchPodcast(title string) string{
    type Response struct {
       Podcasts []struct {
           Title string `json:"title"`
           Url string `json:"url"`
       } `json:"feeds"`
    }

    client := &http.Client{}
    apiKey := ""
    apiSecret := ""
    authDate := time.Now().Unix()
    authorization := fmt.Sprintf("%s%s%d", apiKey, apiSecret, authDate)
    fmt.Println("auth date", fmt.Sprintf("%d", authDate))

    h := sha1.New()
    h.Write([]byte(authorization))
    authorizationEncrypted := fmt.Sprintf("%x", h.Sum(nil))
    fmt.Println("auth", authorizationEncrypted)

    req, _ := http.NewRequest("GET", "https://api.podcastindex.org/api/1.0/search/bytitle", nil)
    req.Header.Add("X-Auth-Date", fmt.Sprintf("%d", authDate))
    req.Header.Add("X-Auth-Key", apiKey)
    req.Header.Add("Authorization", authorizationEncrypted)

    q := req.URL.Query()
    q.Add("q", title)
    req.URL.RawQuery = q.Encode()

    res, err := client.Do(req)
    if err != nil {
        fmt.Println("Error on podcastindex API")
        fmt.Println(err)
    }
    fmt.Println(res.StatusCode)
    defer res.Body.Close()


    var response Response
    body, err := io.ReadAll(res.Body)
    json.Unmarshal(body, &response)

    fmt.Println(response.Podcasts[0].Url)
    fmt.Println(response.Podcasts[0].Title)
    return response.Podcasts[0].Url
}

func paginate(items []Item, pageSize int) [][]Item{
    itemsSize := len(items)
    fmt.Println("There are", itemsSize, "items")
    pagesQtt := (itemsSize / pageSize)
    rest := itemsSize % pageSize
    if rest > 0 {
        pagesQtt++
    }
    pages := make([][]Item, pagesQtt)
    fmt.Println("It will result in", pagesQtt, "pages with", pageSize, "items")
    for i := 0; i < pagesQtt; i++ {
        if ((i * pageSize) + pageSize) > itemsSize {
            pages[i] = items[(i * pageSize):((i * pageSize) + rest)]
        } else {
            pages[i] = items[(i * pageSize):((i * pageSize) + pageSize)]
        }
    }

    return pages
}

