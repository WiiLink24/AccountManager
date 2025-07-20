package main

import (
	"encoding/json"
	"errors"
	"github.com/WiiLink24/AccountManager/middleware"
	"github.com/logrusorgru/aurora/v4"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const SocketSuccess = `{"success": true}`

type JustEatPayload struct {
	WiiNumber string `json:"wii_number"`
	Auth      string `json:"auth"`
}

type SocketFailResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func socketFail(err error) []byte {
	data, _ := json.Marshal(SocketFailResponse{
		Success: false,
		Error:   err.Error(),
	})
	return data
}

func justEatSocketListen() {
	// Remove if it didn't gracefully exit for some reason
	os.Remove("/tmp/eater.sock")

	socket, err := net.Listen("unix", "/tmp/eater.sock")
	checkError(err)

	defer socket.Close()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove("/tmp/eater.sock")
		os.Exit(0)
	}()

	log.Printf("%s", aurora.Green("UNIX socket connected."))
	log.Printf("%s %s\n", aurora.Green("Listening on UNIX socket:"), socket.Addr())

	// Listen forever
	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Print(aurora.Red("Socket Accept ERROR: "), err.Error(), "\n")
		}

		go func(conn net.Conn) {
			defer conn.Close()
			buf := make([]byte, 4096)

			n, err := conn.Read(buf)
			if err != nil {
				log.Print(aurora.Red("Socket Read ERROR: "), err.Error(), "\n")
				return
			}

			reply := []byte(SocketSuccess)
			payload := strings.Replace(string(buf[:n]), "\n", "", -1)

			var justEatPayload JustEatPayload
			err = json.Unmarshal([]byte(payload), &justEatPayload)
			if err != nil {
				log.Print(aurora.Red("Unmarshal ERROR: "), err.Error(), "\n")
				conn.Write(append(socketFail(err), []byte("\n")...))
				return
			}

			claims, status := middleware.GetClaims(verifier, justEatPayload.Auth)
			if status != http.StatusOK {
				log.Print(aurora.Red("Authentication failure."), "\n")
				conn.Write(append(socketFail(errors.New("authentication error")), []byte("\n")...))
				return
			}

			// Toggle the linkage
			if claims.JustEat[justEatPayload.WiiNumber] {
				claims.JustEat[justEatPayload.WiiNumber] = false
			} else {
				claims.JustEat[justEatPayload.WiiNumber] = true
			}

			// Now send off to authentik.
			newPayload := map[string]any{
				"attributes": map[string]any{
					"wiis":     claims.Wiis,
					"wwfc":     claims.WWFC,
					"dominos":  claims.Dominos,
					"just_eat": claims.JustEat,
				},
			}

			err = updateUserRequest(claims.UserId, newPayload)
			if err != nil {
				log.Print(aurora.Red("Authentication failure."), "\n")
				conn.Write(append(socketFail(err), []byte("\n")...))
				return
			}

			_, err = conn.Write(append(reply, []byte("\n")...))
			if err != nil {
				log.Print(aurora.Red("Socket Write ERROR: "), err.Error(), "\n")
				return
			}
		}(conn)
	}
}
