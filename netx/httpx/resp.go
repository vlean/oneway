package httpx

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
)

type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 0

	Header http.Header

	Body   *bytes.Buffer
	reader *textproto.Reader
}

func badStringError(f, v string) error {
	return errors.New(f + " " + v)
}

func ReadResponse(r *bufio.Reader) (*Response, error) {
	tp := textproto.NewReader(r)
	line, err := tp.ReadLine()
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	resp := &Response{reader: tp}
	proto, status, ok := strings.Cut(line, " ")
	if !ok {
		return nil, badStringError("malformed HTTP response", line)
	}
	resp.Proto = proto
	resp.Status = strings.TrimLeft(status, " ")

	statusCode, _, _ := strings.Cut(resp.Status, " ")
	if len(statusCode) != 3 {
		return nil, badStringError("malformed HTTP status code", statusCode)
	}
	resp.StatusCode, err = strconv.Atoi(statusCode)
	if err != nil || resp.StatusCode < 0 {
		return nil, badStringError("malformed HTTP status code", statusCode)
	}
	if resp.ProtoMajor, resp.ProtoMinor, ok = http.ParseHTTPVersion(resp.Proto); !ok {
		return nil, badStringError("malformed HTTP version", resp.Proto)
	}

	// Parse the response headers.
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	resp.Header = http.Header(mimeHeader)

	fixPragmaCacheControl(resp.Header)

	if err != nil {
		return nil, err
	}

	resp.Body = &bytes.Buffer{}
	if resp.Header.Get("Content-Encoding") == "gzip" {
		resp.Header.Del("Content-Encoding")
		rd, err := gzip.NewReader(tp.R)
		if err != nil {
			return nil, err
		}
		tmp := make([]byte, 512)
		for {
			n, err := rd.Read(tmp)
			if err != nil {
				if err == io.EOF {
					resp.Body.Write(tmp[:n])
					break
				}
				return nil, err
			}
			resp.Body.Write(tmp[:n])
		}
	} else {
		io.Copy(resp.Body, tp.R)
	}
	return resp, nil
}

func fixPragmaCacheControl(header http.Header) {
	if hp, ok := header["Pragma"]; ok && len(hp) > 0 && hp[0] == "no-cache" {
		if _, presentcc := header["Cache-Control"]; !presentcc {
			header["Cache-Control"] = []string{"no-cache"}
		}
	}
}
