package main

import (
	"strings"
	// import standard libraries
	"fmt"
	"log"

	// import third party libraries
	"github.com/PuerkitoBio/goquery"
)

func postScrape() {
	doc, err := goquery.NewDocument("https://www.ptt.cc/bbs/MacShop/index.html")
	if err != nil {
		log.Fatal(err)
	}

	// use CSS selector found with the browser inspector
	// for each, use index and item
	doc.Find(".r-ent").Each(func(index int, item *goquery.Selection) {
		titleItem := item.Find(".title")
		title := strings.TrimSpace(titleItem.Text())
		linkTag := titleItem.Find("a")
		link, _ := linkTag.Attr("href")
		meta := item.Find(".meta")
		date := meta.Find(".date").Text()
		author := meta.Find(".author").Text()
		fmt.Printf("%s: %s - %s https://www.ptt.cc%v\n", title, date, author, link)
	})
}

func main() {
	postScrape()
}
