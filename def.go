package utils

import (
	"encoding/json"
	"sync"
)

const (
	// PROD 正式环境
	PROD = "prod"
	// DEV 开发环境，一般用于本机测试
	DEV = "dev"
	// TEST 测试环境
	TEST = "test"
	// PRE 预发布环境
	PRE = "pre"
)

type CtxDoneType string

const (
	CtxDone CtxDoneType = "done"
)

// Code 用于定义返回码
type Code int

const (
	CodeOk   Code = 1
	CodeSrv  Code = 2 // 服务错误
	CodePara Code = 3 // 参数错误
)

// CodeMap 定义返回码对应的描述
var CodeMap = map[Code]string{
	CodeSrv:  "服务错误",
	CodePara: "参数错误",
}

// Resp 用于定义返回数据格式(json)
type Resp struct {
	Ret  Code        `json:"ret"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

type TraceProc struct {
	Name string        `json:"name,omitempty"`
	Dur  string        `json:"dur,omitempty"` // time.Duration stringer
	Args []interface{} `json:"args,omitempty"`
}

// Trace 用于定义trace信息
type Trace struct {
	ID      string       `json:"id,omitempty"`
	Mid     int64        `json:"mid,omitempty"`
	SrvSrc  string       `json:"srv_src,omitempty"`
	SrvDst  string       `json:"srv_dst,omitempty"`
	NameSrc string       `json:"name_src,omitempty"`
	NameDst string       `json:"name_dst,omitempty"`
	Procs   []*TraceProc `json:"procs,omitempty"`
	Mu      sync.Mutex   `json:"-"`
}

func (trace *Trace) String() string {
	if trace == nil {
		return ""
	}

	trace.Mu.Lock()
	defer trace.Mu.Unlock()

	bs, _ := json.Marshal(trace)
	return string(bs)
}

func (trace *Trace) Encode() string {
	if trace == nil {
		return ""
	}

	traceNew := &Trace{
		ID:      trace.ID,
		Mid:     trace.Mid,
		SrvSrc:  trace.SrvDst,
		NameSrc: trace.NameDst,
	}
	bs, _ := json.Marshal(traceNew)
	return string(bs)
}

func (trace *Trace) Decode(s string) {
	if trace == nil {
		return
	}

	json.Unmarshal([]byte(s), trace)
}
