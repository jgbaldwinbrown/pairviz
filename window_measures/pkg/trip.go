package windif

type TripletCounts struct {
	Reg11Counts [11]int64
	Reg12Counts [12]int64
}

func BaseMatch(a, b byte) bool {
	return a == b && a != 'n' && a != 'N'
}

func BasesMatch(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range []byte(a) {
		if !BaseMatch(a[i], b[i]) {
			return false
		}
	}
	return true
}

func CountTriplets(seq1, seq2 string) TripletCounts {
	c := TripletCounts{}
	for i := 0; i < 11; i++ {
		s1 := seq1[i:]
		s2 := seq2[i:]
		l := len(s1) - 3
		for j := 0; j < l; j += 11 {
			if BasesMatch(s1[j:j+3], s2[j:j+3]) {
				c.Reg11Counts[i]++
			}
		}
	}
	for i := 0; i < 12; i++ {
		s1 := seq1[i:]
		s2 := seq2[i:]
		l := len(s1) - 3
		for j := 0; j < l; j += 12 {
			if BasesMatch(s1[j:j+3], s2[j:j+3]) {
				c.Reg12Counts[i]++
			}
		}
	}
	return c
}

func BestTriplets(cs *TripletCounts) int64 {
	var best int64 = 0
	for _, c := range cs.Reg11Counts {
		if c > best {
			best = c
		}
	}
	for _, c := range cs.Reg12Counts {
		if c > best {
			best = c
		}
	}
	return best
}

func TripletWinAvg(wp Winpair, winsize, winstep int64) float64 {
	l := int64(len(wp.Fa1.Seq)) - winsize
	sum := 0.0
	var i int64
	for i = 0; i < l; i += winstep {
		c := CountTriplets(wp.Fa1.Seq[i:i + winsize], wp.Fa2.Seq[i:i + winsize])
		sum += float64(BestTriplets(&c))
	}
	return sum / float64(l)
}

func TripletsPerBp(wp Winpair, winsize, winstep int64) float64 {
	avg := TripletWinAvg(wp, winsize, winstep)
	return avg / float64(winsize)
}
