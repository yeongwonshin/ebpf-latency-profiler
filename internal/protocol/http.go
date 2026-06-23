package protocol

import (
	"bytes"
	"strconv"
	"strings"
)

type HTTPHint struct {
	IsRequest  bool
	IsResponse bool
	Method     string
	Path       string
	Status     int
}

var methods = [][]byte{
	[]byte("GET "),
	[]byte("POST "),
	[]byte("PUT "),
	[]byte("PATCH "),
	[]byte("DELETE "),
	[]byte("HEAD "),
	[]byte("OPTIONS "),
}

func ParseHTTP1Prefix(prefix []byte) HTTPHint {
	line := firstLine(prefix)
	if len(line) == 0 {
		return HTTPHint{}
	}
	for _, method := range methods {
		if bytes.HasPrefix(line, method) {
			parts := strings.SplitN(string(line), " ", 3)
			if len(parts) >= 2 {
				return HTTPHint{IsRequest: true, Method: parts[0], Path: sanitizePath(parts[1])}
			}
		}
	}
	if bytes.HasPrefix(line, []byte("HTTP/1.")) {
		parts := strings.SplitN(string(line), " ", 3)
		if len(parts) >= 2 {
			status, _ := strconv.Atoi(parts[1])
			return HTTPHint{IsResponse: true, Status: status}
		}
	}
	return HTTPHint{}
}

func firstLine(prefix []byte) []byte {
	if idx := bytes.IndexByte(prefix, '\n'); idx >= 0 {
		return bytes.TrimSpace(prefix[:idx])
	}
	return bytes.TrimSpace(prefix)
}

func sanitizePath(path string) string {
	if path == "" {
		return "/"
	}
	if q := strings.IndexByte(path, '?'); q >= 0 {
		path = path[:q]
	}
	return path
}
