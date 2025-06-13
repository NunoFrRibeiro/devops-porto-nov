package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

var (
	counterPort        = "8081"
	counterContextRoot = "/increment"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addHandler)

	log.Println("Adder backend running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	backendHost := fmt.Sprintf("%s:%s", getHostWithoutPort(host), counterPort)
	counterBackend := fmt.Sprintf("http://%s%s", backendHost, counterContextRoot)
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	resp, err := http.Post(counterBackend, "application/json", nil)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		log.Println("Error contacting counter backend: ", err)
		return
	}
	err = resp.Body.Close()
	if err != nil {
		log.Panic()
	}

	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprintln(w, "Increment sent to counter backend.")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Write failed: %v\n", err)
	}
}

func getHostWithoutPort(host string) string {
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
	<title>Adder</title>
	<script src="https://unpkg.com/htmx.org"></script>
</head>
<body>
	<h1>Adder Backend</h1>
	<form hx-post="/add" hx-target="#message">
		<button type="submit">Send +1</button>
	</form>
	<p id="message"></p>
</body>
</html>
  `

	t := template.Must(template.New("index").Parse(tmpl))
	err := t.Execute(w, nil)
	if err != nil {
		return
	}
}
