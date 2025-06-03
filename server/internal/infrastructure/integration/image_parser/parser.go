package image_parser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/chai2010/webp"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"golang.org/x/net/context"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type ImagesParser struct {
	timeout time.Duration
	g       *geziyor.Geziyor
}

func NewImagesParser(timeout time.Duration) *ImagesParser {
	p := &ImagesParser{
		timeout: timeout,
	}
	p.UpdateGeziyor()
	return p
}

func (p *ImagesParser) UpdateGeziyor() {
	p.g = geziyor.NewGeziyor(&geziyor.Options{
		Timeout:           p.timeout,
		RobotsTxtDisabled: true,
		UserAgent:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	})

}

type attemptsError struct {
	errs []error
}

func (e attemptsError) Error() string {
	s := "parse image errors:\n"
	for i, err := range e.errs {
		s += fmt.Sprintf("\tattempt %d. %s\n", i+1, err)
	}
	return s
}

func (p *ImagesParser) LoadWebpImageByGTIN(ctx context.Context, gtin uint64) (io.ReadCloser, error) {
	p.UpdateGeziyor()

	attempts := []func(ctx context.Context, gtin uint64) (io.ReadCloser, error){
		p.loadWebpImageByGTINv1,
		p.loadWebpImageByGTINv2,
		p.loadWebpImageByGTINv3,
	}

	var attemptsErr attemptsError

	for _, attempt := range attempts {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		r, err := attempt(ctx, gtin)
		if err != nil {
			attemptsErr.errs = append(attemptsErr.errs, err)
			continue
		}
		return r, nil
	}

	return nil, attemptsErr
}

func (p *ImagesParser) loadWebpImageByGTINv3(ctx context.Context, gtin uint64) (io.ReadCloser, error) {
	cookie := &http.Cookie{
		Name:  "SGID_UID",
		Value: strconv.Itoa(rand.Int()),
	}

	req, err := client.NewRequest("GET", fmt.Sprintf("https://аптека-омск.рф/?digiSearch=true&term=%d&params=%%7Csort%%3DDEFAULT", gtin), nil)
	if err != nil {
		return nil, fmt.Errorf("request creating error %v", err)
	}
	req.Request = req.WithContext(ctx)
	req.AddCookie(cookie)
	req.Synchronized = true

	var prodPageUrl string
	var parserErr error

	p.g.Do(req, func(g *geziyor.Geziyor, r *client.Response) {
		if r.StatusCode != 200 {
			parserErr = fmt.Errorf("status %s", r.Status)
			return
		}

		selection := r.HTMLDoc.Find("a.digi-product__image-wrapper")

		var exist bool
		prodPageUrl, exist = selection.Attr("href")
		if !exist {
			parserErr = fmt.Errorf("cannot find attr 'href' in a.digi-product__image-wrapper")
		}
	})

	if parserErr != nil {
		return nil, parserErr
	}

	req, err = client.NewRequest("GET", "https://аптека-омск.рф"+prodPageUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("request creating error %v", err)
	}
	req.AddCookie(cookie)
	req.Synchronized = true

	var url string

	p.g.Do(req, func(g *geziyor.Geziyor, r *client.Response) {
		if r.StatusCode != 200 {
			parserErr = fmt.Errorf("status %s", r.Status)
			return
		}

		selection := r.HTMLDoc.Find("img.image-style-slide")

		var exist bool
		url, exist = selection.Attr("src")
		if !exist {
			parserErr = fmt.Errorf("cannot find attr 'src' in img.image-style-slide")
		}
	})

	r, err := p.LoadWEBPImage(url)
	return r, nil
}

