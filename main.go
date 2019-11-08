package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "keepalive",
			Usage: "set keepalive or not",
		},
		cli.IntFlag{
			Name:  "keepalive-time",
			Value: 75,
			Usage: "keepalive-time in seconds",
		},
	}

	server_flags := []cli.Flag{
		cli.StringFlag{
			Name:  "bind",
			Value: "0.0.0.0",
			Usage: "Bind host for server, default 0.0.0.0",
		},
		cli.IntFlag{
			Name:  "port",
			Value: 8081,
			Usage: "Port to listen/connect",
		},
		cli.StringFlag{
			Name:  "hello-server",
			Value: "hello-server string",
			Usage: "hello-server string",
		},
	}

	client_flags := []cli.Flag{
		cli.StringFlag{
			Name:  "connect",
			Value: "127.0.0.1",
			Usage: "Connect to host for client, default 127.0.0.1",
		},
		cli.IntFlag{
			Name:  "port",
			Value: 8081,
			Usage: "Port to listen/connect",
		},
		cli.StringFlag{
			Name:  "hello-client",
			Value: "hello-client string",
			Usage: "hello-client string",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "run as server",
			Flags:   server_flags,
			Action: func(c *cli.Context) error {
				fmt.Println("added task: ", c.Args(), c.GlobalBool("keepalive"))
				main_server(c)
				return nil
			},
		},
		{
			Name:    "client",
			Aliases: []string{"c"},
			Usage:   "run as client",
			Flags:   client_flags,
			Action: func(c *cli.Context) error {
				fmt.Println("completed task: ", c.Args())
				main_client(c)
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func main_server(c *cli.Context) {

	listenAddr := fmt.Sprintf("%s:%d", c.String("bind"), c.Int("port"))
	log.Println("Launching server... on addr ", listenAddr)

	// listen on all interfaces
	ln, _ := net.Listen("tcp", listenAddr)

	// accept connection on port
	conn, _ := ln.Accept()
	tcpConn := conn.(*net.TCPConn)

	if c.GlobalBool("keepalive") {
		log.Println("TCP-KEEPALIVE :: Enable tcp-keepalive")
		tcpConn.SetKeepAlive(true)

		durStr := fmt.Sprintf("%ds", c.GlobalInt("keepalive-time"))
		log.Println("TCP-KEEPALIVE :: Set tcp socket keepalive as ", durStr)
		m, _ := time.ParseDuration(durStr)
		tcpConn.SetKeepAlivePeriod(m)
	} else {
		log.Println("Disable tcp-keepalive")
		tcpConn.SetKeepAlive(false)
	}

	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		// output message received
		log.Println("Message Received:", string(message))
		// sample process for string received
		newmessage := strings.ToUpper(message)
		// send new string back to client
		conn.Write([]byte(newmessage + "\n"))
	}
}

func main_client(c *cli.Context) {

	connectAddr := fmt.Sprintf("%s:%d", c.String("connect"), c.Int("port"))
	log.Println("Launching Client to addr ", connectAddr)

	// connect to this socket
	conn, _ := net.Dial("tcp", connectAddr)
	tcpConn := conn.(*net.TCPConn)

	if c.GlobalBool("keepalive") {
		log.Println("TCP-KEEPALIVE :: Enable tcp-keepalive")
		tcpConn.SetKeepAlive(true)

		durStr := fmt.Sprintf("%ds", c.GlobalInt("keepalive-time"))
		log.Println("TCP-KEEPALIVE :: Set tcp socket keepalive as ", durStr)
		m, _ := time.ParseDuration(durStr)
		tcpConn.SetKeepAlivePeriod(m)
	} else {
		log.Println("Disable tcp-keepalive")
		tcpConn.SetKeepAlive(false)
	}

	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Text to send: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		// tcpConn.Write(text + "\n")

		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		log.Println("Message from server: " + message)
	}
}
