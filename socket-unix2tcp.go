/**
 * Basic server to read tcp/ip and put all content into unix socket.
 *
 * Inspired by https://coderwall.com/p/wohavg/creating-a-simple-tcp-server-in-go
 * curl localhost:4444/v1.28/version -X GET
 *
 */
package main

import (
    "fmt"
    "net"
    "os"
)

const (
    CONN_TYPE = "tcp"
    DOCKER_UNIX_SOCKET = "/var/run/docker.sock"
    BUFFER_SIZE = 10
)

func main() {
  if len(os.Args) < 3 {
    fmt.Fprintln(os.Stderr, "usage:", os.Args[0], "<ip> <port> [/var/run/docker.sock]")
    return
  }

  ip := os.Args[1]
  port := os.Args[2]
  var docker_unix_socket string = DOCKER_UNIX_SOCKET

  if len(os.Args) == 4 {
    docker_unix_socket = os.Args[3]
  }

  // Listen for incoming connections.
  l, err := net.Listen(CONN_TYPE, ip + ":" + port)
  if err != nil {
    fmt.Println("Error listening:", err.Error())
    os.Exit(1)
  }
  // Close the listener when the application closes.
  defer l.Close()
  fmt.Println("Listening on " + ip + ":" + port + " and write into " + docker_unix_socket)
  for {
    // Listen for an incoming connection.
    conn, err := l.Accept()
    if err != nil {
      fmt.Println("Error accepting: ", err.Error())
      os.Exit(1)
    }
    // Handle connections in a new goroutine.
    go handleRequest(conn, docker_unix_socket)
  }
}

/**
 * Read content from socket to put into another socket.
 *
 * @param inputSocket read socket
 * @param outputSocket write socket
 */
func readFromWriteTo(inputSocket net.Conn, outputSocket net.Conn) {
  // Make a buffer to hold incoming data.
  buf := make([]byte, BUFFER_SIZE)

  for {
    // Read the incoming connection into the buffer.
    inLen, err := inputSocket.Read(buf)

    if inLen > 0 {
      fmt.Printf("%s", buf[0:inLen])
      //fmt.Println("Nb:", inLen)

      outLen, err := outputSocket.Write(buf[0:inLen])

      if err != nil {
        // Error, stop
        fmt.Println("Error writing:", err.Error())
      } else if (outLen != inLen) {
        fmt.Println("Error write != read", outLen, inLen)
      }
    }

    if err != nil {
      // Error, stop
      fmt.Println("Error reading:", err.Error())
      break
    } else if inLen < BUFFER_SIZE {
      // End of request
      fmt.Println("End of request")
      break
    }
  }
}

// Handles incoming requests.
func handleRequest(conn net.Conn, docker_unix_socket string) {
  // Open unix socket
  sockerConnection, err := net.Dial("unix", docker_unix_socket)
  if err != nil {
    // Error
    fmt.Println("Error socker:", err.Error())
  }

  readFromWriteTo(conn, sockerConnection)

  readFromWriteTo(sockerConnection, conn)

  sockerConnection.Close()

  // Close the connection when you're done with it.
  conn.Close()
}
