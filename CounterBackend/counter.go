package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"text/template"
)

var (
	counter int
	mutex   sync.Mutex
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/increment", counterHandler)
	http.HandleFunc("/counter", counterPartialHandler)

	log.Println("Counter backend with frontend running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func counterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
		return
	}

	mutex.Lock()
	counter++
	currentCount := counter
	mutex.Unlock()

	fmt.Fprintf(w, "%d", currentCount)
}

func counterPartialHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	currentCount := counter
	mutex.Unlock()

	fmt.Fprintf(w, "%d", currentCount)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
	<title>Counter</title>
	<script src="https://unpkg.com/htmx.org"></script>
</head>
<body>
	<h1>Counter Backend</h1>
  <p>Current Count: <span id="count" hx-get="/counter" hx-trigger="every 1s" hx-swap="innerHTML">{{.Count}}</span></p>
</body>
</html>
  `
	t := template.Must(template.New("index").Parse(tmpl))
	mutex.Lock()
	data := struct {
		Count int
	}{
		Count: counter,
	}

	mutex.Unlock()
	err := t.Execute(w, data)
	if err != nil {
		return
	}
}
