package physics

import (
	"pitchle/shared/models"
)

type Point3D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	T float64 `json:"t"`
}

// CalculateTrajectory computes a 3D trajectory from a pitch profile.
// It returns a list of 3D points representing the flight path.
func CalculateTrajectory(p *models.PitchProfile) []Point3D {
	// Release point: y0 is release_extension away from the pitcher rubber (60.5 ft)
	x0 := p.ReleasePosX
	y0 := 60.5 - p.ReleaseExtension
	z0 := p.ReleasePosZ

	// Target point (at the plate): yf is home plate (which is at y = 1.417 ft, which represents home plate front)
	// We'll use yf = 1.417 as specified in the plan
	xf := p.PlateX
	yf := 1.417
	zf := p.PlateZ

	// Flight duration T in seconds:
	// velocity is in mph, convert to ft/s by multiplying by 1.467
	velFtSec := p.Velocity * 1.467
	distY := y0 - yf
	if distY < 0 {
		distY = -distY
	}
	T := distY / velFtSec
	if T <= 0 {
		T = 0.4 // fallback flight duration if extension/velocity is invalid
	}

	// Break forces in feet. The CSV break_x and break_z represent total deviation.
	// We map them to constant accelerations.
	// a_x = 2 * break_x / T^2
	// a_z = -2 * break_z / T^2  (break_z is the drop including gravity, so it accelerates downward)
	ax := (2.0 * p.BreakX) / (T * T)
	az := (-2.0 * p.BreakZ) / (T * T)

	// Solve initial velocities to satisfy boundary conditions at t = T
	// x(T) = x0 + vx0 * T + 0.5 * ax * T^2 = xf  =>  vx0 = (xf - x0 - 0.5 * ax * T^2) / T
	vx0 := (xf - x0 - 0.5*ax*T*T) / T
	// z(T) = z0 + vz0 * T + 0.5 * az * T^2 = zf  =>  vz0 = (zf - z0 - 0.5 * az * T^2) / T
	vz0 := (zf - z0 - 0.5*az*T*T) / T
	// y(T) = y0 + vy0 * T = yf  =>  vy0 = (yf - y0) / T
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
