package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
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
