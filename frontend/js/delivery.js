/**
 * PitchDeliveryEngine - Biomechanically grounded 3D MLB Pitching Motion Kinematics
 * 
 * Supports:
 * - 5 Sequential Delivery Phases (Set, Leg Lift, Stride/Cocking, Release, Follow-Through)
 * - Statcast Parameterization (arm_angle 0°-90°, release_extension 5.0'-7.5', pitch_hand R/L)
 * - Archetype slot blending (Submarine, Sidearm, Low 3/4, High 3/4, Overhand, Over-the-Top)
 * - Accurate timing synchronization with ball release hand-off at t = 1.25s
 */

(function (root, factory) {
    if (typeof module === 'object' && module.exports) {
        let three = null;
        try {
            three = require('three');
        } catch (e) {
            three = {
                Vector3: class { constructor(x=0, y=0, z=0) { this.x = x; this.y = y; this.z = z; } copy(v) { this.x = v.x; this.y = v.y; this.z = v.z; return this; } set(x, y, z) { this.x = x; this.y = y; this.z = z; return this; } },
                MathUtils: { degToRad: (d) => (d * Math.PI) / 180.0 }
            };
        }
        module.exports = factory(three);
    } else {
        root.PitchDeliveryEngine = factory(root.THREE);
    }
}(typeof self !== 'undefined' ? self : this, function (THREE) {
    'use strict';

    // Interpolation helpers
    function clamp(val, min, max) {
        return Math.max(min, Math.min(max, val));
    }

    function lerp(a, b, t) {
        return a + (b - a) * t;
    }

    function smoothstep(t) {
        t = clamp(t, 0, 1);
        return t * t * (3 - 2 * t);
    }

    function smootherstep(t) {
        t = clamp(t, 0, 1);
        return t * t * t * (t * (t * 6 - 15) + 10);
    }

    function easeInQuad(t) {
        t = clamp(t, 0, 1);
        return t * t;
    }

    function easeOutQuad(t) {
        t = clamp(t, 0, 1);
        return t * (2 - t);
    }

    function easeInOutCubic(t) {
        t = clamp(t, 0, 1);
        return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
    }

    class PitchDeliveryEngine {
        constructor() {
            // Delivery Timing Constants (in seconds)
            this.TIMINGS = {
                START: 0.0,
                SET_END: 0.35,
                LEG_LIFT_PEAK: 0.70,
                HANDS_BREAK: 0.85,
                STRIDE_FOOT_PLANT: 1.12,
                MAX_COCKING: 1.18,
                RELEASE: 1.25,      // Exact moment of ball detachment into flight trajectory
                DECEL_PEAK: 1.45,
                FOLLOW_THROUGH_END: 1.80
            };

            // Rubber placement anchor
            this.RUBBER_POS = new THREE.Vector3(0, 0.84, 60.5);
            this.mocapClips = {};
        }

        async loadMocapClip(mlbId) {
            if (!mlbId || this.mocapClips[mlbId]) return this.mocapClips[mlbId];
            try {
                const res = await fetch(`/animations/pitchers/${mlbId}.json`);
                if (res.ok) {
                    const data = await res.json();
                    this.mocapClips[mlbId] = data;
                    return data;
                }
            } catch (e) {
                // Ignore missing mocap
            }
            return null;
        }

        /**
         * Get delivery metadata from pitch data or compute reasonable defaults
         */
        parsePitchParams(pitchData) {
            let armAngle = 45.0; // standard 3/4 default
            let isLHP = false;
            let extension = 6.2; // standard MLB stride (ft)
            let releasePos = { x: -1.8, y: 54.3, z: 5.8 };

            if (pitchData) {
                if (typeof pitchData.arm_angle === 'number' && !isNaN(pitchData.arm_angle)) {
                    armAngle = pitchData.arm_angle;
                } else if (pitchData.armAngle) {
                    armAngle = pitchData.armAngle;
                }

                if (typeof pitchData.pitch_hand === 'string') {
                    isLHP = pitchData.pitch_hand.toUpperCase() === 'L';
                } else if (pitchData.isLHP !== undefined) {
                    isLHP = !!pitchData.isLHP;
                }

                if (typeof pitchData.release_extension === 'number' && pitchData.release_extension > 0) {
                    extension = pitchData.release_extension;
                } else if (pitchData.releaseExtension) {
                    extension = pitchData.releaseExtension;
                }

                // If trajectory points exist, deduce start position and hand
                if (Array.isArray(pitchData) && pitchData.length > 0) {
                    const p0 = pitchData[0];
                    releasePos = { x: p0.x, y: p0.y, z: p0.z };
                    extension = clamp(60.5 - p0.y, 4.5, 8.0);
                    if (pitchData.pitch_hand === undefined) {
                        isLHP = p0.x < -0.2; // LHP releases on screen LEFT (x0 < 0)
                    }
                    if (pitchData.arm_angle === undefined) {
                        // Estimate arm angle from release height z0 vs extension
                        // Height 4.0-5.2 -> sidearm/low 3/4 (20-35 deg), 5.5-6.2 -> high 3/4 (45-55 deg), >6.3 -> overhand (>60 deg)
                        const estimatedAngle = (releasePos.z - 4.5) * 25.0 + 30.0;
                        armAngle = clamp(estimatedAngle, 10.0, 75.0);
                    }
                } else if (pitchData.points && pitchData.points.length > 0) {
                    const p0 = pitchData.points[0];
                    releasePos = { x: p0.x, y: p0.y, z: p0.z };
                    extension = clamp(60.5 - p0.y, 4.5, 8.0);
                    if (pitchData.pitch_hand === undefined) {
                        isLHP = p0.x < -0.2;
                    }
                }
            }

            return {
                armAngle: clamp(armAngle, 0.0, 90.0),
                isLHP: isLHP,
                extension: clamp(extension, 4.5, 7.8),
                releasePos: releasePos
            };
        }

        /**
         * Classify arm angle into MLB slot archetypes
         */
        getArmSlotName(angle) {
            if (angle < 15.0) return 'Submarine';
            if (angle < 30.0) return 'Sidearm';
            if (angle < 42.0) return 'Low 3/4';
            if (angle < 54.0) return 'High 3/4';
            if (angle < 65.0) return 'Overhand';
            return 'Over-the-Top';
        }

        /**
         * Applies skeletal kinematics to pitcher character model at time `t` (seconds)
         */
        applyDeliveryPose(pitcher, t, pitchParams = {}) {
            if (!pitcher || !pitcher.bones) return;

            const params = this.parsePitchParams(pitchParams);
            const { armAngle, isLHP, extension } = params;
            const bones = pitcher.bones;

            // Set handedness on model
            pitcher.setHandedness(isLHP);

            // Handedness multiplier: RHP has h = -1.0, LHP has h = 1.0
            const h = isLHP ? 1.0 : -1.0;
            // Normalize arm angle (0 = horizontal sidearm / submarine, 1 = direct vertical overhand)
            const angleRad = (armAngle * Math.PI) / 180.0;
            const slotFactor = clamp(armAngle / 70.0, 0.0, 1.2); // 0.0 for low sidearm, 1.0+ for overhand

            // Delivery phase time anchors
            const T = this.TIMINGS;
            t = clamp(t, 0.0, T.FOLLOW_THROUGH_END);

            // -------------------------------------------------------------
            // 1. ROOT & PELVIS TRANSLATION & ORIENTATION
            // -------------------------------------------------------------
            // Initial position on rubber
            bones.root.position.copy(this.RUBBER_POS);
            bones.root.rotation.set(0, 0, 0);

            let pelvisPosX = 0;
            let pelvisPosY = 3.25;
            let pelvisPosZ = 0;
            let pelvisRotY = (Math.PI * 0.48) * h; // Sideways stretch setup facing dugout
            let pelvisRotX = 0;
            let pelvisRotZ = 0;

            // Stride progress along Z (towards home plate at -Z)
            const maxStrideDist = extension * 0.72; // Pitcher pelvis advances ~72% of total extension

            if (t <= T.SET_END) {
                // Phase 1: Set Position - Still, slight breathing sway
                const setP = t / T.SET_END;
                pelvisPosY = 3.25 + Math.sin(setP * Math.PI) * 0.02;
                pelvisRotY = (Math.PI * 0.48 + Math.sin(setP * Math.PI) * 0.02) * h;
            } else if (t <= T.LEG_LIFT_PEAK) {
                // Phase 2A: Leg Lift - Weight gathers on back pivot foot
                const p = (t - T.SET_END) / (T.LEG_LIFT_PEAK - T.SET_END);
                const sp = smootherstep(p);

                pelvisPosY = 3.25 + sp * 0.12; // Slight rise on back leg
                pelvisPosZ = sp * 0.15;        // Gather slightly back
                pelvisRotY = (Math.PI * 0.48 + sp * 0.18) * h; // Coil back showing numbers to batter
                pelvisRotZ = -sp * 0.08 * h;   // Balance lean over rubber
            } else if (t <= T.STRIDE_FOOT_PLANT) {
                // Phase 2B & 3: Stride towards home plate
                const p = (t - T.LEG_LIFT_PEAK) / (T.STRIDE_FOOT_PLANT - T.LEG_LIFT_PEAK);
                const sp = easeInOutCubic(p);

                pelvisPosY = 3.37 - sp * 0.45; // Center of gravity lowers into powerful stride
                pelvisPosZ = 0.15 - sp * (maxStrideDist + 0.15); // Advancing forward along -Z
                pelvisPosX = -sp * 0.25 * h;   // Slight cross-body stride step

                // Pelvis opens towards home plate as stride lands
                pelvisRotY = lerp((Math.PI * 0.66) * h, (Math.PI * 0.15) * h, smoothstep(p));
                pelvisRotX = sp * 0.15; // Forward athletic hinge
                pelvisRotZ = (0.08 - sp * 0.12) * h;
            } else if (t <= T.RELEASE) {
                // Phase 4: Explosive Chest & Hip Rotation to Release
                const p = (t - T.STRIDE_FOOT_PLANT) / (T.RELEASE - T.STRIDE_FOOT_PLANT);
                const sp = easeOutQuad(p);

                pelvisPosY = 2.92 - sp * 0.08;
                pelvisPosZ = -maxStrideDist - sp * 0.35;
                pelvisPosX = (-0.25 - sp * 0.1) * h;

                // Pelvis square to target
                pelvisRotY = lerp((Math.PI * 0.15) * h, 0.0, sp);
                pelvisRotX = 0.15 + sp * 0.22;
                pelvisRotZ = -sp * 0.08 * h;
            } else {
                // Phase 5: Follow-Through & Energy Dissipation
                const p = (t - T.RELEASE) / (T.FOLLOW_THROUGH_END - T.RELEASE);
                const sp = easeOutQuad(p);

                pelvisPosY = 2.84 - sp * 0.25;
                pelvisPosZ = (-maxStrideDist - 0.35) - sp * 0.55;
                pelvisPosX = (-0.35 - sp * 0.35) * h;

                // Pelvis over-rotates slightly to glove side during deceleration
                pelvisRotY = (-sp * 0.35) * h;
                pelvisRotX = 0.37 + sp * 0.25; // Deep forward trunk flexion
                pelvisRotZ = (-0.08 - sp * 0.22) * h;
            }

            bones.pelvis.position.set(pelvisPosX, pelvisPosY, pelvisPosZ);
            bones.pelvis.rotation.set(pelvisRotX, pelvisRotY, pelvisRotZ);

            // -------------------------------------------------------------
            // 2. SPINE & CHEST (Torso Rotation & Lateral Tilt)
            // -------------------------------------------------------------
            // Arm slot tilt: submarine/sidearm tilts heavily sideways towards throwing arm, overhand stays vertical
            const lateralTilt = lerp(0.48, -0.15, slotFactor) * h;

            let spineRotX = 0;
            let spineRotY = 0;
            let spineRotZ = 0;
            let chestRotX = 0;
            let chestRotY = 0;
            let chestRotZ = 0;

            if (t <= T.SET_END) {
                // Set position: chest square to pelvis, hands joined
                spineRotX = 0.05;
                spineRotY = 0;
                spineRotZ = 0;
            } else if (t <= T.LEG_LIFT_PEAK) {
                const p = (t - T.SET_END) / (T.LEG_LIFT_PEAK - T.SET_END);
                const sp = smootherstep(p);

                spineRotY = sp * 0.15 * h; // Counter-rotation
                chestRotY = sp * 0.15 * h;
                spineRotX = -sp * 0.06;
            } else if (t <= T.STRIDE_FOOT_PLANT) {
                const p = (t - T.LEG_LIFT_PEAK) / (T.STRIDE_FOOT_PLANT - T.LEG_LIFT_PEAK);
                const sp = smoothstep(p);

                // Torso stays closed while pelvis opens (hip-shoulder separation!)
                spineRotY = lerp(0.15 * h, 0.40 * h, sp);
                chestRotY = lerp(0.15 * h, 0.45 * h, sp);
                spineRotZ = lerp(0, lateralTilt * 0.4, sp);
                chestRotZ = lerp(0, lateralTilt * 0.6, sp);
                spineRotX = sp * 0.12;
            } else if (t <= T.RELEASE) {
                const p = (t - T.STRIDE_FOOT_PLANT) / (T.RELEASE - T.STRIDE_FOOT_PLANT);
                const sp = easeInOutCubic(p);

                // Explosive torso uncoiling towards home plate
                spineRotY = lerp(0.40 * h, -0.15 * h, sp);
                chestRotY = lerp(0.45 * h, -0.25 * h, sp);
                spineRotZ = lerp(lateralTilt * 0.4, lateralTilt, sp);
                chestRotZ = lerp(lateralTilt * 0.6, lateralTilt * 1.1, sp);
                spineRotX = lerp(0.12, 0.48, sp); // Powerful trunk flexion
                chestRotX = lerp(0.0, 0.35, sp);
            } else {
                // Follow-through trunk flexion & glove side rotation
                const p = (t - T.RELEASE) / (T.FOLLOW_THROUGH_END - T.RELEASE);
                const sp = easeOutQuad(p);

                spineRotY = lerp(-0.15 * h, -0.55 * h, sp);
                chestRotY = lerp(-0.25 * h, -0.65 * h, sp);
                spineRotZ = lerp(lateralTilt, lateralTilt * 1.25, sp);
                chestRotZ = lerp(lateralTilt * 1.1, lateralTilt * 1.35, sp);
                spineRotX = lerp(0.48, 0.85, sp);
                chestRotX = lerp(0.35, 0.45, sp);
            }

            bones.spine.rotation.set(spineRotX, spineRotY, spineRotZ);
            bones.chest.rotation.set(chestRotX, chestRotY, chestRotZ);

            // -------------------------------------------------------------
            // 3. HEAD & NECK (Always locked on home plate target)
            // -------------------------------------------------------------
            // Counteract torso rotation so pitcher eyes remain focused on catcher
            const totalTorsoY = pelvisRotY + spineRotY + chestRotY;
            const totalTorsoX = pelvisRotX + spineRotX + chestRotX;
            const totalTorsoZ = pelvisRotZ + spineRotZ + chestRotZ;

            bones.neck.rotation.set(-totalTorsoX * 0.5 + 0.1, -totalTorsoY * 0.85, -totalTorsoZ * 0.7);
            bones.head.rotation.set(-totalTorsoX * 0.4, -totalTorsoY * 0.2, -totalTorsoZ * 0.3);

            // -------------------------------------------------------------
            // 4. LEGS KINEMATICS (Stride leg & Pivot leg)
            // -------------------------------------------------------------
            // For RHP: Left leg is stride leg, Right leg is rubber pivot leg
            // For LHP: Right leg is stride leg, Left leg is rubber pivot leg
            const strideThigh = isLHP ? bones.thighR : bones.thighL;
            const strideKnee = isLHP ? bones.kneeR : bones.kneeL;
            const strideFoot = isLHP ? bones.footR : bones.footL;

            const pivotThigh = isLHP ? bones.thighL : bones.thighR;
            const pivotKnee = isLHP ? bones.kneeL : bones.kneeR;
            const pivotFoot = isLHP ? bones.footL : bones.footR;

            if (t <= T.SET_END) {
                // Set Position: Legs standing upright, slight knee flex
                strideThigh.rotation.set(0.1, 0.05 * h, -0.05 * h);
                strideKnee.rotation.set(0.15, 0, 0);
                strideFoot.rotation.set(-0.25, 0, 0);

                pivotThigh.rotation.set(0.05, -0.05 * h, 0.05 * h);
                pivotKnee.rotation.set(0.12, 0, 0);
                pivotFoot.rotation.set(-0.17, 0, 0);
            } else if (t <= T.LEG_LIFT_PEAK) {
                // Phase 2: High Leg Kick!
                const p = (t - T.SET_END) / (T.LEG_LIFT_PEAK - T.SET_END);
                const sp = smootherstep(p);

                // Lead thigh flexes up high (~85 degrees) and adducts inwards
                strideThigh.rotation.set(sp * 1.55, sp * 0.45 * h, -sp * 0.35 * h);
                strideKnee.rotation.set(-sp * 1.65, 0, 0); // Knee flexed
                strideFoot.rotation.set(sp * 0.35, 0, 0);

                // Pivot leg solid on rubber, knee flexes slightly to absorb balance
                pivotThigh.rotation.set(-sp * 0.15, -sp * 0.10 * h, sp * 0.12 * h);
                pivotKnee.rotation.set(sp * 0.25, 0, 0);
                pivotFoot.rotation.set(-sp * 0.10, 0, 0);
            } else if (t <= T.STRIDE_FOOT_PLANT) {
                // Phase 3: Leg Drives Forward & Plants
                const p = (t - T.LEG_LIFT_PEAK) / (T.STRIDE_FOOT_PLANT - T.LEG_LIFT_PEAK);
                const sp = easeInOutCubic(p);

                // Stride leg kicks out and extends to plant foot firmly
                strideThigh.rotation.set(
                    lerp(1.55, -0.75, sp),
                    lerp(0.45 * h, -0.15 * h, sp),
                    lerp(-0.35 * h, 0.10 * h, sp)
                );
                strideKnee.rotation.set(lerp(-1.65, 0.65, sp), 0, 0);
                strideFoot.rotation.set(lerp(0.35, 0.10, sp), 0, 0);

                // Pivot leg pushes explosively off the rubber
                pivotThigh.rotation.set(
                    lerp(-0.15, 0.65, sp),
                    lerp(-0.10 * h, 0.35 * h, sp),
                    lerp(0.12 * h, -0.20 * h, sp)
                );
                pivotKnee.rotation.set(lerp(0.25, 0.75, sp), 0, 0);
                pivotFoot.rotation.set(lerp(-0.10, 0.55, sp), 0, 0);
            } else if (t <= T.RELEASE) {
                // Phase 4: Lead Leg Braces (Firm front leg block for energy transfer)
                const p = (t - T.STRIDE_FOOT_PLANT) / (T.RELEASE - T.STRIDE_FOOT_PLANT);
                const sp = easeOutQuad(p);

                strideThigh.rotation.set(lerp(-0.75, -0.60, sp), -0.15 * h, 0.10 * h);
                strideKnee.rotation.set(lerp(0.65, 0.40, sp), 0, 0); // Lead knee extends/braces
                strideFoot.rotation.set(0.10, 0, 0);

                // Trail leg peels off rubber and begins forward swing
                pivotThigh.rotation.set(lerp(0.65, 1.15, sp), 0.35 * h, -0.20 * h);
                pivotKnee.rotation.set(lerp(0.75, 1.10, sp), 0, 0);
                pivotFoot.rotation.set(0.55, 0, 0);
            } else {
                // Phase 5: Trail Leg Swings Over Rubber into Athletic Landing
                const p = (t - T.RELEASE) / (T.FOLLOW_THROUGH_END - T.RELEASE);
                const sp = easeOutQuad(p);

                strideThigh.rotation.set(lerp(-0.60, -0.45, sp), -0.15 * h, 0.10 * h);
                strideKnee.rotation.set(lerp(0.40, 0.55, sp), 0, 0);
                strideFoot.rotation.set(0.10, 0, 0);

                // Back leg kicks high in the air and lands
                pivotThigh.rotation.set(lerp(1.15, 0.45, sp), lerp(0.35 * h, -0.25 * h, sp), -0.35 * h);
                pivotKnee.rotation.set(lerp(1.10, 1.45, sp), 0, 0);
                pivotFoot.rotation.set(lerp(0.55, 0.20, sp), 0, 0);
            }

            // -------------------------------------------------------------
            // 5. THROWING ARM & GLOVE ARM KINEMATICS
            // -------------------------------------------------------------
            const throwShoulder = isLHP ? bones.shoulderL : bones.shoulderR;
            const throwUpperArm = isLHP ? bones.upperArmL : bones.upperArmR;
            const throwElbow = isLHP ? bones.elbowL : bones.elbowR;
            const throwForearm = isLHP ? bones.forearmL : bones.forearmR;
            const throwWrist = isLHP ? bones.wristL : bones.wristR;

            const gloveShoulder = isLHP ? bones.shoulderR : bones.shoulderL;
            const gloveUpperArm = isLHP ? bones.upperArmR : bones.upperArmL;
            const gloveElbow = isLHP ? bones.elbowR : bones.elbowL;
            const gloveForearm = isLHP ? bones.forearmR : bones.forearmL;
            const gloveWrist = isLHP ? bones.wristR : bones.wristL;

            // Throwing shoulder abduction angle driven by Statcast arm_angle
            // armAngle: 0 deg (submarine/horizontal) -> 90 deg (vertical overhand)
            const armAbduction = THREE.MathUtils.degToRad(armAngle);

            if (t <= T.SET_END) {
                // Set: Hands joined together at chest
                throwShoulder.rotation.set(0, 0, 0);
                throwUpperArm.rotation.set(0.65, 0.45 * h, -0.65 * h);
                throwElbow.rotation.set(0, 0, 0);
                throwForearm.rotation.set(1.45, 0, 0.35 * h);
                throwWrist.rotation.set(0.2, 0.1 * h, 0);

                gloveShoulder.rotation.set(0, 0, 0);
                gloveUpperArm.rotation.set(0.65, -0.45 * h, 0.65 * h);
                gloveElbow.rotation.set(0, 0, 0);
                gloveForearm.rotation.set(1.45, 0, -0.35 * h);
                gloveWrist.rotation.set(0.2, -0.1 * h, 0);
            } else if (t <= T.HANDS_BREAK) {
                // Phase 2: Hands Break & Arm Drawback (Scapular Load)
                const p = (t - T.SET_END) / (T.HANDS_BREAK - T.SET_END);
                const sp = smootherstep(p);

                // Throwing arm breaks down and swings back into cocking arc
                throwUpperArm.rotation.set(
                    lerp(0.65, -0.85, sp),
                    lerp(0.45 * h, 0.85 * h, sp),
                    lerp(-0.65 * h, (0.35 + armAbduction * 0.4) * h, sp)
                );
                throwForearm.rotation.set(lerp(1.45, 0.65, sp), 0, lerp(0.35 * h, -0.25 * h, sp));
                throwWrist.rotation.set(0.2, 0.1 * h, 0);

                // Glove arm extends toward target to guide direction
                gloveUpperArm.rotation.set(
                    lerp(0.65, 0.95, sp),
                    lerp(-0.45 * h, -0.65 * h, sp),
                    lerp(0.65 * h, 0.45 * h, sp)
                );
                gloveForearm.rotation.set(lerp(1.45, 0.95, sp), 0, -0.35 * h);
                gloveWrist.rotation.set(0.2, -0.1 * h, 0);
            } else if (t <= T.MAX_COCKING) {
                // Phase 3: Max Arm Layback (Shoulder External Rotation & Arm Slot Setting)
                const p = (t - T.HANDS_BREAK) / (T.MAX_COCKING - T.HANDS_BREAK);
                const sp = easeInOutCubic(p);

                // Set exact arm slot angle
                const targetUpperX = lerp(-0.85, -0.25, sp);
                const targetUpperY = lerp(0.85 * h, 0.35 * h, sp);
                // The Z-rotation sets the shoulder height/angle relative to spine
                const targetUpperZ = lerp((0.35 + armAbduction * 0.4) * h, (armAbduction * 1.15) * h, sp);

                throwUpperArm.rotation.set(targetUpperX, targetUpperY, targetUpperZ);
                // Forearm lays back horizontally (~90 deg elbow flex with full external rotation)
                throwForearm.rotation.set(lerp(0.65, 1.65, sp), 0, lerp(-0.25 * h, -0.85 * h, sp));
                throwWrist.rotation.set(lerp(0.2, -0.35, sp), 0, 0);

                // Glove tucks tightly into front ribcage for rotational acceleration
                gloveUpperArm.rotation.set(lerp(0.95, 0.35, sp), lerp(-0.65 * h, -0.25 * h, sp), lerp(0.45 * h, 0.75 * h, sp));
                gloveForearm.rotation.set(lerp(0.95, 1.75, sp), 0, -0.45 * h);
                gloveWrist.rotation.set(0.2, -0.1 * h, 0);
            } else if (t <= T.RELEASE) {
                // Phase 4: Explosive Whip Forward to Release Point (t = 1.25s)
                const p = (t - T.MAX_COCKING) / (T.RELEASE - T.MAX_COCKING);
                const sp = easeInQuad(p); // Rapid acceleration into ball release

                // Arm unleashes forward towards target
                const releaseUpperX = lerp(-0.25, 1.35, sp);
                const releaseUpperY = lerp(0.35 * h, -0.20 * h, sp);
                const releaseUpperZ = lerp((armAbduction * 1.15) * h, (armAbduction * 0.95) * h, sp);

                throwUpperArm.rotation.set(releaseUpperX, releaseUpperY, releaseUpperZ);
                // Forearm extends rapidly and pronates
                throwForearm.rotation.set(lerp(1.65, 0.15, sp), 0, lerp(-0.85 * h, 0.10 * h, sp));
                throwWrist.rotation.set(lerp(-0.35, 0.45, sp), 0, 0);

                // Glove firmly tucked
                gloveUpperArm.rotation.set(0.35, -0.25 * h, 0.75 * h);
                gloveForearm.rotation.set(1.75, 0, -0.45 * h);
            } else {
                // Phase 5: Arm Deceleration & Follow-Through across opposite knee
                const p = (t - T.RELEASE) / (T.FOLLOW_THROUGH_END - T.RELEASE);
                const sp = easeOutQuad(p);

                // Throwing arm sweeps smoothly across the torso and opposite hip
                throwUpperArm.rotation.set(
                    lerp(1.35, 1.75, sp),
                    lerp(-0.20 * h, -0.95 * h, sp),
                    lerp((armAbduction * 0.95) * h, -0.65 * h, sp)
                );
                throwForearm.rotation.set(lerp(0.15, 1.45, sp), 0, lerp(0.10 * h, 0.65 * h, sp));
                throwWrist.rotation.set(lerp(0.45, 0.15, sp), 0, 0);

                // Glove arm stays stabilized near hip
                gloveUpperArm.rotation.set(0.45, -0.35 * h, 0.65 * h);
                gloveForearm.rotation.set(1.65, 0, -0.45 * h);
            }
        }
    }

    return PitchDeliveryEngine;
}));
