package wfc

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
)

var (
	localID int = 0
)

// Defines a final cell state
type CellState struct {
	ID     int
	Pixels []color.Color
}

type Cell struct {
	Pt        image.Point
	ID        int
	Wt        float64
	Collapsed bool
	Pixels    []color.Color
	State     *CellState
}

type WFModelSpatialFS struct {
	Width, Height             int
	RegionWidth, RegionHeight int
	Square                    bool
}

type WFModelPartition struct {
	Size, Position image.Point
	Region         image.Rectangle
	Data           *image.RGBA
	CollapsedCount int
	FullyCollapsed bool
	Cells          []*Cell
	Colors         map[color.RGBA]int
}

func (mp *WFModelPartition) FindColors() {
	for _, c := range mp.Cells {
		col := GetPixelColor(mp.Data, c.Pt)
		_, ok := mp.Colors[col]
		if !ok {
			mp.Colors[col] = 1
		}
	}
}

type WFModelSet struct {
	ID           int
	Name         string
	Size         *WFModelSpatialFS
	Subdivisions image.Point
	BaseImage    *image.RGBA
	Partitions   []*WFModelPartition
	States       []*CellState
	Colors       []color.Color
}

// Create a new model set from the provided image at path
func NewWFModelSet(path string, subdivX, subdivY int) (*WFModelSet, error) {
	localID++
	img, err := LoadImage(path)
	if err != nil {
		return nil, err
	}

	f, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	sX, sY := bounds.Dx(), bounds.Dy()

	spif := &WFModelSpatialFS{
		Width:        sX,
		Height:       sY,
		Square:       sX == sY,
		RegionWidth:  sX / subdivX,
		RegionHeight: sY / subdivY,
	}

	model := &WFModelSet{
		ID:           localID,
		Name:         f.Name(),
		Size:         spif,
		Subdivisions: image.Point{Y: subdivY, X: subdivX},
		BaseImage:    ImageToRGBA(img),
		Partitions:   []*WFModelPartition{},
	}

	for y := 0; y < subdivY; y++ {
		for x := 0; x < subdivX; x++ {
			partition := &WFModelPartition{
				Size:     image.Point{Y: spif.RegionHeight, X: spif.RegionWidth},
				Position: image.Point{Y: y, X: x},
				Region: image.Rectangle{
					Min: image.Pt(x*spif.RegionWidth, y*spif.RegionHeight),
					Max: image.Pt((x+1)*spif.RegionWidth, (y+1)*spif.RegionHeight),
				},
				CollapsedCount: 0,
				FullyCollapsed: false,
				Cells:          []*Cell{},
				Colors:         make(map[color.RGBA]int),
			}
			partition.Data = CopyImageRegionData(model.BaseImage, partition.Region)
			model.Partitions = append(model.Partitions, partition)
		}
	}

	return model, nil
}

// Save the models data to an output png file
func (model *WFModelSet) Save() error {
	if err := SaveImage(fmt.Sprintf("output/generated_%s", model.Name), model.BaseImage); err != nil {
		return err
	}
	return nil
}

// Creates a new sample image
// @TODO partition analysis and propagation to collapse the wave function
func NewSample(name string, sizeX int, sizeY int, subdivisionsX, subdivisionsY int) *WFModelSet {
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

			SetRegionColor(model.BaseImage, GetRandomColor(), partition.Region)
			model.Partitions = append(model.Partitions, partition)
		}
	}

	return model
}

// Creates and saves a new sample image with provided name and size
// outputs to output/samples in png format.
func NewSampleImage(name string, size int) error {

	fullPath := fmt.Sprintf("output/samples/%s.png", name)

	// seems like sizes need to be fairly even or the fractional bit adds up.
	// the more unevenly they divide leaves more of the edge unprocessed (alpha)
	wfModel := NewSample(fullPath, size, size, 32, 32)

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
