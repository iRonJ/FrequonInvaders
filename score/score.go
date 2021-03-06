package score

import (
	"fmt"
	"github.com/ArchRobison/Gophetica/nimble"
)

const nLight = 6

type rgb struct {
	r, g, b float32
}

// Colors from left to right
var lightColor [nLight]rgb = [nLight]rgb{
	rgb{1, 0, 1},      // Purple
	rgb{0, .5, 1},     // Blue
	rgb{0, 1, 0},      // Green
	rgb{1, 1, 0},      // Yellow
	rgb{1, 2. / 3, 0}, // Orange
	rgb{1, 0, 0},      // Red
}

var lightStorage []nimble.Pixel
var lightWidth, lightHeight int32 // Width and height of a light in pixels

type state int

// Get linearized slice of pixels for displaying kth light in state s.
func getLight(k int, s state) []nimble.Pixel {
	size := lightHeight * lightWidth
	j := int32(2*k+int(s)) * lightHeight * lightWidth
	return lightStorage[j : j+size]
}

// Init prepares state used by Draw.
func Init(width, height int32) {
	if width < 3 || height < 3 {
		panic(fmt.Sprintf("score.Init: width=%v height=%v\n", width, height))
	}
	lightWidth = width / 6
	lightHeight = height
	lightStorage = make([]nimble.Pixel, nLight*2*lightHeight*lightWidth)
	xScale := 2.0 / float32(lightWidth-2)
	yScale := 2.0 / float32(lightHeight-2)
	xOffset := -.5 * xScale * float32(lightWidth)
	yOffset := -.5 * yScale * float32(lightHeight)
	for k, color := range lightColor {
		for s := state(0); s < state(2); s++ {
			pm := nimble.MakePixMap(lightWidth, lightHeight, getLight(k, s), lightWidth)
			for i := int32(0); i < lightHeight; i++ {
				for j := int32(0); j < lightWidth; j++ {
					x := float32(j)*xScale + xOffset
					y := float32(i)*yScale + yOffset
					factor := 1 - x*x - y*y
					if factor < 0 {
						factor = 0
					}
					if s == 0 {
						// The light is off, so dim the image
						factor *= .25
					}
					pm.SetPixel(j, i, nimble.RGB(color.r*factor, color.g*factor, color.b*factor))
				}
			}
		}
	}
}

// Draw draws the score (as binary lights) on the given PixMap.
func Draw(pm nimble.PixMap, scoreValue int) {
	if pm.Height() != lightHeight {
		panic(fmt.Sprintf("score.Draw: pm.Height()=%v lightHeight=%v\n", pm.Height(), lightHeight))
	}
	for k := range lightColor {
		s := state(scoreValue >> uint(nLight-k-1) & 1)
		src := nimble.MakePixMap(lightWidth, lightHeight, getLight(k, s), lightWidth)
		pm.Copy(lightWidth*int32(k), 0, &src)
	}
}
