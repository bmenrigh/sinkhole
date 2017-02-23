package main


import (
	"fmt"
	"net"
	"os"
	"time"
	"encoding/hex"
	"encoding/json"
	"bytes"
	"bufio"
	"regexp"
)


const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)


type log_req struct {
	F_src_ip          string `json:"src_ip,omitempty"`
	F_src_port        uint16 `json:"src_port,omitempty"`
	F_dst_name        string `json:"dst_name,omitempty"`
	F_url_path        string `json:"url_path,omitempty"`
	F_bytes_client    uint32 `json:"bytes_client,omitempty"`
	F_http_method     string `json:"http_method,omitempty"`
	F_http_version    string `json:"http_version,omitempty"`
	F_x_forwarded_for string `json:"x_forwarded_for,omitempty"`
	F_http_referer    string `json:"http_referer,omitempty"`
	F_http_user_agent string `json:"http_referer,omitempty"`
}


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

	var  req log_req

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

	req_re := regexp.MustCompile(`^(GET)\s(\S+)\s(HTTP\/1\.[01])$`)

	// read first line of HTTP request
	bufline, lineprefix, err := bufreader.ReadLine()
	if (err == nil) {
		if (lineprefix == false) {
			fmt.Printf("Got first line: %s\n", bufline)
			matches := req_re.FindStringSubmatch(string(bufline))

			if (matches != nil) {
				fmt.Printf("method: %s; path: %s; ver: %s\n",
					matches[1], matches[2], matches[3])

				req.F_http_method = string(matches[1])
				req.F_url_path = string(matches[2])
				req.F_http_version = string(matches[3])
			} else {
				fmt.Printf("Got nil matches!\n")
			}
		} else {
			fmt.Printf("Got truncated first line: %s\n", bufline)
		}
	} else {
		return; // Couldn't read first line
	}

	header_re := regexp.MustCompile(`^([A-Za-z][A-Za-z0-9-]*):\s(.*)$`)
	// Read any headers
	for {
		bufline, lineprefix, err := bufreader.ReadLine()

		if (err != nil) {
			break;
		}

		if (lineprefix == true) {
			break;
		}

		bufstr := string(bufline)
		if (bufstr == "") {
			break;
		}

		matches := header_re.FindStringSubmatch(bufstr)
		if (matches != nil) {
			fmt.Printf("Header: %s; Value: %s\n",
				matches[1], matches[2])
		} else {
			break;
		}

	}

	fmt.Printf("Struct method is %s\n", req.F_http_method)
	json_req, err :=  json.Marshal(req)
	if (err == nil) {
		fmt.Printf("JSON: %s\n", json_req)
	}

	// Send a response back to person contacting us.
	conn.Write([]byte("Message received.\r\n"))
}
