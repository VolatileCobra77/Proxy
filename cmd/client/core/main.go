package core

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/armon/go-socks5"
)

type Config struct {
	ControlServer         string    `json:"server"`
	HTTPProxyAddress      string    `json:"http_proxy"`
	SOCKSProxyAddress     string    `json:"socks_proxy"`
	ProxyDiscoveryAddress string    `json:"proxy_discovery_address"`
	Mode                  ProxyMode `json:"mode"`
	AutoSetSystem         bool      `json:"auto_set_system_proxy"`
	FilteringEnabled      bool      `json:"filtering_enabled"`
	WhitelistMode         bool      `json:"use_whitelist"`
	Blacklist             []string  `json:"blocked_sites"`
	Whitelist             []string  `json:"allowed_sites"`
	//Autostart     bool   `json:"autostart"`
}

// ProxyMode is a string type for your proxy mode enum
type ProxyMode string

// Define constants for allowed values
const (
	ModeWindows ProxyMode = "WINDOWS"
	ModeBrowser ProxyMode = "BROWSER"
)

func (m ProxyMode) String() string {
	switch m {
	case ModeWindows:
		return "WINDOWS"
	case ModeBrowser:
		return "BROWSER"
	default:
		return "UNKNOWN"
	}
}

var CONFIGS *Config

const HAS_ROOT_ACCESS bool = false

type ProxyManager struct {
	httpListener  net.Listener
	socksListener net.Listener
	backendAddr   string

	running bool
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	log.Println("Loading configs from " + path)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("Configs file not present, creating one")
			cfg = Config{
				"localhost:8080",
				":8081",
				":8082",
				":8083",
				ModeBrowser,
				false,
				false,
				false,
				[]string{},
				[]string{},
			}
			if err := SaveConfig(path, &cfg); err != nil {
				return nil, err
			}
			return &cfg, nil
		}
		return nil, err
	}
	// File exists → parse JSON
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
func SaveConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	log.Println("Creating configs at " + path)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func main() {
	fmt.Println("Loading from CLI, default configs used")

	proxyManager := ProxyManager{}
	cfg, err := LoadConfig("clientConfig.json")
	fmt.Println("Address: " + cfg.ControlServer)
	if err != nil {
		log.Fatal(err)
	}
	CONFIGS = cfg
	proxyManager.Start(cfg)
	for proxyManager.running {

	}
}

func (manager *ProxyManager) Start(configs *Config) error {
	if CONFIGS == nil {
		CONFIGS = configs
	}
	if manager.running {
		return nil
	}
	manager.backendAddr = configs.ControlServer
	manager.running = true
	if CONFIGS.Mode == ModeBrowser {

		go func() {
			manager.httpListener = startHttps(manager)
		}()
	}
	go func() {
		manager.socksListener = startSocks5(manager)

	}()
	go func() {
		startWebserver()
	}()
	return nil
}

func (manager *ProxyManager) Stop() error {
	if !manager.running {
		return nil
	}
	if manager.httpListener != nil {
		err := manager.httpListener.Close()
		if err != nil {
			return err
		}
	}
	if manager.socksListener != nil {
		err := manager.socksListener.Close()
		if err != nil {
			return err
		}
	}
	manager.running = false
	return nil
}

func startWebserver() {
	http.HandleFunc("/proxy.pac", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		var proxyAddr string

		if CONFIGS.Mode == ModeWindows {
			proxyAddr = CONFIGS.SOCKSProxyAddress
		} else {
			proxyAddr = CONFIGS.HTTPProxyAddress
		}

		pac := fmt.Sprintf(`function FindProxyForURL(url, host) {
        return "PROXY %s";
    }`, proxyAddr)
		w.Write([]byte(pac))
	})

	go http.ListenAndServe(CONFIGS.ProxyDiscoveryAddress, nil)

}

func startHttps(manager *ProxyManager) net.Listener {
	fmt.Println("client starting...")
	listener, err := net.Listen("tcp", CONFIGS.HTTPProxyAddress)
	log.Println("HTTP Proxy listnening on :8081")
	if err != nil {
		log.Fatal(err)
	}

	for manager.running {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		go handleHTTP(conn)

	}

	return listener

}

func startSocks5(manager *ProxyManager) net.Listener {
	conf := &socks5.Config{
		Dial: socksDial,
	}

	server, err := socks5.New(conf)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("SOCKS5 listening on 127.0.0.1:8082")

	listener, err := net.Listen("tcp", CONFIGS.SOCKSProxyAddress)
	if err != nil {
		log.Fatal(err)
	}
	for manager.running {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			err := server.ServeConn(conn)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	return listener
}

func socksDial(ctx context.Context, network, addr string) (net.Conn, error) {
	// addr is "host:port"
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	if CONFIGS.FilteringEnabled {
		if CONFIGS.WhitelistMode {
			if !slices.Contains(CONFIGS.Whitelist, host) {
				ctx.Done()
				return nil, nil
			}
		} else {
			if slices.Contains(CONFIGS.Blacklist, host) {
				ctx.Done()
				return nil, nil
			}
		}
	}
	tlsConfigs := &tls.Config{
		InsecureSkipVerify: true, // only for testing/self-signed certs
	}

	serverConn, err := tls.Dial("tcp", CONFIGS.ControlServer, tlsConfigs)
	if err != nil {
		return nil, err
	}

	// Send your control protocol
	_, err = serverConn.Write([]byte(
		"SOCKS\n" + host + "\n" + port + "\n\n",
	))
	if err != nil {
		serverConn.Close()
		return nil, err
	}

	return serverConn, nil
}

func handleHTTP(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic:", r)
		}
		conn.Close()
	}()
	tlsConfigs := &tls.Config{
		InsecureSkipVerify: true, // only for testing/self-signed certs
	}

	reader := bufio.NewReader(conn)
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
	}
	log.Println("First line recieved ", firstLine)
	if firstLine == "EOF" {
		return
	}

	list := strings.Split(firstLine, " ")
	method := list[0]
	hostPortCombo := strings.Split(list[1], ":")
	host := strings.TrimSpace(hostPortCombo[0])
	if CONFIGS.FilteringEnabled {
		if CONFIGS.WhitelistMode {
			if !slices.Contains(CONFIGS.Whitelist, host) {
				conn.Close()
			}
		} else {
			if slices.Contains(CONFIGS.Blacklist, host) {
				conn.Close()
			}
		}
	}
	port := strings.TrimSpace(hostPortCombo[1])
	outbound, err := tls.Dial("tcp", CONFIGS.ControlServer, tlsConfigs)
	if err != nil {
		log.Println(err)
		return
	}
	if method == "CONNECT" {
		outbound.Write([]byte(fmt.Sprintf("HTTPS\n%s\n%s\n", host, port)))

	} else {
		outbound.Write([]byte(fmt.Sprintf("HTTP\n%s\n%s\n", host, port)))

	}
	go func() {
		_, _ = io.Copy(outbound, conn)
		conn.Close()
		outbound.Close()
	}()
	_, _ = io.Copy(conn, outbound)

}
