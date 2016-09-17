package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

// Get credentials here: // https://console.eu-gb.bluemix.net/catalog/services/text-to-speech
//
// Price: 	Standard	The first million characters are free, €0.0151 EUR/THOUSAND CHAR
// wc --chars /tmp/text | ruby -ne 'chars_cnt = Integer($_.split.first); estimated_price = chars_cnt * 0.0151 / 1000; printf("Estimated price: %.2f €\n", estimated_price)'

// Data corresponds to the data required by the Watson API
// See: - https://www.ibm.com/watson/developercloud/text-to-speech/api/v1/#synthesize audio
//      - https://watson-api-explorer.mybluemix.net/apis/text-to-speech-v1#!/synthesize/postSynthesize
type Data struct {
	Text string `json:"text"`
}

var endPoint = "https://stream.watsonplatform.net/text-to-speech/api/v1/synthesize?voice=en-US_AllisonVoice"

var validFormats = []string{"ogg", "flac", "wav"}
var format = flag.String("format", "ogg", fmt.Sprintf("The output file format (smallest to largest): %v.", validFormats))
var out = flag.String("out", "out", "The output filename")
var in = ""

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s: [OPTIONS] TEXT_FILEPATH\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	checkFormat()
	finalFormat()
	if flag.Arg(0) == "" {
		log.Fatal("Must give filepath")
	}
	in = flag.Arg(0)
}

func finalFormat() {
	if *format == "ogg" {
		*format = "ogg;codecs=opus"
	}
}

func checkFormat() {
	for _, v := range validFormats {
		if *format == v {
			return
		}
	}
	log.Fatal("Invalid format: " + *format)
}

func main() {

	client := &http.Client{}
	content, err := ioutil.ReadFile(in)
	if err != nil {
		log.Fatal(err)
	}
	text := Data{Text: string(content)}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(text)
	req, err := http.NewRequest("POST", endPoint, b)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/"+*format)
	req.SetBasicAuth(os.Getenv("TTS_USERNAME"), os.Getenv("TTS_PASSWORD"))
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(body))
		log.Fatal("HTTP request failed: " + resp.Status)
	}
	fmt.Println(resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(*out, body, 0644)
}
