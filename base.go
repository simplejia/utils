package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync/atomic"
)

const (
	// KeyBody request body内容会存在此key对应params参数里
	KeyBody = "_body_"
	// KeyHeader 公共头内容会存在此key对应params参数里
	KeyHeader = "_header_"
	// KeyResp resp body内容会存在此key对应params参数里
	KeyResp = "_resp_"
	// KeyTrace trace内容会存在此key对应params参数里
	KeyTrace = "_trace_"
	// KeyTimeout timeout flag会存在此key对应params参数里
	KeyTimeout = "_timeout_"
	// KeyTimeoutMutex timeout mutex会存在此key对应params参数里
	KeyTimeoutMutex = "_timeout_mutex_"
	// KeyTimeoutDone timeout done chan会存在此key对应params参数里
	KeyTimeoutDone = "_timeout_done_"
	// KeyTimeoutOccur timeout occur flag会存在此key对应params参数里
	KeyTimeoutOccur = "_timeout_occur_"
	// KeyCtxDone request context done value会存在此key对应params参数里
	KeyCtxDone = "_ctx_done_"
)

// IBase 所有Controller必须实现此接口
type IBase interface {
	SetParam(string, interface{})
	GetParam(string) (interface{}, bool)
	ReadBody(*http.Request) []byte
	ReplyRaw(http.ResponseWriter, []byte)
	Reply(http.ResponseWriter, interface{})
	ReplyOk(http.ResponseWriter, interface{})
	ReplyFail(http.ResponseWriter, Code)
	ReplyFailWithMsg(http.ResponseWriter, Code, string)
	ReplyFailWithResult(http.ResponseWriter, *Resp)
}

type Base struct {
	params map[string]interface{}
}

func (base *Base) SetParam(key string, value interface{}) {
	if base.params == nil {
		base.params = make(map[string]interface{})
	}
	base.params[key] = value
}

func (base *Base) GetParam(key string) (value interface{}, ok bool) {
	value, ok = base.params[key]
	return
}

func (base *Base) ReadBody(r *http.Request) (body []byte) {
	value, ok := base.GetParam(KeyBody)
	if ok {
		body = value.([]byte)
		return
	}

	body, _ = ioutil.ReadAll(r.Body)

	base.SetParam(KeyBody, body)

	if ctxDone, ok := r.Context().Value(CtxDone).(chan struct{}); ok {
		base.SetParam(KeyCtxDone, ctxDone)
	}
	return
}

func (base *Base) ReplyRaw(w http.ResponseWriter, data []byte) {
	base.SetParam(KeyResp, data)

	if _, ok := base.GetParam(KeyTimeout); ok {
		muI, _ := base.GetParam(KeyTimeoutMutex)
		mu := muI.(*int32)
		ok := atomic.CompareAndSwapInt32(mu, 0, 1)
		if !ok {
			return
		}

		doneI, _ := base.GetParam(KeyTimeoutDone)
		done := doneI.(chan struct{})
		close(done)
	}

	w.Write(data)

	if ctxDoneI, ok := base.GetParam(KeyCtxDone); ok {
		ctxDone := ctxDoneI.(chan struct{})
		close(ctxDone)
	}

	return
}

func (base *Base) Reply(w http.ResponseWriter, data interface{}) {
	d, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	base.ReplyRaw(w, d)
	return
}

func (base *Base) ReplyOk(w http.ResponseWriter, data interface{}) {
	base.Reply(w, &Resp{
		Ret:  CodeOk,
		Data: data,
	})
	return
}

func (base *Base) ReplyFail(w http.ResponseWriter, code Code) {
	base.Reply(w, &Resp{
		Ret: code,
		Msg: CodeMap[code],
	})
	return
}

func (base *Base) ReplyFailWithMsg(w http.ResponseWriter, code Code, msg string) {
	base.Reply(w, &Resp{
		Ret: code,
		Msg: msg,
	})
	return
}

func (base *Base) ReplyFailWithResult(w http.ResponseWriter, result *Resp) {
	if result != nil {
		base.Reply(w, result)
		return
	}

	base.ReplyFail(w, CodeSrv)
	return
}
