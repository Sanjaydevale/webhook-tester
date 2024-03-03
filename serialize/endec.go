package serialize

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/http"
)

func EncodeRequest(req *http.Request) []byte {
	buf := bytes.NewBuffer([]byte{})
	req.ParseForm()

	encoder := gob.NewEncoder(buf)
	encoder.Encode(req.Method)
	encoder.Encode(req.URL)
	encoder.Encode(req.Proto)
	encoder.Encode(req.ProtoMajor)
	encoder.Encode(req.ProtoMinor)
	encoder.Encode(req.Header)
	data, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(data))
	encoder.Encode(data)
	encoder.Encode(req.ContentLength)
	encoder.Encode(req.TransferEncoding)
	encoder.Encode(req.Host)
	encoder.Encode(req.Form)
	encoder.Encode(req.PostForm)
	encoder.Encode(req.Trailer)
	encoder.Encode(req.RemoteAddr)
	encoder.Encode(req.RequestURI)
	return buf.Bytes()
}

func DecodeRequest(buf []byte) *http.Request {
	req := http.Request{}
	decoder := gob.NewDecoder(bytes.NewBuffer(buf))
	decoder.Decode(&req.Method)
	decoder.Decode(&req.URL)
	decoder.Decode(&req.Proto)
	decoder.Decode(&req.ProtoMajor)
	decoder.Decode(&req.ProtoMinor)
	decoder.Decode(&req.Header)
	body := []byte{}
	decoder.Decode(&body)
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	decoder.Decode(&req.ContentLength)
	decoder.Decode(&req.TransferEncoding)
	decoder.Decode(&req.Host)
	decoder.Decode(&req.Form)
	decoder.Decode(&req.PostForm)
	decoder.Decode(&req.Trailer)
	decoder.Decode(&req.RemoteAddr)
	decoder.Decode(&req.RequestURI)
	return &req
}
