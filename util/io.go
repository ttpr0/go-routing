package util

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
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
