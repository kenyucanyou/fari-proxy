package http

import (
	"bytes"
	"strconv"
	"strings"
	"time"
)

var httpRequest = []byte("GET /blog.html HTTP/1.1\r\n" +
	"Accept:image/gif.image/jpeg,*/*\r\n" +
	"Accept-Language:zh-cn\r\n" +
	"Connection:Keep-Alive\r\n" +
	"Host:localhost\r\n" +
	"User-Agent:Mozila/4.0(compatible;MSIE5.01;Window NT5.0)\r\n" +
	"Accept-Encoding:gzip,deflate\r\n" +
	"Content-Length:")

func GMT() string {
	utcTime := time.Now().UTC().Format(time.RFC1123)
	return strings.Replace(utcTime, "UTC", "GMT", -1)
}


var httpResponse = []byte("HTTP/1.1 200 OK\r\n" +
	"Content-Type: text/html\r\n" +
	"Date: " +  GMT() + "\r\n" +
	"Server: Microsoft-IIS/6.0\r\n" +
	"Content-Type: text/html\r\n" +
	"Content-Length:")

var RequestLength = len(httpRequest)
var ResponseLength = len(httpResponse)

var ctrf = []byte("\r\n\r\n")
var ctrfLength = len(ctrf)


func NewHttpRequest(ciphertext []byte) []byte {
	httpMsg := append(httpRequest, []byte(strconv.Itoa(len(ciphertext)))...)
	httpMsg = append(httpMsg, ctrf...)
	httpMsg = append(httpMsg, ciphertext...)
	return httpMsg
}


func NewHttpResponse(ciphertext []byte) []byte {
	httpMsg := append(httpResponse, []byte(strconv.Itoa(len(ciphertext)))...)
	httpMsg = append(httpMsg, ctrf...)
	httpMsg = append(httpMsg, ciphertext...)
	return httpMsg
}

func ParseHttpRequest(msg []byte) []byte {
	header := bytes.Split(msg, []byte("\r\n"))
	if len(header) > 7 {
		contentLength := bytes.Split(header[7], []byte(":"))
		if len(contentLength) == 2 {
			lengthName := bytes.Split(header[7], []byte(":"))[0]
			length := bytes.Split(header[7], []byte(":"))[1]
			if string(lengthName) == "Content-Length" {
				contentLength, _ := strconv.Atoi(string(length))
				return msg[RequestLength+len(length)+ctrfLength : RequestLength+len(length)+ctrfLength+contentLength]
			}
		}
	}
	return nil
}

func ParseHttpResponse(msg []byte) []byte {
	header := bytes.Split(msg, []byte("\r\n"))
	if len(header) > 5 {
		contentLength := bytes.Split(header[5], []byte(":"))
		if len(contentLength) == 2 {
			lengthName := bytes.Split(header[5], []byte(":"))[0]
			length := bytes.Split(header[5], []byte(":"))[1]
			if string(lengthName) == "Content-Length" {
				contentLength, _ := strconv.Atoi(string(length))
				return msg[ResponseLength+len(length)+ctrfLength : ResponseLength+len(length)+ctrfLength+contentLength]
			}
		}
	}
	return nil
}