package spine

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"strconv"
)

func New(r io.Reader, scale float32) (*SkeletonData, error) {
	var data interface{}
	err := json.NewDecoder(r).Decode(&data)
	if err != nil {
		return nil, err
	}

	skeletonData := NewSkeletonData()
	root := data.(map[string]interface{})

	// Bones
	for _, bones := range root["bones"].([]interface{}) {
		boneMap := bones.(map[string]interface{})
		boneName := boneMap["name"].(string)
		var boneParent *BoneData
		if parentName, ok := boneMap["parent"].(string); ok {
			_, boneParent = skeletonData.findBone(parentName)
			if boneParent == nil {
				return nil, errors.New("Parent bone not found: " + parentName)
			}
		}

		boneData := NewBoneData(boneName, boneParent)

		if length, ok := boneMap["length"].(float64); ok {
			boneData.Length = float32(length) * scale
		}

		if x, ok := boneMap["x"].(float64); ok {
			boneData.x = float32(x) * scale
		}

		if y, ok := boneMap["y"].(float64); ok {
			boneData.y = float32(y) * scale
		}

		if rotation, ok := boneMap["rotation"].(float64); ok {
			boneData.rotation = float32(rotation)
		}

		boneData.scaleX = 1
		if scaleX, ok := boneMap["scaleX"].(float64); ok {
			boneData.scaleX = float32(scaleX)
		}

		boneData.scaleY = 1
		if scaleY, ok := boneMap["scaleY"].(float64); ok {
			boneData.scaleY = float32(scaleY)
		}

		skeletonData.bones = append(skeletonData.bones, boneData)
	}

	// Slots
	for _, slot := range root["slots"].([]interface{}) {
		slotMap := slot.(map[string]interface{})
		boneName := slotMap["bone"].(string)
		_, boneData := skeletonData.findBone(boneName)
		if boneData == nil {
			return nil, errors.New("Slot bone not found: " + boneName)
		}
		slotData := NewSlotData(slotMap["name"].(string), boneData)

		if color, ok := slotMap["color"].(string); ok {
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

		slotData.attachmentName = slotMap["attachment"].(string)

		skeletonData.slots = append(skeletonData.slots, slotData)
	}

	if skinsMap, ok := root["skins"].(map[string]interface{}); ok {
		for skinName, skinMap := range skinsMap {
			skin := NewSkin(skinName)
			for slotName, slotMap := range skinMap.(map[string]interface{}) {
				slotIndex, _ := skeletonData.findSlot(slotName)
				for atName, atMap := range slotMap.(map[string]interface{}) {
					attachmentMap := atMap.(map[string]interface{})

					if name, ok := attachmentMap["name"].(string); ok {
						atName = name
					}

					attachment := NewAttachment(atName)

					if x, ok := attachmentMap["x"].(float64); ok {
						attachment.X = float32(x) * scale
					}

					if y, ok := attachmentMap["y"].(float64); ok {
						attachment.Y = float32(y) * scale
					}

					if rotation, ok := attachmentMap["rotation"].(float64); ok {
						attachment.Rotation = float32(rotation)
					}

					attachment.ScaleX = 1
					if scaleX, ok := attachmentMap["scaleX"].(float64); ok {
						attachment.ScaleX = float32(scaleX)
					}

					attachment.ScaleY = 1
					if scaleY, ok := attachmentMap["scaleY"].(float64); ok {
						attachment.ScaleY = float32(scaleY)
					}

					attachment.Width = 32
					if width, ok := attachmentMap["width"].(float64); ok {
						attachment.Width = float32(width) * scale
					}

					attachment.Height = 32
					if height, ok := attachmentMap["height"].(float64); ok {
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

		if animsMap, ok := root["animations"].(map[string]interface{}); ok {
			for animName, bonesMap := range animsMap {
				timelines := make([]Timeline, 0)
				duration := float32(0)
				boneMap := bonesMap.(map[string]interface{})
				bones := boneMap["bones"].(map[string]interface{})
				for boneName, timelinesMap := range bones {
					boneIndex, _ := skeletonData.findBone(boneName)
					timelineMap := timelinesMap.(map[string]interface{})
					for timelineType, timelinesData := range timelineMap {
						timelineData := timelinesData.([]interface{})
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
		}
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
