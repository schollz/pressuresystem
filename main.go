package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"

	log "github.com/schollz/logger"
	"github.com/tarm/serial"
	stserial "go.bug.st/serial.v1"
)

var serialConfig *serial.Config
var s *serial.Port
var mu sync.Mutex

func main() {
	log.SetLevel("debug")
	err := run()
	if err != nil {
		log.Error(err)
	}
}

func run() (err error) {

	csig := make(chan os.Signal, 1)
	signal.Notify(csig, os.Interrupt)
	go func() {
		for sig := range csig {
			log.Debug("shutdown")
			log.Debug(sig)
			if s != nil {
				write(s, "voltage0")
				write(s, "sol1off")
				write(s, "sol2off")
				s.Close()
			}
			os.Exit(1)
		}
	}()

	go func() {
		exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:8080").Start()
	}()

	log.Info("running on port 8080")
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
	// 		log.Error(err)
	// 		c.JSON(200, gin.H{
	// 			"success": false,
	// 			"message": err.Error(),
	// 		})
	// 		return
	// 	}
	// 	if msg == "read" {
	// 		reply, err := read(s)
	// 		if err != nil {
	// 			log.Error(err)
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
	// log.Infof("running on port 8080")

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

type Response struct {
	Message string
	Success bool
	Ports   []string
}

func writeJSON(re Response, w http.ResponseWriter, r *http.Request) (err error) {
	b, err := json.Marshal(re)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return
}

// handle is the function for the main routing logic of share.
func handle(w http.ResponseWriter, r *http.Request) (err error) {
	jsonData := true
	var re Response
	defer func() {
		if jsonData {
			if err != nil {
				re.Message = err.Error()
			}
			re.Success = err == nil
			err = writeJSON(re, w, r)
		}
	}()
	if strings.HasPrefix(r.URL.Path, "/COM") {
		if s != nil {
			s.Flush()
			s.Close()
			s = nil
		}
		com := strings.TrimPrefix(r.URL.Path, "/")
		serialConfig = &serial.Config{Name: com, Baud: 9600, ReadTimeout: time.Second * 1}
		s, err = serial.OpenPort(serialConfig)
		if err != nil {
			s = nil
			return
		}
		err = s.Flush()
		if err != nil {
			return
		}
		re.Message = "connected"
	} else if strings.HasPrefix(r.URL.Path, "/sol") || strings.HasPrefix(r.URL.Path, "/voltage") {
		if s == nil {
			err = fmt.Errorf("no com port!")
			return
		}
		err = write(s, strings.TrimPrefix(r.URL.Path, "/"))
		re.Message = strings.TrimPrefix(r.URL.Path, "/")
	} else if strings.HasPrefix(r.URL.Path, "/read") {
		if s == nil {
			err = fmt.Errorf("no com port!")
			return
		}
		err = write(s, "read")
		if err != nil {
			return
		}
		re.Message, err = read(s)
		re.Message = strings.TrimSpace(re.Message)
	} else if strings.HasPrefix(r.URL.Path, "/stop") {
		if s != nil {
			write(s, "voltage0")
			write(s, "sol1off")
			write(s, "sol2off")
			s.Flush()
			s.Close()
		}
		s = nil
		re.Message = "stopped"
	} else if strings.HasPrefix(r.URL.Path, "/coms") {
		re.Ports, err = stserial.GetPortsList()
		re.Message = "found ports"
	} else {
		jsonData = false
		p := r.URL.Path
		if p == "/" {
			p = "/static/index.html"
		}
		p = strings.TrimPrefix(p, "/")
		if strings.Contains(p, ".js") {
			w.Header().Set("Content-Type", "text/javascript")
		} else if strings.Contains(p, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.Contains(p, ".png") {
			w.Header().Set("Content-Type", "image/png")
		} else if strings.Contains(p, ".json") {
			w.Header().Set("Content-Type", "application/json")
		} else {
			w.Header().Set("Content-Type", "text/html")
		}
		var b []byte
		b, err = Asset(p)
		// b, err = ioutil.ReadFile(p)
		if err != nil {
			log.Error(err)
		} else {
			_, err = w.Write(b)
		}
	}
	return
}

func write(s *serial.Port, data string) (err error) {
	mu.Lock()
	defer mu.Unlock()
	log.Tracef("writing '%s'", data)
	_, err = s.Write([]byte(data + "\n"))
	if err != nil {
		return
	}
	err = s.Flush()
	return
}

func read(s *serial.Port) (reply string, err error) {
	// log.Debug("locking")
	// mu.Lock()
	// defer func() {
	// 	mu.Unlock()
	// 	log.Debug("unlocking")
	// }()
	for {
		log.Trace("waiting for byte")
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
	log.Tracef("read '%s'", strings.TrimSpace(reply))
	return
}
