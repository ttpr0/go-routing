package util

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"strconv"
)

func NewBufferReader(data []byte) BufferReader {
	reader := bytes.NewReader(data)
	return BufferReader{
		reader: reader,
	}
}

type BufferReader struct {
	reader *bytes.Reader
}

func Read[T any](reader BufferReader) T {
	var value T
	binary.Read(reader.reader, binary.LittleEndian, &value)
	return value
}

func ReadArray[T any](reader BufferReader) Array[T] {
	var size int32
	binary.Read(reader.reader, binary.LittleEndian, &size)
	value := NewArray[T](int(size))
	binary.Read(reader.reader, binary.LittleEndian, &value)
	return value
}

func NewBufferWriter() BufferWriter {
	buffer := bytes.Buffer{}
	return BufferWriter{
		buffer: &buffer,
	}
}

type BufferWriter struct {
	buffer *bytes.Buffer
}

func (self *BufferWriter) Bytes() []byte {
	return self.buffer.Bytes()
}

func Write[T any](writer BufferWriter, value T) {
	binary.Write(writer.buffer, binary.LittleEndian, value)
}
func WriteArray[T any](writer BufferWriter, value Array[T]) {
	binary.Write(writer.buffer, binary.LittleEndian, int32(value.Length()))
	binary.Write(writer.buffer, binary.LittleEndian, value)
}

func WriteToFile[T any](value T, file string) {
	writer := NewBufferWriter()

	Write[T](writer, value)

	shcfile, _ := os.Create(file)
	defer shcfile.Close()
	shcfile.Write(writer.Bytes())
}

func WriteArrayToFile[T any](value Array[T], file string) {
	writer := NewBufferWriter()

	WriteArray[T](writer, value)

	shcfile, _ := os.Create(file)
	defer shcfile.Close()
	shcfile.Write(writer.Bytes())
}

func WriteJSONToFile[T any](value T, file string) {
	data, _ := json.Marshal(value)

	shcfile, _ := os.Create(file)
	defer shcfile.Close()
	shcfile.Write(data)
}

func ReadFromFile[T any](file string) T {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	shortcutdata, _ := os.ReadFile(file)
	reader := NewBufferReader(shortcutdata)

	value := Read[T](reader)
	return value
}

func ReadArrayFromFile[T any](file string) Array[T] {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	shortcutdata, _ := os.ReadFile(file)
	reader := NewBufferReader(shortcutdata)

	value := ReadArray[T](reader)
	return value
}

func ReadJSONFromFile[T any](file string) T {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		panic("file not found: " + file)
	}

	data, _ := os.ReadFile(file)

	var value T
	json.Unmarshal(data, &value)

	return value
}

func ReadCSVFromFile[T any](filename string, delimiter rune) func(yield func(T) bool) {
	return func(yield func(T) bool) {
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.Comma = delimiter
		header, err := reader.Read()
		if err != nil {
			panic(err)
		}
		name_row_mapping := NewDict[string, int](10)
		for i, name := range header {
			name_row_mapping[name] = i
		}

		var val T
		typ := reflect.TypeOf(val)
		num_field := typ.NumField()
		fields := NewList[Triple[int, int, reflect.Kind]](num_field)
		for i := 0; i < num_field; i++ {
			field := typ.Field(i)
			tag := field.Tag.Get("csv")
			if tag == "" {
				continue
			}
			if !name_row_mapping.ContainsKey(tag) {
				continue
			}
			row := name_row_mapping[tag]
			switch field.Type.Kind() {
			case reflect.Bool:
				fields.Add(MakeTriple(i, row, reflect.Bool))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fields.Add(MakeTriple(i, row, reflect.Int))
			case reflect.Float32, reflect.Float64:
				fields.Add(MakeTriple(i, row, reflect.Float64))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fields.Add(MakeTriple(i, row, reflect.Uint))
			case reflect.String:
				fields.Add(MakeTriple(i, row, reflect.String))
			}
		}
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			} else if err == csv.ErrFieldCount {
				continue
			} else if err != nil {
				continue
			}
			t := reflect.New(typ).Elem()
			for _, field := range fields {
				index := field.A
				row := field.B
				typ := field.C
				value := record[row]
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
			value := t.Interface().(T)
			if !yield(value) {
				break
			}
		}
	}
}
