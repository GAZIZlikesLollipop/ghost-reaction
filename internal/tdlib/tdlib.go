package tdlib

//#include "include/td_json_client.h"
//#cgo LDFLAGS: -L${SRCDIR}/../../lib -ltdjson -Wl,-rpath,${SRCDIR}/../../lib
import "C"

func CreateClientId() int {
	return int(C.td_create_client_id())
}

func Send(clientId int, req string) {
	C.td_send(C.int(clientId), C.CString(req))
}

func Receive(timeout float64) []byte {
	result := C.td_receive(C.double(timeout))
	if result == nil {
		return nil
	}
	return []byte(C.GoString(result))
}

func Execute(req string) []byte {
	result := C.td_execute(C.CString(req))
	if result == nil {
		return nil
	}
	return []byte(C.GoString(result))
}
