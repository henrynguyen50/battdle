# Pitchle: 3D Pitching Motion & Trajectory Physics Specification
**STATUS: LOCKED & AUTHORITATIVE — DO NOT MODIFY**

---

## 1. 3D Coordinate System & Hitter's Perspective

The 3D visualization is rendered from the **Hitter's / Catcher's Perspective** looking out at the pitcher on the mound ($Z = 60.5\text{ ft}$):

```
                           [ CENTERFIELD / MOUND ]
                                  (Z = 60.5)

           [ SCREEN LEFT: (-X) ]                 [ SCREEN RIGHT: (+X) ]
           ========================================================================
           ⚾ RIGHT-HANDED PITCHERS (RHP)        ⚾ LEFT-HANDED PITCHERS (LHP)
             - (Paul Skenes, José Soriano)         - (Framber Valdez, Chris Sale)
             - Right throwing arm: Screen LEFT     - Left throwing arm: Screen RIGHT
             - Glove arm: Screen RIGHT             - Glove arm: Screen LEFT
             - Releases on Screen LEFT ($X_0 < 0$)   - Releases on Screen RIGHT ($X_0 > 0$)
             - Sinker / Fastball:                  - Sinker / Fastball:
               Fades to the LEFT (arm-side)          Fades to the RIGHT (arm-side)
             - Slider / Sweeper:                   - Slider / Sweeper:
               Breaks to the RIGHT (glove-side)      Breaks to the LEFT (glove-side)
           ========================================================================

                             [ HOME PLATE & STRIKE ZONE ]
                                      (Z = 1.42)
                              [ HITTER'S PERSPECTIVE ]
```

---

## 2. 3D Character Model Joints (`frontend/js/pitcher_model.js`)
- `shoulderR` (Right Arm / RHP Throwing Arm): $X = -0.78\text{ ft}$ (Screen **LEFT**).
- `shoulderL` (Left Arm / RHP Glove Arm): $X = +0.78\text{ ft}$ (Screen **RIGHT**).
- `hipR` (RHP Pivot Leg on Rubber): $X = -0.32\text{ ft}$ (Screen **LEFT**).
- `hipL` (RHP Stride Leg): $X = +0.32\text{ ft}$ (Screen **RIGHT**).

---

## 3. Kinematic Acceleration & Movement Formula (`services/pitch/internal/physics/trajectory.go`)
- Universal acceleration formula for both RHP and LHP:
  $$a_x = \frac{-2.0 \cdot \text{breakXFt}}{T^2}$$
  $$a_z = \frac{2.0 \cdot \text{breakZFt}}{T^2} - \text{humpBoost}$$
- **Movement Outcomes**:
  - **Sinkers / Fastballs / Changeups**: Curve **OUTWARDS** (arm-side run).
  - **Sliders / Sweepers / Cutters**: Curve **INWARDS** (glove-side break across zone).
  - **Curveballs / Slurves / Sweepers**: Feature the **upward pop (hump arc)** in early flight ($Z_{\text{apex}} > Z_0$) before 12-6 hammer drop.
