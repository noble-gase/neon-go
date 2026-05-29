package helper

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/noble-gase/neon/protokit"
	"github.com/noble-gase/neon/validkit"
	"google.golang.org/protobuf/proto"
)

// BindJSON 解析JSON请求体并校验
func BindJSON(r *http.Request, obj any) error {
	if r.Body != nil && r.Body != http.NoBody {
		defer io.Copy(io.Discard, r.Body)
		if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
			return err
		}
	}
	return validkit.ValidateStruct(obj)
}

// BindProto 解析Proto请求体并校验
func BindProto(r *http.Request, msg proto.Message, protovalidate bool) error {
	if err := ParseProto(r, msg); err != nil {
		return err
	}
	if protovalidate {
		return protokit.Validate(msg)
	}
	return validkit.ValidateStruct(msg)
}

// ParseProto 解析Proto请求体
func ParseProto(r *http.Request, msg proto.Message) error {
	if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
		switch ContentType(r.Header) {
		case ContentForm, ContentMultipartForm:
			if r.PostForm == nil {
				r.ParseMultipartForm(1 << 20)
			}
			return protokit.ValuesToMessage(msg, r.PostForm)
		case ContentJSON:
			if r.Body == nil || r.Body == http.NoBody {
				return nil
			}
			defer io.Copy(io.Discard, r.Body)
			return json.NewDecoder(r.Body).Decode(msg)
		default:
			return errors.New("unsupported Content-Type")
		}
	}
	return protokit.ValuesToMessage(msg, r.URL.Query())
}
