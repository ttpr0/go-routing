package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	. "github.com/ttpr0/go-routing/util"
)

type none struct{}

func ReadRequestBody[T any](r *http.Request) T {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	var req T
	err = json.Unmarshal(data, &req)
	if err != nil {
		fmt.Println(err.Error())
	}
	return req
}

func WriteResponse[T any](w http.ResponseWriter, resp T, status int) {
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

type Result2[T any] struct {
	result T
	status int
}

func OK[T any](value T) Result2[T] {
	return Result2[T]{
		result: value,
		status: http.StatusOK,
	}
}

func BadRequest[T any](value T) Result2[T] {
	return Result2[T]{
		result: value,
		status: http.StatusBadRequest,
	}
}

func MapPost[F any, T any](app *http.ServeMux, path string, handler func(F) Result2[T]) {
	app.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		body := ReadRequestBody[F](r)
		res := handler(body)
		WriteResponse[T](w, res.result, res.status)
	})
}

func MapGet[F any, T any](app *http.ServeMux, path string, handler func(F) Result2[T]) {
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
		WriteResponse[T](w, res.result, res.status)
	})
}
