package shadowsocks

import (

	// "io"
	"net"
	"sync"
	"time"
)

const (
	NO_TIMEOUT = iota
	SET_TIMEOUT
)

var pool = &sync.Pool{New: func() interface{} {
	return make([]byte, 4096)
}}

func SetReadTimeout(c net.Conn) {
	if readTimeout != 0 {
		c.SetReadDeadline(time.Now().Add(readTimeout))
	}
}

// PipeThenClose copies data from src to dst, closes dst when done.
func PipeThenClose(src, dst net.Conn, timeoutOpt int, port, dir string) {
	defer dst.Close()
	buf := pool.Get().([]byte)
	defer pool.Put(buf)
	for {
		if timeoutOpt == SET_TIMEOUT {
			SetReadTimeout(src)
		}
		n, err := src.Read(buf)
		// read may return EOF with n > 0
		// should always process n > 0 bytes before handling error
		if n > 0 {
			_, err := dst.Write(buf[0:n])
			if port != "" {
				var ip string
				if dir == "out" {
					ip = src.RemoteAddr().(*net.TCPAddr).IP.String()
				}
				upTraffic(port, n, ip)
			}
			if err != nil {
				Debug.Println("write:", err)
				break
			}
		}
		if err != nil {
			// Always "use of closed network connection", but no easy way to
			// identify this specific error. So just leave the error along for now.
			// More info here: https://code.google.com/p/go/issues/detail?id=4373
			/*
				if err != io.EOF {
					Debug.Println("read:", err)
				}
			*/
			break
		}
	}
}
