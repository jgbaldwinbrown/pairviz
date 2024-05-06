package windif

import (
	"encoding/json"
	"math"
)

type Float struct {
	F float64
}

func (f Float) MarshalJSON() ([]byte, error) {
	if math.IsNaN(f.F) {
		return []byte(`"NaN"`), nil
	}
	if math.IsInf(f.F, 0) {
		if math.IsInf(f.F, -1) {
			return []byte(`"-Inf"`), nil
		}
		return []byte(`"Inf"`), nil
	}
	return json.Marshal(f.F)
}

func (f *Float) UnmarshalJSON(b []byte) error {
	s := string(b)
	switch s {
	case `"NaN"`:
		f.F = math.NaN()
	case `"Inf"`:
		f.F = math.Inf(1)
	case `"-Inf"`:
		f.F = math.Inf(-1)
	default:
		return json.Unmarshal(b, &f.F)
	}
	return nil
}
