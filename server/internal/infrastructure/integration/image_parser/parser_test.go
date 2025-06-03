package image_parser

import (
	"fmt"
	"golang.org/x/net/context"
	"io"
	"os"
	"testing"
	"time"
)

const GTIN = 4602193015753

var GTINs = []uint64{
	4602193015753,
	5414789001403,
	4630121500309,
}

func TestLoadWEBPImage(t *testing.T) {
	p := NewImagesParser(time.Second * 10)

	r, err := p.LoadWEBPImage(fmt.Sprintf("https://apteka-standart.ru/uploads/catalog/%d_01.jpg.webp", GTIN))

	if err != nil {
		t.Error(err)
	}

	f, err := os.Create("img.webp")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		t.Error(err)
	}
}

func TestLoadWebpImageByGTIN(t *testing.T) {
	p := NewImagesParser(time.Second * 10)

	for _, gtin := range GTINs {
		func() {
			r, err := p.LoadWebpImageByGTIN(context.Background(), gtin)
			if err != nil {
				t.Error(err)
				return
			}
			defer r.Close()

			f, err := os.Create(fmt.Sprintf("%d.webp", gtin))
			if err != nil {
				t.Error(err)
				return
			}
			defer f.Close()

			if _, err := io.Copy(f, r); err != nil {
				t.Error(err)
				return
			}
		}()
	}
}

//https://apteka-standart.ru/uploads/catalog/5414789001403_2.jpg.webp --ok
//https://apteka-standart.ru/uploads/catalog/5414789001403_02.jpg.webp
