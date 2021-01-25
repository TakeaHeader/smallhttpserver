package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

const (
	OK uint = 200
	NOT_FOUND uint = 404
	Moved_permanent  uint = 301
	Moved_Temporarily  uint = 302
	Internal_Server_Error uint = 500
)

var Context = &HttpContext{
	mapping:make(map[string] Handler),
}

type HttpContext struct {
	mapping map[string] Handler
}

type Req struct {
	Method,Version,Path string
	br *bufio.Reader
	Header map[string]string
}

type Rep struct {
	Version,Msg string
	Code uint
	bw *bufio.Writer
}

type DefaultHandler struct {

}

func (handler *DefaultHandler) Service(req *Req,rep *Rep) (e error) {
	return nil
}

func main() {
	Context.registerPathHandler("/",&DefaultHandler{})
	Context.serve()
}

func (context *HttpContext) serve(){
	li ,_ := net.Listen("tcp",":18990")
	for{
		conn ,_ := li.Accept()
		go func() {
			context.handler(conn)
		}()
	}
}

func (context *HttpContext) handler(conn net.Conn)  {
	defer conn.Close()
	req,_ := parseRequest(conn)
	handler := context.mapping[req.Path]
	rep := &Rep{}
	rep.bw = bufio.NewWriter(conn)
	rep.Version = req.Version
	if handler != nil {
		handler.Service(req,rep)
	}
	rep.Code = NOT_FOUND
	rep.printSteam()
}

func parseRequest(conn net.Conn)(*Req,error) {
	req := &Req{}
	req.br = bufio.NewReader(conn)
	req.Method,req.Path,req.Version = parseFirstLine(req)
	req.Header = parseUpcomeHeaders(req)
	return req,nil
}

func parseFirstLine(req *Req) (method,path,version string) {
	by, _, _ := req.br.ReadLine()
	line := string(by)
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:]
}

func parseUpcomeHeaders(req *Req) map[string] string {
	header := make(map[string]string)
	for {
		line, _, _ := req.br.ReadLine()
		if len(line) == 0{
			break
		}
		sep := strings.Index(string(line),":")
		header[string(line[:sep])] = string(line[sep+1:])
	}
	return header
}

func (req *Req) getInputSteam() *bufio.Reader {
	return req.br
}

func (rep *Rep) getOutputSteam() *bufio.Writer {
	return rep.bw
}

func (rep *Rep) printSteam()  {
	rep.getOutputSteam().Write([]byte(fmt.Sprintf("%s %d %s\r\n",rep.Version,rep.Code,"ok")))
	rep.getOutputSteam().Write([]byte(fmt.Sprintf("Content-Type: %s\r\n","text/html")))
	rep.getOutputSteam().Write([]byte(string("\r\n")))
	rep.getOutputSteam().Write([]byte(string("<html><header><meta charset='utf-8'/></header><b>你好</b></html>")))
	rep.getOutputSteam().Flush()
}

type Handler interface {
	Service(*Req,*Rep) (error)
}

func (context *HttpContext) registerPathHandler(path string,handler Handler)(ok bool,err error) {
	context.mapping[path] = handler
	return true,nil
}

