package main

import (
	"regexp"
	"strconv"
	"strings"
	// import standard libraries
	"fmt"
	"log"

	// import third party libraries
	"github.com/PuerkitoBio/goquery"
	"github.com/jasonlvhit/gocron"
)

//"https://www.ptt.cc/bbs/MacShop/index.html"
var hyperLinkPreFix = "https://www.ptt.cc"
var webPrefix = hyperLinkPreFix + "/bbs/"
var emailMap = make(map[string]string)

func postScrape(queryHTML string) {
	doc, err := goquery.NewDocument(queryHTML)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(queryHTML)
	doc.Find(".r-ent").Each(func(index int, item *goquery.Selection) {
		titleItem := item.Find(".title")
		title := strings.TrimSpace(titleItem.Text())
		if strings.Contains(title, "徵") && strings.Contains(strings.ToUpper(title), "IPHONE") {
			linkTag := titleItem.Find("a")
			link, _ := linkTag.Attr("href")
			meta := item.Find(".meta")
			date := meta.Find(".date").Text()
			author := meta.Find(".author").Text()
			link = "https://www.ptt.cc" + link
			//fmt.Printf("%s: %s - %s https://www.ptt.cc%v\n", title, date, author, link)
			if _, exist := emailMap[author]; !exist {
				emailMap[author] = date + " : " + author + " " + title + " " + link
			}
		}
	})
}

func postPages(board string) {
	var prevPage int64
	website := webPrefix + board + "/index.html"
	doc, err := goquery.NewDocument(website)
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile("[0-9]+")

	doc.Find("a.btn").Each(func(index int, item *goquery.Selection) {
		title := item.Text()

		if strings.Contains(title, "上頁") {
			linkTag := item
			link, _ := linkTag.Attr("href")
			//fmt.Printf("%s : %v\n", title, hyperLinkPreFix+link)
			//fmt.Println(re.FindAllString(link, -1))
			pageIndex := strings.Join(re.FindAllString(link, -1), "")
			prevPage, err = strconv.ParseInt(pageIndex, 0, 64)
		}
	})

	var queryList [6]string
	queryList[0] = website
	for i := 5; i > 0; i-- {
		queryHTML := webPrefix + board + "/index" + strconv.FormatInt(prevPage-int64(i-1), 10) + ".html"
		queryList[i] = queryHTML
	}

	for _, queryHTML := range queryList {
		postScrape(queryHTML)
	}

	if len(emailMap) > 0 {
		var email Email
		email.sendEmail()
	}
}

func main() {
	//reader := bufio.NewReader(os.Stdin)
	//board, _ := reader.ReadString('\n')
	postPages("MacShop")
	//postScrape()
	gocron.Every(10).Seconds().Do(postPages, "MacShop")
	_, time := gocron.NextRun()
	fmt.Println(time)
	<-gocron.Start()

}
