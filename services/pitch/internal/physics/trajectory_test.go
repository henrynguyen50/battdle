package physics

import (
	"math"
	"pitchle/shared/models"
	"testing"
)

// TestCalculateTrajectory_Boundaries verifies boundary conditions:
// - The trajectory starts at the release position.
// - The trajectory ends at the plate position.
// - The number of points matches the expected steps (60 steps = 61 points).
func TestCalculateTrajectory_Boundaries(t *testing.T) {
	tests := []struct {
		name     string
		profile  *models.PitchProfile
		expected struct {
			x0, y0, z0 float64
			xf, yf, zf float64
		}
	}{
		{
			name: "Standard Fastball",
			profile: &models.PitchProfile{
				ReleasePosX:      -2.0,
				ReleasePosZ:      6.0,
				ReleaseExtension: 6.5,
				PlateX:           0.5,
				PlateZ:           2.5,
				Velocity:         95.0,
				BreakX:           -1.0,
				BreakZ:           4.0,
			},
			expected: struct {
				x0, y0, z0 float64
				xf, yf, zf float64
			}{
				x0: -2.0,
				y0: 60.5 - 6.5,
				z0: 6.0,
				xf: 0.5,
				yf: 1.417,
				zf: 2.5,
			},
		},
		{
			name: "Slider with extreme break",
			profile: &models.PitchProfile{
				ReleasePosX:      -2.2,
				ReleasePosZ:      5.8,
				ReleaseExtension: 6.0,
				PlateX:           -0.8,
				PlateZ:           1.8,
				Velocity:         85.0,
				BreakX:           8.0,
				BreakZ:           14.0,
			},
			expected: struct {
				x0, y0, z0 float64
				xf, yf, zf float64
			}{
				x0: -2.2,
				y0: 60.5 - 6.0,
				z0: 5.8,
				xf: -0.8,
				yf: 1.417,
				zf: 1.8,
			},
		},
	}

	const epsilon = 1e-9

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			points := CalculateTrajectory(tc.profile)

			// Verify steps
			if len(points) != 61 {
				t.Fatalf("expected 61 trajectory points, got %d", len(points))
			}

			// Verify start point
			start := points[0]
			if start.T != 0.0 {
				t.Errorf("expected start time 0, got %f", start.T)
			}
			if math.Abs(start.X-tc.expected.x0) > epsilon {
				t.Errorf("start X mismatch: expected %f, got %f", tc.expected.x0, start.X)
			}
			if math.Abs(start.Y-tc.expected.y0) > epsilon {
				t.Errorf("start Y mismatch: expected %f, got %f", tc.expected.y0, start.Y)
			}
			if math.Abs(start.Z-tc.expected.z0) > epsilon {
				t.Errorf("start Z mismatch: expected %f, got %f", tc.expected.z0, start.Z)
			}

			// Verify end point
			end := points[60]
			if math.Abs(end.X-tc.expected.xf) > epsilon {
				t.Errorf("end X mismatch: expected %f, got %f", tc.expected.xf, end.X)
			}
			if math.Abs(end.Y-tc.expected.yf) > epsilon {
				t.Errorf("end Y mismatch: expected %f, got %f", tc.expected.yf, end.Y)
			}
			if math.Abs(end.Z-tc.expected.zf) > epsilon {
				t.Errorf("end Z mismatch: expected %f, got %f", tc.expected.zf, end.Z)
			}
		})
	}
}

