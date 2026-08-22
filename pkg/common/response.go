package common

import (
	"encoding/json"
	"net/http"
)

type Response[T any] struct {
	Code string `json:"code"`
	Info string `json:"info"`
	Data T      `json:"data,omitempty"`
}

func Success[T any](data T) *Response[T] {
	return &Response[T]{
		Code: SuccessCode.Code,
		Info: SuccessCode.Info,
		Data: data,
	}
}

func Fail[T any](code string, info string) *Response[T] {
	return &Response[T]{
		Code: code,
		Info: info,
	}
}

func FailWithCode[T any](rc ResponseCode) *Response[T] {
	return &Response[T]{
		Code: rc.Code,
		Info: rc.Info,
	}
}

func WriteJSON(w http.ResponseWriter, status int, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func WriteSuccess[T any](w http.ResponseWriter, data T) {
	WriteJSON(w, http.StatusOK, Success(data))
}

func WriteError[T any](w http.ResponseWriter, status int, rc ResponseCode) {
	WriteJSON(w, status, FailWithCode[T](rc))
}

func WriteErrorf[T any](w http.ResponseWriter, status int, code, info string) {
	WriteJSON(w, status, Fail[T](code, info))
}
