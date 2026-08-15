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

// TestCalculateTrajectory_Kinematics_RHP verifies that for an RHP (ReleasePosX >= 0 mapped to screen LEFT x0 < 0):
// - Sinker (BreakX < 0) curves to the LEFT (midPt.X < linearX).
// - Sweeper (BreakX > 0) curves to the RIGHT (midSweeper.X > linearSweeperX).
func TestCalculateTrajectory_Kinematics_RHP(t *testing.T) {
	rhpSinker := &models.PitchProfile{
		ReleasePosX:      -2.4, // RHP on screen LEFT (< 0)
		ReleasePosZ:      5.5,
		ReleaseExtension: 6.2,
		PlateX:           -0.48,
		PlateZ:           1.65,
		Velocity:         96.4,
		BreakX:           -15.0, // Negative BreakX (arm-side run)
		BreakZ:           -26.0,
	}

	points := CalculateTrajectory(rhpSinker)
	if len(points) != 61 {
		t.Fatalf("expected 61 points, got %d", len(points))
	}

	// Release point must be on screen LEFT (x0 < 0)
	if points[0].X >= 0 {
		t.Errorf("RHP release point must be on screen LEFT: points[0].X=%f", points[0].X)
	}

	midPt := points[30]
	linearX := (points[0].X + points[60].X) / 2.0
	// Sinker curves to screen LEFT (midPt.X < linearX)
	if midPt.X >= linearX {
		t.Errorf("RHP Sinker should curve to the LEFT: midPt.X=%f >= linearX=%f", midPt.X, linearX)
	}
	rhpSweeper := &models.PitchProfile{
		ReleasePosX:      -2.4, // RHP on screen LEFT (< 0)
		ReleasePosZ:      5.5,
		ReleaseExtension: 6.2,
		PlateX:           0.75,
		PlateZ:           1.60,
		Velocity:         84.0,
		BreakX:           14.0, // Positive BreakX (glove-side sweep)
		BreakZ:           -18.0,
	}

	sweeperPoints := CalculateTrajectory(rhpSweeper)
	midSweeper := sweeperPoints[30]
	linearSweeperX := (sweeperPoints[0].X + sweeperPoints[60].X) / 2.0
	// Sweeper curves to screen RIGHT (midSweeper.X > linearSweeperX)
	if midSweeper.X <= linearSweeperX {
		t.Errorf("RHP Sweeper should curve to the RIGHT: midSweeper.X=%f <= linearSweeperX=%f", midSweeper.X, linearSweeperX)
	}
}

// TestCalculateTrajectory_Kinematics_LHP verifies that for an LHP (ReleasePosX < 0):
// - Sinker (BreakX > 0 in Statcast) curves to the RIGHT (midPt.X > linearX).
// - Sweeper (BreakX < 0 in Statcast) curves to the LEFT (midSweeper.X < linearSweeperX).
func TestCalculateTrajectory_Kinematics_LHP(t *testing.T) {
	lhpSinker := &models.PitchProfile{
		ReleasePosX:      -2.2, // LHP on screen LEFT (< 0)
		ReleasePosZ:      5.5,
		ReleaseExtension: 6.0,
		PlateX:           -0.48,
		PlateZ:           1.65,
		Velocity:         94.0,
		BreakX:           15.0, // Positive BreakX for LHP in Statcast (arm-side run)
		BreakZ:           -26.0,
	}

	points := CalculateTrajectory(lhpSinker)
	if points[0].X >= 0 {
		t.Errorf("LHP release point must be on screen LEFT: points[0].X=%f", points[0].X)
	}

	midPt := points[30]
	linearX := (points[0].X + points[60].X) / 2.0
	// Sinker curves to screen RIGHT (midPt.X > linearX)
	if midPt.X <= linearX {
		t.Errorf("LHP Sinker should curve to the RIGHT: midPt.X=%f <= linearX=%f", midPt.X, linearX)
	}

	lhpSweeper := &models.PitchProfile{
		ReleasePosX:      -2.2,
		ReleasePosZ:      5.5,
		ReleaseExtension: 6.0,
		PlateX:           0.55,
		PlateZ:           1.60,
		Velocity:         82.0,
		BreakX:           -14.0, // Negative BreakX for LHP in Statcast (glove-side sweep)
		BreakZ:           -18.0,
	}

	sweeperPoints := CalculateTrajectory(lhpSweeper)
	midSweeper := sweeperPoints[30]
	linearSweeperX := (sweeperPoints[0].X + sweeperPoints[60].X) / 2.0
	// Sweeper curves to screen LEFT (midSweeper.X < linearSweeperX)
	if midSweeper.X >= linearSweeperX {
		t.Errorf("LHP Sweeper should curve to the LEFT: midSweeper.X=%f >= linearSweeperX=%f", midSweeper.X, linearSweeperX)
	}
}
// TestCalculateTrajectory_CurveballHump verifies that a curveball pops upward in early flight (hump arc)
// before crashing down into the lower strike zone.
func TestCalculateTrajectory_CurveballHump(t *testing.T) {
	curveball := &models.PitchProfile{
		PitchType:        "Curveball",
		ReleasePosX:      2.4,
		ReleasePosZ:      5.57,
		ReleaseExtension: 6.2,
		PlateX:           -0.10,
		PlateZ:           1.26,
		Velocity:         81.9,
		BreakX:           12.0,
		BreakZ:           -55.0,
	}

	points := CalculateTrajectory(curveball)
	if len(points) != 61 {
		t.Fatalf("expected 61 points, got %d", len(points))
	}

	// In early flight (around step 8-12), Z should rise above release height z0
	earlyPt := points[10]
	if earlyPt.Z <= curveball.ReleasePosZ {
		t.Errorf("Curveball failed to pop upward (hump): earlyPt.Z=%f <= ReleasePosZ=%f", earlyPt.Z, curveball.ReleasePosZ)
	}

	// At the plate, Z must finish in the lower zone
	platePt := points[60]
	if platePt.Z > 2.0 {
		t.Errorf("Curveball failed to drop through lower zone: platePt.Z=%f > 2.0", platePt.Z)
	}
}

// TestCalculateTrajectory_SweeperHump verifies that a sweeper pops up slightly before sweeping horizontally.
func TestCalculateTrajectory_SweeperHump(t *testing.T) {
	sweeper := &models.PitchProfile{
		PitchType:        "Sweeper",
		ReleasePosX:      2.4,
		ReleasePosZ:      5.57,
		ReleaseExtension: 6.2,
		PlateX:           -0.75,
		PlateZ:           1.50,
		Velocity:         84.9,
		BreakX:           18.0,
		BreakZ:           -40.0,
	}

	points := CalculateTrajectory(sweeper)
	earlyPt := points[8]
	if earlyPt.Z <= sweeper.ReleasePosZ {
		t.Errorf("Sweeper failed to pop upward in early flight: earlyPt.Z=%f <= ReleasePosZ=%f", earlyPt.Z, sweeper.ReleasePosZ)
	}
}
