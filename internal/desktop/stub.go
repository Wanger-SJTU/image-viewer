//go:build !desktop

package desktop

import (
	"fmt"
	"net/http"
	"os"
)

// Run is a stub — desktop mode requires building with "-tags desktop".
func Run(handler http.Handler, startPort int) {
	fmt.Fprintln(os.Stderr, "Desktop mode not available in this build.")
	fmt.Fprintln(os.Stderr, "Rebuild with: go build -tags desktop -o viewer .")
	fmt.Fprintln(os.Stderr, "Falling back to server mode at http://127.0.0.1:8080")
	http.ListenAndServe("127.0.0.1:8080", handler)
}
