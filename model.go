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

	COLOR_RED     = color.RGBA{R: 255, G: 0, B: 0, A: 255}   // Red
	COLOR_GREEN   = color.RGBA{R: 0, G: 255, B: 0, A: 255}   // Green
	COLOR_BLUE    = color.RGBA{R: 0, G: 0, B: 255, A: 255}   // Blue
	COLOR_YELLOW  = color.RGBA{R: 255, G: 255, B: 0, A: 255} // Yellow
	COLOR_CYAN    = color.RGBA{R: 0, G: 255, B: 255, A: 255} // Cyan
	COLOR_MAGENTA = color.RGBA{R: 255, G: 0, B: 255, A: 255} // Magenta
	COLOR_ORANGE  = color.RGBA{R: 255, G: 165, B: 0, A: 255} // Orange
	COLOR_PURPLE  = color.RGBA{R: 128, G: 0, B: 128, A: 255} // Purple
	COLOR_BROWN   = color.RGBA{R: 165, G: 42, B: 42, A: 255} // Brown

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
	Size   *WFModelSpatialFS
	Region image.Rectangle
}

type WFModelSet struct {
	Name         string
	Size         *WFModelSpatialFS
	Subdivisions int
	BaseImage    *image.RGBA
	SubRegions   [][]image.Rectangle
}

// Creates a new WFModelSet using the parameters given
func NewWFModelSet(name string, sizeX int, sizeY int, subdivisions int) *WFModelSet {
	spif := &WFModelSpatialFS{
		Width:        sizeX,
		Height:       sizeY,
		Square:       sizeX == sizeY,
		RegionWidth:  sizeX / subdivisions,
		RegionHeight: sizeY / subdivisions,
	}

	model := &WFModelSet{
		Name:         name,
		Size:         spif,
		Subdivisions: subdivisions,
		BaseImage:    image.NewRGBA(image.Rect(0, 0, sizeX, sizeY)),
		SubRegions:   [][]image.Rectangle{},
	}

	subdivisionCt := subdivisions / 2
	regions := make([][]image.Rectangle, subdivisions)

	for y := 0; y < subdivisionCt; y++ {
		regions[y] = make([]image.Rectangle, subdivisionCt)
		for x := 0; x < subdivisionCt; x++ {
			minX := x * spif.RegionWidth
			minY := y * spif.RegionHeight
			maxX := (x + 1) * spif.RegionWidth
			maxY := (y + 1) * spif.RegionHeight

			if x == subdivisionCt-1 {
				maxX = sizeX
			}

			if y == subdivisionCt-1 {
				maxY = sizeY
			}

			regions[y][x] = image.Rect(minX, minY, maxX, maxY)
		}
	}

	model.SubRegions = regions
	return model
}

// type WFCModel struct {
// 	ID     int
// 	height int
// 	width  int
// }

// func NewWFCModel(h, w int) *WFCModel {
// 	localID++
// 	return &WFCModel{
// 		ID:     localID,
// 		height: h,
// 		width:  w,
// 	}
// }

