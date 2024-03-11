package main

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strconv"

	. "github.com/ttpr0/go-routing/util"
	"golang.org/x/exp/slog"
)

type none struct{}

func ReadRequestBody[T any](r *http.Request) (T, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
		var t T
		return t, err
	}
	var req T
	err = json.Unmarshal(data, &req)
	if err != nil {
		slog.Error(err.Error())
		var t T
		return t, err
	}
	return req, nil
}

func WriteResponse[T any](w http.ResponseWriter, resp T, status int) {
	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

type Result struct {
	result any
	status int
}

func OK[T any](value T) Result {
	return Result{
		result: value,
		status: http.StatusOK,
	}
}

func BadRequest[T any](value T) Result {
	return Result{
		result: value,
		status: http.StatusBadRequest,
	}
}

func MapPost[F any](app *http.ServeMux, path string, handler func(F) Result) {
	app.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		slog.Info("POST " + path)
		body, err := ReadRequestBody[F](r)
		if err != nil {
			slog.Error("failed POST " + err.Error())
			WriteResponse(w, NewErrorResponse(path, err.Error()), http.StatusInternalServerError)
		}
		res := handler(body)
		if res.status != http.StatusOK {
			slog.Error("failed POST " + path)
			WriteResponse(w, NewErrorResponse(path, res.result), res.status)
		} else {
			slog.Info("successfully finished POST")
			WriteResponse(w, res.result, res.status)
		}
	})
}

func MapGet[F any](app *http.ServeMux, path string, handler func(F) Result) {
	var val F
	typ := reflect.TypeOf(val)
	num_field := typ.NumField()
	fields := NewList[Triple[int, string, reflect.Kind]](num_field)
	for i := 0; i < num_field; i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" {
			continue
		}
		switch field.Type.Kind() {
		case reflect.Bool:
			fields.Add(MakeTriple(i, tag, reflect.Bool))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fields.Add(MakeTriple(i, tag, reflect.Int))
		case reflect.Float32, reflect.Float64:
			fields.Add(MakeTriple(i, tag, reflect.Float64))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fields.Add(MakeTriple(i, tag, reflect.Uint))
		case reflect.String:
			fields.Add(MakeTriple(i, tag, reflect.String))
		}
	}
	app.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		slog.Info("GET " + path)
		query := r.URL.Query()
		t := reflect.New(typ).Elem()
		for _, field := range fields {
			index := field.A
			name := field.B
			typ := field.C
			value := query.Get(name)
			if value == "" {
				continue
			}
			f := t.Field(index)
			switch typ {
			case reflect.Bool:
				num, _ := strconv.ParseBool(value)
				f.SetBool(num)
			case reflect.Int:
				num, _ := strconv.ParseInt(value, 10, 64)
				f.SetInt(num)
			case reflect.Uint:
				num, _ := strconv.ParseUint(value, 10, 64)
				f.SetUint(num)
			case reflect.Float64:
				num, _ := strconv.ParseFloat(value, 64)
				f.SetFloat(num)
			case reflect.String:
				f.SetString(value)
			}
		}
		value := t.Interface().(F)
		res := handler(value)
		if res.status != http.StatusOK {
			slog.Error("failed GET " + path)
			WriteResponse(w, NewErrorResponse(path, res.result), res.status)
		} else {
			slog.Info("successfully finished GET")
			WriteResponse(w, res.result, res.status)
		}
	})
}
