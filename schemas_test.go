package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/docker/distribution/context"
)

func TestRequest_DataString(t *testing.T) {
	emptyReq := Request{}
	if str := emptyReq.DataString(); str != "" {
		t.Errorf("\nRequest.DataString method error\nExpected: %s\nGot: %s\n", "", str)
	}
	req := Request{
		Data: map[string]interface{}{"test": "test"},
	}
	expected := `{"test":"test"}`
	if str := req.DataString(); reflect.DeepEqual(str, expected) {
		t.Errorf("\nRequest.DataString method error\nExpected: %s\nGot: %s\n", expected, str)
	}
}

func TestRequest_UnmarshalJSON_WrongData(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(`{"test": "test"}`)
	err := getRequestError(&buf)
	if err == nil {
		t.Error("Wrong data")
	}
}

func TestRequest_UnmarshalJSON_NoSchemaVersion(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(`{
		"schema_version": "",
		"model_version": "1.0",
		"timestamp": "2016-08-24T12:35:25.391293168Z",
		"data": {"test": "test"}
	}`)
	err := getRequestError(&buf)
	if err == nil {
		t.Error("No error when no schema is provided")
	} else {
		if e, ok := err.(*requestError); ok {
			if e.SchemaVersionError == "" {
				t.Error("Wrong error message")
			}
		} else {
			t.Error("Wrong error")
		}
	}
}

func TestRequest_UnmarshalJSON_WrongSchemaVersion(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(`{
		"schema_version": "0.0",
		"model_version": "1.0",
		"timestamp": "2016-08-24T12:35:25.391293168Z",
		"data": {"test": "test"}
	}`)
	err := getRequestError(&buf)
	if err == nil {
		t.Error("No error when wrong schema is provided")
	} else {
		if e, ok := err.(*requestError); ok {
			if e.SchemaVersionError == "" {
				t.Error("Wrong error message")
			}
		} else {
			t.Error("Wrong error")
		}
	}
}

func TestRequest_UnmarshalJSON_NoModelVersion(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(`{
		"schema_version": "0.1",
		"model_version": "",
		"timestamp": "2016-08-24T12:35:25.391293168Z",
		"data": {"test": "test"}
	}`)
	err := getRequestError(&buf)
	if err == nil {
		t.Error("No error when no model is provided")
	} else {
		if e, ok := err.(*requestError); ok {
			if e.ModelVersionError == "" {
				t.Error("Wrong error message")
			}
		} else {
			t.Error("Wrong error")
		}
	}
}

func TestRequest_UnmarshalJSON_WrongModelVersion(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(`{
		"schema_version": "0.0",
		"model_version": "0.0",
		"timestamp": "2016-08-24T12:35:25.391293168Z",
		"data": {"test": "test"}
	}`)
	err := getRequestError(&buf)
	if err == nil {
		t.Error("No error when wrong model is provided")
	} else {
		if e, ok := err.(*requestError); ok {
			if e.SchemaVersionError == "" {
				t.Error("Wrong error message")
			}
		} else {
			t.Error("Wrong error")
		}
	}
}

func TestRequest_UnmarshalJSON_WrongTimestamp(t *testing.T) {
	var buf bytes.Buffer

	buf.WriteString(`{
		"schema_version": "0.1",
		"model_version": "1.0",
		"timestamp": "2016-08-24 12:35:25.391293168Z",
		"data": {"test": "test"}
	}`)
	err := getRequestError(&buf)
	if err == nil {
		t.Error("No error when wrong timestamp is provided")
	} else {
		if e, ok := err.(*requestError); ok {
			if e.TimestampError == "" {
				t.Error("Wrong error message")
			}
		} else {
			t.Error("Wrong error")
		}
	}
}

func TestRequest_UnmarshalJSON_UndefinedFields(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(`{
		"schema_version": "0.1",
		"model_version": "1.0",
		"timestampx": "2016-08-24T12:35:25.391293168Z",
		"data": {"test": "test"}
	}`)
	err := getRequestError(&buf)
	if err == nil {
		t.Error("No error when undefined fields are provided")
	} else {
		if e, ok := err.(*requestError); ok {
			if len(e.UnknownFields) == 0 {
				t.Error("Wrong error message")
			}
		} else {
			t.Error("Wrong error")
		}
	}
}

