package render

import "github.com/kjkrol/gokg/geom"

// Camera is the visible window into the world: position and size in world
// coordinates. Renderers use it to cull off-screen entities and to
// translate world coordinates into screen coordinates.
type Camera struct {
	X, Y          uint32
	Width, Height uint32
}

func NewCamera(width, height uint32) *Camera {
	return &Camera{Width: width, Height: height}
}

// ToScreen converts world-space coordinates into screen-space, relative to
// the camera's current position.
func (c *Camera) ToScreen(x, y float32) (float32, float32) {
	return x - float32(c.X), y - float32(c.Y)
}

// Visible reports whether box intersects the camera's current viewport.
func (c *Camera) Visible(box geom.AABB[uint32]) bool {
	return box.BottomRight.X > c.X && box.TopLeft.X < c.X+c.Width &&
		box.BottomRight.Y > c.Y && box.TopLeft.Y < c.Y+c.Height
}

// MoveTo repositions the camera's top-left corner in world space.
func (c *Camera) MoveTo(x, y uint32) { c.X, c.Y = x, y }

// Translate shifts the camera by a delta.
func (c *Camera) Translate(dx, dy int32) {
	c.X = uint32(int64(c.X) + int64(dx))
	c.Y = uint32(int64(c.Y) + int64(dy))
}
