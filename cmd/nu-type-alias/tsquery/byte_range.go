package tsquery

type ByteRange struct {
	Start uint
	End   uint
}

func NewByteRange(start, end uint) ByteRange {
	return ByteRange{
		Start: start,
		End:   end,
	}
}

func (r ByteRange) GetString(buff []byte) string {
	return string(buff[r.Start:r.End])
}
