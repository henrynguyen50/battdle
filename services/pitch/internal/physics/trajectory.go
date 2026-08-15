package physics

import (
	"math"
	"pitchle/shared/models"
)

type Point3D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	T float64 `json:"t"`
}

// CalculateTrajectory computes a 3D trajectory from a pitch profile.
//
// Hitter's Perspective Coordinate Alignment:
// - RHP: Right throwing arm is on screen LEFT (x0 < 0, e.g. -2.46 ft).
// - LHP: Left throwing arm is on screen RIGHT (x0 > 0, e.g. +2.20 ft).
func CalculateTrajectory(p *models.PitchProfile) []Point3D {
	x0 := p.ReleasePosX
	xf := p.PlateX

	y0 := 60.5 - p.ReleaseExtension
	z0 := p.ReleasePosZ
	yf := 1.417
	zf := p.PlateZ

	velFtSec := p.Velocity * 1.467
	distY := y0 - yf
	if distY < 0 {
		distY = -distY
	}
	T := distY / velFtSec
	if T <= 0 {
		T = 0.4 // fallback flight duration if extension/velocity is invalid
	}

	// -------------------------------------------------------------
	// Convert Statcast BreakX and BreakZ from Inches to Feet
	// -------------------------------------------------------------
	breakXFt := p.BreakX
	if math.Abs(breakXFt) > 2.0 {
		breakXFt = breakXFt / 12.0 // Convert inches to feet
	}

	breakZFt := p.BreakZ
	if math.Abs(breakZFt) > 3.5 {
		breakZFt = breakZFt / 12.0 // Convert inches to feet
	}

	// Lateral acceleration (ax):
	// - For RHP (x0 <= 0): ax = -2 * breakXFt / T^2 => BreakX < 0 fades LEFT (-X, arm-side)
	// - For LHP (x0 > 0): ax = 2 * breakXFt / T^2  => BreakX > 0 fades RIGHT (+X, arm-side)
	isRHP := x0 <= 0
	var ax float64
	if isRHP {
		ax = (-2.0 * breakXFt) / (T * T)
	} else {
		ax = (2.0 * breakXFt) / (T * T)
	}

	// -------------------------------------------------------------
	// Vertical Acceleration (az) & Hump Arc:
	// -------------------------------------------------------------
	var az float64
	humpBoost := 0.0
	switch p.PitchType {
	case "Curveball", "Slurve", "Sweeper":
		humpBoost = 18.0 // Produces upward pop (hump) out of the hand
	}

	az = (2.0*breakZFt)/(T*T) - humpBoost

	// Solve initial velocities to satisfy boundary conditions at t = T
	vx0 := (xf - x0 - 0.5*ax*T*T) / T
	vz0 := (zf - z0 - 0.5*az*T*T) / T
	vy0 := (yf - y0) / T

	steps := 60
	points := make([]Point3D, steps+1)
	for i := range steps + 1 {
		t := T * (float64(i) / float64(steps))
		x := x0 + vx0*t + 0.5*ax*t*t
		y := y0 + vy0*t
		z := z0 + vz0*t + 0.5*az*t*t
		points[i] = Point3D{
			X: x,
			Y: y,
			Z: z,
			T: t,
		}
	}

	return points
}
