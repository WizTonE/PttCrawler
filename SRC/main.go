package main

import (
	"crypto/tls"
	"net/smtp"
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

type Mail struct {
	senderID string
	toIds    []string
	subject  string
	body     string
}

type SmtpServer struct {
	host string
	port string
}

func (s *SmtpServer) ServerName() string {
	return s.host + ":" + s.port
}

func (mail *Mail) BuildMessage() string {
	message := ""
	message += fmt.Sprintf("From: %s\r\n", mail.senderID)
	if len(mail.toIds) > 0 {
		message += fmt.Sprintf("To: %s\r\n", strings.Join(mail.toIds, ";"))
	}

	message += fmt.Sprintf("Subject: %s\r\n", mail.subject)
	message += "\r\n" + mail.body

	return message
}

func sendEmail() {
	var emailContent string

	for _, w := range emailMap {
		emailContent = emailContent + w + "\n"
	}
	fmt.Println(emailContent)
	mail := Mail{}
	mail.senderID = "013a0.tw@gmail.com"
	mail.toIds = []string{"rice.fan@gmail.com"}
	mail.subject = "Check it out"
	mail.body = emailContent

	messageBody := mail.BuildMessage()

	smtpServer := SmtpServer{host: "smtp.gmail.com", port: "465"}

	log.Println(smtpServer.host)
	//build an auth
	auth := smtp.PlainAuth("", mail.senderID, "", smtpServer.host)

	// Gmail will reject connection if it's not secure
	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer.host,
	}

	conn, err := tls.Dial("tcp", smtpServer.ServerName(), tlsconfig)
	if err != nil {
		log.Panic(err)
	}

	client, err := smtp.NewClient(conn, smtpServer.host)
	if err != nil {
		log.Panic(err)
	}

	// step 1: Use Auth
	if err = client.Auth(auth); err != nil {
		log.Panic(err)
	}

	// step 2: add all from and to
	if err = client.Mail(mail.senderID); err != nil {
		log.Panic(err)
	}
	for _, k := range mail.toIds {
		if err = client.Rcpt(k); err != nil {
			log.Panic(err)
		}
	}

	// Data
	w, err := client.Data()
	if err != nil {
		log.Panic(err)
	}

	_, err = w.Write([]byte(messageBody))
	if err != nil {
		log.Panic(err)
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
	}

	client.Quit()

	log.Println("Mail sent successfully")

}

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
		sendEmail()
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
