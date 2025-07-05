# vhunter v1.1

```
																				
   .   , .           .                  
   |  /  |           |                  
   | /   |-. . . ;-. |-  ,-. ;-.        
   |/    | | | | | | |   |-' |          
   '     ' ' `-` ' ' `-' `-' '    v1.1
  @cypherguard                        
```

🎯 A fast and efficient virtual host (vhost) discovery tool written in Go. vhunter helps security researchers and penetration testers discover virtual hosts by bruteforcing hostnames against a target server.

## ✨ Features

- ⚡ **Fast Multi-threaded Scanning**: Configurable number of concurrent threads
- 🎯 **Baseline Detection**: Automatically detects and filters out default/baseline responses
- 🔧 **Flexible Filtering**: Filter by status codes, response size, or keywords
- 🚦 **Rate Limiting**: Built-in rate limiting to avoid overwhelming targets
- 📊 **Multiple Output Formats**: Console output with colored results and optional file output
- 🌐 **HTTP Method Support**: Supports different HTTP methods (GET, POST, etc.)
- 🔒 **TLS Support**: Handles HTTPS targets with configurable TLS settings
## 🎯 Baseline Detection

vhunter automatically detects baseline responses by:
- 🎲 Generating random hostnames and sending requests
- 📊 Analyzing response patterns (status code, title, server headers)
- 🔍 Filtering out responses that match baseline signatures
- ⚙️ Configurable number of baseline requests (3-10, recommended 5-7 for accuracy)

The baseline detection helps eliminate false positives by identifying the server's default response pattern.

## 🚀 Performance Tips

- 🧵 **Threads**: Start with 10-20 threads, increase gradually based on target capacity
- 🐌 **Rate Limiting**: Use 5-10 requests/second for stealthy scanning
- ⏰ **Timeout**: Adjust timeout based on target response time
- 📊 **Baseline Requests**: Use 5-7 baseline requests for better accuracy
## 📥 Installation

### From Source
```bash
git clone <repository-url>
cd vhunter
go build -o vhunter vhunter.go
```

### Direct Run
```bash
go run vhunter.go [options]
```

### 📦 Download Binary
Download the latest pre-compiled binary from the [releases page](https://github.com/your-username/vhunter/releases)

## 🚀 Usage

### Basic Usage
```bash
./vhunter -u http://example.com -w wordlist.txt
```


## ⚙️ Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `-u` | Target URL (required) | - |
| `-w` | Path to vhost wordlist (required) | - |
| `-t` | Number of threads | 10 |
| `-X` | HTTP method to use | GET |
| `-timeout` | Request timeout in seconds | 5 |
| `-rate` | Requests per second (0 = unlimited) | 0 |
| `-fc` | Ignore these status codes (e.g., 403,404) | - |
| `-sc` | Only show these status codes (overrides -fc) | - |
| `-fs` | Filter out responses by exact size (bytes) | 0 |
| `-o` | Output file to write results | - |
| `-bc` | Number of baseline requests (3-10) | 3 |
| `-bk` | Ignore responses containing this keyword | - |
| `-v` | Verbose mode (print all matches) | false |
| `-h` | Show help message | false |

## 💡 Examples

### 🔍 Basic Virtual Host Discovery
```bash
./vhunter -u http://target.com -w common-vhosts.txt
```
### 📝 Use POST Method with Custom Baseline
```bash
./vhunter -u https://api.target.com -w api-endpoints.txt -X POST -bc 7
```

### 🔤 Ignore Responses with Specific Keywords
```bash
./vhunter -u http://target.com -w wordlist.txt -bk "not found"
```
## 📊 Output Format
- 🎨 **Colored Status Codes**: 
  - 🟢 Green (2xx): Successful responses
  - 🟡 Yellow (3xx): Redirects
  - 🔴 Red (4xx/5xx): Client/Server errors
- 📈 **Real-time Progress**: Shows current progress and status code distribution
- 📋 **Result Summary**: Final count of responses by status code

### 📁 File Output
Plain text format suitable for further processing:
```
[200] admin.target.com size: 1024 bytes
[301] dev.target.com size: 512 bytes
[200] api.target.com size: 2048 bytes
```
## 🔒 Security Considerations

- ✅ Always obtain proper authorization before scanning
- 🚦 Use appropriate rate limiting to avoid DoS conditions
- ⚖️ Be aware of legal and ethical implications
- 🔐 Consider using VPN or proxy for anonymity
- 👀 Monitor your scanning to avoid overwhelming the target
- 🛡️ Some targets may have intrusion detection systems

## 🛠️ Troubleshooting

### ❗ Common Issues

**"Connection refused" errors**:
- 🔗 Check if the target URL is accessible
- 🌐 Verify the protocol (HTTP vs HTTPS)
- 🔥 Check firewall/network restrictions

**Too many false positives**:
- 📊 Increase baseline requests with `-bc 7`
- 📏 Use response size filtering with `-fs`
- 🔤 Add keyword filtering with `-bk`

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👨‍💻 Author

**@cypherguard**

## ⚠️ Disclaimer

This tool is for educational and authorized security testing purposes only. Users are responsible for complying with applicable laws and regulations. The author is not responsible for any misuse of this tool.

---

**🚨 Important**: Always ensure you have proper authorization before testing any systems you do not own. Unauthorized scanning may be illegal in your jurisdiction.
