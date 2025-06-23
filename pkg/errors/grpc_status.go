package errors

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	grpccode "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	UseGobMarshaler  = 1
	UseJsonMarshaler = 2
)

var marshalerCtn = UseJsonMarshaler

func SetErrorMarshalType(mte int) {
	switch mte {
	case UseGobMarshaler:
		marshalerCtn = UseGobMarshaler
	case UseJsonMarshaler:
		marshalerCtn = UseJsonMarshaler
	default:
		marshalerCtn = UseJsonMarshaler
	}
}

type codePayload struct {
	Code     int    `json:"code"`
	HttpCode int    `json:"http_code"`
	GrpcCode uint32 `json:"grpc_code"`
	Message  string `json:"message"`
}

func (e *Kerror) marshalJSON() ([]byte, error) {
	data := codePayload{
		Code:     e.code,
		HttpCode: e.httpCode,
		GrpcCode: e.grpcCode,
		Message:  fmt.Sprintf("%+v", e),
	}
	return json.Marshal(data)
}

func (e *Kerror) marshalGOB() ([]byte, error) {
	var buf bytes.Buffer
	msr := gob.NewEncoder(&buf)
	data := codePayload{
		Code:     e.code,
		HttpCode: e.httpCode,
		GrpcCode: e.grpcCode,
		Message:  fmt.Sprintf("%+v", e),
	}
	if err := msr.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *Kerror) unmarshalJSON(data []byte) error {
	var temp codePayload
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	e.code = temp.Code
	e.httpCode = temp.HttpCode
	e.grpcCode = temp.GrpcCode
	e.msg = temp.Message
	e.stackTrace = nil
	return nil
}

func (e *Kerror) unmarshalGOB(data []byte) error {
	var temp codePayload
	d := gob.NewDecoder(bytes.NewBuffer(data))
	if err := d.Decode(&temp); err != nil {
		return err
	}
	e.code = temp.Code
	e.httpCode = temp.HttpCode
	e.grpcCode = temp.GrpcCode
	e.msg = temp.Message
	e.stackTrace = nil
	return nil
}

func (e *Kerror) grpcStatus() *status.Status {
	var msg []byte
	var err error
	switch marshalerCtn {
	case UseGobMarshaler:
		msg, err = e.marshalGOB()
	case UseJsonMarshaler:
		msg, err = e.marshalJSON()
	}
	if err != nil {
		return status.New(grpccode.Code(e.grpcCode), fmt.Sprintf("failed to marshal error: %v", err))
	}
	return status.New(grpccode.Code(e.grpcCode), string(msg))
}

func (e *Kerror) GRPCStatus() *status.Status {
	return e.grpcStatus()
}

func UnmarshalKerror(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	e := &Kerror{}
	switch marshalerCtn {
	case UseGobMarshaler:
		_ = e.unmarshalGOB(data)
	case UseJsonMarshaler:
		_ = e.unmarshalJSON(data)
	}
	return e
}

func MarshalKerror(err error) string {
	if ke, ok := err.(*Kerror); ok {
		data, _ := ke.marshalJSON()
		return string(data)
	}
	// fallback 包装
	err = NewWithStack(err.Error())
	data, _ := err.(*Kerror).marshalJSON()
	return string(data)
}

func ExtractKerrorFromGRPC(err error) error {
	if st, ok := status.FromError(err); ok {
		e := &Kerror{}
		switch marshalerCtn {
		case UseGobMarshaler:
			_ = e.unmarshalGOB([]byte(st.Message()))
		case UseJsonMarshaler:
			_ = e.unmarshalJSON([]byte(st.Message()))
		}
		return e
	}
	return err
}
