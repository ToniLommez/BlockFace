package nether

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
)

const (
	IMAGE_WIDTH  = 910
	IMAGE_HEIGHT = 512
	IMAGE_SIZE   = IMAGE_WIDTH * IMAGE_HEIGHT * 4 // RGBA: 4 bytes por pixel
)

type Image struct {
	data [IMAGE_SIZE]byte
}

func resizeAndCrop(img image.Image) *image.RGBA {
	cropped := cropToAspectRatio(img, IMAGE_WIDTH, IMAGE_HEIGHT)
	return resizeImage(cropped, IMAGE_WIDTH, IMAGE_HEIGHT)
}

func cropToAspectRatio(img image.Image, targetWidth, targetHeight int) image.Image {
	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	targetRatio := float64(targetWidth) / float64(targetHeight)
	srcRatio := float64(srcWidth) / float64(srcHeight)

	var cropX, cropY, cropWidth, cropHeight int

	if srcRatio > targetRatio {
		cropWidth = int(float64(srcHeight) * targetRatio)
		cropHeight = srcHeight
		cropX = (srcWidth - cropWidth) / 2
		cropY = 0
	} else {
		cropWidth = srcWidth
		cropHeight = int(float64(srcWidth) / targetRatio)
		cropX = 0
		cropY = (srcHeight - cropHeight) / 2
	}

	return img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(cropX, cropY, cropX+cropWidth, cropY+cropHeight))
}

func resizeImage(img image.Image, width, height int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	srcBounds := img.Bounds()
	xRatio := float64(srcBounds.Dx()) / float64(width)
	yRatio := float64(srcBounds.Dy()) / float64(height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := int(float64(x) * xRatio)
			srcY := int(float64(y) * yRatio)
			srcColor := img.At(srcX, srcY)
			dst.Set(x, y, srcColor)
		}
	}

	return dst
}

func newImage(encodedImg string) (*Image, error) {
	imgBytes, err := base64.StdEncoding.DecodeString(encodedImg)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar string Base64: %v", err)
	}

	reader := bytes.NewReader(imgBytes)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("os bytes não representam uma imagem válida: %v", err)
	}

	finalImg := resizeAndCrop(img)

	finalBytes, err := imageToRawBytes(finalImg)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter para RAW: %v", err)
	}

	var imgArray [IMAGE_SIZE]byte
	copy(imgArray[:], finalBytes)

	return &Image{data: imgArray}, nil
}

func imageToRawBytes(img image.Image) ([]byte, error) {
	// Garantir que a imagem seja RGBA
	rgbaImg, ok := img.(*image.RGBA)
	if !ok {
		bounds := img.Bounds()
		rgbaImg = image.NewRGBA(bounds)
		draw.Draw(rgbaImg, bounds, img, bounds.Min, draw.Src)
	}

	if len(rgbaImg.Pix) > IMAGE_SIZE {
		return nil, fmt.Errorf("imagem RAW excede o tamanho máximo permitido")
	}

	return rgbaImg.Pix, nil
}

func (img *Image) Serialize() ([]byte, error) {
	buffer := new(bytes.Buffer)

	if err := binary.Write(buffer, binary.LittleEndian, img.data); err != nil {
		return nil, fmt.Errorf("erro ao serializar imagem: %v", err)
	}

	return buffer.Bytes(), nil
}

func (img *Image) Deserialize(data []byte) error {
	if len(data) != IMAGE_SIZE {
		return fmt.Errorf("tamanho dos dados inválido: esperado %d bytes, recebido %d bytes", IMAGE_SIZE, len(data))
	}

	copy(img.data[:], data)

	return nil
}

func generateRandomImage() (*Image, error) {
	// Criar uma imagem RGBA com dimensões fixas
	img := image.NewRGBA(image.Rect(0, 0, IMAGE_WIDTH, IMAGE_HEIGHT))

	// Preencher a imagem com valores aleatórios
	for y := 0; y < IMAGE_HEIGHT; y++ {
		for x := 0; x < IMAGE_WIDTH; x++ {
			r := uint8(rand.Intn(256))
			g := uint8(rand.Intn(256))
			b := uint8(rand.Intn(256))
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// Converter a imagem para RAW
	rawBytes, err := imageToRawBytes(img)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter para RAW: %v", err)
	}

	// Garantir que os dados cabem em IMAGE_SIZE
	if len(rawBytes) > IMAGE_SIZE {
		return nil, fmt.Errorf("a imagem RAW excede o tamanho máximo permitido")
	}

	// Criar um array fixo e copiar os dados
	var imgArray [IMAGE_SIZE]byte
	copy(imgArray[:], rawBytes)

	return &Image{data: imgArray}, nil
}
