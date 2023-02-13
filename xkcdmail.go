package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

type newestComic struct {
	Title       string `json:"safe_title"`
	AltText     string `json:"alt"`
	ImgUrl      string `json:"img"`
	ComicNumber int    `json:"num"`
}

func main() {
	// readDotEnvFile()
	jsonUrl := "https://xkcd.com/info.0.json"

	to := os.Getenv("TO")
	if to == "" {
		log.Fatal("Missing TO environment variable.")
	}
	toEmails := strings.Split(to, ",")

	fromEmail := os.Getenv("FROM")
	if fromEmail == "" {
		log.Fatal("Missing FROM environment variable.")
	}
	fromPassword := os.Getenv("PASSWORD")
	if fromPassword == "" {
		log.Fatal("Missing PASSWORD environment variable.")
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	auth := smtp.PlainAuth("", fromEmail, fromPassword, smtpHost)

	xkcdApiResults, err := getXkcdJson(jsonUrl)
	if err != nil {
		sendErrorEmail(smtpHost, smtpPort, auth, fromEmail, toEmails, err)
		os.Exit(-1)
	}

	comicResults := newestComic{}

	jsonErr := json.Unmarshal(xkcdApiResults, &comicResults)
	if jsonErr != nil {
		sendErrorEmail(smtpHost, smtpPort, auth, fromEmail, toEmails, jsonErr)
		os.Exit(-1)
	}

	sendXkcdEmail(smtpHost, smtpPort, auth, fromEmail, toEmails, comicResults)
}

func getXkcdJson(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func sendXkcdEmail(host string, port string, auth smtp.Auth, from string, to []string, comic newestComic) {
	toHeader := "To: "
	for i, emailAddress := range to {
		if i == 0 {
			toHeader = toHeader + emailAddress
		} else {
			toHeader = toHeader + ", " + emailAddress
		}
	}
	message := []byte(toHeader + "\r\n" +
		"Subject: XKCD Mail - " + strconv.Itoa(comic.ComicNumber) + ": " + comic.Title + "\r\n" +
		"Reply-To: " + from + "\r\n" +
		"MIME-Version: 1.0" + "\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"" +
		"\r\n\r\n" +
		"<!doctype html>" +
		"<html>" +
		"<head>" +
		"<meta name=\"viewport\" content=\"width=device-width\" />" +
		"<meta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\" />" +
		"<title>XKCD Mail</title>" +
		"</head>" +
		"<body>" +
		"<img src=\"" + comic.ImgUrl + "\" alt=\"" + comic.AltText + "\" style=\"width:100%;\">" +
		"<p style=\"text-align:center;\">Alt Text: " + comic.AltText + "</p>" +
		"</body>" +
		"</html>" +
		"\r\n")
	mailErr := smtp.SendMail(host+":"+port, auth, from, to, message)
	if mailErr != nil {
		sendErrorEmail(host, port, auth, from, to, mailErr)
		os.Exit(-1)
	}
}

func sendErrorEmail(host string, port string, auth smtp.Auth, from string, to []string, err error) {
	toHeader := "To: "
	for i, emailAddress := range to {
		if i == 0 {
			toHeader = toHeader + emailAddress
		} else {
			toHeader = toHeader + ", " + emailAddress
		}
	}
	message := []byte(toHeader + "\r\n" +
		"Subject: XKCD Error!\r\n" +
		"\r\n" +
		"An error occurred in XKCD Mail.\r\nError: " + err.Error() + "\r\n")
	mailErr := smtp.SendMail(host+":"+port, auth, from, to, message)
	if mailErr != nil {
		log.Fatal(mailErr)
	}

	os.Exit(-1)
}
