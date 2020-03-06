package scene

import "math"

func NewCamera() *Camera {
	return &Camera{
		x:        0.0,
		y:        0.0,
		z:        0.0,
		angle:    0.0,
		angleCos: 1.0,
		angleSin: 0.0,
		skew:     0.0,
	}
}

type Camera struct {
	x        float32
	y        float32
	z        float32
	angle    float32
	angleCos float32
	angleSin float32
	skew     float32
}

func (c *Camera) X() float32 {
	return c.x
}

func (c *Camera) Y() float32 {
	return c.y
}

func (c *Camera) Z() float32 {
	return c.z
}

func (c *Camera) SetPosition(x, y, z float32) {
	c.x = x
	c.y = y
	c.z = z
}

func (c *Camera) SetRotation(angle float32) {
	c.angle = angle
	c.updateAngleCosSin()
}

func (c *Camera) MoveForward(amount float32) {
	c.x -= c.angleSin * amount
	c.z += c.angleCos * amount
}

func (c *Camera) MoveBackward(amount float32) {
	c.x += c.angleSin * amount
	c.z -= c.angleCos * amount
}

func (c *Camera) MoveLeft(amount float32) {
	c.x -= c.angleCos * amount
	c.z -= c.angleSin * amount
}

func (c *Camera) MoveRight(amount float32) {
	c.x += c.angleCos * amount
	c.z += c.angleSin * amount
}

func (c *Camera) MoveUp(amount float32) {
	c.y -= amount
}

func (c *Camera) MoveDown(amount float32) {
	c.y += amount
}

func (c *Camera) TurnLeft(amount float32) {
	c.angle += amount
	c.updateAngleCosSin()
}

func (c *Camera) TurnRight(amount float32) {
	c.angle -= amount
	c.updateAngleCosSin()
}

func (c *Camera) LookUp(amount float32) {
	c.skew -= amount
}

func (c *Camera) LookDown(amount float32) {
	c.skew += amount
}

func (c *Camera) updateAngleCosSin() {
	c.angleCos = float32(math.Cos(math.Pi * (float64(c.angle) / 180.0)))
	c.angleSin = float32(math.Sin(math.Pi * (float64(c.angle) / 180.0)))
}
