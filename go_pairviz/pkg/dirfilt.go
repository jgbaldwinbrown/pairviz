package pairviz

//	Dir int
//
//
//type Facing int
//
//const (
//	Unknown Facing = iota
//	In
//	Out
//	Match
//)
//
//
//func (p Pair) Face() Facing {
//	if p.Read1.Dir < 0 && p.Read2.Dir < 0 {
//		return Match
//	}
//	if p.Read1.Dir > 0 && p.Read2.Dir > 0 {
//		return Match
//	}
//
//	lread := p.Read1
//	rread := p.Read2
//
//	if p.Read2.Pos < p.Read1.Pos {
//		lread = p.Read2
//		rread = p.Read1
//	}
//
//	if lread.Dir < 0 && rread.Dir > 0 {
//		return Out
//	}
//
//	if lread.Dir > 0 && rread.Dir < 0 {
//		return In
//	}
//
//	return Unknown
//}
//
//
//	switch fields[2] {
//		case "+": read.Dir = 1
//		case "-": read.Dir = -1
//		default: read.Dir = 0
//	}
//
//
//	pair.Read1 = ParseRead(append([]string{}, line[1], line[2], line[5]))
//	pair.Read2 = ParseRead(append([]string{}, line[3], line[4], line[6]))
