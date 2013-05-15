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
}

func NewRotateTimeline(l int) *RotateTimeline {
	timeline := new(RotateTimeline)
	timeline.frames = make([]float32, l*2)
	return timeline
}

func (t *RotateTimeline) Apply(skeleton *Skeleton, time, alpha float32) {
	frames := t.frames
	if time < frames[0] {
		return
	}

	bone := skeleton.bones[t.boneIndex]
	if time >= frames[len(frames)-2] {
		amount := bone.data.rotation + frames[len(frames)-1] - bone.Rotation
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
	// TODO: curves
	amount := frames[frameIndex+1] - lastFrameValue
	for amount > 180 {
		amount -= 360
	}
	for amount < -180 {
		amount += 360
	}
	amount = bone.data.rotation + (lastFrameValue + amount*percent) - bone.Rotation
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
}

func NewTranslateTimeline(l int) *TranslateTimeline {
	timeline := new(TranslateTimeline)
	timeline.frames = make([]float32, l*3)
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

	bone := skeleton.bones[t.boneIndex]

	if time >= frames[len(frames)-3] {
		bone.X += (bone.data.x + frames[len(frames)-2] - bone.X) * alpha
		bone.Y += (bone.data.y + frames[len(frames)-1] - bone.Y) * alpha
		return
	}

	frameIndex := binarySearch(frames, time, 3)
	lastFrameX := frames[frameIndex-2]
	lastFrameY := frames[frameIndex-1]
	frameTime := frames[frameIndex]
	percent := 1 - (time-frameTime)/(frames[frameIndex-3]-frameTime)

	// TODO: percent = this.curves.getCurvePercent(frameIndex / 3 - 1, percent)

	bone.X += (bone.data.x + lastFrameX + (frames[frameIndex+1]-lastFrameX)*percent - bone.X) * alpha
	bone.Y += (bone.data.y + lastFrameY + (frames[frameIndex+2]-lastFrameY)*percent - bone.Y) * alpha
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
		for _, timeline := range a.timelines {
			timeline.Apply(skeleton, time, 1)
		}
	}
}
