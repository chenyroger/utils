package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
	"sync"
	"math/rand"
	"errors"
)

const (
	METHOD_GET      = "GET"
	METHOD_POST     = "POST"
	DEFAULT_THREADS = 1
	DefaultTimeout  = 5 * time.Second
	ContentTypeJson = "application/json;charset=utf-8"
	ContentTypeXml  = "text/xml; charset=utf-8"
)

var requestAgents = []string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.11; rv:46.0) Gecko/20100101 Firefox/46.0",
	"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:46.0) Gecko/20100101 Firefox/46.0",
	"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
	"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; WOW64; Trident/6.0)",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.87 Safari/537.36 OPR/37.0.2178.31",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.87 Safari/537.36 OPR/37.0.2178.31",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.104 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.104 Safari/537.36",
}

type Request struct {
	Url            string
	Method         string
	Body           string
	ContentType    string
	result         []byte
	Header         string
	RequestTimeout time.Duration
	Error          error
}

func (r *Request) SetResult(result []byte) {
	r.result = result
}

func (r *Request) SetAgent() {
	if r.Header == "" {
		rand.Seed(time.Now().UnixNano())
		r.Header = requestAgents[rand.Intn(len(requestAgents))]
	}
}

func (r *Request) GetResult() []byte {
	return r.result
}

type MultiRequest struct {
	Requests       []*Request
	Threads        int
	ProcessThreads chan int
}

func (m *MultiRequest) AddRequest(request *Request) {
	m.Requests = append(m.Requests, request)
}

func (m *MultiRequest) Reset() {
	m.Requests = []*Request{}
}

func (m *MultiRequest) Run() {
	requestLen := len(m.Requests)
	if requestLen > 0 {
		if m.Threads <= 0 {
			m.Threads = DEFAULT_THREADS
		}
		m.ProcessThreads = make(chan int, m.Threads)
		var wg sync.WaitGroup
		for _, v := range m.Requests {
			m.ProcessThreads <- 1
			wg.Add(1)
			switch v.Method {
			case METHOD_GET:
				go m.httpGet(v, &wg)
				break
			case METHOD_POST:
				go m.httpPost(v, &wg)
				break
			}
		}
		wg.Wait()
	}
}

func (m *MultiRequest) GetResult() []byte {
	resultLen := len(m.Requests)
	if resultLen == 1 {
		return m.Requests[0].GetResult()
	} else {
		_result := ""
		for _, v := range m.Requests {
			_result += string(v.GetResult()) + "\n"
		}
		return []byte(_result)
	}
	return []byte("")
}
func (m *MultiRequest) httpGet(request *Request, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		<-m.ProcessThreads
	}()
	if request.RequestTimeout == 0 {
		request.RequestTimeout = DefaultTimeout
	}
	client := &http.Client{Timeout: request.RequestTimeout}
	req, err := http.NewRequest("GET", request.Url, nil)
	if err != nil {
		request.Error = err
		return
	}
	request.SetAgent();
	req.Header.Set("User-Agent", request.Header)

	response, err := client.Do(req)
	if err != nil {
		request.Error = err
		return
	}
	response.Close = true
	defer response.Body.Close()
	res, err := ioutil.ReadAll(response.Body)
	if err != nil {
		request.Error = err
		return
	}
	request.SetResult(res)
}

func (m *MultiRequest) httpPost(request *Request, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		<-m.ProcessThreads
	}()
	if request.Body == "" {
		request.Error =  errors.New("no body found")
		return
	}
	body := bytes.NewBuffer([]byte(request.Body))
	req, err := http.NewRequest(request.Method, request.Url, body)
	if err != nil {
		request.Error = err
		return
	}
	request.SetAgent();
	req.Header.Set("User-Agent", request.Header)
	req.Close = true
	if request.ContentType != "" {
		req.Header.Set("Content-Type", request.ContentType)
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if request.RequestTimeout == 0 {
		request.RequestTimeout = DefaultTimeout
	}

	client := &http.Client{Timeout: request.RequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		request.Error = err
		return
	}
	defer resp.Body.Close()

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		request.Error = err
		return
	}
	request.SetResult(res)
}
