# 🎯 VHunter v1.1

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Version](https://img.shields.io/badge/Version-1.1-blue?style=for-the-badge)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

> 🚀 **Fast and efficient virtual host discovery tool written in Go**

VHunter is a powerful command-line tool designed for virtual host enumeration and discovery. It performs concurrent HTTP requests with custom Host headers to identify hidden virtual hosts on web servers.

## Features

- **⚡ Concurrent Processing**: Multi-threaded architecture for high-speed virtual host enumeration
- **🎯 Intelligent Baseline Detection**: Automated false positive filtering through baseline response comparison
- **🔧 Advanced Filtering Options**: Comprehensive status code and response size filtering capabilities
- **📊 Real-time Progress Monitoring**: Live scan progress with detailed status code statistics
- **🌐 Protocol Flexibility**: Support for multiple HTTP methods (GET, POST, PUT, DELETE, etc.)
- **🚦 Traffic Control**: Built-in rate limiting to ensure responsible scanning practices
- **📈 Performance Optimization**: Configurable timeout and thread management for optimal performance
- **📋 Comprehensive Reporting**: Detailed output with file export capabilities for further analysis
- **🎨 Enhanced User Experience**: Color-coded terminal output with professional formatting

## 🛠️ Installation

### Prerequisites
- Go 1.19 or higher

### Build from Source
```bash
git clone https://github.com/cypherguard/vhunter.git
cd vhunter
go build -o vhunter vhunter.go
```

### Download Binary
Check the [Releases](https://github.com/cypherguard/vhunter/releases) page for pre-compiled binaries.

## 🚀 Usage

### Basic Usage
```bash
./vhunter -u target.com -w wordlist.txt
```

### Advanced Usage
```bash
./vhunter -u https://target.com -w wordlist.txt -t 20 -o results.txt -v
```

## 📋 Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-u`  |  Target IP or domain (required)
| `-w`  |  Path to vhost wordlist (required) 
| `-t`  |  Number of concurrent threads |   10 |
| `-fc` |  Comma-separated response codes to ignore 
| `-sc` |  Comma-separated response codes to ONLY show
| `-fs` |  Filter out responses with exact body size (bytes)
| `-rate` |  Requests per second rate limit (0 = unlimited) 
| `-timeout` |  HTTP client timeout in seconds |   5 |
| `-X`  |  HTTP method to use (GET, POST, etc.) | GET |
| `-o`  |  Output file to save matched results 
| `-v`  |  Verbose mode: print each matched vhost 
| `-h`  |  Show help message 

##  Examples

### Basic Virtual Host Discovery
```bash
./vhunter -u example.com -w subdomains.txt
```

### Filter Out Common False Positives
```bash
./vhunter -u example.com -w subdomains.txt -fc 404,403
```

### Show Only Successful Responses
```bash
./vhunter -u example.com -w subdomains.txt -sc 200,301,302
```

### High-Speed Scanning with Rate Limiting
```bash
./vhunter -u example.com -w subdomains.txt -t 50 -rate 100
```

### Save Results to File
```bash
./vhunter -u example.com -w subdomains.txt -o discovered_vhosts.txt
```

### POST Method with Custom Settings
```bash
./vhunter -u example.com -w subdomains.txt -X POST -timeout 10 -v
```

##  Output Format

VHunter provides colored output for better readability:

- 🟢 **Green** (2xx): Successful responses
- 🟡 **Yellow** (3xx): Redirection responses  
- 🔴 **Red** (4xx-5xx): Client/Server errors

Example output:
```
[200] admin.example.com size: 2048 bytes
[301] api.example.com size: 156 bytes
[403] secure.example.com size: 1024 bytes
```



## 📊 Performance Tips

- **Threads**: Increase `-t` for faster scanning (be mindful of target server limits)
- **Rate Limiting**: Use `-rate` to avoid triggering rate limits or WAF
- **Size Filtering**: Use `-fs` to filter out common response sizes
- **Code Filtering**: Use `-fc` or `-sc` to focus on relevant responses

##  Responsible Usage

-  Only test on systems you own or have explicit permission to test
-  Respect rate limits and server resources
-  Follow responsible disclosure practices
-  Use appropriate rate limiting in production environments

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by various virtual host discovery tools

##  Support

If you encounter any issues or have questions:
- 🐛 [Open an issue](https://github.com/cypherguard/vhunter/issues)
- 💬 [Start a discussion](https://github.com/cypherguard/vhunter/discussions)

---


