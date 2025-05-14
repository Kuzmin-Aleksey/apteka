package file

import (
	"io"
	"log"
	"os"
)

type UnloadFile struct {
	filepath string
}

func NewUnloadFile(filepath string) *UnloadFile {
	return &UnloadFile{
		filepath: filepath,
	}
}

func (u *UnloadFile) Unload(data io.Reader) error {
	f, err := os.OpenFile(u.filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 666)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := io.Copy(f, data)
	if err != nil {
		return err
	}

	log.Printf("Unload: %d bytes", n)

	return nil
}
