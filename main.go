package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"main/client"
	_ "main/memory"
	"main/session"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

var globalSessions *session.Manager
func main() {
    globalSessions, _ = session.NewManager("memory","gosessionid",3600)
    go globalSessions.GC()
    http.HandleFunc("/index", viewHandler)
    http.HandleFunc("/load", loadHandler)
    http.HandleFunc("/search", searchHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
    session := globalSessions.SessionStart(w, r)
    session.Set("cur_page", 0)
    tpl, err := template.ParseFiles("index.html")
    if err != nil {
        log.Fatal("Error parsing html file", err)
    }

    podcastNames := []string{"podcast para tudo", "diva da diva", "eu tava la", "radio escafandro"}
    var wg sync.WaitGroup
    var items []client.Item
    c := make(chan []client.Item)
    for _, url := range podcastNames {
        wg.Add(1)
        go func(url string) {
            defer wg.Done()
            result:= client.SearchPodcast(url).Podcasts[0].Url
            res := client.ParseRssFeed(result)
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

    pages := paginate(items, 10)
    items = pages[0]
    session.Set("pages", pages)

    check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

    err = tpl.Execute(w, items)
	check(err)
}

func loadHandler(w http.ResponseWriter, r *http.Request){
    switch r.Method {
        case http.MethodPost:
            session := globalSessions.SessionStart(w, r)
            pages := session.Get("pages")
            currentPage := session.Get("cur_page").(int)
            session.Set("cur_page", currentPage + 1)
            tpl, _ := template.ParseFiles("partial.html")
            tpl.Execute(w, pages.([][]client.Item)[currentPage + 1])
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func searchHandler(w http.ResponseWriter, r *http.Request){
    switch r.Method {
        case http.MethodPost:
            defer r.Body.Close()
            body, err := io.ReadAll(r.Body)
            if err != nil {
                fmt.Println("Error reading body")
            }
            split := strings.Split(string(body), "=")
            var response client.Response
            if len(split) > 0 {
                fmt.Println(string(body))
                searchedTerm := split[1]
                s, _ := url.QueryUnescape(searchedTerm)
                fmt.Println("searchedTerm", s)
                response = client.SearchPodcast(s)
            }
            tpl, _ := template.ParseFiles("result.html")
            tpl.Execute(w, response.Podcasts)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func paginate(items []client.Item, pageSize int) [][]client.Item{
    itemsSize := len(items)
    fmt.Println("There are", itemsSize, "items")
    pagesQtt := (itemsSize / pageSize)
    rest := itemsSize % pageSize
    if rest > 0 {
        pagesQtt++
    }
    pages := make([][]client.Item, pagesQtt)
    fmt.Println("It will result in", pagesQtt, "pages with", pageSize, "items")
    for i := 0; i < pagesQtt; i++ {
        if ((i * pageSize) + pageSize) > itemsSize {
            pages[i] = items[(i * pageSize):((i * pageSize) + rest)]
        } else {
            pages[i] = items[(i * pageSize):((i * pageSize) + pageSize)]
            pages[i][pageSize - 1].LastElement = true
        }
    }

    return pages
}

