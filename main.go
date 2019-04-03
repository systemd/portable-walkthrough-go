package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func ListenFds() []*os.File {
	// Minimal implementation of systemd's socket activation protocol
	pid, err := strconv.Atoi(os.Getenv("LISTEN_PID"))
	if err != nil || pid != os.Getpid() {
		return nil
	}
	nfds, err := strconv.Atoi(os.Getenv("LISTEN_FDS"))
	if err != nil || nfds == 0 {
		return nil
	}
	files := []*os.File(nil)
	for fd := 3; fd < 3+nfds; fd++ {
		syscall.CloseOnExec(fd)
		files = append(files, os.NewFile(uintptr(fd), ""))
	}
	return files
}

func handler(w http.ResponseWriter, r *http.Request) {
	fn := os.Getenv("STATE_DIRECTORY") + "/counter"

	file, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Let's take a BSD lock to synchronize access to the counter file
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		log.Fatal(err)
	}

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	counter := 0

	trimmed := strings.TrimSpace(string(contents))
	if trimmed != "" {
		counter, err = strconv.Atoi(trimmed)
		if err != nil {
			log.Fatal(err)
		}
	}

	counter++

	fmt.Fprintf(w, "Hello! You are visitor #%d!\n", counter)

	file.Truncate(0)
	file.Seek(0, 0)
	fmt.Fprintf(file, "%d\n", counter)
}

func run_http(f *os.File) {
	s, err := net.FileListener(f)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.Serve(s, nil))
}

func main() {
	http.HandleFunc("/", handler)

	listeners := ListenFds()
	if len(listeners) == 0 {
		log.Fatal(http.ListenAndServe(":8080", nil))
	} else {
		// If multiple sockets are passed, spawn all but the last one as go-routine
		for _, l := range listeners[:len(listeners)-1] {
			go run_http(l)
		}

		run_http(listeners[len(listeners)-1])
	}
}