// TestCalculateTrajectory_FlightTime verifies the flight time calculation.
func TestCalculateTrajectory_FlightTime(t *testing.T) {
	// Normal flight time formula: T = distY / (Velocity * 1.467)
	// where distY = |(60.5 - ReleaseExtension) - 1.417|
	tests := []struct {
		name      string
		profile   *models.PitchProfile
		expectedT float64
	}{
		{
			name: "Normal Fastball 95mph, extension 6.5ft",
			profile: &models.PitchProfile{
				Velocity:         95.0,
				ReleaseExtension: 6.5,
			},
			expectedT: ((60.5 - 6.5) - 1.417) / (95.0 * 1.467),
		},
		{
			name: "Normal Slider 85mph, extension 6.0ft",
			profile: &models.PitchProfile{
				Velocity:         85.0,
				ReleaseExtension: 6.0,
			},
			expectedT: ((60.5 - 6.0) - 1.417) / (85.0 * 1.467),
		},
		{
			name: "Negative Velocity (fallback)",
			profile: &models.PitchProfile{
				Velocity:         -90.0,
				ReleaseExtension: 6.5,
			},
			expectedT: 0.4,
		},
		{
			name: "Zero Velocity (infinite flight time)",
			profile: &models.PitchProfile{
				Velocity:         0.0,
				ReleaseExtension: 6.5,
			},
			expectedT: math.Inf(1),
		},
		{
			name: "Zero Distance approximation (y0 ~= yf)",
			profile: &models.PitchProfile{
				Velocity:         90.0,
				ReleaseExtension: 60.5 - 1.417, // makes y0 = 1.4170000000000007
			},
			expectedT: ((60.5 - (60.5 - 1.417)) - 1.417) / (90.0 * 1.467),
		},
		{
			name: "Negative Distance before absolute value (y0 < yf) (non-fallback)",
			profile: &models.PitchProfile{
				Velocity:         90.0,
				ReleaseExtension: 60.0, // y0 = 0.5, yf = 1.417, distY = |-0.917| = 0.917
			},
			expectedT: 0.917 / (90.0 * 1.467),
		},
	}

	const epsilon = 1e-9

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			points := CalculateTrajectory(tc.profile)
			if len(points) == 0 {
				t.Fatalf("no points returned")
			}
			actualT := points[len(points)-1].T
			if math.Abs(actualT-tc.expectedT) > epsilon {
				t.Errorf("flight time T mismatch: expected %f, got %f", tc.expectedT, actualT)
			}
		})
	}
}

// TestCalculateTrajectory_Kinematics verifies that intermediate trajectory coordinates
// match the expected gravity drop and spin break offsets.
// Specifically, for any point at t = f * T:
// - Deviation in X from a linear path is BreakX * (f^2 - f)
// - Deviation in Z from a linear path is BreakZ * (f - f^2)
func TestCalculateTrajectory_Kinematics(t *testing.T) {
	profile := &models.PitchProfile{
		ReleasePosX:      -2.0,
		ReleasePosZ:      6.0,
		ReleaseExtension: 6.5,
		PlateX:           0.5,
		PlateZ:           2.5,
		Velocity:         95.0,
		BreakX:           -1.5,
		BreakZ:           5.0,
	}

	points := CalculateTrajectory(profile)
	if len(points) != 61 {
		t.Fatalf("expected 61 points, got %d", len(points))
	}

	x0 := profile.ReleasePosX
	xf := profile.PlateX
	z0 := profile.ReleasePosZ
	zf := profile.PlateZ

	const epsilon = 1e-9

	// Test intermediate points at fractions f = 0.25 (idx 15), f = 0.5 (idx 30), f = 0.75 (idx 45)
	fractions := map[int]float64{
		15: 0.25,
		30: 0.50,
		45: 0.75,
	}

	for idx, f := range fractions {
		pt := points[idx]

		// Linear interpolation values
		xLinear := (1-f)*x0 + f*xf
		zLinear := (1-f)*z0 + f*zf

		// Expected deviation
		expectedXOffset := profile.BreakX * (f*f - f)
		expectedZOffset := profile.BreakZ * (f - f*f)

		// Actual deviation
		actualXOffset := pt.X - xLinear
		actualZOffset := pt.Z - zLinear

		if math.Abs(actualXOffset-expectedXOffset) > epsilon {
			t.Errorf("X offset mismatch at f=%f (idx %d): expected %f, got %f", f, idx, expectedXOffset, actualXOffset)
		}
		if math.Abs(actualZOffset-expectedZOffset) > epsilon {
			t.Errorf("Z offset mismatch at f=%f (idx %d): expected %f, got %f", f, idx, expectedZOffset, actualZOffset)
		}
	}
}
