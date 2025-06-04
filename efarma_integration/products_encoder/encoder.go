package products_encoder

import (
	"efarma_integration/models"
	"encoding/binary"
	"fmt"
	"io"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

func (e *Encoder) Encode(storeId int, products []models.Product) error {
	if err := e.writeUint32(uint32(storeId)); err != nil {
		return err
	}
	if err := e.writeUint32(uint32(len(products))); err != nil {
		return err
	}

	for _, product := range products {
		if err := e.writeUint32(uint32(product.CodeSTU)); err != nil {
			return err
		}
		if err := e.writeString(product.Name); err != nil {
			return err
		}
		if err := e.writeUint64(product.GTIN); err != nil {
			return err
		}
		if err := e.writeString(product.Description); err != nil {
			return err
		}
		if err := e.writeUint32(uint32(product.Count)); err != nil {
			return err
		}
		if err := e.writeUint32(uint32(product.Price)); err != nil {
			return err
		}
		if err := e.writeString(product.Country); err != nil {
			return err
		}
		if err := e.writeString(product.Producer); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) writeAndCheck(p []byte) error {
	n, err := e.w.Write(p)
	if err != nil {
		return err
	}
	if n != len(p) {
		return fmt.Errorf("expected to write %d bytes, wrote %d", len(p), n)
	}
	return nil
}

func (e *Encoder) writeUint64(n uint64) error {
	p := make([]byte, 8)
	binary.BigEndian.PutUint64(p, n)
	return e.writeAndCheck(p)
}

func (e *Encoder) writeUint32(n uint32) error {
	p := make([]byte, 4)
	binary.BigEndian.PutUint32(p, n)
	return e.writeAndCheck(p)
}

func (e *Encoder) writeUint16(n uint16) error {
	p := make([]byte, 2)
	binary.BigEndian.PutUint16(p, n)
	return e.writeAndCheck(p)
}

func (e *Encoder) writeString(s string) error {
	lenStr := len(s)
	if err := e.writeUint16(uint16(lenStr)); err != nil {
		return err
	}
	if lenStr == 0 {
		return nil
	}
	if err := e.writeAndCheck([]byte(s)); err != nil {
		return err
	}
	return nil
}
