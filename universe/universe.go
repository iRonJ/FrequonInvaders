package universe

import (
	"fmt"
	"github.com/ArchRobison/FrequonInvaders/math32"
	"math/rand"
)

type Critter struct {
	Sx, Sy    float32    // Position of particle (units=pixels) - frequency on "fourier" view
	vx, vy    float32    // Velocity of particle (units=pixels/sec)
	Amplitude float32    // Between 0 and 1 - horizontal position on "fall" view and amplitude in "fourier" view.  -1 for "self"
	Progress  float32    // Vertical position on "fall" view, scaled [0,1]
	fallRate  float32    // rate of maturation in maturity/sec
	health    healthType // Initially initialHealth.  Subtracted down to 0. Negative values are death sequence values.  Jumps down to finalHealth at end of sequence
	Show      bool       // If true, show in space domain
	id        int8       // Index into pastels
}

type healthType int16

const (
	initialHealth healthType = 0x7FFF
	// FIXME - decouple the death animation from the unverse model.
	// Then a negative Amplitude can denote death and -1 can denote a dying Critter
	deathThreshold healthType = -0x8000
)

const killTime = 0.1 // Time it takes being close to kill

const amplitudeDieTime = 2.0 // Time in sec for Frequon to die at full amplitude.

const MaxCritter = 16 // Maximum allowed critters (including self)

var zooStorage [MaxCritter]Critter

var Zoo []Critter

func Init(width, height int32) {
	xSize, ySize = float32(width), float32(height)
	Zoo = zooStorage[0:1]
	for k := range zooStorage {
		zooStorage[k].id = int8(k)
	}
	// Original sources used (ySize/32) for the square-root of the kill radius.
	// The formula here gives about same answer for widescreen monitors,
	// while adjusting more sensible for other aspect ratios.
	killRadius2 = (xSize * ySize) * ((9. / 16.) / (32 * 32))
}

// Update advances the universe forwards by time interval dt,
// using (selfX,selfY) as the coordinates the player.
func Update(dt float32, selfX, selfY int32) {
	c := Zoo
	if len(c) < 1 {
		panic("universe.Zoo is empty")
	}
	c[0].Sx = float32(selfX)
	c[0].Sy = float32(selfY)
	c[0].Amplitude = -1

	updateLive(dt)
	cullDead()
	tryBirth(dt)
}

var (
	xSize, ySize float32 // Width and height of fourier port, units = pixels
	killRadius2  float32
)

func bounce(sref, vref *float32, limit, dt float32) {
	v := *vref
	s := *sref + v*dt
	if s < 0 || s > limit {
		panic(fmt.Sprintf("s=%v v=%v limit=%v dt=%v\n", s, v, limit, dt))
	}
	for {
		if s < 0 {
			s = -s
		} else if s > limit {
			s = 2*limit - s
		} else {
			break
		}
		v = -v
	}
	*sref = s
	*vref = v
}

// Update state of live aliens
func updateLive(dt float32) {
	x0, y0 := Zoo[0].Sx, Zoo[0].Sy
	for k := 1; k < len(Zoo); k++ {
		c := &Zoo[k]
		// Update S and v
		bounce(&c.Sx, &c.vx, xSize-1, dt)
		bounce(&c.Sy, &c.vy, ySize-1, dt)
		// Update Progress
		c.Progress += c.fallRate * dt
		// Update health
		if c.health > 0 {
			// Healthy alien
			dx := c.Sx - x0
			dy := c.Sy - y0
			if dx*dx+dy*dy <= killRadius2 {
				const killTime = 0.1
				c.health -= healthType(dt * (float32(initialHealth) / killTime))
				if c.health <= 0 {
					c.health = -1 // Transition to death sequence
					// FIXME - play sound here
				}
				c.Show = true
			} else {
				c.Show = false
			}
		} else {
			// Dying alien
			c.health -= 1
			c.Show = true
		}
		// Update amplitude
		if c.health > 0 {
			c.Amplitude = math32.Sqrt(c.Progress)
		} else {
			c.Amplitude -= dt * (1 / amplitudeDieTime)
			if c.Amplitude < 0 {
				// Mark alien as dead
				c.health = deathThreshold
			}
		}
	}
}

func cullDead() {
	for j := 0; j < len(Zoo); j++ {
		if Zoo[j].health > deathThreshold {
			// Surivor
		} else {
			// Cull it by moving id to end and shrinking slice
			k := len(Zoo) - 1
			deadId := Zoo[j].id
			Zoo[j] = Zoo[k]
			Zoo[k].id = deadId
			Zoo = zooStorage[:k]
		}
	}
}

var (
	birthRate   float32 = 1 // units = per second, only an average
	maxLive             = 1
	velocityMax float32 = 1
)

const π = math32.Pi

// Initialize alien Critter to its birth state.
func (c *Critter) initAlien() {
	// Choose random position
	c.Sx = rand.Float32() * xSize
	c.Sy = rand.Float32() * ySize

	// Choose random velocity
	// This peculiar 2D distibution was in the original sources.
	// It was probably an error, but is now traditional in Frequon Invaders.
	// It biases towards diagonal directions
	c.vx = math32.Cos(2*π*rand.Float32()) * velocityMax
	c.vy = math32.Sin(2*π*rand.Float32()) * velocityMax

	c.Amplitude = 0

	// Choose random fall rate.  The 0.256 is derived from the original sources.
	c.Progress = 0
	c.fallRate = (rand.Float32() + 1) * 0.0256

	c.health = initialHealth
	c.Show = false
}

func tryBirth(dt float32) {
	j := len(Zoo)
	if maxLive > len(zooStorage)-1 {
		panic(fmt.Sprintf("birth: maxLive=%v > len(zooStorage)-1=%v\n", maxLive, len(zooStorage)-1))
	}
	nLive := j - 1
	if nLive > maxLive {
		panic(fmt.Sprintf("birth: nLive=%v > maxLive=%v\n", nLive, maxLive))
	}
	if nLive >= maxLive {
		// Already reached populatino limit
		return
	}
	if rand.Float32() > dt*birthRate {
		return
	}
	// Swap in random id
	avail := len(zooStorage) - len(Zoo)
	k := rand.Intn(avail) + j
	zooStorage[j].id, zooStorage[k].id = zooStorage[k].id, zooStorage[j].id
	Zoo = zooStorage[:j+1]

	// Initialize the alien
	Zoo[j].initAlien()
}

// FIXME - try to avoid coupling unverse Model to View this way.
func (c *Critter) ImageIndex() int {
	if c.health >= 0 {
		return 0
	} else {
		return -int(c.health)
	}
}