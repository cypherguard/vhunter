package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	wordlistPath   string
	targetURL      string
	threads        int
	filterCodesStr string
	singleCodesStr string
	outputPath     string
	timeoutSecs    int
	filterSize     int
	verbose        bool
	showHelp       bool
	rateLimit      int
	httpMethod     string
	baselineCount  int
	ignoreKeyword  string
)

var (
	filterCodes   = make(map[int]bool)
	singleCodes   = make(map[int]bool)
	statusCounter = make(map[int]int)
	attemptCount  int
	totalWords    int
	attemptsMutex sync.Mutex
	baselineSet   bool
)

type baselineSignature struct {
	Code   int
	Title  string
	Server string
}

var baselines []baselineSignature

func colorize(text string, colorCode string) string {
	return "\033[" + colorCode + "m" + text + "\033[0m"
}

func getColorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "32" // green
	case code >= 300 && code < 400:
		return "33" // yellow
	case code >= 400 && code < 600:
		return "31" // red
	default:
		return "0"
	}
}

func getColorForBannerField(field string) string {
	switch field {
	case "Target URL":
		return "36"
	case "Threads":
		return "35"
	case "Wordlist":
		return "34"
	case "Timeout":
		return "33"
	case "Output File":
		return "32"
	case "Verbose Mode":
		return "31"
	case "Show Codes":
		return "36"
	case "Filter Size":
		return "33"
	case "Rate Limit":
		return "35"
	case "HTTP Method":
		return "34"
	default:
		return "0"
	}
}

func banner() {
	fmt.Println(colorize("																				", "36"))
	fmt.Println(colorize("   .   , .           .                  ", "32"))
	fmt.Println(colorize("   |  /  |           |                  ", "31"))
	fmt.Println(colorize("   | /   |-. . . ;-. |-  ,-. ;-.        ", "31"))
	fmt.Println(colorize("   |/    | | | | | | |   |-' |          ", "32"))
	fmt.Println(colorize("   '     ' ' `-` ' ' `-' `-' '    v1.1", "32"))
	fmt.Println(colorize("  @cypherguard                        ", "31"))
	fmt.Println("─────────────────────┬──────────────────────")
	fmt.Printf("     %s │ %s\n", colorize("Target URL     ", getColorForBannerField("Target URL")), colorize(targetURL, "32"))
	fmt.Printf("     %s │ %s\n", colorize("Threads        ", getColorForBannerField("Threads")), colorize(fmt.Sprintf("%d", threads), "35"))
	fmt.Printf("     %s │ %s\n", colorize("Wordlist       ", getColorForBannerField("Wordlist")), colorize(wordlistPath, "34"))
	fmt.Printf("     %s │ %s\n", colorize("HTTP Method    ", getColorForBannerField("HTTP Method")), colorize(httpMethod, "34"))
	fmt.Printf("     %s │ %s\n", colorize("Timeout (sec)  ", getColorForBannerField("Timeout")), colorize(fmt.Sprintf("%d", timeoutSecs), "33"))
	fmt.Printf("     %s │ %s\n", colorize("Output File    ", getColorForBannerField("Output File")), colorize(func() string {
		if outputPath == "" {
			return "none"
		}
		return outputPath
	}(), "32"))
	fmt.Printf("     %s │ %s\n", colorize("Verbose Mode   ", getColorForBannerField("Verbose Mode")), colorize(fmt.Sprintf("%v", verbose), "31"))

	if filterSize > 0 {
		fmt.Printf("     %s │ %s\n", colorize("Filter Size    ", getColorForBannerField("Filter Size")), colorize(fmt.Sprintf("%d bytes", filterSize), "33"))
	}

	if rateLimit > 0 {
		fmt.Printf("     %s │ %s\n", colorize("Rate Limit     ", getColorForBannerField("Rate Limit")), colorize(fmt.Sprintf("%d req/sec", rateLimit), "35"))
	}

	if len(singleCodes) > 0 {
		var sc []string
		for code := range singleCodes {
			sc = append(sc, colorize(fmt.Sprintf("%d", code), getColorForStatus(code)))
		}
		fmt.Printf("     %s │ [%s]\n", colorize("Show Codes     ", getColorForBannerField("Show Codes")), strings.Join(sc, ", "))
	} else if len(filterCodes) > 0 {
		var fc []string
		for code := range filterCodes {
			fc = append(fc, colorize(fmt.Sprintf("%d", code), getColorForStatus(code)))
		}
		fmt.Printf("     %s │ [%s]\n", colorize("Filter Codes   ", getColorForBannerField("Show Codes")), strings.Join(fc, ", "))
	} else {
		fmt.Printf("     %s │ %s\n", colorize("Show Codes     ", getColorForBannerField("Show Codes")), colorize("all", "32"))
	}
	fmt.Println("─────────────────────┴──────────────────────")
}