func TestRequest_UnmarshalJSON_WrongError(t *testing.T) {
	var buf bytes.Buffer

	buf.WriteString(`{
		"schema_version": "0.1",
		"model_version": "1.0",
		"timestamp": "2016-08-24T12:35:25.391293168Z",
		"data": {"test": "test"}
	}`)
	err := getRequestError(&buf)
	if err != nil {
		t.Error("Error when good data is provided")
	}
}

func getRequestError(r io.Reader) error {
	req := Request{}
	return json.NewDecoder(r).Decode(&req)
}

func TestAppError_Error(t *testing.T) {
	errorMsg := "test"
	ae := AppError{Reason: errorMsg}
	if ae.Error() != ae.Reason {
		t.Error("Error message is not a reason field")
	}
	ae.Reason = ""
	ae.Err = errors.New(errorMsg)
	if ae.Error() != ae.Err.Error() {
		t.Error("Error message is wrong")
	}
	ae.Err = nil
	ae.StatusCode = http.StatusInternalServerError
	if ae.Error() != http.StatusText(http.StatusInternalServerError) {
		t.Error("Wrong error message")
	}
}

func TestAppError_Write(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()
	ae := AppError{StatusCode: http.StatusInternalServerError}
	ae.Write(ctx, w)
	if w.Code != ae.StatusCode {
		t.Error("Wrong status code")
	}
	resp := Response{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateResponseFromRequest(t *testing.T) {
	req := Request{
		SchemaVersion: "1.0",
		ModelVersion:  "0.1",
	}
	resp := CreateResponseFromRequest(&req)
	if resp.ModelVersion != req.ModelVersion {
		t.Error("Wrong model version")
	}
	if resp.SchemaVersion != req.SchemaVersion {
		t.Error("Wrong schema version")
	}
}

func TestResponse_MarshalJSON(t *testing.T) {
	sr := Response{}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&sr)
	if err != nil {
		t.Error(err)
	}
	buf.Reset()

	sr.err = &AppError{}
	sr.ExecutionTime = time.Duration(5)
	sr.Data = []byte("test")
	err = json.NewEncoder(&buf).Encode(&sr)
	if err != nil {
		t.Error(err)
	}
	if sr.ExecutionTime != 0 {
		t.Error("Wrong execution time")
	}
}

func TestResponse_Write_Success(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()
	expectedData := map[string]string{"test": "test"}
	expected := Response{
		SchemaVersion: "1.0",
		ModelVersion:  "0.1",
		Status:        "ok",
		ExecutionTime: 10 * time.Second,
	}
	expected.Data, _ = json.Marshal(&expectedData)
	expected.Write(ctx, w)
	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Error("Wrong content type")
	}
	actual := Response{}
	err := json.NewDecoder(w.Body).Decode(&actual)
	if err != nil {
		t.Errorf("Error decoding response: %s", err)
	}
	actualData := make(map[string]string)
	err = json.Unmarshal(actual.Data, &actualData)
	if err != nil {
		t.Errorf("Error decoding data: %s", err)
	}
	for k, v := range actualData {
		if expectedData[k] != v {
			t.Errorf("Wrong data.\nExpected: %s\n Actual: %s", expectedData[k], v)
		}
	}
}

func TestResponse_Write_ScriptError(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()
	badData := []byte(`test`)
	bad := Response{
		SchemaVersion: "1.0",
		ModelVersion:  "0.1",
		Status:        "ok",
		ExecutionTime: 10 * time.Second,
		Data:          badData,
	}
	bad.Write(ctx, w)
	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Error("Wrong content type")
	}
	errResp := Response{}
	err := json.NewDecoder(w.Body).Decode(&errResp)
	if err != nil {
		t.Errorf("Error decoding response: %s", err)
	}
	if errResp.Reason != "Script does not return valid json string." {
		t.Error("Wrong reason")
	}
	if errResp.Status != "error" {
		t.Error("Wrong status")
	}
}
