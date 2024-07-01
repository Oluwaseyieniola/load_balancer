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
	isAlve() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}
type SimpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func newSimpleServer(addr string) *SimpleServer {
	serverUrl, err := url.Parse(addr)

	handleErr(err)

	return &SimpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type LoadBalancer struct {
	port           string
	rounRobinCount int
	servers        []Server
}

func newLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		rounRobinCount: 0,
		port:           port,
		servers:        servers,
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)

	}
}

func (s *SimpleServer) Address() string { return s.addr }

func (s *SimpleServer) isAlve() bool { return true }

func (s *SimpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}
func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.rounRobinCount%len(lb.servers)]
	for !server.isAlve() {
		lb.rounRobinCount++
		server = lb.servers[lb.rounRobinCount%len(lb.servers)]
	}
	lb.rounRobinCount++
	return server
}

func (lb *LoadBalancer) ServerProxy(rw http.ResponseWriter, req *http.Request) {
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("forwarding request%q\n", targetServer.Address())
	targetServer.Serve(rw, req)
}

func main() {
	servers := []Server{
		newSimpleServer("https://www.spotify.com"),
		newSimpleServer("https://www.bing.com"),
		newSimpleServer("https://www.x.com"),
	}
	lb := newLoadBalancer("8000", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.ServerProxy(rw, req)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("server serving request at' localhost:%s'\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
