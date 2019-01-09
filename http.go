package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpTransportMap sync.Map

type GPP struct {
	Uri            string
	Timeout        time.Duration
	ConnectTimeout time.Duration
	Proxy          string
	Headers        map[string]string
	HttpHeader     http.Header
	HttpHeaderRet  *http.Header
	StatusCodeRet  *int
	Params         interface{}
	Reader         io.Reader
	isForm         bool
}

func Get(gpp *GPP) (body []byte, err error) {
	if params := gpp.Params; params != nil {
		v, ok := params.(map[string]string)
		if !ok {
			return nil, errors.New("params invalid")
		}

		if len(v) > 0 {
			u, err := url.Parse(gpp.Uri)
			if err != nil {
				return nil, err
			}
			values := u.Query()
			for key, value := range v {
				values.Set(key, value)
			}
			u.RawQuery = values.Encode()
			gpp.Uri = u.String()
		}
	}

	return sendHttpRequest(http.MethodGet, gpp)
}

func Post(gpp *GPP) (body []byte, err error) {
	if gpp.Reader == nil {
		if params := gpp.Params; params != nil {
			switch v := params.(type) {
			case string:
				gpp.Reader = strings.NewReader(v)
			case []byte:
				gpp.Reader = bytes.NewReader(v)
			default:
				bs, err := json.Marshal(v)
				if err != nil {
					return nil, err
				}
				gpp.Reader = bytes.NewReader(bs)
			}
		}
	}

	return sendHttpRequest(http.MethodPost, gpp)
}

func PostForm(gpp *GPP) (body []byte, err error) {
	gpp.isForm = true

	if gpp.Reader == nil {
		if params := gpp.Params; params != nil {
			v, ok := params.(map[string]string)
			if !ok {
				return nil, errors.New("params invalid")
			}

			values := url.Values{}
			for key, value := range v {
				values.Set(key, value)
			}
			gpp.Reader = strings.NewReader(values.Encode())
		}
	}

	return sendHttpRequest(http.MethodPost, gpp)
}

func sendHttpRequest(method string, gpp *GPP) (body []byte, err error) {
	httpHeader := gpp.HttpHeader
	if httpHeader == nil {
		httpHeader = make(http.Header)
	}
	for key, value := range gpp.Headers {
		httpHeader.Add(key, value)
	}

	if method == http.MethodPost && httpHeader.Get("Content-Type") == "" {
		if gpp.isForm {
			httpHeader.Set("Content-Type", "application/x-www-form-urlencoded")
		} else {
			httpHeader.Set("Content-Type", "application/json")
		}
	}

	req, err := http.NewRequest(method, gpp.Uri, gpp.Reader)
	if err != nil {
		return
	}
	if host := httpHeader.Get("Host"); host != "" {
		req.Host = host
		httpHeader.Del("Host")
	}
	if connection := httpHeader.Get("Connection"); connection != "" {
		if strings.ToLower(connection) == "close" {
			req.Close = true
			httpHeader.Del("Connection")
		}
	}

	req.Header = httpHeader

	client := &http.Client{}
	if timeout := gpp.Timeout; timeout > 0 {
		client.Timeout = timeout
	}

	if connectTimeout, proxy := gpp.ConnectTimeout, gpp.Proxy; connectTimeout > 0 || proxy != "" {
		httpTransport := &http.Transport{ // from http.DefaultTransport
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		if proxy != "" {
			httpTransport.Proxy = func(*http.Request) (*url.URL, error) {
				return url.Parse(proxy)
			}
		}
		if connectTimeout > 0 {
			httpTransport.DialContext = (&net.Dialer{
				Timeout:   connectTimeout,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext
		}

		mkey := fmt.Sprintf("%s,%s", connectTimeout, proxy)
		if httpTransportInf, loaded := httpTransportMap.LoadOrStore(mkey, httpTransport); loaded {
			httpTransport = httpTransportInf.(*http.Transport)
		}

		client.Transport = httpTransport
	}

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if statusCodeRet := gpp.StatusCodeRet; statusCodeRet != nil {
		*statusCodeRet = resp.StatusCode
	} else {
		if g, e := resp.StatusCode, http.StatusOK; g != e {
			err = fmt.Errorf("http resp code: %d, body: %s", g, body)
			return
		}
	}

	if httpHeaderRet := gpp.HttpHeaderRet; httpHeaderRet != nil {
		*httpHeaderRet = resp.Header
	}

	return
}
