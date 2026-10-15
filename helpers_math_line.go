package shapes

type lineF32 struct {
	Origin PointF32
	Dir    PointF32
}

type segmentF32 struct {
	Origin, End PointF32
}

func (s segmentF32) Dir() PointF32 {
	return s.End.Sub(s.Origin)
}
