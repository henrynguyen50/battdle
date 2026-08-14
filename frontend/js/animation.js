/**
 * PitchleAnimation - 3D Three.js Pitch Visualization with Realistic Pitcher Delivery Motion & 3D Ball Path
 * 
 * Features:
 * - 3D Rigged Pitcher Character delivering the ball on the mound
 * - 5-Phase Skeletal Kinematics (Windup, Leg Kick, Stride/Cocking, Release, Follow-Through)
 * - Statcast Parameterized Arm Slot IK (0°-90°), Extension (5.0'-7.5'), LHP/RHP Mirroring
 * - Seamless Release Hand-Off: Ball held during windup, detaches at t = 1.25s into flight trajectory
 * - 3D Glowing Ball Path Trajectory Ribbon & Tube showing flight arc, break, and drop
 * - Multi-Angle POV Switching (RH Batter, LH Batter, Catcher/Umpire, Pitcher Mound)
 * - Variable Playback Speed Controls (1.0x Normal, 0.5x Slow-Mo, 0.25x Super Slow-Mo)
 */

const PitchleAnimation = {
    scene: null,
    camera: null,
    renderer: null,
    baseball: null,
    tracerLine: null,
    ballPathGroup: null,
    ballPathTube: null,
    ballPathCore: null,
    pitcher: null,
    deliveryEngine: null,

    trajectoryPoints: [],
    pitchParams: {
        armAngle: 45.0,
        isLHP: false,
        extension: 6.2,
        releasePos: { x: -1.8, y: 54.3, z: 5.8 }
    },

    isPlaying: false,
    startTime: null,
    playbackSpeed: 1.0,      // 1.0, 0.5, 0.25
    deliveryDuration: 1.25,   // Time in seconds to reach ball release
    flightDuration: 0.40,    // Ball flight duration in seconds
    totalDuration: 1.65,     // deliveryDuration + flightDuration
    followThroughEnd: 1.80,  // Full follow-through finish
    initialized: false,
    currentView: 'rhb',      // 'rhb', 'lhb', 'catcher', 'pitcher'

    init() {
        const watchBtn = document.getElementById('btn-watch');
        const replayBtn = document.getElementById('btn-replay');

        if (watchBtn) {
            watchBtn.addEventListener('click', () => {
                this.handleWatchClick();
            });
        }

        if (replayBtn) {
            replayBtn.addEventListener('click', () => {
                this.startPlayback();
            });
        }

        // Wire Perspective Toggle Buttons (RH Batter, LH Batter, Catcher, Pitcher)
        const viewButtons = document.querySelectorAll('#camera-view-toggle .btn-toggle');
        viewButtons.forEach(btn => {
            btn.addEventListener('click', () => {
                const view = btn.getAttribute('data-view');
                if (view) {
                    this.setCameraView(view);
                }
            });
        });

        // Wire Playback Speed Toggle Buttons (1.0x, 0.5x, 0.25x)
        const speedButtons = document.querySelectorAll('#playback-speed-toggle .btn-toggle');
        speedButtons.forEach(btn => {
            btn.addEventListener('click', () => {
                const speed = parseFloat(btn.getAttribute('data-speed')) || 1.0;
                this.setPlaybackSpeed(speed);
            });
        });

        // Initialize 3D scene immediately
        if (typeof THREE !== 'undefined') {
            this.initThreeJS();
        }
    },

    setPlaybackSpeed(speed) {
        this.playbackSpeed = speed;

        // Update button active state
        document.querySelectorAll('#playback-speed-toggle .btn-toggle').forEach(btn => {
            const btnSpeed = parseFloat(btn.getAttribute('data-speed')) || 1.0;
            if (Math.abs(btnSpeed - speed) < 0.05) {
                btn.classList.add('active');
            } else {
                btn.classList.remove('active');
            }
        });
    },

    setPitchParams(params = {}) {
        if (!this.deliveryEngine) return;
        this.pitchParams = this.deliveryEngine.parsePitchParams(params);

        if (this.pitcher) {
            this.deliveryEngine.applyDeliveryPose(this.pitcher, 0.0, this.pitchParams);
            this.positionBallInHand();
            if (this.renderer && this.scene && this.camera) {
                this.renderer.render(this.scene, this.camera);
            }
        }
    },

    async handleWatchClick() {
        const watchBtn = document.getElementById('btn-watch');
        const replayBtn = document.getElementById('btn-replay');

        if (watchBtn) {
            watchBtn.disabled = true;
            watchBtn.textContent = 'Loading...';
        }

        try {
            const animationData = await window.PitchleAPI.getAnimation();

            // Support both array response and object payload with metadata
            if (Array.isArray(animationData)) {
                this.trajectoryPoints = animationData;
            } else if (animationData && Array.isArray(animationData.points)) {
                this.trajectoryPoints = animationData.points;
            } else if (animationData && Array.isArray(animationData.trajectory)) {
                this.trajectoryPoints = animationData.trajectory;
            } else {
                this.trajectoryPoints = [];
            }

            // Extract pitch delivery parameters (arm angle, hand, extension)
            if (this.deliveryEngine) {
                this.pitchParams = this.deliveryEngine.parsePitchParams(animationData);
            }

            // Calculate flight duration T from last point
            if (this.trajectoryPoints.length > 0) {
                this.flightDuration = this.trajectoryPoints[this.trajectoryPoints.length - 1].t;
                this.totalDuration = this.deliveryDuration + this.flightDuration;
            }

            if (!this.initialized) {
                this.initThreeJS();
            }

            // Build glowing 3D trajectory ball path
            this.buildBallPath();

            // Apply set position for target pitcher params
            if (this.pitcher && this.deliveryEngine) {
                this.deliveryEngine.applyDeliveryPose(this.pitcher, 0.0, this.pitchParams);
                this.positionBallInHand();
            }

            this.startPlayback();

            if (replayBtn) {
                replayBtn.disabled = false;
            }
        } catch (error) {
            alert('Failed to load pitch animation: ' + error.message);
        } finally {
            if (watchBtn) {
                watchBtn.disabled = false;
                watchBtn.textContent = 'Watch Pitch';
            }
        }
    },

    initThreeJS() {
        const container = document.getElementById('canvas-container');
        if (!container) return;

        const width = container.clientWidth || 600;
        const height = container.clientHeight || 350;

        // 1. Create Scene
        this.scene = new THREE.Scene();
        this.scene.background = new THREE.Color(0x0a0f1d);

        // 2. Camera Setup
        this.camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 300);
        this.updateCameraPosition();

        // 3. Renderer Setup
        this.renderer = new THREE.WebGLRenderer({ antialias: true, alpha: false });
        this.renderer.setSize(width, height);
        this.renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
        this.renderer.shadowMap.enabled = true;
        this.renderer.shadowMap.type = THREE.PCFSoftShadowMap;
        container.innerHTML = '';
        container.appendChild(this.renderer.domElement);

        // 4. Lighting Rig
        const ambientLight = new THREE.AmbientLight(0xffffff, 0.55);
        this.scene.add(ambientLight);

        const stadiumLight1 = new THREE.DirectionalLight(0xffffff, 0.85);
        stadiumLight1.position.set(25, 40, 30);
        stadiumLight1.castShadow = true;
        stadiumLight1.shadow.mapSize.width = 1024;
        stadiumLight1.shadow.mapSize.height = 1024;
        stadiumLight1.shadow.camera.near = 1;
        stadiumLight1.shadow.camera.far = 120;
        stadiumLight1.shadow.camera.left = -20;
        stadiumLight1.shadow.camera.right = 20;
        stadiumLight1.shadow.camera.top = 20;
        stadiumLight1.shadow.camera.bottom = -10;
        this.scene.add(stadiumLight1);

        const fillLight = new THREE.DirectionalLight(0x90cdf4, 0.35);
        fillLight.position.set(-25, 20, -15);
        this.scene.add(fillLight);

        // 5. Ground Field (Green grass extending beyond outfield)
        const groundGeo = new THREE.PlaneGeometry(200, 300);
        const groundMat = new THREE.MeshStandardMaterial({ color: 0x1e3f20, roughness: 0.85 });
        const ground = new THREE.Mesh(groundGeo, groundMat);
        ground.rotation.x = -Math.PI / 2;
        ground.position.set(0, 0, 80);
        ground.receiveShadow = true;
        this.scene.add(ground);

        // 6. Home Plate Dirt Circle
        const dirtMat = new THREE.MeshStandardMaterial({ color: 0x6e473b, roughness: 0.9 });
        const hpDirtGeo = new THREE.CylinderGeometry(13.0, 13.0, 0.01, 32);
        const hpDirt = new THREE.Mesh(hpDirtGeo, dirtMat);
        hpDirt.position.set(0, 0.005, 1.417);
        hpDirt.receiveShadow = true;
        this.scene.add(hpDirt);

        // 7. Home Plate (White slab at Z = 1.417)
        const plateGeo = new THREE.BoxGeometry(1.417, 0.02, 1.417);
        const plateMat = new THREE.MeshStandardMaterial({ color: 0xffffff, roughness: 0.4 });
        const plate = new THREE.Mesh(plateGeo, plateMat);
        plate.position.set(0, 0.015, 1.417);
        this.scene.add(plate);

        // 8. Pitcher Mound (18-ft diameter dirt mound sloping up 10 inches, seated flush on field)
        const moundGeo = new THREE.CylinderGeometry(4.5, 9.0, 0.833, 32);
        const mound = new THREE.Mesh(moundGeo, dirtMat);
        mound.position.set(0, 0.4165, 60.5);
        mound.receiveShadow = true;
        this.scene.add(mound);

        // 9. Pitcher's Rubber (24" x 6" white slab on top of mound)
        const rubberGeo = new THREE.BoxGeometry(2.0, 0.05, 0.5);
        const rubberMat = new THREE.MeshStandardMaterial({ color: 0xffffff, roughness: 0.3 });
        const rubber = new THREE.Mesh(rubberGeo, rubberMat);
        rubber.position.set(0, 0.84, 60.5);
        this.scene.add(rubber);

        // 10. Strike Zone Box (Crisp Wireframe Box + Grid)
        // Width: 17 inches = 1.417 ft, Height: 2 ft (centered at Y = 2.5, Z = 1.417)
        const szGeo = new THREE.BoxGeometry(1.417, 2.0, 0.04);
        const szEdgeGeo = new THREE.EdgesGeometry(szGeo);
        const szMat = new THREE.LineBasicMaterial({ color: 0x38bdf8, linewidth: 2 });
        const strikeZone = new THREE.LineSegments(szEdgeGeo, szMat);
        strikeZone.position.set(0, 2.5, 1.417);
        this.scene.add(strikeZone);

        // 11. Baseball Sphere
        const ballGeo = new THREE.SphereGeometry(0.14, 24, 24);
        const ballMat = new THREE.MeshStandardMaterial({
            color: 0xf8fafc,
            roughness: 0.35,
            metalness: 0.05
        });
        this.baseball = new THREE.Mesh(ballGeo, ballMat);
        this.baseball.castShadow = true;
        this.scene.add(this.baseball);

        // 12. Pitcher Character 3D Model & Kinematics Engine
        if (typeof PitcherModel !== 'undefined' && typeof PitchDeliveryEngine !== 'undefined') {
            this.deliveryEngine = new PitchDeliveryEngine();
            this.pitcher = new PitcherModel();
            this.scene.add(this.pitcher.mesh);

            // Set initial position in Set Stance on the mound rubber
            this.deliveryEngine.applyDeliveryPose(this.pitcher, 0.0, this.pitchParams);
            this.positionBallInHand();
        }

        // Resize Listener
        window.addEventListener('resize', () => {
            const w = container.clientWidth;
            const h = container.clientHeight;
            if (w > 0 && h > 0) {
                this.camera.aspect = w / h;
                this.camera.updateProjectionMatrix();
                this.renderer.setSize(w, h);
            }
        });

        this.initialized = true;
        this.renderer.render(this.scene, this.camera);
    },

    positionBallInHand() {
        if (!this.baseball || !this.pitcher) return;

        const tempVec = new THREE.Vector3();
        this.pitcher.getThrowingHandWorldPosition(tempVec);

        if (tempVec.lengthSq() > 0.01) {
            this.baseball.position.copy(tempVec);
        } else {
            // Fallback placement on rubber
            this.baseball.position.set(0, 4.5, 60.5);
        }
        this.baseball.visible = true;
    },

    /**
     * Builds complete 3D glowing Ball Path Trajectory Ribbon & Tube
     */
    buildBallPath() {
        if (!this.scene || this.trajectoryPoints.length < 2) return;

        // Remove old ball path group if exists
        if (this.ballPathGroup) {
            this.scene.remove(this.ballPathGroup);
            this.ballPathGroup.traverse(child => {
                if (child.geometry) child.geometry.dispose();
                if (child.material) {
                    if (Array.isArray(child.material)) {
                        child.material.forEach(m => m.dispose());
                    } else {
                        child.material.dispose();
                    }
                }
            });
            this.ballPathGroup = null;
        }

        this.ballPathGroup = new THREE.Group();
        this.ballPathGroup.name = 'BallPathGroup';

        // 1. Vector3 Points mapped: physics (x, y, z) -> Three.js (x, z, y)
        const curvePoints = this.trajectoryPoints.map(pt => new THREE.Vector3(pt.x, pt.z, pt.y));
        const curve = new THREE.CatmullRomCurve3(curvePoints);

        // 2. 3D Glowing Trajectory Tube
        const tubeSegments = Math.max(64, this.trajectoryPoints.length * 2);
        const tubeGeo = new THREE.TubeGeometry(curve, tubeSegments, 0.042, 8, false);
        const tubeMat = new THREE.MeshStandardMaterial({
            color: 0x38bdf8,
            emissive: 0x0284c7,
            emissiveIntensity: 0.95,
            roughness: 0.25,
            metalness: 0.1,
            transparent: true,
            opacity: 0.88
        });
        this.ballPathTube = new THREE.Mesh(tubeGeo, tubeMat);
        this.ballPathGroup.add(this.ballPathTube);

        // 3. Bright Inner Core Line
        const coreGeo = new THREE.BufferGeometry().setFromPoints(curvePoints);
        const coreMat = new THREE.LineBasicMaterial({
            color: 0xf0f9ff,
            linewidth: 3,
            transparent: true,
            opacity: 0.95
        });
        this.ballPathCore = new THREE.Line(coreGeo, coreMat);
        this.ballPathGroup.add(this.ballPathCore);

        // 4. Release Point Ring Marker (at p0)
        const p0 = curvePoints[0];
        const ringGeo = new THREE.RingGeometry(0.12, 0.22, 24);
        const ringMat = new THREE.MeshBasicMaterial({
            color: 0x38bdf8,
            side: THREE.DoubleSide,
            transparent: true,
            opacity: 0.85
        });
        const releaseRing = new THREE.Mesh(ringGeo, ringMat);
        releaseRing.position.copy(p0);
        this.ballPathGroup.add(releaseRing);

        // 5. Plate Target Point Marker (at pf)
        const pf = curvePoints[curvePoints.length - 1];
        const targetRing = new THREE.Mesh(ringGeo, ringMat);
        targetRing.position.copy(pf);
        this.ballPathGroup.add(targetRing);
        this.ballPathGroup.visible = false; // Hidden until pitch is released
        this.scene.add(this.ballPathGroup);
    },

    startPlayback() {
        if (!this.initialized) return;

        this.isPlaying = true;
        this.startTime = performance.now();

        // Ensure ball path is constructed and hidden during windup
        if (!this.ballPathGroup && this.trajectoryPoints.length >= 2) {
            this.buildBallPath();
        }

        if (this.ballPathGroup) {
            this.ballPathGroup.visible = false;
        }

        // Remove old dynamic tracer line if exists
        if (this.tracerLine) {
            this.scene.remove(this.tracerLine);
            this.tracerLine.geometry.dispose();
            this.tracerLine = null;
        }

        // Dynamic Real-time Glowing Tracer Line
        const totalPointsCount = (this.trajectoryPoints.length > 0 ? this.trajectoryPoints.length : 60) + 10;
        const tracerGeo = new THREE.BufferGeometry();
        const positions = new Float32Array(totalPointsCount * 3);
        tracerGeo.setAttribute('position', new THREE.BufferAttribute(positions, 3));
        tracerGeo.setDrawRange(0, 0);

        const tracerMat = new THREE.LineBasicMaterial({
            color: 0x38bdf8,
            linewidth: 3,
            transparent: true,
            opacity: 0.95
        });
        this.tracerLine = new THREE.Line(tracerGeo, tracerMat);
        this.scene.add(this.tracerLine);

        // Position pitcher and baseball at initial windup pose
        if (this.pitcher && this.deliveryEngine) {
            this.deliveryEngine.applyDeliveryPose(this.pitcher, 0.0, this.pitchParams);
            this.positionBallInHand();
        }

        this.animate();
    },

    animate() {
        if (!this.isPlaying) return;

        const now = performance.now();
        const rawElapsed = (now - this.startTime) / 1000.0;
        const animTime = rawElapsed * this.playbackSpeed;

        const T_release = this.deliveryDuration; // 1.25s
        const T_total = T_release + this.flightDuration; // ~1.65s
        const T_end = Math.max(T_total, this.followThroughEnd); // ~1.80s

        // -------------------------------------------------------------
        // Delivery Phase 1 to 4: Windup, Leg Kick & Arm Cocking (0 -> 1.25s)
        // -------------------------------------------------------------
        if (animTime < T_release) {
            if (this.pitcher && this.deliveryEngine) {
                this.deliveryEngine.applyDeliveryPose(this.pitcher, animTime, this.pitchParams);
                this.positionBallInHand();
            }

            // Hide ball path until pitch is released from hand
            if (this.ballPathGroup) {
                this.ballPathGroup.visible = false;
            }
            // No active dynamic tracer line during windup
            if (this.tracerLine) {
                this.tracerLine.geometry.setDrawRange(0, 0);
            }

            this.renderer.render(this.scene, this.camera);
            if (this.isPlaying) {
                requestAnimationFrame(() => this.animate());
            }
            return;
        }

        // -------------------------------------------------------------
        // Delivery Phase 5 & Flight: Ball Release and Flight (1.25s -> End)
        // -------------------------------------------------------------
        const flightElapsed = animTime - T_release;

        // Animate pitcher follow-through mechanics
        if (this.pitcher && this.deliveryEngine) {
            const followTime = Math.min(animTime, this.followThroughEnd);
            this.deliveryEngine.applyDeliveryPose(this.pitcher, followTime, this.pitchParams);
        }

        if (flightElapsed >= this.flightDuration || this.trajectoryPoints.length === 0) {
            // Ball reached home plate / glove
            if (this.trajectoryPoints.length > 0) {
                const endPt = this.trajectoryPoints[this.trajectoryPoints.length - 1];
                this.baseball.position.set(endPt.x, endPt.z, endPt.y);

                // Final full trajectory tracer draw
                if (this.tracerLine) {
                    const posAttr = this.tracerLine.geometry.attributes.position;
                    this.trajectoryPoints.forEach((pt, i) => {
                        posAttr.setXYZ(i, pt.x, pt.z, pt.y);
                    });
                    this.tracerLine.geometry.setDrawRange(0, this.trajectoryPoints.length);
                    posAttr.needsUpdate = true;
                }
            }

            // Ensure ball path remains visible at end
            if (this.ballPathGroup) {
                this.ballPathGroup.visible = true;
            }

            // Stop playback once both flight and full follow-through complete
            if (animTime >= T_end) {
                this.isPlaying = false;
            }
            this.renderer.render(this.scene, this.camera);
            if (this.isPlaying) {
                requestAnimationFrame(() => this.animate());
            }
            return;
        }

        // Interpolate ball position along trajectory
        const pos = this.getInterpolatedPosition(flightElapsed);
        this.baseball.position.set(pos.x, pos.z, pos.y);

        // Spin ball
        this.baseball.rotation.x += 0.35 * this.playbackSpeed;
        this.baseball.rotation.y += 0.25 * this.playbackSpeed;

        // Update real-time glowing tracer trail
        if (this.tracerLine && this.trajectoryPoints.length > 0) {
            const currentIdx = Math.min(
                Math.floor((flightElapsed / this.flightDuration) * this.trajectoryPoints.length),
                this.trajectoryPoints.length - 1
            );

            const posAttr = this.tracerLine.geometry.attributes.position;
            for (let i = 0; i <= currentIdx; i++) {
                const pt = this.trajectoryPoints[i];
                posAttr.setXYZ(i, pt.x, pt.z, pt.y);
            }
            posAttr.setXYZ(currentIdx + 1, pos.x, pos.z, pos.y);
            this.tracerLine.geometry.setDrawRange(0, currentIdx + 2);
            posAttr.needsUpdate = true;
        }

        if (this.isPlaying) {
            requestAnimationFrame(() => this.animate());
        }
        this.renderer.render(this.scene, this.camera);
    },

    getInterpolatedPosition(time) {
        if (this.trajectoryPoints.length === 0) return { x: 0, y: 0, z: 0 };

        let idx = 0;
        for (let i = 0; i < this.trajectoryPoints.length - 1; i++) {
            if (time >= this.trajectoryPoints[i].t && time <= this.trajectoryPoints[i + 1].t) {
                idx = i;
                break;
            }
        }

        const p0 = this.trajectoryPoints[idx];
        const p1 = this.trajectoryPoints[idx + 1] || p0;

        const tDiff = p1.t - p0.t;
        const ratio = tDiff === 0 ? 0 : (time - p0.t) / tDiff;

        return {
            x: p0.x + (p1.x - p0.x) * ratio,
            y: p0.y + (p1.y - p0.y) * ratio,
            z: p0.z + (p1.z - p0.z) * ratio
        };
    },

    updateCameraPosition() {
        if (!this.camera) return;

        const isLHP = this.pitchParams ? this.pitchParams.isLHP : false;

        if (this.currentView === 'lhb') {
            // Left-Handed Batter's POV (in LH batter box looking out at pitcher on mound)
            this.camera.position.set(-2.4, 3.6, -5.0);
            this.camera.lookAt(new THREE.Vector3(0.0, 3.2, 30.0));
        } else if (this.currentView === 'catcher') {
            // Catcher / Umpire POV (directly behind home plate looking at delivery and zone)
            this.camera.position.set(0.0, 3.6, -6.0);
            this.camera.lookAt(new THREE.Vector3(0.0, 3.2, 30.0));
        } else if (this.currentView === 'pitcher') {
            // Pitcher's Perspective (Mound POV over pitcher's throwing shoulder looking in to Home Plate)
            const shoulderX = isLHP ? -1.8 : 1.8;
            this.camera.position.set(shoulderX, 6.2, 65.0);
            this.camera.lookAt(new THREE.Vector3(0.0, 2.4, 1.417));
        } else {
            // Right-Handed Batter's POV (in RH batter box looking out at pitcher on mound)
            this.camera.position.set(2.4, 3.6, -5.0);
            this.camera.lookAt(new THREE.Vector3(0.0, 3.2, 30.0));
        }

        this.camera.updateProjectionMatrix();
    },

    setCameraView(viewName) {
        this.currentView = viewName;
        this.updateCameraPosition();

        // Update button active state
        document.querySelectorAll('#camera-view-toggle .btn-toggle').forEach(btn => {
            if (btn.getAttribute('data-view') === viewName) {
                btn.classList.add('active');
            } else {
                btn.classList.remove('active');
            }
        });

        if (this.renderer && this.scene && this.camera) {
            this.renderer.render(this.scene, this.camera);
        }
    }
};

// Expose globally
window.PitchleAnimation = PitchleAnimation;

// Initialize
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        PitchleAnimation.init();
    });
} else {
    PitchleAnimation.init();
}
