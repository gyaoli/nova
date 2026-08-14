package wire

import (
	"encoding/binary"
	"errors"
)

const (
	HeaderSize     = 16
	CurrentVersion = 1
	MaxFrameSize   = 64 << 10
)

var (
	ErrFrameTooShort      = errors.New("frame too short")
	ErrFrameTooLarge      = errors.New("frame too large")
	ErrFrameLength        = errors.New("frame length mismatch")
	ErrUnsupportedVersion = errors.New("unsupported protocol version")
)

type Header struct {
	Version uint8
	Flags   uint8
	MsgID   uint32
	Seq     uint32
	BodyLen uint32
}

type Frame struct {
	Header Header
	Body   []byte
}

func Decode(data []byte) (Frame, error) {
	if len(data) < HeaderSize {
		return Frame{}, ErrFrameTooShort
	}
	total := int(binary.BigEndian.Uint32(data[:4]))
	if total != len(data) || total < HeaderSize {
		return Frame{}, ErrFrameLength
	}
	if total > MaxFrameSize {
		return Frame{}, ErrFrameTooLarge
	}
	header := Header{
		Version: data[4],
		Flags:   data[5],
		MsgID:   binary.BigEndian.Uint32(data[8:12]),
		Seq:     binary.BigEndian.Uint32(data[12:16]),
		BodyLen: uint32(total - HeaderSize),
	}
	if data[6] != 0 || data[7] != 0 {
		return Frame{}, ErrFrameLength
	}
	if header.Version != CurrentVersion {
		return Frame{}, ErrUnsupportedVersion
	}
	body := append([]byte(nil), data[HeaderSize:]...)
	return Frame{Header: header, Body: body}, nil
}

func Encode(msgID, seq uint32, body []byte) ([]byte, error) {
	total := HeaderSize + len(body)
	if total > MaxFrameSize {
		return nil, ErrFrameTooLarge
	}
	data := make([]byte, total)
	binary.BigEndian.PutUint32(data[:4], uint32(total))
	data[4] = CurrentVersion
	binary.BigEndian.PutUint32(data[8:12], msgID)
	binary.BigEndian.PutUint32(data[12:16], seq)
	copy(data[HeaderSize:], body)
	return data, nil
}
