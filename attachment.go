package spine

// Attachment
type Attachment struct {
	Name        string
	X           float32
	Y           float32
	Rotation    float32
	ScaleX      float32
	ScaleY      float32
	Width       float32
	Height      float32
	WidthRatio  float32
	HeightRatio float32
	OriginX     float32
	OriginY     float32
}

func NewAttachment(name string) *Attachment {
	attachment := new(Attachment)
	attachment.Name = name
	return attachment
}
