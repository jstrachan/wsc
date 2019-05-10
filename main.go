package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/tidwall/pretty"
	"golang.org/x/net/websocket"
)

type headers []string

func (h *headers) String() string {
	return strings.Join(*h, ", ")
}

func (h *headers) Set(value string) error {
	*h = append(*h, value)
	return nil
}

var useColor = true

func main() {
	var (
		target   = flag.String("u", "ws://localhost:8080/api/v1/kubeql", "The URL to connect to")
		origin   = flag.String("o", "http://localhost:8080", "The origin to use in the WS request")
		cmdText  = flag.String("c", "", "The initial command to end to the WS request")
		noColour = flag.Bool("p", false, "Use plain (no colour) output")
		h        headers
		origURL  *url.URL
	)
	flag.Var(&h, "H", `Headers to use in the WS request, can be used to multiple times to specify multiple headers.`+
		` Example: -H "Sample-Header-1: foo" -H "Sample-Header-2: bar"`)
	flag.Parse()

	if noColour != nil && *noColour {
		useColor = false
	}

	if *target == "" {
		fmt.Fprintf(os.Stderr, "missing url\n")
		os.Exit(1)
	}

	if *origin != "" {
		var err error
		origURL, err = url.Parse(*origin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse origin URL: %s", err.Error())
			os.Exit(1)
		}
	}
	ws := connect(*target, makeHeader(h), origURL)
	trapCtrlC(ws)
	if *cmdText != "" {
		writeText(ws, *cmdText)
	}
	go write(ws)
	read(ws)
}

func makeHeader(h headers) http.Header {
	httpH := make(http.Header)
	for _, hv := range h {
		splits := strings.SplitN(hv, ":", 2)
		httpH.Add(strings.TrimSpace(splits[0]), strings.TrimSpace(splits[1]))
	}
	return httpH
}

func connect(addr string, h http.Header, origin *url.URL) *websocket.Conn {
	log.Printf("connecting to %s...", addr)
	conf, err := websocket.NewConfig(addr, addr)
	if err != nil {
		log.Fatal(err)
	}
	conf.Header = h
	conf.Origin = origin
	ws, err := websocket.DialConfig(conf)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("ready, exit with CTRL+C.")
	return ws
}

// Graceful shutdown
func trapCtrlC(c *websocket.Conn) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		for range ch {
			fmt.Println("\nexiting")
			c.Close()
			os.Exit(0)
		}
	}()
}

// Send STDIN lines to websocket server.
func write(ws *websocket.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		t := scanner.Text()
		writeText(ws, t)
	}
}

func writeText(ws *websocket.Conn, t string) {
	ws.Write([]byte(t))
	fmt.Printf(">> %s\n", t)
}

// Read from websocket and print messages to STDOUT
func read(ws *websocket.Conn) {
	msg := make([]byte, 16384)
	for {
		n, err := ws.Read(msg)
		if err != nil {
			log.Fatal(err)
		}
		text := msg[:n]
		result := pretty.Pretty(text)
		if useColor {
			result = pretty.Color(result, nil)
		}
		fmt.Printf("<< %s\n", result)
	}
}