// split the image into 9 sub regions
func SplitImage(img *image.RGBA) [9]*image.RGBA {
	var regions [9]*image.RGBA
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	regionWidth := width / 3
	regionHeight := height / 3

	// Define the 9 regions (top-left, top-middle, top-right, middle-left, center, middle-right, bottom-left, bottom-middle, bottom-right)
	regionCoordinates := [9]image.Rectangle{
		image.Rect(0, 0, regionWidth, regionHeight),                          // Top-left
		image.Rect(regionWidth, 0, 2*regionWidth, regionHeight),              // Top-middle
		image.Rect(2*regionWidth, 0, width, regionHeight),                    // Top-right
		image.Rect(0, regionHeight, regionWidth, 2*regionHeight),             // Middle-left
		image.Rect(regionWidth, regionHeight, 2*regionWidth, 2*regionHeight), // Center
		image.Rect(2*regionWidth, regionHeight, width, 2*regionHeight),       // Middle-right
		image.Rect(0, 2*regionHeight, regionWidth, height),                   // Bottom-left
		image.Rect(regionWidth, 2*regionHeight, 2*regionWidth, height),       // Bottom-middle
		image.Rect(2*regionWidth, 2*regionHeight, width, height),             // Bottom-right
	}

	for i, rect := range regionCoordinates {
		subImg := image.NewRGBA(rect)
		draw.Draw(subImg, subImg.Bounds(), img, rect.Min, draw.Src)
		regions[i] = subImg
	}

	return regions
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

// fill a region from [xMin, yMin] to [xMax, yMax] with a solid color
// going to use for debugging visual. helps to visualize thge grid when subdividing it.
func fillReigonWithColor(img *image.RGBA, col color.Color, xMin, yMin, xMax, yMax int) {
	for y := yMin; y < yMax; y++ {
		for x := xMin; x < xMax; x++ {
			img.Set(x, y, col)
		}
	}
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

// func GenModelFromInputImage(imgPath string) *WFCModel {
// 	img, err := LoadImage(imgPath)
// 	if err != nil {
// 		panic("failed to load image " + imgPath)
// 	}

// 	fmt.Printf("IMG:\n%+v\n", img)
// 	return &WFCModel{
// 		ID: 1,
// 	}
// }

// make a new sample image
func NewSampleImage(name string, size int) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	halfGridX := size / 2
	qtrGridX := halfGridX / 2
	eighthGridX := qtrGridX / 2
	sixTeenth := eighthGridX / 2

	fmt.Printf("Grid Size: %d. Half: %d. Qtr: %d. Eighth: %d. Sixteenth: %d \n", size, halfGridX, qtrGridX, eighthGridX, sixTeenth)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var col color.Color

			r := uint8(rng.Intn(256))
			g := uint8(rng.Intn(256))
			b := uint8(rng.Intn(256))
			a := uint8(255)
			col = color.RGBA{R: r, G: g, B: b, A: a}

			img.Set(x, y, col)
		}
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	fmt.Printf("Image Size: width %d, height %d \n", width, height)

	// Define the coordinates for the 9 regions
	regionCoordinates := [9]image.Rectangle{
		image.Rect(0, 0, size/3, size/3),               // Top-left
		image.Rect(size/3, 0, 2*size/3, size/3),        // Top-middle
		image.Rect(2*size/3, 0, size, size/3),          // Top-right
		image.Rect(0, size/3, size/3, 2*size/3),        // Middle-left
		image.Rect(size/3, size/3, 2*size/3, 2*size/3), // Center
		image.Rect(2*size/3, size/3, size, 2*size/3),   // Middle-right
		image.Rect(0, 2*size/3, size/3, size),          // Bottom-left
		image.Rect(size/3, 2*size/3, 2*size/3, size),   // Bottom-middle
		image.Rect(2*size/3, 2*size/3, size, size),     // Bottom-right
	}

	for i, rect := range regionCoordinates {
		fillImageRectWithColor(img, COLOR_MAP[i], rect)
	}

	mfModelSetItem := NewWFModelSet("output/samples/coolset.png", 64, 64, 9)

	for i, subreg := range mfModelSetItem.SubRegions {
		for j, rect := range subreg {
			fmt.Printf("[%d-%d] RECT: %+v\n", i, j, rect)
			fillImageRectWithColor(mfModelSetItem.BaseImage, COLOR_MAP[i+j], rect)
		}
	}

	file1, err := os.Create(mfModelSetItem.Name)
	if err != nil {
		panic(err)
	}
	defer file1.Close()

	err = png.Encode(file1, mfModelSetItem.BaseImage)
	if err != nil {
		panic(err)
	}

	file, err := os.Create(fmt.Sprintf("output/samples/%s.png", name))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Random sample image %q.png created in output/samples.\n", name)
	return nil

}