func extractDomain(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	return strings.Split(url, "/")[0]
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	rand.Seed(time.Now().UnixNano())
	sb := strings.Builder{}
	for i := 0; i < n; i++ {
		sb.WriteByte(letters[rand.Intn(len(letters))])
	}
	return sb.String()
}

func extractTitle(body []byte) string {
	re := regexp.MustCompile("(?i)<title>(.*?)</title>")
	match := re.FindSubmatch(body)
	if len(match) > 1 {
		return string(match[1])
	}
	return ""
}

func doRequest(vhost string, client *http.Client) (*http.Response, []byte, int) {
	req, err := http.NewRequest(httpMethod, targetURL, nil)
	if err != nil {
		return nil, nil, 0
	}
	req.Host = vhost
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, 0
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp, body, len(body)
}

func getBaselines(n int, client *http.Client) {
	domain := extractDomain(targetURL)
	for i := 0; i < n; i++ {
		randomHost := generateRandomString(12) + "." + domain
		resp, body, _ := doRequest(randomHost, client)
		if resp != nil {
			sig := baselineSignature{
				Code:   resp.StatusCode,
				Title:  extractTitle(body),
				Server: resp.Header.Get("Server"),
			}
			baselines = append(baselines, sig)
		}
	}
	baselineSet = true
}

func isBaselineResponse(resp *http.Response, body []byte) bool {
	title := extractTitle(body)
	server := resp.Header.Get("Server")
	for _, b := range baselines {
		if resp.StatusCode == b.Code && title == b.Title && server == b.Server {
			return true
		}
	}
	return false
}

func readWordlist(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	totalWords = len(lines)
	return lines, scanner.Err()
}

func printProgress() {
	attemptsMutex.Lock()
	defer attemptsMutex.Unlock()
	var sc []string
	for code, count := range statusCounter {
		color := getColorForStatus(code)
		sc = append(sc, fmt.Sprintf("%s%d:%d%s", "\033["+color+"m", code, count, "\033[0m"))
	}
	fmt.Printf("\rTesting vhosts... %d/%d (%.2f%%) | [%s]", attemptCount, totalWords, float64(attemptCount)/float64(totalWords)*100, strings.Join(sc, " "))
}

func worker(wg *sync.WaitGroup, jobs <-chan string, client *http.Client, outFile *os.File, limiter <-chan time.Time) {
	defer wg.Done()
	for vhost := range jobs {
		if rateLimit > 0 {
			<-limiter
		}
		fullHost := vhost + "." + extractDomain(targetURL)
		resp, body, size := doRequest(fullHost, client)

		attemptsMutex.Lock()
		attemptCount++
		attemptsMutex.Unlock()

		if resp == nil || (baselineSet && isBaselineResponse(resp, body)) || (filterSize > 0 && size == filterSize) || (ignoreKeyword != "" && strings.Contains(strings.ToLower(string(body)), strings.ToLower(ignoreKeyword))) {
			printProgress()
			continue
		}

		status := resp.StatusCode
		attemptsMutex.Lock()
		statusCounter[status]++
		attemptsMutex.Unlock()

		show := (len(singleCodes) > 0 && singleCodes[status]) || (len(singleCodes) == 0 && !filterCodes[status])
		if show {
			colorCode := getColorForStatus(status)
			coloredStatus := colorize(fmt.Sprintf("%d", status), colorCode)
			coloredDomain := colorize(fullHost, colorCode)
			coloredSize := colorize(fmt.Sprintf("%d bytes", size), colorCode)
			line := fmt.Sprintf("\n[%s] %s size: %s\n", coloredStatus, coloredDomain, coloredSize)
			if verbose {
		fmt.Print(line)
			}
			if outFile != nil {
				outFile.WriteString(fmt.Sprintf("[%d] %s size: %d bytes\n", status, fullHost, size))
			}
		}
		printProgress()
	}
}

