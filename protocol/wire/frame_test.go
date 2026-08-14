package wire

import (
	"errors"
	"testing"
)

func TestFrameRoundTrip(t *testing.T) {
	encoded, err := Encode(1001, 7, []byte("body"))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if frame.Header.MsgID != 1001 || frame.Header.Seq != 7 || string(frame.Body) != "body" {
		t.Fatalf("unexpected frame: %+v", frame)
	}
}

func TestDecodeRejectsInvalidFrames(t *testing.T) {
	if _, err := Decode([]byte{1}); !errors.Is(err, ErrFrameTooShort) {
		t.Fatalf("short frame error = %v", err)
	}
	encoded, _ := Encode(1, 1, nil)
	encoded[4] = 2
	if _, err := Decode(encoded); !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("version error = %v", err)
	}
}
