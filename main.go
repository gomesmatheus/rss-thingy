package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

type Rss struct {
    Channel struct {
        Title string `xml:"title"`
        Items []struct {
            Title string `xml:"title"`
            Link string `xml:"link"`
            Id string
            Enclosure struct {
                Url string `xml:"url,attr"`
            } `xml:"enclosure"`
        } `xml:"item"`
    } `xml:"channel"`
}

func main() {
    http.HandleFunc("/index", viewHandler)
    http.HandleFunc("/ping", pingHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("bateu aqui!")
    w.Write([]byte("pong"))
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
    tpl := `
        <!DOCTYPE html>
        <html>
            <head>
                <script>
                    function buildAudioComponent(id) {
                        const audioElement = document.getElementById(id);
                        audioElement.addEventListener("pause", (event) => {
                            console.log("pausou:", event.target.currentTime);
                            localStorage.setItem(id, event.target.currentTime);
                        });

                        const currentTime = localStorage.getItem(id);
                        if(currentTime){
                            audioElement.currentTime = currentTime;
                        }
                    }

                    addEventListener("beforeunload", (event) => {
                        const audios = document.getElementsByTagName("audio");
                        for (let i = 0; i < audios.length; i++){
                            const audio = audios[i]
                            localStorage.setItem(audio.getAttribute("id"), audio.currentTime);
                        }
                        fetch("/ping").then((res) => res);
                    });
                </script>
            </head>
            <body>
                <h1>Title - {{.Channel.Title}}</h1>
                {{range .Channel.Items}}
                    <p>{{.Title}}</p>
                    <audio id="{{ .Id }}" src="{{ .Enclosure.Url }}" controls onloadstart="buildAudioComponent({{ .Id }})"></audio>
                {{else}}
                    <p>no rows</p>
                {{end}}
            </body>
        </html>
    `

    // feedsUrl := []string{"https://radioescafandro.com/feed/", "https://anchor.fm/s/1969eccc/podcast/rss"}
    feedsUrl := []string{"https://radioescafandro.com/feed/"}

    var parsedXml Rss
    for _, url := range feedsUrl {
        parsedXml = *parseRssFeed(url)
    }

    for i := range parsedXml.Channel.Items {
        segments := strings.Split(parsedXml.Channel.Items[i].Link, "/")
        parsedXml.Channel.Items[i].Id = segments[len(segments) - 1]
    }

    check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	t, err := template.New("webpage").Parse(tpl)
	check(err)


    err = t.Execute(w, parsedXml)
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

