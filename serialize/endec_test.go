package serialize_test

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"
	"whtester/serialize"
)

func TestEncoderAndDecoder(t *testing.T) {
	t.Run("encodes and decodes http request into binary", func(t *testing.T) {
		data := []byte("this is the body of the request")
		reqBody := bytes.NewBuffer(data)
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080", reqBody)
		if err != nil {
			t.Errorf("%v", err)
		}

		buf := serialize.EncodeRequest(req)
		got := serialize.DecodeRequest(buf)

		assertRequest(t, *got, *req)
	})
}

func assertRequest(t testing.TB, got, want http.Request) {
	t.Helper()
	fields := []string{"Method", "Proto", "ProtoMajor", "ProtoMinor", "Header", "URL", "RequestURI",
		"ContentLength", "TransferEncoding", "Host", "PostForm", "Form", "RemoteAddr"}
	assertStruct(t, fields, got, want)

	// compare bodies
	if got.Body == nil {
		got.Body = io.NopCloser(bytes.NewBuffer([]byte{}))
	}
	if want.Body == nil {
		want.Body = io.NopCloser(bytes.NewBuffer([]byte{}))
	}
	gotBody, _ := io.ReadAll(got.Body)
	wantBody, _ := io.ReadAll(want.Body)

	got.Body = io.NopCloser(bytes.NewBuffer(gotBody))
	want.Body = io.NopCloser(bytes.NewBuffer(wantBody))

	if string(gotBody) != string(wantBody) {
		t.Errorf("different Bodies, got %q, want %q", string(gotBody), string(wantBody))
	}

}

func assertStruct(t testing.TB, fields []string, structA, structB interface{}) {
	t.Helper()
	valA := reflect.ValueOf(structA)
	valB := reflect.ValueOf(structB)

	for _, field := range fields {
		if !reflect.DeepEqual(valA.FieldByName(field).Interface(), valB.FieldByName(field).Interface()) {
			t.Errorf("different %s, got %v, want %v", field, valA.FieldByName(field), valB.FieldByName(field))
		}
	}
}
