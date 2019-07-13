package net

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// SPLITER split data
const SPLITER = "||||"

// Packer packer Class
type Packer struct {
	Code   int16  // 指令编码
	Length int32  // 数据部分长度
	Msg    []byte // 数据部分长度
}

// Pack pack data
func (p *Packer) Pack(writer io.Writer) error {
	var err error
	err = binary.Write(writer, binary.LittleEndian, []byte(SPLITER))
	err = binary.Write(writer, binary.LittleEndian, &p.Code)
	err = binary.Write(writer, binary.LittleEndian, &p.Length)
	err = binary.Write(writer, binary.LittleEndian, &p.Msg)
	return err
}

// PackToByte get bytes of pack data
func (p *Packer) PackToByte() ([]byte, error) {
	buf := new(bytes.Buffer)
	e := p.Pack(buf)
	if e != nil {
		return nil, e
	}
	return buf.Bytes(), nil
}

// Unpack unpack data
func (p *Packer) Unpack(reader io.Reader) error {
	var err error
	spliter := make([]byte, 4)
	err = binary.Read(reader, binary.LittleEndian, &spliter)
	err = binary.Read(reader, binary.LittleEndian, &p.Code)
	err = binary.Read(reader, binary.LittleEndian, &p.Length)
	p.Msg = make([]byte, p.Length)
	err = binary.Read(reader, binary.LittleEndian, &p.Msg)
	return err
}

func (p *Packer) String() string {
	return fmt.Sprintf("Code:%d Length:%d msg:%s",
		p.Code,
		p.Length,
		p.Msg,
	)
}
