package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

type Kerror struct {
	msg        string
	code       int
	httpCode   int
	grpcCode   uint32
	stackTrace *stack
}

func (e *Kerror) HTTPCode() int {
	return e.httpCode
}

func (e *Kerror) GrpcCode() uint32 {
	return e.grpcCode
}

func (e *Kerror) Code() int {
	return e.code
}

// 实现 error 接口
func (e *Kerror) Error() string {
	return e.msg
}

// 实现 fmt.Formatter 接口
func (e *Kerror) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v', 's':
		if s.Flag('+') {
			fmt.Fprint(s, e.msg)
			if e.stackTrace != nil {
				fmt.Fprint(s, "\nStack Trace:\n")
				e.stackTrace.Format(s, verb)
			}
		} else {
			fmt.Fprint(s, e.msg)
		}
	case 'q':
		fmt.Fprintf(s, "%q", e.msg)
	}
}

func New(msg string) error {
	return &Kerror{
		msg:      msg,
		httpCode: 500,
		grpcCode: uint32(codes.Unknown),
		code:     0,
	}
}

func NewWithStack(msg string) error {
	return &Kerror{
		msg:        msg,
		httpCode:   500,
		grpcCode:   uint32(codes.Unknown),
		code:       0,
		stackTrace: callers(),
	}
}

func Errorf(format string, args ...interface{}) error {
	return New(fmt.Sprintf(format, args...))
}

func ErrorfWithStack(format string, args ...interface{}) error {
	return NewWithStack(fmt.Sprintf(format, args...))
}

func NewCodeKerror(msg string, code, httpcode int, grpccode uint32) error {
	return &Kerror{
		msg:      msg,
		code:     code,
		httpCode: httpcode,
		grpcCode: grpccode,
	}
}

func NewCodeKerrorWithStack(msg string, code, httpcode int, grpccode uint32) error {
	return &Kerror{
		msg:        msg,
		code:       code,
		httpCode:   httpcode,
		grpcCode:   grpccode,
		stackTrace: callers(),
	}
}
