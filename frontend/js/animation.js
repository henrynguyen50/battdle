const PitchleAnimation = {
    scene: null,
    camera: null,
    renderer: null,
    baseball: null,
    tracerLine: null,
    
    trajectoryPoints: [],
    isPlaying: false,
    startTime: null,
    duration: 0.4, // flight duration in seconds
    isSlowMo: false,
    initialized: false,
    
    init() {
        const watchBtn = document.getElementById('btn-watch');
        const replayBtn = document.getElementById('btn-replay');
        const slowmoCheckbox = document.getElementById('toggle-slowmo');

        if (!watchBtn) return;

        watchBtn.addEventListener('click', () => {
            this.handleWatchClick();
        });

        if (replayBtn) {
            replayBtn.addEventListener('click', () => {
                this.startPlayback();
            });
        }

        if (slowmoCheckbox) {
            slowmoCheckbox.addEventListener('change', (e) => {
                this.isSlowMo = e.target.checked;
            });
        }
    },

    async handleWatchClick() {
        const watchBtn = document.getElementById('btn-watch');
        const replayBtn = document.getElementById('btn-replay');

        watchBtn.disabled = true;
        watchBtn.textContent = 'Loading...';

        try {
            const animationData = await window.PitchleAPI.getAnimation();
            this.trajectoryPoints = animationData;
            
            // Calculate flight duration T from last point
            if (this.trajectoryPoints.length > 0) {
                this.duration = this.trajectoryPoints[this.trajectoryPoints.length - 1].t;
            }

            if (!this.initialized) {
                this.initThreeJS();
            }

            this.startPlayback();
            
            if (replayBtn) {
                replayBtn.disabled = false;
            }
        } catch (error) {
            alert('Failed to load pitch animation: ' + error.message);
        } finally {
            watchBtn.disabled = false;
            watchBtn.textContent = 'Watch Pitch';
        }
    },

    initThreeJS() {
        const container = document.getElementById('canvas-container');
        if (!container) return;

        const width = container.clientWidth;
        const height = container.clientHeight;

        // 1. Create Scene
        this.scene = new THREE.Scene();
        this.scene.background = new THREE.Color(0x0a0f1d);

        // 2. Camera Setup (Batter box POV)
        // Batter eye position: x=1.5, z=3.0 (which maps to 3D Y=3.0), y=0 (which maps to 3D Z=0)
        // Look at: mound (x=0, z=2.5, y=60.5)
        this.camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 100);
        this.camera.position.set(1.5, 3.0, 0.0);
        this.camera.lookAt(new THREE.Vector3(0.0, 2.5, 60.5));

        // 3. Renderer Setup
        this.renderer = new THREE.WebGLRenderer({ antialias: true });
        this.renderer.setSize(width, height);
        this.renderer.shadowMap.enabled = true;
        container.innerHTML = ''; // clear loading text if any
        container.appendChild(this.renderer.domElement);

        // 4. Lights
        const ambientLight = new THREE.AmbientLight(0xffffff, 0.4);
        this.scene.add(ambientLight);

        const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
        dirLight.position.set(10, 20, 10);
        dirLight.castShadow = true;
        this.scene.add(dirLight);

        // 5. Ground (Field)
        const groundGeo = new THREE.PlaneGeometry(100, 100);
        const groundMat = new THREE.MeshStandardMaterial({ color: 0x1e3f20, roughness: 0.8 });
        const ground = new THREE.Mesh(groundGeo, groundMat);
        ground.rotation.x = -Math.PI / 2;
        ground.receiveShadow = true;
        this.scene.add(ground);

        // 6. Home Plate (approx pentagon at Z = 1.417)
        // Simplification: thin white box centered at (0, 0.01, 1.417)
        const plateGeo = new THREE.BoxGeometry(1.417, 0.02, 1.417);
        const plateMat = new THREE.MeshStandardMaterial({ color: 0xffffff, roughness: 0.5 });
        const plate = new THREE.Mesh(plateGeo, plateMat);
        plate.position.set(0, 0.01, 1.417);
        this.scene.add(plate);

        // 7. Pitcher Mound (Raised circle at Z = 60.5)
        const moundGeo = new THREE.CylinderGeometry(5.0, 5.0, 0.83, 32);
        const moundMat = new THREE.MeshStandardMaterial({ color: 0x6e473b, roughness: 0.9 });
        const mound = new THREE.Mesh(moundGeo, moundMat);
        mound.position.set(0, 0.4, 60.5);
        this.scene.add(mound);

        // 8. Strike Zone Box (Wireframe)
        // Width: 17 inches = 1.417 ft
        // Height: 2 ft (1.5 to 3.5 ft above ground)
        const szGeo = new THREE.BoxGeometry(1.417, 2.0, 0.05);
        const szEdgeGeo = new THREE.EdgesGeometry(szGeo);
        const szMat = new THREE.LineBasicMaterial({ color: 0x1f6feb, linewidth: 2 });
        const strikeZone = new THREE.LineSegments(szEdgeGeo, szMat);
        strikeZone.position.set(0, 2.5, 1.417); // centered at Y = 2.5
        this.scene.add(strikeZone);

        // 9. Baseball Sphere
        // Radius: 1.45 inches = 0.12 ft
        const ballGeo = new THREE.SphereGeometry(0.12, 16, 16);
        const ballMat = new THREE.MeshStandardMaterial({ color: 0xffffff, roughness: 0.3 });
        this.baseball = new THREE.Mesh(ballGeo, ballMat);
        this.baseball.castShadow = true;
        this.scene.add(this.baseball);

        // 10. Window Resize Handler
        window.addEventListener('resize', () => {
            const w = container.clientWidth;
            const h = container.clientHeight;
            this.camera.aspect = w / h;
            this.camera.updateProjectionMatrix();
            this.renderer.setSize(w, h);
        });

        this.initialized = true;

        // Render static scene
        this.renderer.render(this.scene, this.camera);
    },

    startPlayback() {
        if (!this.initialized || this.trajectoryPoints.length === 0) return;

        this.isPlaying = true;
        this.startTime = performance.now();

        // Remove old tracer if exists
        if (this.tracerLine) {
            this.scene.remove(this.tracerLine);
        }

        // Initialize tracer trail lines
        const tracerGeo = new THREE.BufferGeometry();
        const positions = new Float32Array(this.trajectoryPoints.length * 3);
        
        // Populate coordinates: mapping physics (x, y, z) -> 3D (x, z, y) -> (X, Y, Z)
        this.trajectoryPoints.forEach((pt, idx) => {
            positions[idx * 3] = pt.x;     // X is X
            positions[idx * 3 + 1] = pt.z; // Y is physics Z (height)
            positions[idx * 3 + 2] = pt.y; // Z is physics Y (distance)
        });
        
        tracerGeo.setAttribute('position', new THREE.BufferAttribute(positions, 3));
        const tracerMat = new THREE.LineBasicMaterial({ color: 0x8b949e, transparent: true, opacity: 0.4 });
        this.tracerLine = new THREE.Line(tracerGeo, tracerMat);
        this.scene.add(this.tracerLine);

        // Position baseball at start
        const startPt = this.trajectoryPoints[0];
        this.baseball.position.set(startPt.x, startPt.z, startPt.y);
        this.baseball.visible = true;

        // Start requestAnimationFrame loop
        this.animate();
    },

    animate() {
        if (!this.isPlaying) return;

        requestAnimationFrame(() => this.animate());

        const now = performance.now();
        let elapsed = (now - this.startTime) / 1000.0; // in seconds

        if (this.isSlowMo) {
            elapsed *= 0.25; // 4x slower
        }

        // Loop checks
        const currentDuration = this.isSlowMo ? this.duration * 4 : this.duration;
        const progressTime = this.isSlowMo ? elapsed * 0.25 : elapsed;

        if (progressTime >= this.duration) {
            // Animation finished
            const endPt = this.trajectoryPoints[this.trajectoryPoints.length - 1];
            this.baseball.position.set(endPt.x, endPt.z, endPt.y);
            this.isPlaying = false;
            this.renderer.render(this.scene, this.camera);
            return;
        }

        // Interpolate baseball position along keyframes
        const pos = this.getInterpolatedPosition(progressTime);
        this.baseball.position.set(pos.x, pos.z, pos.y);

        this.renderer.render(this.scene, this.camera);
    },

    getInterpolatedPosition(time) {
        if (this.trajectoryPoints.length === 0) return { x: 0, y: 0, z: 0 };
        
        // Find interval
        let idx = 0;
        for (let i = 0; i < this.trajectoryPoints.length - 1; i++) {
            if (time >= this.trajectoryPoints[i].t && time <= this.trajectoryPoints[i+1].t) {
                idx = i;
                break;
            }
        }

        const p0 = this.trajectoryPoints[idx];
        const p1 = this.trajectoryPoints[idx + 1];

        // Linear interpolation ratio
        const tDiff = p1.t - p0.t;
        const ratio = tDiff === 0 ? 0 : (time - p0.t) / tDiff;

        return {
            x: p0.x + (p1.x - p0.x) * ratio,
            y: p0.y + (p1.y - p0.y) * ratio,
            z: p0.z + (p1.z - p0.z) * ratio
        };
    }
};

// Initialize
window.addEventListener('load', () => {
    PitchleAnimation.init();
});
