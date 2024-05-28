package wfc

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"time"
)

var (
	localID int = 0
	rnumgen *rand.Rand

	COLOR_RED     = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	COLOR_GREEN   = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	COLOR_BLUE    = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	COLOR_YELLOW  = color.RGBA{R: 255, G: 255, B: 0, A: 255}
	COLOR_CYAN    = color.RGBA{R: 0, G: 255, B: 255, A: 255}
	COLOR_MAGENTA = color.RGBA{R: 255, G: 0, B: 255, A: 255}
	COLOR_ORANGE  = color.RGBA{R: 255, G: 165, B: 0, A: 255}
	COLOR_PURPLE  = color.RGBA{R: 128, G: 0, B: 128, A: 255}
	COLOR_BROWN   = color.RGBA{R: 165, G: 42, B: 42, A: 255}

	COLOR_MAP = [9]color.RGBA{
		COLOR_RED,
		COLOR_GREEN,
		COLOR_BLUE,
		COLOR_YELLOW,
		COLOR_CYAN,
		COLOR_MAGENTA,
		COLOR_ORANGE,
		COLOR_PURPLE,
		COLOR_BROWN,
	}
)

type WFModelSpatialFS struct {
	Width, Height             int
	RegionWidth, RegionHeight int
	Square                    bool
}

type WFModelPartition struct {
	Size, Position image.Point
	Region         image.Rectangle
}

type WFModelSet struct {
	ID           int
	Name         string
	Size         *WFModelSpatialFS
	Subdivisions image.Point
	BaseImage    *image.RGBA
	Partitions   []*WFModelPartition
}

// Creates a new WFModelSet using the parameters given
func NewWFModelSet(name string, sizeX int, sizeY int, subdivisionsX, subdivisionsY int) *WFModelSet {
	localID++

	spif := &WFModelSpatialFS{
		Width:        sizeX,
		Height:       sizeY,
		Square:       sizeX == sizeY,
		RegionWidth:  sizeX / subdivisionsX,
		RegionHeight: sizeY / subdivisionsY,
	}

	model := &WFModelSet{
		ID:           localID,
		Name:         name,
		Size:         spif,
		Subdivisions: image.Point{Y: subdivisionsY, X: subdivisionsX},
		BaseImage:    image.NewRGBA(image.Rect(0, 0, sizeX, sizeY)),
		Partitions:   []*WFModelPartition{},
	}

	for y := 0; y < subdivisionsY; y++ {
		for x := 0; x < subdivisionsX; x++ {

			partition := &WFModelPartition{
				Size:     image.Point{Y: spif.RegionHeight, X: spif.RegionWidth},
				Position: image.Point{Y: y, X: x},
				Region: image.Rectangle{
					Min: image.Pt(x*spif.RegionWidth, y*spif.RegionHeight),
					Max: image.Pt((x+1)*spif.RegionWidth, (y+1)*spif.RegionHeight),
				},
			}

			fillImageRectWithColor(model.BaseImage, GetRandomColor(), partition.Region)
			model.Partitions = append(model.Partitions, partition)
		}
	}

	// model.SubRegions = regions
	return model
}

// Load image data from path
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

// Save image to local drive
func SaveImage(file string, img image.Image) error {
	_, err := os.Stat(file)
	if err != nil {
		return err
	}

	if err := os.Remove(file); err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()
	png.Encode(f, img)
	return nil
}

// fill a rect within the image a certain color
func fillImageRectWithColor(img draw.Image, col color.Color, rect image.Rectangle) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			img.Set(x, y, col)
		}
	}
}

// Load image from path provided.
// returns a pointer to the loaded image data
func LoadImageFromPath(path string) (*image.Image, error) {
	img, err := LoadImage(path)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

// make a new sample image
func NewSampleImage(name string, size int) error {

	InitRnd(rand.New(rand.NewSource(time.Now().UnixNano())))

	// seems like sizes need to be fairly even or the fractional bit adds up.
	// the more unevenly they divide leaves more of the edge unprocessed (alpha)
	wfModel := NewWFModelSet("output/samples/coolset.png", 512, 512, 32, 32)

	file, err := os.Create(wfModel.Name)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = png.Encode(file, wfModel.BaseImage)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Random sample image %q.png created in output/samples.\n", wfModel.Name)
	return nil
}

// cycles colors from an index
func GetColorFromIndex(index int) color.RGBA {
	ix := index % len(COLOR_MAP)
	if ix <= len(COLOR_MAP) {
		return COLOR_MAP[ix]
	}
	return COLOR_RED
}

// initialize the random number seed
func InitRnd(rng *rand.Rand) {
	rnumgen = rng
}

// get a random color
func GetRandomColor() color.RGBA {
	return color.RGBA{
		R: uint8(rnumgen.Intn(256)),
		G: uint8(rnumgen.Intn(256)),
		B: uint8(rnumgen.Intn(256)),
		A: uint8(255),
	}
}
