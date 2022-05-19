package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

func main() {
	resp, err := http.DefaultClient.Get("http://139.155.4.180:9999")
	if err != nil {
		log.Fatal(err)
	}
	data, _ := ioutil.ReadAll(resp.Body) // Ignoring error for example
	var fingerfilepath string
	flag.StringVar(&fingerfilepath, "ffp", "", "指纹库文件路径，默认为空")
	flag.Parse()
	wappalyzerClient, err := wappalyzer.New(fingerfilepath)
	fingerprints := wappalyzerClient.Fingerprint(resp.Header, data)
	fmt.Printf("%v\n", fingerprints)

	// Output: map[Acquia Cloud Platform:{} Amazon EC2:{} Apache:{} Cloudflare:{} Drupal:{} PHP:{} Percona:{} React:{} Varnish:{}]
}
