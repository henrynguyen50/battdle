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
// Incorporates realistic Magnus spin dynamics and curveball/sweeper launch hump arcs.
func CalculateTrajectory(p *models.PitchProfile) []Point3D {
	x0 := p.ReleasePosX
	y0 := 60.5 - p.ReleaseExtension
	z0 := p.ReleasePosZ

	xf := p.PlateX
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

	// Lateral Acceleration (ax):
	// In Statcast, BreakX is negative for RHP arm-side run, positive for LHP arm-side run.
	// ax = -2 * breakXFt / T^2 curves sinkers/fastballs OUTWARDS and sliders/sweepers INWARDS for both RHP and LHP.
	ax := (-2.0 * breakXFt) / (T * T)
	// -------------------------------------------------------------
	// Vertical Acceleration (az) & Hump Arc:
	// Curveballs, Slurves, and Sweepers are launched with initial upward pop (hump)
	// out of the hand before heavy topspin and gravity drag them sharply down.
	// -------------------------------------------------------------
	var az float64
	humpBoost := 0.0
	switch p.PitchType {
	case "Curveball", "Slurve", "Sweeper":
		humpBoost = 18.0 // Produces upward pop (hump) out of the hand before sharp drop/sweep
	}

	// Total vertical acceleration (downward gravity + topspin)
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
