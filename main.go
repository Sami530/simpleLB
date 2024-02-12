package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() string
	isAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

type simpleserver struct {
	addr  string
	proxy *httputil.ReverseProxy
}

type loadbalacer struct {
	port          string
	roundRobinCnt int
	servers       []Server
}

func NewLoadbalancer(port string, servers []Server) *loadbalacer {
	return &loadbalacer{
		port:          port,
		roundRobinCnt: 0,
		servers:       servers,
	}
}

func newSimpleServer(addr string) *simpleserver {
	serveUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleserver{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serveUrl),
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func (s *simpleserver) Address() string { return s.addr }

func (s *simpleserver) isAlive() bool { return true }

func (s *simpleserver) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

func (lb *loadbalacer) gettheNextAvlServer() Server {
	server := lb.servers[lb.roundRobinCnt%len(lb.servers)]

	for !server.isAlive() {
		lb.roundRobinCnt++
		server = lb.servers[lb.roundRobinCnt%len(lb.servers)]
	}
	lb.roundRobinCnt++
	return server
}

func (lb *loadbalacer) serverProxy(rw http.ResponseWriter, req *http.Request) {
	hittingserver := lb.gettheNextAvlServer()
	fmt.Printf("forwarding your request to the address %q\n", hittingserver.Address())
	hittingserver.Serve(rw, req)
}
func main() {
	servers := []Server{
		newSimpleServer("https://www.google.com"),
		newSimpleServer("https://www.bing.com"),
		newSimpleServer("http://www.facebook.com"),
	}

	lb := NewLoadbalancer("8000", servers)

	handleredirect := func(rw http.ResponseWriter, req *http.Request) {
		fmt.Println("Received request")
		lb.serverProxy(rw, req)
	}
	http.HandleFunc("/", handleredirect)

	fmt.Printf("Serving your request at the 'localhost:%s' hang tight!!\n", lb.port)

	err := http.ListenAndServe(":"+lb.port, nil)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

// func main() {
// 	servers := []Server{
// 		newSimpleServer("https://www.google.com"),
// 		newSimpleServer("https://www.bing.com"),
// 		newSimpleServer("http://www.facebook.com"),
// 	}

// 	lb := NewLoadbalancer("8000", servers)

// 	handleredirect := func(rw http.ResponseWriter, req *http.Request) {
// 		lb.serverProxy(rw, req)
// 	}
// 	http.HandleFunc("/", handleredirect)

// 	fmt.Printf("Serving your request at the 'localhost:%s' hang tight!!\n", lb.port)

// 	http.ListenAndServe(":"+lb.port, nil)
// }
