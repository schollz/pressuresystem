package main

import (
	"bytes"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/schollz/logger"
	"github.com/schollz/miti/src/log"
	"github.com/tarm/serial"
)

var serialConfig *serial.Config
var s *serial.Port
var mu sync.Mutex

func main() {
	logger.SetLevel("debug")
	err := run()
	if err != nil {
		logger.Error(err)
	}
}

func run() (err error) {

	csig := make(chan os.Signal, 1)
	signal.Notify(csig, os.Interrupt)
	go func() {
		for sig := range csig {
			logger.Debug("shutdown")
			logger.Debug(sig)
			write(s, "voltage0")
			write(s, "sol1off")
			write(s, "sol2off")
			s.Close()
			os.Exit(1)
		}
	}()

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)

	// r := gin.Default()
	// r.StaticFile("/", "./index.html")
	// r.GET("/api", func(c *gin.Context) {
	// 	msg := c.DefaultQuery("msg", "")
	// 	if msg == "" {
	// 		c.JSON(200, gin.H{
	// 			"success": false,
	// 			"message": "no message",
	// 		})
	// 		return
	// 	} else if msg == "open" {
	// 		serialConfig = &serial.Config{Name: "COM5", Baud: 9600, ReadTimeout: time.Second * 1}
	// 		s, err = serial.OpenPort(serialConfig)
	// 		m := ""
	// 		if err != nil {
	// 			err = errors.Wrap(err, "no com port")
	// 			m = err.Error()
	// 		}
	// 		s.Flush()
	// 		c.JSON(200, gin.H{
	// 			"success": err == nil,
	// 			"message": m,
	// 		})
	// 		return
	// 	} else if msg == "ports" {
	// 		ports, err := stserial.GetPortsList()
	// 		m := ""
	// 		if err != nil {
	// 			m = err.Error()
	// 		} else {
	// 			m = "found ports"
	// 		}
	// 		log.Debugf("ports: %+v", ports)
	// 		c.JSON(200, gin.H{
	// 			"success": err == nil,
	// 			"message": m,
	// 			"ports":   ports,
	// 		})
	// 		return
	// 	}
	// 	err = write(s, msg)
	// 	if err != nil {
	// 		logger.Error(err)
	// 		c.JSON(200, gin.H{
	// 			"success": false,
	// 			"message": err.Error(),
	// 		})
	// 		return
	// 	}
	// 	if msg == "read" {
	// 		reply, err := read(s)
	// 		if err != nil {
	// 			logger.Error(err)
	// 			c.JSON(200, gin.H{
	// 				"success": false,
	// 				"message": err.Error(),
	// 			})
	// 			return
	// 		}
	// 		c.JSON(200, gin.H{
	// 			"success": true,
	// 			"message": strings.TrimSpace(reply),
	// 		})
	// 	} else {
	// 		c.JSON(200, gin.H{
	// 			"success": true,
	// 		})
	// 	}
	// })
	// logger.Infof("running on port 8080")

	return
}

// handler is the main handler for all requests
func handler(w http.ResponseWriter, r *http.Request) {
	t := time.Now().UTC()
	err := handle(w, r)
	if err != nil {
		log.Error(err)
		w.Write([]byte(err.Error() + "\n"))
	}
	log.Infof("%v %v %v %s", r.RemoteAddr, r.Method, r.URL.Path, time.Since(t))
}

// handle is the function for the main routing logic of share.
func handle(w http.ResponseWriter, r *http.Request) (err error) {
	if strings.HasPrefix(r.URL.Path, "/open") {

	} else {
		page = "assets/" + strings.TrimPrefix(page, "/static/")
		if strings.Contains(page, ".js") {
			w.Header().Set("Content-Type", "text/javascript")
		} else if strings.Contains(page, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.Contains(page, ".png") {
			w.Header().Set("Content-Type", "image/png")
		} else if strings.Contains(page, ".json") {
			w.Header().Set("Content-Type", "application/json")
		} else {
			w.Header().Set("Content-Type", "text/html")
		}
		w.Write(b)
	}
}

func write(s *serial.Port, data string) (err error) {
	mu.Lock()
	defer mu.Unlock()
	logger.Tracef("writing '%s'", data)
	_, err = s.Write([]byte(data + "\n"))
	if err != nil {
		return
	}
	err = s.Flush()
	return
}

func read(s *serial.Port) (reply string, err error) {
	// logger.Debug("locking")
	// mu.Lock()
	// defer func() {
	// 	mu.Unlock()
	// 	logger.Debug("unlocking")
	// }()
	for {
		buf := make([]byte, 128)
		var n int
		n, err = s.Read(buf)
		reply += string(buf[:n])
		if bytes.Contains(buf[:n], []byte("\n")) {
			break
		}
		if err != nil {
			break
		}
	}
	logger.Tracef("read '%s'", strings.TrimSpace(reply))
	return
}
