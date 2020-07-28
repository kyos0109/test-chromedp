package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"github.com/cheggaaa/pb/v3"

	"github.com/briandowns/spinner"
)

var (
	subPathCount  int
	help          bool
	headless      bool
	debug         bool
	configPath    string
	WarningLogger *log.Logger
	LatencyLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	DebugLogger   *log.Logger
	NoticeLogger  *log.Logger
)

type YAMLConfig struct {
	HTML        BodyWaitDomLoad `yaml:"HTML"`
	TestConfig  TestConfig      `yaml:"TestConfig"`
	TestTargets []Target        `yaml:"TestTargets"`
}

type BodyWaitDomLoad struct {
	BodyWaitDomLoad string `yaml:"BodyWaitDomLoad"`
}

type TestConfig struct {
	SubPath    []string `yaml:"SubPath"`
	Remote     bool     `yaml:"Remote,omitempty"`
	ChromedpWS string   `yaml:"ChromedpWS,omitempty"`
}

type Target struct {
	URL         string `yaml:"URL"`
	SkipSubPath bool   `yaml:"SkipSubPath,omitempty"`
	Disable     bool   `yaml:"Disable,omitempty"`
	SkipHTTPs   bool   `yaml:"SkipHTTPs,omitempty"`
}

func init() {
	file, err := os.OpenFile("info.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	errFile, err := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	latencyFile, err := os.OpenFile("latency.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// mwf := io.MultiWriter(os.Stderr, file)
	log.SetOutput(file)

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(errFile, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(errFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	NoticeLogger = log.New(os.Stdout, "NOTICE: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(os.Stdout, "ddd: ", log.Ldate|log.Ltime|log.Lshortfile)
	LatencyLogger = log.New(latencyFile, "LATENCY: ", log.Ldate|log.Ltime|log.Lshortfile)

	flag.BoolVar(&help, "h", false, "This help")
	flag.BoolVar(&debug, "debug", false, "Enable debug response")
	flag.BoolVar(&headless, "less", true, "Enable Chrome headless. (Only Local)")
	flag.StringVar(&configPath, "c", "config.yaml", "YAML config `path`")
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	InfoLogger.Println("Initial Done.")
}

func (yc *YAMLConfig) Init() {
	yc.ReadYAMLConfig()
	yc.FilterEnable()
}

func (yc *YAMLConfig) ReadYAMLConfig() {
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		NoticeLogger.Printf("yamlFile.Get err #%v ", err)
		ErrorLogger.Fatalf("yamlFile.Get err #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, yc)
	if err != nil {
		NoticeLogger.Printf("Unmarshal: %v", err)
		ErrorLogger.Fatalf("Unmarshal: %v", err)
	}
}

func (yc *YAMLConfig) FilterEnable() {
	for i := len(yc.TestTargets) - 1; i >= 0; i-- {
		if yc.TestTargets[i].Disable {
			yc.TestTargets[i] = yc.TestTargets[len(yc.TestTargets)-1]
			yc.TestTargets[len(yc.TestTargets)-1] = Target{}
			yc.TestTargets = yc.TestTargets[:len(yc.TestTargets)-1]
		}
	}

	for _, v := range yc.TestTargets {
		if !v.SkipSubPath {
			subPathCount = subPathCount + (len(yc.TestConfig.SubPath) - 1)
		}
	}

	if len(yc.TestTargets) == 0 {
		NoticeLogger.Println("Not Senders Profile Enable....Bye.")
		os.Exit(0)
	}

	if len(yc.TestTargets) > 0 {
		NoticeLogger.Printf("Enable test target: %d", len(yc.TestTargets))
	}
}

func main() {
	var yc YAMLConfig

	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Start()

	time.Sleep(2 * time.Second)
	yc.Init()

	s.Stop()

	run(&yc)
}

func (t *Target) AppendHttpPerfix(target *Target) {
	if len(target.URL) > 3 {
		if !strings.HasPrefix(target.URL, "http://") && !strings.HasPrefix(target.URL, "https://") {
			if target.SkipHTTPs {
				target.URL = "http://" + target.URL
			} else {
				target.URL = "https://" + target.URL
			}
		}
	}
}

func run(yamlConig *YAMLConfig) {
	var (
		t          Target
		ctx        context.Context
		cancel     context.CancelFunc
		appendPath []string
	)
	targets := yamlConig.TestTargets

	DOM, err := yaml.Marshal(&yamlConig.HTML.BodyWaitDomLoad)
	if err != nil {
		DebugLogger.Println(err)
		log.Fatalf("error: %v", err)
	}

	InfoLogger.Println("Append SubPath", yamlConig.TestConfig.SubPath)

	fmt.Println("===========> Happy Start Run Test <===========")

	webSocketURL := yamlConig.TestConfig.ChromedpWS

	switch yamlConig.TestConfig.Remote && len(webSocketURL) > 3 {
	case true:
		NoticeLogger.Println("Use remote Chrome ->", webSocketURL)
		ctx, cancel = runChromedpRemote(&webSocketURL)
		defer cancel()
	case false:
		NoticeLogger.Println("Use local Chrome")
		ctx, cancel = runChromedpLocal()
		defer cancel()
	default:
		log.Fatal("Unknown Remote Option.")
	}

	bar := pb.StartNew(len(targets) + subPathCount)

	for _, v := range targets {
		t.AppendHttpPerfix(&v)

		if v.SkipSubPath {
			appendPath = []string{"/"}
		} else {
			appendPath = yamlConig.TestConfig.SubPath
		}

		for _, k := range appendPath {
			URL := v.URL + k
			bar.Increment()
			chromedpMain(&URL, &ctx, &DOM)
		}
	}

	bar.Finish()
	InfoLogger.Println("Test Done.")
	NoticeLogger.Println("Test Done.")
}

func runChromedpLocal() (context.Context, context.CancelFunc) {
	dir, err := ioutil.TempDir("", "chromedp-wow")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", headless),
		chromedp.Flag("ignore-certificate-errors", false),
		chromedp.UserDataDir(dir),
	)

	return chromedp.NewExecAllocator(context.Background(), opts...)
}

func runChromedpRemote(ws *string) (context.Context, context.CancelFunc) {
	return chromedp.NewRemoteAllocator(context.Background(), *ws)
}

func chromedpMain(URL *string, allocCtx *context.Context, DOM *[]byte) {
	var StartTime time.Time

	InfoLogger.Println("Run -->", *URL)

	ctx, cancel := chromedp.NewContext(*allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev.(type) {
		case *network.EventResponseReceived:
			go networkEventResponseReceived(ev.(*network.EventResponseReceived), URL, cancel)
			break
		case *network.EventLoadingFailed:
			go func(r *network.EventLoadingFailed) {
				ErrorLogger.Println(*URL, r.RequestID, r.ErrorText)
				cancel()
			}(ev.(*network.EventLoadingFailed))
			break
		case *network.EventRequestWillBeSent:
			go func(r *network.EventRequestWillBeSent) {
				if r.Type == "Document" {
					StartTime = r.Timestamp.Time()
				}
			}(ev.(*network.EventRequestWillBeSent))
			break
		case *page.EventDomContentEventFired:
			go pageEventDomContentEventFired(ev.(*page.EventDomContentEventFired), URL, &StartTime)
			break
		case *page.EventLoadEventFired:
			go pageEventLoadEventFired(ev.(*page.EventLoadEventFired), URL, &StartTime)
			break
		}
	})

	clearDOM := strings.TrimSuffix(string(*DOM), "\n")
	clearDOM = strings.Trim(clearDOM, `'"`)

	err := chromedp.Run(ctx,
		network.Enable(),
		network.ClearBrowserCache(),
		page.Enable(),
		chromedp.Navigate(*URL),
		chromedp.WaitVisible(clearDOM, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
	)
	if err != nil {
		ErrorLogger.Println(*URL, err)
	}
}

func pageEventDomContentEventFired(e *page.EventDomContentEventFired, URL *string, startTime *time.Time) {
	domDiffTime := e.Timestamp.Time().Sub(*startTime)
	InfoLogger.Printf("└→ DomContent: %v", domDiffTime)

	if domDiffTime > (800 * time.Millisecond) {
		WarningLogger.Printf("URL: %s, DomContent: %v", *URL, domDiffTime)
	}
}

func pageEventLoadEventFired(e *page.EventLoadEventFired, URL *string, startTime *time.Time) {
	loadDiffTime := e.Timestamp.Time().Sub(*startTime)
	InfoLogger.Printf("└→ Load: %v", loadDiffTime)

	if loadDiffTime > (2 * time.Second) {
		WarningLogger.Printf("URL: %s, Load: %v", *URL, loadDiffTime)
	}
}

func networkEventResponseReceived(r *network.EventResponseReceived, URL *string, cancel context.CancelFunc) {
	res := r.Response
	if debug {
		DebugLogger.Println(res)
	}
	if res.ConnectionID != 0 && res.RemoteIPAddress != "" {
		switch true {
		case (res.Status > 499):
			ErrorLogger.Printf(
				"site: %v, status: %d, URL: %s, ServerIP: %s, Latency: %vms",
				*URL, res.Status, res.URL, res.RemoteIPAddress, res.Timing.ReceiveHeadersEnd)
			cancel()
			break
		case (res.Status < 200 || res.Status > 399):
			WarningLogger.Printf(
				"site: %v, status: %d, URL: %s, ServerIP: %s, Latency: %vms",
				*URL, res.Status, res.URL, res.RemoteIPAddress, res.Timing.ReceiveHeadersEnd)
			break
		case (res.Timing.ReceiveHeadersEnd > 500):
			LatencyLogger.Printf(
				"site: %v, status: %d, URL: %s, ServerIP: %s, Latency: %vms",
				*URL, res.Status, res.URL, res.RemoteIPAddress, res.Timing.ReceiveHeadersEnd)
			break
		}
	}
}
