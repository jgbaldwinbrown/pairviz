package pairviz

import (
	"github.com/gonum/stat/distuv"
)

type ObservedExpected struct {
	Observed float64
	Expected float64
}

func ChiSq(oes ...ObservedExpected) float64 {
	sum := 0.0
	for _, oe := range oes {
		diff := oe.Observed - oe.Expected
		sum += (diff * diff) / oe.Expected
	}
	return sum
}

func ChiSqNoDivZero(oes ...ObservedExpected) (chisq float64, skipped int) {
	sum := 0.0
	for _, oe := range oes {
		if oe.Expected == 0 {
			skipped++
			continue
		}
		diff := oe.Observed - oe.Expected
		sum += (diff * diff) / oe.Expected
	}
	return sum, skipped
}

func ChiSqP(chisq float64, df float64) float64 {
	dist := distuv.ChiSquared{K: df}
	return 1.0 - dist.CDF(chisq)
}
