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

	cnv_MAP = [8]CellNVector{
		CNV_T,
		CNV_TL,
		CNV_L,
		CNV_BL,
		CNV_B,
		CNV_BR,
		CNV_R,
		CNV_TR,
	}
)

type CellNVector uint

const (
	CNV_T  CellNVector = iota      // 0, -1
	CNV_TL CellNVector = 1 << iota // -1, -1
	CNV_L  CellNVector = 1 << iota // -1, 0
	CNV_BL CellNVector = 1 << iota // -1, 1
	CNV_B  CellNVector = 1 << iota // 0, 1
	CNV_BR CellNVector = 1 << iota // 1, 1
	CNV_R  CellNVector = 1 << iota // 1, 0
	CNV_TR CellNVector = 1 << iota // 1, 1

)

// Returns the offset that should be applied to get the neighboring position
func (cnv CellNVector) GetOffset() (x, y int) {
	switch cnv {
	case CNV_T:
		return 0, -1
	case CNV_TL:
		return -1, -1
	case CNV_L:
		return -1, 0
	case CNV_BL:
		return -1, 1
	case CNV_B:
		return 0, 1
	case CNV_BR:
		return 1, 1
	case CNV_R:
		return 1, 0
	case CNV_TR:
		return 1, 1
	}
	return 0, 0
}

type MacroState struct {
	MicroState map[CellNVector]float64
}

// Defines a cell's states and entropy systems
type CellState struct {
	Pt      image.Point
	Entropy map[color.RGBA]*MacroState
}

// Conducts analysis of the cell state. Tracks colors and their position in relation to others
// to eventually generate a map of weights that can be used to generate rules for outputs
func (cs *CellState) EntropicAnalysis(img *image.RGBA) {
	for _, k := range cnv_MAP {
		ox, oy := k.GetOffset()
		// fmt.Printf("offsets: [%d, %d]", ox, oy)
		// fmt.Printf("bounds %+v", img)
		bxmax, bymax := img.Rect.Max.X, img.Rect.Max.Y
		if cs.Pt.X+ox < 0 || cs.Pt.Y+oy < 0 || ox > bxmax || oy > bymax {
			continue
		}
		// fmt.Printf("PT %+v", cs.Pt)

		pt := image.Pt(cs.Pt.X+ox, cs.Pt.Y+oy)
		if !pt.In(img.Bounds()) {
			continue
		}
		col := GetPixelColor(img, pt)
		_, ok := cs.Entropy[col]
		if !ok {
			cs.Entropy[col] = &MacroState{
				MicroState: make(map[CellNVector]float64, 0),
			}
		}

		_, ok = cs.Entropy[col].MicroState[k]
		if !ok {
			cs.Entropy[col].MicroState[k] = 1
		}

		cs.Entropy[col].MicroState[k] += 1
	}
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
	Cells          []*CellState
	Colors         map[color.RGBA]float64
	Entropy        map[color.RGBA]*MacroState
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

func (mp *WFModelPartition) Analyze() {
	sX, sY := mp.Region.Bounds().Dx(), mp.Region.Bounds().Dy()
	// xmin, xmax := mp.Region.Min.X, mp.Region.Max.X
	// fmt.Printf("[sX: %d], [sY: %d], [xmin: %d] [xmax: %d]", sX, sY, xmin, xmax)

	for y := 0; y < sY; y++ {
		for x := 0; x < sX; x++ {
			cs := &CellState{
				Pt:      image.Pt(x, y),
				Entropy: make(map[color.RGBA]*MacroState),
			}
			cs.EntropicAnalysis(mp.Data)
			// for i, j := range cs.Entropy {
			// 	for mcs, nvd := range j.MicroState {
			// 		if _, ok := mp.Entropy[i]; !ok {
			// 			mp.Entropy[i] = &MacroState{
			// 				MicroState: map[CellNVector]float64{},
			// 			}
			// 		}
			// 		if _, ok := mp.Entropy[i].MicroState[mcs]; !ok {
			// 			mp.Entropy[i].MicroState[mcs] = 0
			// 		}
			// 		mp.Entropy[i].MicroState[mcs] += nvd
			// 	}
			// }
			mp.Cells = append(mp.Cells, cs)

			for col, mc := range cs.Entropy {
				for mis, mas := range mc.MicroState {

					if _, ok := mp.Entropy[col]; !ok {
						mp.Entropy[col] = &MacroState{MicroState: make(map[CellNVector]float64)}
					}

					if _, ok := mp.Entropy[col].MicroState[mis]; !ok {
						mp.Entropy[col].MicroState[mis] = 0
					}

					mp.Entropy[col].MicroState[mis] += mas

					if _, ok := mp.Colors[col]; !ok {
						mp.Colors[col] = 0
					}

					mp.Colors[col] += mas

					// fmt.Printf("color: [%+v], vector: [%+v], weight: %+f\n", col, mis, mas)
				}
			}
		}
	}

	// fmt.Printf("Entropy length")

	// for col, mc := range mp.Entropy {
	// 	for s, v := range mc.MicroState {
	// 		fmt.Printf("E: [%+v] - [%+v]: %f\n", col, s, v)
	// 	}
	// }
	for col, wt := range mp.Colors {
		fmt.Printf("color: %+v wt: %f\n", col, wt)
	}

	// fmt.Printf("Cell Size [%d, %d]. Pixels: %d\n", sX, sY, len(mp.Cells))
	// fmt.Printf("	Averaged Weights %+v", mp.Entropy)
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
				Cells:          []*CellState{},
				Colors:         make(map[color.RGBA]float64),
				Entropy:        make(map[color.RGBA]*MacroState),
			}
			partition.Data = CopyImageRegionData(model.BaseImage, partition.Region)
			model.Partitions = append(model.Partitions, partition)
		}
	}

	for _, mp := range model.Partitions {
		mp.Analyze()
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

	for _, mp := range model.Partitions {
		mp.Analyze()
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
