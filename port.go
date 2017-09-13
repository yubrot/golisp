package golisp

import (
	"bufio"
	"os"
)

type Port interface {
	Reader() *bufio.Reader
	Writer() *bufio.Writer
	Close() error
	Value
}

type portIn struct {
	file   *os.File
	reader *bufio.Reader
}

func (pi *portIn) Reader() *bufio.Reader {
	return pi.reader
}

func (pi *portIn) Writer() *bufio.Writer {
	return nil
}

func (pi *portIn) Close() error {
	if pi.reader == nil {
		return nil
	}
	err := pi.file.Close()
	if err == nil {
		pi.file = nil
		pi.reader = nil
	}
	return err
}

func (_ *portIn) Inspect() string {
	return "<port>"
}

type portOut struct {
	file   *os.File
	writer *bufio.Writer
}

func (po *portOut) Reader() *bufio.Reader {
	return nil
}

func (po *portOut) Writer() *bufio.Writer {
	return po.writer
}

func (po *portOut) Close() error {
	if po.writer == nil {
		return nil
	}
	err := po.writer.Flush()
	if err == nil {
		err = po.file.Close()
		if err == nil {
			po.file = nil
			po.writer = nil
		}
	}
	return err
}

func (_ *portOut) Inspect() string {
	return "<port>"
}

func NewPortIn(file *os.File) Port {
	return &portIn{file, bufio.NewReader(file)}
}

func NewPortOut(file *os.File) Port {
	return &portOut{file, bufio.NewWriter(file)}
}
