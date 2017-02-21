package main


import (
	"fmt"
	"net"
	"os"
	"time"
	"encoding/hex"
	"bytes"
	"bufio"
	"regexp"
)


const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)


func main() {
	// Listen for incoming connections.
	sock, err := net.Listen(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer sock.Close()

	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := sock.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}


// Handles incoming requests.
func handleRequest(conn net.Conn) {

	// Close the socket when we're done
	defer conn.Close()

	// Make a buffer to hold incoming data.
	buf := make([]byte, 4096)

	// Limit the time we'll spend reading
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	// Read the incoming connection into the buffer.
	readlen, err := conn.Read(buf)

	if (err != nil) {
		fmt.Println("Error reading:", err.Error())
		return
	}

	hexstr := make([]byte, hex.EncodedLen(readlen))
	hex.Encode(hexstr, buf[:readlen])

	fmt.Printf("Recieved %s\n", hexstr)

	bufreader := bufio.NewReader(bytes.NewReader(buf[:readlen]))
	bufline, lineprefix, err := bufreader.ReadLine()

	req_re := regexp.MustCompile(`^(GET)\s(\S+)\s(HTTP\/1\.[01])$`)

	if (err == nil) {
		if (lineprefix == false) {
			fmt.Printf("Got first line: %s\n", bufline)
			matches := req_re.FindStringSubmatch(string(bufline))

			if (matches != nil) {
				fmt.Printf("method: %s; path: %s; ver: %s\n",
					matches[1], matches[2], matches[3])
			} else {
				fmt.Printf("Got nil matches!\n")
			}
		} else {
			fmt.Printf("Got truncated first line: %s\n", bufline)
		}
	}

	// Send a response back to person contacting us.
	conn.Write([]byte("Message received.\r\n"))
}
