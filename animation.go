package spine

import (
	"math"
)

type Timeline interface {
	Apply(skeleton *Skeleton, time, alpha float32)
}

type RotateTimeline struct {
	boneIndex int
	frames    []float32
	curve     *Curve
}

func NewRotateTimeline(l int) *RotateTimeline {
	timeline := new(RotateTimeline)
	timeline.frames = make([]float32, l*2)
	timeline.curve = NewCurve(l)
	return timeline
}

func (t *RotateTimeline) Apply(skeleton *Skeleton, time, alpha float32) {
	frames := t.frames
	if time < frames[0] {
		return
	}

	bone := skeleton.Bones[t.boneIndex]
	if time >= frames[len(frames)-2] {
		amount := bone.Data.rotation + frames[len(frames)-1] - bone.Rotation
		for amount > 180 {
			amount -= 360
		}

		for amount < -180 {
			amount += 360
		}
		bone.Rotation += amount * alpha
		return
	}

	frameIndex := binarySearch(frames, time, 2)
	lastFrameValue := frames[frameIndex-1]
	frameTime := frames[frameIndex]
	percent := 1 - (time-frameTime)/(frames[frameIndex-2]-frameTime)
	percent = t.curve.CurvePercent(frameIndex/2-1, percent)
	amount := frames[frameIndex+1] - lastFrameValue
	for amount > 180 {
		amount -= 360
	}
	for amount < -180 {
		amount += 360
	}
	amount = bone.Data.rotation + (lastFrameValue + amount*percent) - bone.Rotation
	for amount > 180 {
		amount -= 360
	}

	for amount < -180 {
		amount += 360
	}
	bone.Rotation += amount * alpha
}

func binarySearch(values []float32, target float32, step int) int {
	low := 0
	high := int(math.Floor(float64(len(values)/step))) - 2
	if high == 0 {
		return step
	}
	current := high >> 1
	for {
		if values[(current+1)*step] <= target {
			low = current + 1
		} else {
			high = current
		}
		if low == high {
			return (low + 1) * step
		}
		current = (low + high) >> 1
	}
}

func (t *RotateTimeline) setFrame(index int, time, angle float32) {
	frameIndex := index * 2
	t.frames[frameIndex] = time
	t.frames[frameIndex+1] = angle
}

func (t *RotateTimeline) frameCount() int {
	return len(t.frames) / 2
}

type TranslateTimeline struct {
	boneIndex int
	frames    []float32
	curve     *Curve
}

func NewTranslateTimeline(l int) *TranslateTimeline {
	timeline := new(TranslateTimeline)
	timeline.frames = make([]float32, l*3)
	timeline.curve = NewCurve(l)
	return timeline
}

func (t *TranslateTimeline) frameCount() int {
	return len(t.frames) / 3
}

func (t *TranslateTimeline) setFrame(index int, time, x, y float32) {
	frameIndex := index * 3
	t.frames[frameIndex] = time
	t.frames[frameIndex+1] = x
	t.frames[frameIndex+2] = y
}

func (t *TranslateTimeline) Apply(skeleton *Skeleton, time, alpha float32) {
	frames := t.frames
	if time < frames[0] {
		return
	}

	bone := skeleton.Bones[t.boneIndex]

	if time >= frames[len(frames)-3] {
		bone.X += (bone.Data.x + frames[len(frames)-2] - bone.X) * alpha
		bone.Y += (bone.Data.y + frames[len(frames)-1] - bone.Y) * alpha
		return
	}

	frameIndex := binarySearch(frames, time, 3)
	lastFrameX := frames[frameIndex-2]
	lastFrameY := frames[frameIndex-1]
	frameTime := frames[frameIndex]
	percent := 1 - (time-frameTime)/(frames[frameIndex-3]-frameTime)
	percent = t.curve.CurvePercent(frameIndex/3-1, percent)

	bone.X += (bone.Data.x + lastFrameX + (frames[frameIndex+1]-lastFrameX)*percent - bone.X) * alpha
	bone.Y += (bone.Data.y + lastFrameY + (frames[frameIndex+2]-lastFrameY)*percent - bone.Y) * alpha
}

type ScaleTimeline struct {
	boneIndex int
	frames    []float32
	curve     *Curve
}

func NewScaleTimeline(l int) *ScaleTimeline {
	timeline := new(ScaleTimeline)
	timeline.frames = make([]float32, l*3)
	timeline.curve = NewCurve(l)
	return timeline
}

func (t *ScaleTimeline) frameCount() int {
	return len(t.frames) / 3
}

func (t *ScaleTimeline) setFrame(index int, time, x, y float32) {
	frameIndex := index * 3
	t.frames[frameIndex] = time
	t.frames[frameIndex+1] = x
	t.frames[frameIndex+2] = y
}

func (t *ScaleTimeline) Apply(skeleton *Skeleton, time, alpha float32) {
	frames := t.frames
	if time < frames[0] {
		return
	}

	bone := skeleton.Bones[t.boneIndex]

	if time >= frames[len(frames)-3] {
		bone.ScaleX += (bone.Data.scaleX - 1 + frames[len(frames)-2] - bone.ScaleX) * alpha
		bone.ScaleY += (bone.Data.scaleY - 1 + frames[len(frames)-1] - bone.ScaleY) * alpha
		return
	}

	frameIndex := binarySearch(frames, time, 3)
	lastFrameX := frames[frameIndex-2]
	lastFrameY := frames[frameIndex-1]
	frameTime := frames[frameIndex]
	percent := 1 - (time-frameTime)/(frames[frameIndex-3]-frameTime)
	percent = t.curve.CurvePercent(frameIndex/3-1, percent)

	bone.ScaleX += (bone.Data.scaleX - 1 + lastFrameX + (frames[frameIndex+1]-lastFrameX)*percent - bone.ScaleX) * alpha
	bone.ScaleY += (bone.Data.scaleY - 1 + lastFrameY + (frames[frameIndex+2]-lastFrameY)*percent - bone.ScaleY) * alpha
}

type Animation struct {
	name      string
	timelines []Timeline
	duration  float32
}

func NewAnimation(name string, timelines []Timeline, duration float32) *Animation {
	anim := new(Animation)
	anim.name = name
	anim.timelines = timelines
	anim.duration = duration
	return anim
}

func (a *Animation) Apply(skeleton *Skeleton, time float32, loop bool) {
	if loop && a.duration != 0 {
		time = float32(math.Mod(float64(time), float64(a.duration)))
	}
	for _, timeline := range a.timelines {
		timeline.Apply(skeleton, time, 1)
	}
}

func (a *Animation) Mix(skeleton *Skeleton, time float32, loop bool, alpha float32) {
	if loop && a.duration != 0 {
		time = float32(math.Mod(float64(time), float64(a.duration)))
	}
	for _, timeline := range a.timelines {
		timeline.Apply(skeleton, time, alpha)
	}
}

func (a *Animation) Duration() float32 {
	return a.duration
}
