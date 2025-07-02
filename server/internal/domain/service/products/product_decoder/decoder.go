package product_decoder

import (
	"encoding/binary"
	"fmt"
	"io"
	"server/internal/domain/entity"
)

func Decode(r io.Reader) ([]entity.Product, int, error) {
	storeId, err := readUint32AsInt(r)
	if err != nil {
		return nil, 0, err
	}

	count, err := readUint32AsInt(r)
	if err != nil {
		return nil, 0, err
	}

	products := make([]entity.Product, count)

	for i := range count {
		var product entity.Product
		var err error

		product.CodeSTU, err = readUint32AsInt(r)
		if err != nil {
			return nil, 0, err
		}
		product.StoreId = storeId
		product.Name, err = readString(r)
		if err != nil {
			return nil, 0, err
		}
		product.GTIN, err = readUint64(r)
		if err != nil {
			return nil, 0, err
		}
		product.Description, err = readString(r)
		if err != nil {
			return nil, 0, err
		}
		product.Count, err = readUint32AsInt(r)
		if err != nil {
			return nil, 0, err
		}
		product.Price, err = readUint32AsInt(r)
		if err != nil {
			return nil, 0, err
		}
		product.Country, err = readString(r)
		if err != nil {
			return nil, 0, err
		}
		product.Producer, err = readString(r)
		if err != nil {
			return nil, 0, err
		}

		products[i] = product
	}

	return products, storeId, nil
}

func readAndCheck(r io.Reader, p []byte) error {
	n, err := r.Read(p)
	if err != nil {
		return err
	}
	if n != len(p) {
		return fmt.Errorf("read %d bytes instead of expected %d bytes; reading: (%s)", n, len(p), string(p))
	}
	return nil
}

func readString(r io.Reader) (string, error) {
	strLen, err := readUint16(r)
	if err != nil {
		return "", err
	}
	if strLen == 0 {
		return "", nil
	}

	strBytes := make([]byte, strLen)

	if err := readAndCheck(r, strBytes); err != nil {
		return "", err
	}

	return string(strBytes), nil
}

func readUint32AsInt(r io.Reader) (int, error) {
	p := make([]byte, 4)
	if err := readAndCheck(r, p); err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(p)), nil
}

func readUint64(r io.Reader) (uint64, error) {
	p := make([]byte, 8)
	if err := readAndCheck(r, p); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(p), nil
}

func readUint16(r io.Reader) (uint16, error) {
	p := make([]byte, 2)
	if err := readAndCheck(r, p); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(p), nil
}
