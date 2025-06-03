package product_decoder

import (
	"encoding/binary"
	"fmt"
	"io"
	"server/internal/domain/entity"
)

func Decode(r io.Reader) ([]entity.Product, error) {
	storeId, err := readUint32AsInt(r)
	if err != nil {
		return nil, err
	}

	count, err := readUint32AsInt(r)
	if err != nil {
		return nil, err
	}

	products := make([]entity.Product, count)

	for i := range count {
		var product entity.Product
		var err error

		product.CodeSTU, err = readUint32AsInt(r)
		if err != nil {
			return nil, err
		}
		product.StoreId = storeId
		product.Name, err = readString(r)
		if err != nil {
			return nil, err
		}
		product.GTIN, err = readUint64(r)
		if err != nil {
			return nil, err
		}
		product.Description, err = readString(r)
		if err != nil {
			return nil, err
		}
		product.Count, err = readUint32AsInt(r)
		if err != nil {
			return nil, err
		}
		product.Price, err = readUint32AsInt(r)
		if err != nil {
			return nil, err
		}
		product.Country, err = readString(r)
		if err != nil {
			return nil, err
		}
		product.Producer, err = readString(r)
		if err != nil {
			return nil, err
		}

		products[i] = product
	}

	return products, nil
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

/*
WITH ost AS (
SELECT
ROW_NUMBER() OVER (PARTITION BY G.CODE ORDER BY convert(DECIMAL(20,2),L.PRICE_SAL)) AS dp,
CAST(G.CODE AS BIGINT ) AS CODE,
CAST(G.[NAME] AS VARCHAR(250)) AS NAME,
L.QUANTITY_REM AS KOL,
convert(DECIMAL(20,2),L.PRICE_SAL) AS SALE_PRICE,

CAST(
(
	SELECT TOP 1
	KIZ.GTIN

	FROM KIZ_2_DOCUMENT_ITEM KDI
	JOIN KIZ ON KIZ.ID_KIZ_GLOBAL=KDI.ID_KIZ_GLOBAL
	WHERE 1=1
	AND KDI.ID_DOCUMENT_ITEM_ADD = L.ID_DOCUMENT_ITEM_ADD
) AS bigint) AS GTIN,


CAST(P.[NAME] AS VARCHAR(80)) AS PRODUCER,
CAST(CR.[NAME] AS VARCHAR(80)) AS COUNTRY,

G.DESCRIPTION

FROM LOT L
inner join GOODS G on L.ID_GOODS = G.ID_GOODS
inner join PRODUCER P on P.ID_PRODUCER = G.ID_PRODUCER
inner join STORE ST on ST.ID_STORE = L.ID_STORE
inner join COUNTRY CR on CR.ID_COUNTRY = P.ID_COUNTRY


WHERE 1=1
AND L.QUANTITY_REM > 0
and ST.ID_CONTRACTOR = DBO.FN_CONST_CONTRACTOR_SELF()
and l.id_store = 270
--and st.MNEMOCODE in ('О-4601')
)


SELECT
(
SELECT TOP 1
	L.NAME
	FROM ost L
WHERE L.CODE = ost.CODE
)


FROM ost
WHERE ost.dp = 1

*/
