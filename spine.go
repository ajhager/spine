package spine

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"strconv"
)

type fileAnim struct {
	Bones map[string]map[string][]interface{} `json:"bones"`
}

type fileSlot struct {
	Bone       string `json:"bone"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Attachment string `json:"attachment"`
}

type fileBone struct {
	Name   string `json:"name"`
	Parent string `json:"parent"`

	Length   interface{} `json:"length"`
	Rotation interface{} `json:"rotation"`
	X        interface{} `json:"x"`
	Y        interface{} `json:"y"`
	ScaleX   interface{} `json:"scaleX"`
	ScaleY   interface{} `json:"scaleY"`
}

type fileAttachment struct {
	Name string `json:"name"`

	Rotation interface{} `json:"rotation"`
	X        interface{} `json:"x"`
	Y        interface{} `json:"y"`
	ScaleX   interface{} `json:"scaleX"`
	ScaleY   interface{} `json:"scaleY"`
	Width    interface{} `json:"width"`
	Height   interface{} `json:"height"`
}

type fileRoot struct {
	Bones      []fileBone                                      `json:"bones"`
	Slots      []fileSlot                                      `json:"slots"`
	Skins      map[string]map[string]map[string]fileAttachment `json:"skins"`
	Animations map[string]fileAnim                             `json:"animations"`
}

func New(r io.Reader, scale float32) (*SkeletonData, error) {
	var root fileRoot
	err := json.NewDecoder(r).Decode(&root)
	if err != nil {
		return nil, errors.New("failed to parse skeleton json: " + err.Error())
	}

	skeletonData := NewSkeletonData()

	// Bones
	for _, bone := range root.Bones {
		boneName := bone.Name
		var boneParent *BoneData
		if parentName := bone.Parent; parentName != "" {
			_, boneParent = skeletonData.findBone(parentName)
			if boneParent == nil {
				return nil, errors.New("Parent bone not found: " + parentName)
			}
		}

		boneData := NewBoneData(boneName, boneParent)

		if length, ok := bone.Length.(float64); ok {
			boneData.Length = float32(length) * scale
		}

		if x, ok := bone.X.(float64); ok {
			boneData.x = float32(x) * scale
		}

		if y, ok := bone.Y.(float64); ok {
			boneData.y = float32(y) * scale
		}

		if rotation, ok := bone.Rotation.(float64); ok {
			boneData.rotation = float32(rotation)
		}

		boneData.scaleX = 1
		if scaleX, ok := bone.ScaleX.(float64); ok {
			boneData.scaleX = float32(scaleX)
		}

		boneData.scaleY = 1
		if scaleY, ok := bone.ScaleY.(float64); ok {
			boneData.scaleY = float32(scaleY)
		}

		skeletonData.bones = append(skeletonData.bones, boneData)
	}

	// Slots
	for _, slot := range root.Slots {
		boneName := slot.Bone
		_, boneData := skeletonData.findBone(boneName)
		if boneData == nil {
			return nil, errors.New("Slot bone not found: " + boneName)
		}
		slotData := NewSlotData(slot.Name, boneData)

		if color := slot.Color; color != "" {
			if red, err := strconv.ParseUint(color[0:2], 16, 8); err != nil {
				slotData.r = float32(red) / 255.0
			}
			if green, err := strconv.ParseUint(color[2:4], 16, 8); err != nil {
				slotData.g = float32(green) / 255.0
			}
			if blue, err := strconv.ParseUint(color[4:6], 16, 8); err != nil {
				slotData.b = float32(blue) / 255.0
			}
			if alpha, err := strconv.ParseUint(color[6:8], 16, 8); err != nil {
				slotData.a = float32(alpha) / 255.0
			}
		}

		slotData.attachmentName = slot.Attachment

		skeletonData.slots = append(skeletonData.slots, slotData)
	}

	for skinName, skinMap := range root.Skins {
		skin := NewSkin(skinName)
		for slotName, slotMap := range skinMap {
			slotIndex, _ := skeletonData.findSlot(slotName)
			for atName, at := range slotMap {

				if name := at.Name; name != "" {
					atName = name
				}

				attachment := NewAttachment(atName)

				if x, ok := at.X.(float64); ok {
					attachment.X = float32(x) * scale
				}

				if y, ok := at.Y.(float64); ok {
					attachment.Y = float32(y) * scale
				}

				if rotation, ok := at.Rotation.(float64); ok {
					attachment.Rotation = float32(rotation)
				}

				attachment.ScaleX = 1
				if scaleX, ok := at.ScaleX.(float64); ok {
					attachment.ScaleX = float32(scaleX)
				}

				attachment.ScaleY = 1
				if scaleY, ok := at.ScaleY.(float64); ok {
					attachment.ScaleY = float32(scaleY)
				}

				attachment.Width = 32
				if width, ok := at.Width.(float64); ok {
					attachment.Width = float32(width) * scale
				}

				attachment.Height = 32
				if height, ok := at.Height.(float64); ok {
					attachment.Height = float32(height) * scale
				}

				skin.AddAttachment(slotIndex, atName, attachment)
			}
		}
		skeletonData.skins = append(skeletonData.skins, skin)
		if skin.name == "default" {
			skeletonData.defaultSkin = skin
		}
	}

	for animName, boneMap := range root.Animations {
		timelines := make([]Timeline, 0)
		duration := float32(0)
		for boneName, timelineMap := range boneMap.Bones {
			boneIndex, _ := skeletonData.findBone(boneName)
			for timelineType, timelineData := range timelineMap {
				if timelineType == "rotate" {
					n := len(timelineData)
					timeline := NewRotateTimeline(n)
					timeline.boneIndex = boneIndex
					for i := 0; i < n; i++ {
						valueMap := timelineData[i].(map[string]interface{})
						time := float32(valueMap["time"].(float64))
						angle := float32(valueMap["angle"].(float64))
						timeline.setFrame(i, time, angle)
						if curve, ok := valueMap["curve"]; ok {
							readCurve(timeline.curve, i, curve)
						}
					}
					duration = float32(math.Max(float64(duration), float64(timeline.frames[timeline.frameCount()*2-2])))

					timelines = append(timelines, timeline)
				} else if timelineType == "translate" {
					n := len(timelineData)
					timeline := NewTranslateTimeline(n)
					timeline.boneIndex = boneIndex
					for i := 0; i < n; i++ {
						valueMap := timelineData[i].(map[string]interface{})
						x := float32(0)
						if xx, ok := valueMap["x"].(float64); ok {
							x = float32(xx) * scale
						}
						y := float32(0)
						if yy, ok := valueMap["y"].(float64); ok {
							y = float32(yy) * scale
						}
						time := float32(valueMap["time"].(float64))

						timeline.setFrame(i, time, x, y)
						if curve, ok := valueMap["curve"]; ok {
							readCurve(timeline.curve, i, curve)
						}
					}
					duration = float32(math.Max(float64(duration), float64(timeline.frames[timeline.frameCount()*3-3])))
					timelines = append(timelines, timeline)
				} else if timelineType == "scale" {
					n := len(timelineData)
					timeline := NewScaleTimeline(n)
					timeline.boneIndex = boneIndex
					for i := 0; i < n; i++ {
						valueMap := timelineData[i].(map[string]interface{})
						x := float32(0)
						if xx, ok := valueMap["x"].(float64); ok {
							x = float32(xx)
						}
						y := float32(0)
						if yy, ok := valueMap["y"].(float64); ok {
							y = float32(yy)
						}
						time := float32(valueMap["time"].(float64))

						timeline.setFrame(i, time, x, y)
						if curve, ok := valueMap["curve"]; ok {
							readCurve(timeline.curve, i, curve)
						}
					}
					duration = float32(math.Max(float64(duration), float64(timeline.frames[timeline.frameCount()*3-3])))
					timelines = append(timelines, timeline)
				}
			}
		}
		anim := NewAnimation(animName, timelines, duration)
		skeletonData.animations = append(skeletonData.animations, anim)
	}

	return skeletonData, nil
}

func readCurve(curve *Curve, frameIndex int, data interface{}) {
	switch t := data.(type) {
	default:
	case string:
		if t == "stepped" {
			curve.SetStepped(frameIndex)
		}
	case []interface{}:
		a := float32(t[0].(float64))
		b := float32(t[1].(float64))
		c := float32(t[2].(float64))
		d := float32(t[3].(float64))
		curve.SetCurve(frameIndex, a, b, c, d)
	}
}
