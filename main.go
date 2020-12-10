package main

import (
	"bytes"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/schollz/logger"
	"github.com/tarm/serial"
)

var mu sync.Mutex

func main() {
	logger.SetLevel("debug")
	err := run()
	if err != nil {
		logger.Error(err)
	}
}

func run() (err error) {
	gin.SetMode(gin.ReleaseMode)
	c := &serial.Config{Name: "COM5", Baud: 9600, ReadTimeout: time.Second * 1}
	s, err := serial.OpenPort(c)
	if err != nil {
		err = errors.Wrap(err, "no com port")
		return
	}
	s.Flush()
	defer s.Close()

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

	r := gin.Default()
	r.StaticFile("/", "./index.html")
	r.GET("/api", func(c *gin.Context) {
		msg := c.DefaultQuery("msg", "")
		if msg == "" {
			c.JSON(200, gin.H{
				"success": false,
				"message": "no message",
			})
			return
		}
		err = write(s, msg)
		if err != nil {
			logger.Error(err)
			c.JSON(200, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		if msg == "read" {
			reply, err := read(s)
			if err != nil {
				logger.Error(err)
				c.JSON(200, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
			c.JSON(200, gin.H{
				"success": true,
				"message": strings.TrimSpace(reply),
			})
		} else {
			c.JSON(200, gin.H{
				"success": true,
			})
		}
	})
	logger.Infof("running on port 8080")

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	return
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
