package wfc

var (
	WFC_FALSE_START_LIMIT     uint = 10000
	WFC_CROSS_REFERENCE_LIMIT uint = 200
)

type WFPosDelta uint8

const (
	WFPD_0  WFPosDelta = iota
	WFPD_N  WFPosDelta = iota
	WFPD_NW WFPosDelta = iota
	WFPD_W  WFPosDelta = iota
	WFPD_SW WFPosDelta = iota
	WFPD_S  WFPosDelta = iota
	WFPD_SE WFPosDelta = iota
	WFPD_E  WFPosDelta = iota
	WFPD_NE WFPosDelta = iota
)

type WFTile struct {
	heuristic float64
}

func NewWFTile() *WFTile {
	return &WFTile{
		heuristic: 1,
	}
}

type WaveFunction struct {
	items []*WFTile
}
