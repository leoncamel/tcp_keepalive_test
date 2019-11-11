package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.0.2"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Xiaolin Zhang",
			Email: "leoncamel@gmail.com",
		},
	}
	app.Copyright = "(c) 2019 Xiaolin Zhang"
	app.Usage = "An tiny program for testing TCP keepalive"

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
		cli.BoolFlag{
			Name:  "interactive,i",
			Usage: "Run client interactive from stdin",
		},
		cli.StringFlag{
			Name:  "seq",
			Value: "1000,1000,1000",
			Usage: "Generate with specific delay, in ms. '100{3}' means delay interval is 100ms, delay for 3 iterations",
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
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "test",
			Action: func(c *cli.Context) error {
				fmt.Println("Run test command, with args", c.Args())
				main_test(c.Args()[0])
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

	var data_channel chan string

	if c.Bool("interactive") {
		data_channel = data_from_stdin()
	} else {
		data_channel = data_from_seq(c.String("seq"))
	}

	for text := range data_channel {
		data_to_send := "DATA:" + time.Now().Format("2006-01-02-15 04:05") + ": " + text + "\n"

		// send to socket
		fmt.Fprintf(conn, data_to_send)
		// tcpConn.Write(text + "\n")
		log.Printf("Msg sent: '%s'", data_to_send)

		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		log.Println("Message from server: " + message)
	}
}

func data_from_stdin() chan string {
	c := make(chan string)

	go func() {
		for {
			// read in input from stdin
			reader := bufio.NewReader(os.Stdin)
			fmt.Println("Text to send: ")
			text, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}

			c <- text
		}
	}()

	return c
}

func data_from_seq(seq_delay_defs string) chan string {
	delays_array := strings.Split(seq_delay_defs, ",")

	c := make(chan string)

	go func() {
		for idx, delay_str := range delays_array {
			retMap := parseDelay(delay_str)
			var iternum int64 = 0
			var err error
			if retMap["iternum"] != "" {
				iternum, err = strconv.ParseInt(retMap["iternum"], 10, 32)
				if err != nil {
					iternum = 0
				}
			}

			// fmt.Println("iternum %v", iternum)
			delay, _ := strconv.ParseInt(retMap["delay"], 10, 32)
			for i := 0; i < int(iternum) + 1; i++ {
				time.Sleep(time.Duration(delay) * time.Millisecond)
				log.Printf("Delay idx:%8d\n", i)
			}
			c <- fmt.Sprint(idx, " ", delay_str, "\n")
		}

		close(c)
	}()

	return c
}

func main_test(arg string) {
	fmt.Println("hello from main_test")
	fmt.Println(arg)
	delays_array := strings.Split(arg, ",")
	fmt.Println(delays_array)

	for idx, delay_str := range delays_array {
		fmt.Println(idx, delay_str)
		retMap := parseDelay(delay_str)
		fmt.Println(retMap)
		fmt.Println(retMap["delay"], retMap["iternum"])
		if retMap["iternum"] != "" {
			fmt.Println("not nil")
		}
	}
}

func parseDelay(arg string) (paramsMap map[string]string) {

	var compRegEx = regexp.MustCompile(`(?P<delay>\d+)({(?P<iternum>\d+)?})?`)
	match := compRegEx.FindStringSubmatch(arg)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	return
}
