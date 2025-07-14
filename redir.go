package main

import (
        "fmt"
        "io"
        "log"
        "net"
        "net/http"

        "github.com/fatih/color"
)

var (
        red    = color.New(color.FgRed).SprintFunc()
        green  = color.New(color.FgGreen).SprintFunc()
        yellow = color.New(color.FgYellow).SprintFunc()
        blue   = color.New(color.FgBlue).SprintFunc()
        cyan   = color.New(color.FgCyan).SprintFunc()
)

func main() {
        protocols := []string{"TCP", "UDP", "HTTP"}
        fmt.Println(cyan("\n== Multi-Protocol Redirector ==\n"))

        // Show protocol options
        fmt.Println(blue("Select the protocol to redirect:"))
        for i, p := range protocols {
                fmt.Printf("%d) %s\n", i+1, p)
        }

        // Get user input for protocol choice
        var protocolChoice int
        fmt.Print("Enter the protocol number (1-3): ")
        fmt.Scanln(&protocolChoice)

        if protocolChoice < 1 || protocolChoice > len(protocols) {
                log.Fatal(red("Invalid protocol choice. Exiting..."))
        }

        protocol := protocols[protocolChoice-1]

        // Get destination details
        var localPort, destIP, destPort string
        fmt.Print("Enter local port to listen on: ")
        fmt.Scanln(&localPort)
        fmt.Print("Enter destination IP: ")
        fmt.Scanln(&destIP)
        fmt.Print("Enter destination port: ")
        fmt.Scanln(&destPort)

        // Call the appropriate redirect function
        switch protocol {
        case "TCP":
                startTCPRedirect(localPort, destIP, destPort)
        case "UDP":
                startUDPRedirect(localPort, destIP, destPort)
        case "HTTP":
                startHTTPRedirect(localPort, destIP, destPort)
        default:
                log.Fatal(red("Protocol not supported"))
        }
}

// TCP redirector
func startTCPRedirect(localPort, destIP, destPort string) {
        listenAddr := ":" + localPort
        destAddr := destIP + ":" + destPort

        listener, err := net.Listen("tcp", listenAddr)
        if err != nil {
                log.Fatal(red("Failed to start TCP listener: %v", err))
        }
        defer listener.Close()

        fmt.Printf(green("\nTCP redirector started on %s -> %s\n", listenAddr, destAddr))

        for {
                client, err := listener.Accept()
                if err != nil {
                        fmt.Printf(yellow("Error accepting connection: %v\n", err))
                        continue
                }
                go redirectTCP(client, destAddr)
        }
}

// Handle TCP redirection
func redirectTCP(client net.Conn, destAddr string) {
        server, err := net.Dial("tcp", destAddr)
        if err != nil {
                fmt.Printf(red("Failed to connect to destination: %v\n", err))
                client.Close()
                return
        }

        go io.Copy(server, client)
        go io.Copy(client, server)
}

// UDP redirector
func startUDPRedirect(localPort, destIP, destPort string) {
        localAddr, err := net.ResolveUDPAddr("udp", ":"+localPort)
        if err != nil {
                log.Fatal(red("Failed to resolve UDP address: %v", err))
        }

        remoteAddr, err := net.ResolveUDPAddr("udp", destIP+":"+destPort)
        if err != nil {
                log.Fatal(red("Failed to resolve remote UDP address: %v", err))
        }

        conn, err := net.ListenUDP("udp", localAddr)
        if err != nil {
                log.Fatal(red("Failed to start UDP listener: %v", err))
        }
        defer conn.Close()

        fmt.Printf(green("\nUDP redirector started on %s -> %s\n", localAddr, remoteAddr))

        buf := make([]byte, 4096)
        for {
                n, addr, err := conn.ReadFromUDP(buf)
                if err != nil {
                        fmt.Printf(yellow("Error reading UDP: %v\n", err))
                        continue
                }

                // Redirect UDP data
                _, err = conn.WriteToUDP(buf[:n], remoteAddr)
                if err != nil {
                        fmt.Printf(red("Error sending UDP data: %v\n", err))
                        continue
                }

                fmt.Printf(blue("UDP data redirected from %s -> %s\n", addr, remoteAddr))
        }
}

// HTTP redirector (basic)
func startHTTPRedirect(localPort, destIP, destPort string) {
        httpAddr := ":" + localPort
        destAddr := destIP + ":" + destPort

        fmt.Printf(green("\nHTTP redirector started on %s -> %s\n", httpAddr, destAddr))

        // Listen for HTTP requests
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                client := &http.Client{}
                req, err := http.NewRequest(r.Method, "http://"+destAddr+r.URL.Path, r.Body)
                if err != nil {
                        http.Error(w, "Failed to create request", http.StatusInternalServerError)
                        return
                }

                // Forward headers
                req.Header = r.Header

                resp, err := client.Do(req)
                if err != nil {
                        http.Error(w, "Failed to connect to destination", http.StatusInternalServerError)
                        return
                }
                defer resp.Body.Close()

                // Forward response
                for key, values := range resp.Header {
                        for _, value := range values {
                                w.Header().Add(key, value)
                        }
                }

                w.WriteHeader(resp.StatusCode)
                io.Copy(w, resp.Body)
        })

        log.Fatal(http.ListenAndServe(httpAddr, nil))
}