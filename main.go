package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func init() {

	val := strings.ToLower(os.Getenv("LOGLEVEL"))

	switch val {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	if val = os.Getenv("LOGFORMAT"); strings.ToLower(val) == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			ForceColors:               true,
			PadLevelText:              true,
			DisableLevelTruncation:    true,
			EnvironmentOverrideColors: true,
			FullTimestamp:             true,
		})
	}
}

func handler(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()

	data, err := ioutil.ReadAll(rq.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug(string(data))

	_hmac := hmac.New(sha256.New, []byte(getEnv("GITHUB_SECRET")))
	_hmac.Write(data)

	if "sha256="+hex.EncodeToString(_hmac.Sum(nil)) != rq.Header["X-Hub-Signature-256"][0] {
		log.Warn("Unknown request sender [IPAddress: " + strings.Split(rq.RemoteAddr, ":")[0] + "]. Skipping.")
		return
	}
	log.Info("Signatures matched.")

	// To understand the following line better
	// https://stackoverflow.com/a/48728155
	var objmap map[string]interface{}
	if err = json.Unmarshal(data, &objmap); err != nil {
		log.Fatal(err)
	}

	go func() {
		notification := objmap["issue"].(map[string]interface{})["title"].(string)

		matches := regexp.MustCompile(
			`In \[\W(.+)\W\]\((.+)\n\).+`).FindStringSubmatch(
			objmap["issue"].(map[string]interface{})["body"].(string))

		body := "<ul>" +
			"<li><a href=" + matches[2] + ">Commit <code>" + matches[1] + "</code></a></li>" +
			"<li><a href=" + objmap["issue"].(map[string]interface{})["html_url"].(string) + ">Issue</a></li>" +
			"</ul>"

		if err = sendMails(notification, body); err != nil {
			log.Error(err)
		} else {
			log.Info(notification + ", notification successfully sent.")
		}
	}()

}

func getEnv(vari string) (val string) {
	var e bool
	if val, e = os.LookupEnv(vari); !e {
		log.Fatal("Environment variable " + vari + " is not set.")
	}
	return
}

func sendMails(subject, body string) error {

	var (
		user,
		password,
		host,
		port,
		from string
		to []string
	)

	if _, e := os.LookupEnv("NOTIFICATION_EMAIL_SENDGRID"); e {
		user = "apikey"
		password = getEnv("NOTIFICATION_EMAIL_SENDGRID_API_KEY")
		host = "smtp.sendgrid.net"
		port = "587"
	} else {
		user = getEnv("NOTIFICATION_EMAIL_SMTP_USERNAME")
		password = getEnv("NOTIFICATION_EMAIL_SMTP_PASSWORD")
		host = getEnv("NOTIFICATION_EMAIL_SMTP_HOST")
		port = getEnv("NOTIFICATION_EMAIL_SMTP_PORT")
	}

	from = getEnv("NOTIFICATION_EMAIL_FROM")
	to = strings.Fields(getEnv("NOTIFICATION_EMAIL_TO"))

	auth := smtp.PlainAuth("", user, password, host)
	msg := []byte("From: " + from + "\r\n" +
		"To: " + to[0] + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		"\r\n" +
		body + "\r\n")

	log.Debug("NOTIFICATION_EMAIL_SMTP_USERNAME: " + user + "\n" +
		"NOTIFICATION_EMAIL_SMTP_PASSWORD: " + password + "\n" +
		"NOTIFICATION_EMAIL_SMTP_HOST: " + host + "\n" +
		"NOTIFICATION_EMAIL_SMTP_PORT: " + port + "\n" +
		"NOTIFICATION_EMAIL_FROM: " + from + "\n" +
		"NOTIFICATION_EMAIL_TO: " + strings.Join(to, " "))

	if err := smtp.SendMail(
		host+":"+port, auth, from, to, msg); err != nil {
		return fmt.Errorf("sendMails failed. %v", err)
	}

	return nil
}

func main() {

	var endpoint string
	var e bool

	if endpoint, e = os.LookupEnv("ENDPOINT"); !e {
		log.Warn("ENDPOINT variable not set. Default \"/issue\".")
		endpoint = "issue"
	}

	log.Debug("ENDPOINT: " + endpoint)

	http.HandleFunc("/"+endpoint, handler)

	log.Fatal(
		http.ListenAndServe(":8080", nil))
}
