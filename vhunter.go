package main

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
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
)

var (
	filterCodes   = make(map[int]bool)
	singleCodes   = make(map[int]bool)
	statusCounter = make(map[int]int)
	attemptCount  int
	totalWords    int
	attemptsMutex sync.Mutex

	baselineHash [32]byte
	baselineSize int
	baselineSet  bool
)

func colorize(text string, colorCode string) string {
	return "\033[" + colorCode + "m" + text + "\033[0m"
}

func getColorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "32"
	case code >= 300 && code < 400:
		return "33"
	case code >= 400 && code < 600:
		return "31"
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
	fmt.Println(colorize("   '     ' ' `-` ' ' `-' `-' '    v1.1      ", "32"))
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

func parseFlags() {
	flag.StringVar(&wordlistPath, "w", "", "Path to vhost wordlist(required)")
	flag.StringVar(&targetURL, "u", "", "Target IP or domain")
	flag.IntVar(&threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&filterCodesStr, "fc", "", "Comma-separated list of response codes to ignore")
	flag.StringVar(&singleCodesStr, "sc", "", "Comma-separated list of response codes to ONLY show (overrides -fc)")
	flag.StringVar(&outputPath, "o", "", "Output file to save matched results")
	flag.IntVar(&timeoutSecs, "timeout", 5, "HTTP client timeout in seconds")
	flag.IntVar(&filterSize, "fs", 0, "Filter out responses with body size exactly this value (bytes)")
	flag.IntVar(&rateLimit, "rate", 0, "Requests per second rate limit (0 = unlimited)")
	flag.StringVar(&httpMethod, "X", "GET", "HTTP method to use (GET, POST, etc.)")
	flag.BoolVar(&verbose, "v", false, "Verbose mode: print each matched vhost line")
	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.Parse()

	if showHelp || targetURL == "" {
		fmt.Println(`Usage: vhunter -u <target> -w <wordlist> [options]

Options:
  -u        Target IP or domain
  -w        Path to vhost wordlist 
  -t        Number of concurrent threads [default: 10]
  -fc       Comma-separated list of response codes to ignore
  -sc       Comma-separated list of response codes to ONLY show (overrides -fc)
  -fs       Filter out responses with body size exactly this value (bytes)
  -rate     Requests per second rate limit (0 = unlimited)
  -timeout  HTTP client timeout in seconds [default: 5]
  -X        HTTP method to use (GET, POST, etc.) [default: GET]
  -o        Output file to save matched results
  -v        Verbose mode: print each matched vhost line
  -h        Show help message`)
		os.Exit(0)
	}
	if wordlistPath == "" {
		fmt.Println("Error: -w <wordlist> is required")
		os.Exit(0)
	}
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "http://" + targetURL
	}
	if filterCodesStr != "" {
		for _, codeStr := range strings.Split(filterCodesStr, ",") {
			var code int
			fmt.Sscanf(codeStr, "%d", &code)
			filterCodes[code] = true
		}
	}

	if singleCodesStr != "" {
		for _, codeStr := range strings.Split(singleCodesStr, ",") {
			var code int
			fmt.Sscanf(codeStr, "%d", &code)
			singleCodes[code] = true
		}
	}
}

// readWordlist reads lines from the wordlist file
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

// doRequestWithBody sends a request with custom Host header and returns response + body + size
func doRequestWithBody(vhost string, client *http.Client) (*http.Response, []byte, int) {
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

// extractDomain strips scheme and path from url
func extractDomain(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimSuffix(url, "/")
	parts := strings.Split(url, "/")
	return parts[0]
}

// printProgress shows progress with status code counts
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

// generateRandomString returns a random string of given length for baseline host
func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	rand.Seed(time.Now().UnixNano())
	sb := strings.Builder{}
	for i := 0; i < n; i++ {
		sb.WriteByte(letters[rand.Intn(len(letters))])
	}
	return sb.String()
}

func worker(wg *sync.WaitGroup, jobs <-chan string, client *http.Client, outFile *os.File, limiter <-chan time.Time) {
	defer wg.Done()
	for vhost := range jobs {
		if rateLimit > 0 {
			<-limiter
		}
		fullHost := vhost + "." + extractDomain(targetURL)

		resp, body, bodySize := doRequestWithBody(fullHost, client)
		if resp == nil {
			attemptsMutex.Lock()
			attemptCount++
			attemptsMutex.Unlock()
			printProgress()
			continue
		}

		hash := sha256.Sum256(body)

		// Skip if matches baseline hash or baseline size
		if baselineSet && (hash == baselineHash || bodySize == baselineSize) {
			attemptsMutex.Lock()
			attemptCount++
			attemptsMutex.Unlock()
			printProgress()
			continue
		}

		// Filter size if set explicitly (to exclude)
		if filterSize > 0 && bodySize == filterSize {
			attemptsMutex.Lock()
			attemptCount++
			attemptsMutex.Unlock()
			printProgress()
			continue
		}

		attemptsMutex.Lock()
		attemptCount++
		statusCounter[resp.StatusCode]++
		attemptsMutex.Unlock()

		show := (len(singleCodes) > 0 && singleCodes[resp.StatusCode]) || (len(singleCodes) == 0 && !filterCodes[resp.StatusCode])

		if show {
			colorCode := getColorForStatus(resp.StatusCode)
			coloredStatus := colorize(fmt.Sprintf("%d", resp.StatusCode), colorCode)
			coloredDomain := colorize(fullHost, colorCode)
			coloredSize := colorize(fmt.Sprintf("%d bytes", bodySize), colorCode)
			line := fmt.Sprintf("\n[%s] %s size: %s\n", coloredStatus, coloredDomain, coloredSize)

			if verbose {
				fmt.Print(line)
			} else {
				// Print minimal info if not verbose, but only on found vhost (not filtered)
				fmt.Printf("\r[%s] %s size: %s\n", coloredStatus, coloredDomain, coloredSize)
			}

			if outFile != nil {
				outFile.WriteString(fmt.Sprintf("[%d] %s size│ %d bytes\n", resp.StatusCode, fullHost, bodySize))
			}
		} else if verbose {
			// If verbose but filtered by codes, still print something minimal
		//	fmt.Printf("\r[%d] %s (filtered by code)\n", resp.StatusCode, fullHost)
		}

		printProgress()
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

	// Generate a random hostname for baseline (very unlikely to exist)
	randomHost := generateRandomString(12) + "." + extractDomain(targetURL)
	baseDomain := extractDomain(targetURL)
	fmt.Printf("[*] Sending baseline request to %s ...\n", colorize(baseDomain, "31"))

	resp, body, size := doRequestWithBody(randomHost, client)
	if resp == nil {
		fmt.Println("Error: Failed to get baseline response")
		return
	}
	baselineHash = sha256.Sum256(body)
	baselineSize = size
	baselineSet = true

	vhosts, err := readWordlist(wordlistPath)
	if err != nil {
		fmt.Printf("Error reading wordlist│ %v\n", err)
		return
	}

	jobs := make(chan string, threads)
	var wg sync.WaitGroup

	var outFile *os.File
	if outputPath != "" {
		outFile, err = os.Create(outputPath)
		if err != nil {
			fmt.Printf("Error creating output file│ %v\n", err)
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