func (p *ImagesParser) loadWebpImageByGTINv2(ctx context.Context, gtin uint64) (io.ReadCloser, error) {
	cookie := &http.Cookie{
		Name:  "SGID_UID",
		Value: strconv.Itoa(rand.Int()),
	}

	req, err := client.NewRequest("GET", fmt.Sprintf("https://скидкагид.рф/s/?search_query=%d", gtin), nil)
	if err != nil {
		return nil, fmt.Errorf("request creating error %v", err)
	}
	req.Request = req.WithContext(ctx)
	req.AddCookie(cookie)
	req.Synchronized = true

	var prodPageUrl string
	var parserErr error

	p.g.Do(req, func(g *geziyor.Geziyor, r *client.Response) {
		if r.StatusCode != 200 {
			parserErr = fmt.Errorf("status %s", r.Status)
			return
		}

		selection := r.HTMLDoc.Find("a.products-grid_pic")

		var exist bool
		prodPageUrl, exist = selection.Attr("href")
		if !exist {
			parserErr = fmt.Errorf("cannot find attr 'href' in a.products-grid_pic")
		}
	})

	if parserErr != nil {
		return nil, fmt.Errorf("parser error: %v", parserErr)
	}

	req, err = client.NewRequest("GET", "https://скидкагид.рф"+prodPageUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("request creating error %v", err)
	}
	req.AddCookie(cookie)
	req.Synchronized = true

	var url string

	p.g.Do(req, func(g *geziyor.Geziyor, r *client.Response) {
		if r.StatusCode != 200 {
			parserErr = fmt.Errorf("status %s", r.Status)
			return
		}

		selection := r.HTMLDoc.Find("a.fbig")

		var exist bool
		url, exist = selection.Attr("href")
		if !exist {
			parserErr = fmt.Errorf("cannot find attr 'href' in a.fbig")
		}
	})

	r, err := p.LoadWEBPImage(url)
	return r, err
}

func (p *ImagesParser) loadWebpImageByGTINv1(ctx context.Context, gtin uint64) (io.ReadCloser, error) {
	urlTemplates := []string{
		"https://apteka-standart.ru/uploads/catalog/%d_1.jpg.webp",
		"https://apteka-standart.ru/uploads/catalog/%d_01.jpg.webp",
		"https://apteka-standart.ru/uploads/catalog/%d_1.JPG.webp",
		"https://apteka-standart.ru/uploads/catalog/%d_01.JPG.webp",
	}

	var r io.ReadCloser
	var err error

	for _, urlTemplate := range urlTemplates {
		var resp *http.Response

		req, err := http.NewRequest("GET", fmt.Sprintf(urlTemplate, gtin), nil)
		if err != nil {
			continue
		}
		req = req.WithContext(ctx)

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			err = errors.New(strconv.Itoa(resp.StatusCode) + " " + http.StatusText(resp.StatusCode))
			continue
		}

		r = resp.Body
		break
	}

	if r == nil {
		return nil, fmt.Errorf("cannot load image by gtin %d: %v", gtin, err)
	}

	return r, nil
}

func (p *ImagesParser) LoadWEBPImage(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get image error: %v", err)
	}

	imageType := resp.Header.Get("Content-Type")

	img, err := ConvertToWebp(resp.Body, imageType)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func ConvertToWebp(r io.ReadCloser, mimeType string) (io.ReadCloser, error) {
	switch mimeType {
	case "image/webp":
		return r, nil
	case "image/png":
		img, err := PNG2WEBP(r)
		if err != nil {
			return nil, fmt.Errorf("converting png to webp error: %v", err)
		}
		return img, nil
	case "image/jpeg", "image/jpg":
		img, err := JPEG2WEBP(r)
		if err != nil {
			return nil, fmt.Errorf("converting jpeg to webp error: %v", err)
		}
		return img, nil
	default:
		return nil, fmt.Errorf("unknown image mime type: %s", mimeType)
	}
}

func PNG2WEBP(r io.Reader) (io.ReadCloser, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}
	data, err := webp.EncodeRGBA(img, 100)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func JPEG2WEBP(r io.Reader) (io.ReadCloser, error) {
	img, err := jpeg.Decode(r)
	if err != nil {
		return nil, err
	}
	data, err := webp.EncodeRGBA(img, 100)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}