func parseFlags() {
	flag.StringVar(&wordlistPath, "w", "", "Path to vhost wordlist (required)")
	flag.StringVar(&targetURL, "u", "", "Target URL (required)")
	flag.IntVar(&threads, "t", 10, "Number of threads")
	flag.StringVar(&filterCodesStr, "fc", "", "Ignore these status codes")
	flag.StringVar(&singleCodesStr, "sc", "", "Only show these status codes")
	flag.StringVar(&outputPath, "o", "", "Save output to file")
	flag.IntVar(&timeoutSecs, "timeout", 5, "Request timeout (sec)")
	flag.IntVar(&filterSize, "fs", 0, "Filter by response size")
	flag.IntVar(&rateLimit, "rate", 0, "Requests per second (0=unlimited)")
	flag.StringVar(&httpMethod, "X", "GET", "HTTP method to use")
	flag.IntVar(&baselineCount, "bc", 3, "Number of baseline requests")
	flag.StringVar(&ignoreKeyword, "bk", "", "Ignore responses containing this keyword")
	flag.BoolVar(&verbose, "v", false, "Verbose mode")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.Parse()
	if baselineCount < 3 {
	baselineCount = 3
} else if baselineCount > 10 {
	baselineCount = 10
}


	if showHelp || wordlistPath == "" || targetURL == "" {
		fmt.Println(`Usage:vhunter -u <target> -w <wordlist> [options]

Options:
  -u        Target URL (required)
  -w        Path to vhost wordlist (required)
  -t        Number of threads (default: 10)
  -X        HTTP method to use (default: GET)
  -timeout  Request timeout in seconds (default: 5)
  -rate     Requests per second (0 = unlimited)
  -fc       Ignore these status codes (e.g., 403,404)
  -sc       Only show these status codes (overrides -fc)
  -fs       Filter out responses by exact size (bytes)
  -o        Output file to write results
  -bc       Number of baseline requests (default: 3,Recommend 5–7 for accuracy)
  -bk       Ignore responses containing this keyword
  -v        Verbose mode (print all matches)
  -h        Show this help message`)
	 os.Exit(0)
	}

	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "http://" + targetURL
	}

	for _, c := range strings.Split(filterCodesStr, ",") {
		var code int
		fmt.Sscanf(c, "%d", &code)
		if code > 0 {
			filterCodes[code] = true
		}
	}
	for _, c := range strings.Split(singleCodesStr, ",") {
		var code int
		fmt.Sscanf(c, "%d", &code)
		if code > 0 {
			singleCodes[code] = true
		}
	}
}

func main() {
	parseFlags()
	banner()

	client := &http.Client{
		Timeout: time.Duration(timeoutSecs) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	fmt.Printf("[*] Sending %d baseline request...\n", baselineCount)
	getBaselines(baselineCount, client)

	vhosts, err := readWordlist(wordlistPath)
	if err != nil {
		fmt.Printf("Error reading wordlist: %v\n", err)
		return
	}

	jobs := make(chan string, threads)
	var wg sync.WaitGroup

	var outFile *os.File
	if outputPath != "" {
		outFile, err = os.Create(outputPath)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			return
		}
		defer outFile.Close()
	}

	var limiter <-chan time.Time
	if rateLimit > 0 {
		limiter = time.Tick(time.Second / time.Duration(rateLimit))
	}

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go worker(&wg, jobs, client, outFile, limiter)
	}

	for _, vhost := range vhosts {
		jobs <- vhost
	}
	close(jobs)
	wg.Wait()

	fmt.Println("\n\n  Total results:")
	for code, count := range statusCounter {
		color := getColorForStatus(code)
		fmt.Printf("  %s%d%s : %d\n", "\033["+color+"m", code, "\033[0m", count)
	}
}
