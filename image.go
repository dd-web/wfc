package wfc

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"time"
)

var (
	colorRNG *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	c_RED     = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	c_GREEN   = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	c_BLUE    = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	c_YELLOW  = color.RGBA{R: 255, G: 255, B: 0, A: 255}
	c_CYAN    = color.RGBA{R: 0, G: 255, B: 255, A: 255}
	c_MAGENTA = color.RGBA{R: 255, G: 0, B: 255, A: 255}
	c_ORANGE  = color.RGBA{R: 255, G: 165, B: 0, A: 255}
	c_PURPLE  = color.RGBA{R: 128, G: 0, B: 128, A: 255}
	c_BROWN   = color.RGBA{R: 165, G: 42, B: 42, A: 255}

	c_MAP = [9]color.RGBA{
		c_RED,
		c_GREEN,
		c_BLUE,
		c_YELLOW,
		c_CYAN,
		c_MAGENTA,
		c_ORANGE,
		c_PURPLE,
		c_BROWN,
	}
)

// Load image data from path
// note: you'll need to load whichever encoding type you want to use by importing preifxed by "_"
// returns image data or any error that occurred while attempting to load it
func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Save image data to local disc at path.
// returns any error if one occurred.
func SaveImage(file string, img *image.RGBA) error {
	_, err := os.Stat(file)
	if err == nil {
		if err := os.Remove(file); err != nil {
			return err
		}
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = png.Encode(f, img); err != nil {
		return err
	}

	return nil
}

// Set rectangular region of image to a specified color
// uses rect for size and area information
func SetRegionColor(img draw.Image, col color.Color, rect image.Rectangle) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			img.Set(x, y, col)
		}
	}
}

// Convert image.Image to image.RGBA
func ImageToRGBA(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba
}

func CopyImageRegionData(src *image.RGBA, region image.Rectangle) *image.RGBA {
	regionRGBA := image.NewRGBA(region.Sub(region.Min))
	draw.Draw(regionRGBA, regionRGBA.Bounds(), src, region.Min, draw.Src)
	return regionRGBA
}

// Set a source image region's data
func SetImageRegion(src *image.RGBA, region image.Rectangle, regionData *image.RGBA) {
	draw.Draw(src, region, regionData, region.Min, draw.Src)
}

// Get random RGBA color value with a fixed alpha of 1 (or 255)
func GetRandomColor() color.RGBA {
	return color.RGBA{
		R: uint8(colorRNG.Intn(256)),
		G: uint8(colorRNG.Intn(256)),
		B: uint8(colorRNG.Intn(256)),
		A: uint8(255),
	}
}

// Get a random RGBA color value with a specified alpha value
// alpha is in uint8 255 = full opacity, 0 = completely transparent
func GetRandomColorWithAlpha(alpha uint8) color.RGBA {
	return color.RGBA{
		R: uint8(colorRNG.Intn(256)),
		G: uint8(colorRNG.Intn(256)),
		B: uint8(colorRNG.Intn(256)),
		A: uint8(alpha),
	}
}

// Get the color value at a given pixel.
// if point is outside of supplied image bounds, a default color is returned instead.
func GetPixelColor(img *image.RGBA, pt image.Point) color.RGBA {
	if !pt.In(img.Bounds()) {
		return c_RED
	}
	ix := img.PixOffset(pt.X, pt.Y)
	return color.RGBA{
		R: img.Pix[ix],
		G: img.Pix[ix+1],
		B: img.Pix[ix+2],
		A: img.Pix[ix+3],
	}
}

// Gets a color that cooresponds to the index of the modulus of the seed
func GetSeededColor(seed int) color.RGBA {
	return c_MAP[seed%len(c_MAP)]
}

// Defines one fouth of a region's area.
//
//	0 = top left
//	2 = top right
//	4 = bottom left
//	8 = bottom right
//
// The area can be retreived using one of the constant's GetImageRegionSize() methods.
type RectQuadRegion uint

const (
	RQR_TL RectQuadRegion = iota
	RQR_TR RectQuadRegion = 1 << iota
	RQR_BL RectQuadRegion = 1 << iota
	RQR_BR RectQuadRegion = 1 << iota
)

// Returns a quad of the provided image as an image.Rectangle.
// Example:
//
//	myImage := image.NewRGBA(image.Rect(0, 0, 100, 100))
//	TL := RQR_TL.GetImageRegionSize(myImage)
//	SetRegionColor(myImage, red, TL) // top left corner of the image is now red.
//
// Note that this only returns an area and has no knowledge of the source data.
func (r RectQuadRegion) GetImageRegionSize(src *image.RGBA) image.Rectangle {
	bx, by := src.Bounds().Dx(), src.Bounds().Dy()
	switch r {
	case RQR_TL:
		return image.Rectangle{Min: image.Pt(0, 0), Max: image.Pt(bx/2, by/2)}
	case RQR_TR:
		return image.Rectangle{Min: image.Pt(bx/2, 0), Max: image.Pt(bx, bx/2)}
	case RQR_BL:
		return image.Rectangle{Min: image.Pt(0, bx/2), Max: image.Pt(bx/2, bx)}
	case RQR_BR:
		return image.Rectangle{Min: image.Pt(bx/2, bx/2), Max: image.Pt(bx, bx)}
	}
	return image.Rectangle{}
}
