//go:build desktop

package desktop

import (
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/webview/webview_go"
)

// Run starts the HTTP server on localhost and opens a webview window.
// It blocks until the window is closed.
func Run(handler http.Handler, startPort int) {
	port := findPort(startPort)

	go func() {
		addr := "127.0.0.1:" + strconv.Itoa(port)
		log.Printf("desktop: server on %s", addr)
		if err := http.ListenAndServe(addr, handler); err != nil {
			log.Fatalf("desktop: server error: %v", err)
		}
	}()

	url := "http://127.0.0.1:" + strconv.Itoa(port)

	w := webview.New(false)
	defer w.Destroy()

	w.SetTitle("Image Viewer")
	w.SetSize(1400, 900, webview.HintFixed)
	w.Navigate(url)
	w.Run()
}

func findPort(start int) int {
	for port := start; port < start+100; port++ {
		addr := "127.0.0.1:" + strconv.Itoa(port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			return port
		}
	}
	log.Fatalf("desktop: no available port in range %d-%d", start, start+100)
	return 0
}
