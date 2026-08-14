/**
 * PitcherModel - Procedural 3D Baseball Pitcher Character Rig for Three.js
 * 
 * Creates a lightweight, stylized, fully articulated low-poly MLB pitcher model
 * with hierarchical skeletal pivots for realistic windup, stride, and delivery.
 */

(function (root, factory) {
    if (typeof module === 'object' && module.exports) {
        let three = null;
        try {
            three = require('three');
        } catch (e) {
            three = {
                Group: class { constructor() { this.children = []; this.position = { set() {} }; this.rotation = { set() {} }; } add(c) { this.children.push(c); } },
                Mesh: class { constructor() {} },
                MeshStandardMaterial: class { constructor(opts) { this.color = { setHex() {} }; } },
                CylinderGeometry: class {},
                BoxGeometry: class {},
                SphereGeometry: class {},
                Vector3: class { constructor(x=0,y=0,z=0){this.x=x;this.y=y;this.z=z;} set(x,y,z){this.x=x;this.y=y;this.z=z;return this;} }
            };
        }
        module.exports = factory(three);
    } else {
        root.PitcherModel = factory(root.THREE);
    }
}(typeof self !== 'undefined' ? self : this, function (THREE) {
    'use strict';

    class PitcherModel {
        constructor(options = {}) {
            this.options = Object.assign({
                jerseyColor: 0xf1f5f9,    // Home crisp jersey
                jerseyPinstripe: 0x1e293b,
                pantsColor: 0xe2e8f0,     // Baseball pants
                undershirtColor: 0x1e3a8a,// Navy undershirt
                capColor: 0x1e3a8a,       // Navy cap
                capBillColor: 0x172554,   // Darker bill
                gloveColor: 0x78350f,     // Leather brown mitt
                cleatsColor: 0x0f172a,    // Dark cleats
                skinColor: 0xd4a373,      // Natural skin tone
                beltColor: 0x0f172a,      // Black belt
                socksColor: 0x1e3a8a,     // Navy high socks
                scale: 1.0,
                isLHP: false              // Left-handed pitcher flag
            }, options);

            this.mesh = new THREE.Group();
            this.mesh.name = 'PitcherCharacter';

            // Bone references for kinematic articulation
            this.bones = {};
            this.materials = {};

            this._initMaterials();
            this._buildRig();
            this.setHandedness(this.options.isLHP);
        }

        _initMaterials() {
            this.materials.jersey = new THREE.MeshStandardMaterial({
                color: this.options.jerseyColor,
                roughness: 0.75,
                metalness: 0.05
            });

            this.materials.pants = new THREE.MeshStandardMaterial({
                color: this.options.pantsColor,
                roughness: 0.8,
                metalness: 0.05
            });

            this.materials.undershirt = new THREE.MeshStandardMaterial({
                color: this.options.undershirtColor,
                roughness: 0.7,
                metalness: 0.1
            });

            this.materials.cap = new THREE.MeshStandardMaterial({
                color: this.options.capColor,
                roughness: 0.65,
                metalness: 0.1
            });

            this.materials.capBill = new THREE.MeshStandardMaterial({
                color: this.options.capBillColor,
                roughness: 0.55,
                metalness: 0.15
            });

            this.materials.skin = new THREE.MeshStandardMaterial({
                color: this.options.skinColor,
                roughness: 0.6,
                metalness: 0.0
            });

            this.materials.glove = new THREE.MeshStandardMaterial({
                color: this.options.gloveColor,
                roughness: 0.5,
                metalness: 0.15
            });

            this.materials.cleats = new THREE.MeshStandardMaterial({
                color: this.options.cleatsColor,
                roughness: 0.4,
                metalness: 0.2
            });

            this.materials.belt = new THREE.MeshStandardMaterial({
                color: this.options.beltColor,
                roughness: 0.4,
                metalness: 0.3
            });

            this.materials.socks = new THREE.MeshStandardMaterial({
                color: this.options.socksColor,
                roughness: 0.8,
                metalness: 0.05
            });

            this.materials.stripe = new THREE.MeshStandardMaterial({
                color: 0x3b82f6,
                roughness: 0.5
            });
        }

        _createSegment(geo, mat, shadow = true) {
            const mesh = new THREE.Mesh(geo, mat);
            mesh.castShadow = shadow;
            mesh.receiveShadow = shadow;
            return mesh;
        }

        _buildRig() {
            // Root node seated on mound rubber
            const root = new THREE.Group();
            root.name = 'root';
            this.mesh.add(root);
            this.bones.root = root;

            // Pitcher dimensions (scaled in feet: 1 unit = 1 ft, height ~ 6.25 ft)
            // Pelvis / Hips (Y = 3.25 ft above cleats)
            const pelvis = new THREE.Group();
            pelvis.name = 'pelvis';
            pelvis.position.set(0, 3.25, 0);
            root.add(pelvis);
            this.bones.pelvis = pelvis;

            // Hip / Pants Upper Mesh
            const hipGeo = new THREE.CylinderGeometry(0.55, 0.48, 0.5, 12);
            const hipMesh = this._createSegment(hipGeo, this.materials.pants);
            hipMesh.position.set(0, 0, 0);
            pelvis.add(hipMesh);

            // Belt & Buckle
            const beltGeo = new THREE.CylinderGeometry(0.56, 0.56, 0.1, 12);
            const beltMesh = this._createSegment(beltGeo, this.materials.belt);
            beltMesh.position.set(0, 0.2, 0);
            pelvis.add(beltMesh);

            const buckleGeo = new THREE.BoxGeometry(0.12, 0.08, 0.06);
            const buckleMat = new THREE.MeshStandardMaterial({ color: 0xd1d5db, metalness: 0.8, roughness: 0.2 });
            const buckleMesh = this._createSegment(buckleGeo, buckleMat);
            buckleMesh.position.set(0, 0.2, 0.56);
            pelvis.add(buckleMesh);

            // -------------------------------------------------------------
            // Spine & Torso Hierarchy
            // -------------------------------------------------------------
            const spine = new THREE.Group();
            spine.name = 'spine';
            spine.position.set(0, 0.25, 0);
            pelvis.add(spine);
            this.bones.spine = spine;

            const torsoGeo = new THREE.CylinderGeometry(0.68, 0.54, 0.9, 12);
            const torsoMesh = this._createSegment(torsoGeo, this.materials.jersey);
            torsoMesh.position.set(0, 0.45, 0);
            spine.add(torsoMesh);

            // Chest / Shoulders pivot
            const chest = new THREE.Group();
            chest.name = 'chest';
            chest.position.set(0, 0.9, 0);
            spine.add(chest);
            this.bones.chest = chest;

            const chestGeo = new THREE.BoxGeometry(1.4, 0.45, 0.75);
            const chestMesh = this._createSegment(chestGeo, this.materials.jersey);
            chestMesh.position.set(0, 0.1, 0);
            chest.add(chestMesh);

            // Neck & Head
            const neck = new THREE.Group();
            neck.name = 'neck';
            neck.position.set(0, 0.35, 0);
            chest.add(neck);
            this.bones.neck = neck;

            const neckGeo = new THREE.CylinderGeometry(0.2, 0.22, 0.25, 10);
            const neckMesh = this._createSegment(neckGeo, this.materials.skin);
            neckMesh.position.set(0, 0.1, 0);
            neck.add(neckMesh);

            const head = new THREE.Group();
            head.name = 'head';
            head.position.set(0, 0.22, 0);
            neck.add(head);
            this.bones.head = head;

            // Head shape
            const headGeo = new THREE.SphereGeometry(0.38, 16, 14);
            const headMesh = this._createSegment(headGeo, this.materials.skin);
            headMesh.position.set(0, 0.32, 0);
            headMesh.scale.set(0.9, 1.05, 0.95);
            head.add(headMesh);

            // Baseball Cap (Crown + Visor)
            const capCrownGeo = new THREE.SphereGeometry(0.40, 16, 12, 0, Math.PI * 2, 0, Math.PI * 0.55);
            const capCrownMesh = this._createSegment(capCrownGeo, this.materials.cap);
            capCrownMesh.position.set(0, 0.40, 0);
            capCrownMesh.scale.set(0.95, 0.95, 1.0);
            head.add(capCrownMesh);

            // Cap Button on top
            const capButtonGeo = new THREE.SphereGeometry(0.04, 8, 8);
            const capButtonMesh = this._createSegment(capButtonGeo, this.materials.capBill);
            capButtonMesh.position.set(0, 0.80, 0);
            head.add(capButtonMesh);

            // Cap Bill / Visor (Extending forward along -Z towards home plate)
            const capBillGeo = new THREE.BoxGeometry(0.55, 0.04, 0.42);
            const capBillMesh = this._createSegment(capBillGeo, this.materials.capBill);
            capBillMesh.position.set(0, 0.44, -0.42);
            capBillMesh.rotation.x = -0.15; // slight downward slant
            head.add(capBillMesh);

            // -------------------------------------------------------------
            // Right Arm (Throwing Arm for RHP, Glove Arm for LHP)
            // -------------------------------------------------------------
            const shoulderR = new THREE.Group();
            shoulderR.name = 'shoulderR';
            shoulderR.position.set(0.78, 0.2, 0);
            chest.add(shoulderR);
            this.bones.shoulderR = shoulderR;

            // Shoulder Joint socket
            const shoulderJointGeo = new THREE.SphereGeometry(0.24, 10, 10);
            const shoulderJointRMesh = this._createSegment(shoulderJointGeo, this.materials.undershirt);
            shoulderR.add(shoulderJointRMesh);

            const upperArmR = new THREE.Group();
            upperArmR.name = 'upperArmR';
            upperArmR.position.set(0, 0, 0);
            shoulderR.add(upperArmR);
            this.bones.upperArmR = upperArmR;

            // Upper arm segment (0.95 ft length, hangs down towards -Y)
            const armGeo = new THREE.CylinderGeometry(0.18, 0.15, 0.95, 10);
            const upperArmRMesh = this._createSegment(armGeo, this.materials.undershirt);
            upperArmRMesh.position.set(0, -0.475, 0);
            upperArmR.add(upperArmRMesh);

            // Right Elbow
            const elbowR = new THREE.Group();
            elbowR.name = 'elbowR';
            elbowR.position.set(0, -0.95, 0);
            upperArmR.add(elbowR);
            this.bones.elbowR = elbowR;

            const elbowJointRMesh = this._createSegment(new THREE.SphereGeometry(0.16, 10, 8), this.materials.skin);
            elbowR.add(elbowJointRMesh);

            const forearmR = new THREE.Group();
            forearmR.name = 'forearmR';
            elbowR.add(forearmR);
            this.bones.forearmR = forearmR;

            // Forearm segment (0.9 ft length)
            const forearmGeo = new THREE.CylinderGeometry(0.15, 0.12, 0.9, 10);
            const forearmRMesh = this._createSegment(forearmGeo, this.materials.skin);
            forearmRMesh.position.set(0, -0.45, 0);
            forearmR.add(forearmRMesh);

            // Right Wrist & Hand
            const wristR = new THREE.Group();
            wristR.name = 'wristR';
            wristR.position.set(0, -0.9, 0);
            forearmR.add(wristR);
            this.bones.wristR = wristR;

            const handRGeo = new THREE.BoxGeometry(0.18, 0.25, 0.12);
            const handRMesh = this._createSegment(handRGeo, this.materials.skin);
            handRMesh.position.set(0, -0.12, 0);
            wristR.add(handRMesh);

            // Ball Socket (Held in hand until release)
            const ballSocketR = new THREE.Group();
            ballSocketR.name = 'ballSocketR';
            ballSocketR.position.set(0, -0.22, 0.05);
            wristR.add(ballSocketR);
            this.bones.ballSocketR = ballSocketR;

            // Glove mesh placeholder on right hand (used when LHP)
            const gloveR = this._buildGloveMesh();
            gloveR.position.set(0, -0.16, 0.08);
            gloveR.visible = false;
            wristR.add(gloveR);
            this.bones.gloveR = gloveR;

            // -------------------------------------------------------------
            // Left Arm (Glove Arm for RHP, Throwing Arm for LHP)
            // -------------------------------------------------------------
            const shoulderL = new THREE.Group();
            shoulderL.name = 'shoulderL';
            shoulderL.position.set(-0.78, 0.2, 0);
            chest.add(shoulderL);
            this.bones.shoulderL = shoulderL;

            const shoulderJointLMesh = this._createSegment(shoulderJointGeo, this.materials.undershirt);
            shoulderL.add(shoulderJointLMesh);

            const upperArmL = new THREE.Group();
            upperArmL.name = 'upperArmL';
            shoulderL.add(upperArmL);
            this.bones.upperArmL = upperArmL;

            const upperArmLMesh = this._createSegment(armGeo, this.materials.undershirt);
            upperArmLMesh.position.set(0, -0.475, 0);
            upperArmL.add(upperArmLMesh);

            // Left Elbow
            const elbowL = new THREE.Group();
            elbowL.name = 'elbowL';
            elbowL.position.set(0, -0.95, 0);
            upperArmL.add(elbowL);
            this.bones.elbowL = elbowL;

            const elbowJointLMesh = this._createSegment(new THREE.SphereGeometry(0.16, 10, 8), this.materials.skin);
            elbowL.add(elbowJointLMesh);

            const forearmL = new THREE.Group();
            forearmL.name = 'forearmL';
            elbowL.add(forearmL);
            this.bones.forearmL = forearmL;

            const forearmLMesh = this._createSegment(forearmGeo, this.materials.skin);
            forearmLMesh.position.set(0, -0.45, 0);
            forearmL.add(forearmLMesh);

            // Left Wrist & Hand / Glove
            const wristL = new THREE.Group();
            wristL.name = 'wristL';
            wristL.position.set(0, -0.9, 0);
            forearmL.add(wristL);
            this.bones.wristL = wristL;

            const handLGeo = new THREE.BoxGeometry(0.18, 0.25, 0.12);
            const handLMesh = this._createSegment(handLGeo, this.materials.skin);
            handLMesh.position.set(0, -0.12, 0);
            wristL.add(handLMesh);

            // Glove mesh on left hand (default for RHP)
            const gloveL = this._buildGloveMesh();
            gloveL.position.set(0, -0.16, 0.08);
            gloveL.visible = true;
            wristL.add(gloveL);
            this.bones.gloveL = gloveL;

            const ballSocketL = new THREE.Group();
            ballSocketL.name = 'ballSocketL';
            ballSocketL.position.set(0, -0.22, 0.05);
            ballSocketL.visible = false;
            wristL.add(ballSocketL);
            this.bones.ballSocketL = ballSocketL;

            // -------------------------------------------------------------
            // Right Leg (Pivot rubber leg for RHP)
            // -------------------------------------------------------------
            const hipR = new THREE.Group();
            hipR.name = 'hipR';
            hipR.position.set(0.32, -0.2, 0);
            pelvis.add(hipR);
            this.bones.hipR = hipR;

            const thighR = new THREE.Group();
            thighR.name = 'thighR';
            hipR.add(thighR);
            this.bones.thighR = thighR;

            // Thigh (1.4 ft length)
            const thighGeo = new THREE.CylinderGeometry(0.24, 0.20, 1.4, 10);
            const thighRMesh = this._createSegment(thighGeo, this.materials.pants);
            thighRMesh.position.set(0, -0.7, 0);
            thighR.add(thighRMesh);

            // Right Knee
            const kneeR = new THREE.Group();
            kneeR.name = 'kneeR';
            kneeR.position.set(0, -1.4, 0);
            thighR.add(kneeR);
            this.bones.kneeR = kneeR;

            const kneeJointRMesh = this._createSegment(new THREE.SphereGeometry(0.20, 10, 8), this.materials.pants);
            kneeR.add(kneeJointRMesh);

            const shinR = new THREE.Group();
            shinR.name = 'shinR';
            kneeR.add(shinR);
            this.bones.shinR = shinR;

            // Shin / Sock (1.45 ft length)
            const shinGeo = new THREE.CylinderGeometry(0.19, 0.15, 1.45, 10);
            const shinRMesh = this._createSegment(shinGeo, this.materials.socks);
            shinRMesh.position.set(0, -0.725, 0);
            shinR.add(shinRMesh);

            // Right Foot / Cleat
            const footR = new THREE.Group();
            footR.name = 'footR';
            footR.position.set(0, -1.45, 0);
            shinR.add(footR);
            this.bones.footR = footR;

            const cleatGeo = new THREE.BoxGeometry(0.32, 0.22, 0.75);
            const cleatRMesh = this._createSegment(cleatGeo, this.materials.cleats);
            cleatRMesh.position.set(0, -0.11, -0.15);
            footR.add(cleatRMesh);

            // -------------------------------------------------------------
            // Left Leg (Lead stride leg for RHP)
            // -------------------------------------------------------------
            const hipL = new THREE.Group();
            hipL.name = 'hipL';
            hipL.position.set(-0.32, -0.2, 0);
            pelvis.add(hipL);
            this.bones.hipL = hipL;

            const thighL = new THREE.Group();
            thighL.name = 'thighL';
            hipL.add(thighL);
            this.bones.thighL = thighL;

            const thighLMesh = this._createSegment(thighGeo, this.materials.pants);
            thighLMesh.position.set(0, -0.7, 0);
            thighL.add(thighLMesh);
            // Left Knee
            const kneeL = new THREE.Group();
            kneeL.name = 'kneeL';
            kneeL.position.set(0, -1.4, 0);
            thighL.add(kneeL);
            this.bones.kneeL = kneeL;

            const kneeJointLMesh = this._createSegment(new THREE.SphereGeometry(0.20, 10, 8), this.materials.pants);
            kneeL.add(kneeJointLMesh);

            const shinL = new THREE.Group();
            shinL.name = 'shinL';
            kneeL.add(shinL);
            this.bones.shinL = shinL;

            const shinLMesh = this._createSegment(shinGeo, this.materials.socks);
            shinLMesh.position.set(0, -0.725, 0);
            shinL.add(shinLMesh);

            // Left Foot / Cleat
            const footL = new THREE.Group();
            footL.name = 'footL';
            footL.position.set(0, -1.45, 0);
            shinL.add(footL);
            this.bones.footL = footL;

            const cleatLMesh = this._createSegment(cleatGeo, this.materials.cleats);
            cleatLMesh.position.set(0, -0.11, -0.15);
            footL.add(cleatLMesh);
        }

        _buildGloveMesh() {
            const gloveGroup = new THREE.Group();
            gloveGroup.name = 'BaseballGlove';

            // Glove Pocket / Palm
            const palmGeo = new THREE.BoxGeometry(0.38, 0.42, 0.28);
            const palmMesh = this._createSegment(palmGeo, this.materials.glove);
            gloveGroup.add(palmMesh);

            // Glove Webbing
            const webGeo = new THREE.BoxGeometry(0.12, 0.32, 0.35);
            const webMesh = this._createSegment(webGeo, this.materials.glove);
            webMesh.position.set(0.18, 0.05, 0.1);
            webMesh.rotation.y = 0.3;
            gloveGroup.add(webMesh);

            // Glove Fingers Arch
            const fingerGeo = new THREE.CylinderGeometry(0.18, 0.18, 0.38, 8);
            const fingerMesh = this._createSegment(fingerGeo, this.materials.glove);
            fingerMesh.position.set(-0.1, 0.18, 0);
            fingerMesh.rotation.z = -0.2;
            gloveGroup.add(fingerMesh);

            return gloveGroup;
        }

        /**
         * Switch between Right-Handed Pitcher (RHP) and Left-Handed Pitcher (LHP)
         */
        setHandedness(isLHP) {
            this.options.isLHP = !!isLHP;

            if (this.bones.gloveL && this.bones.gloveR) {
                this.bones.gloveL.visible = !this.options.isLHP;
                this.bones.gloveR.visible = this.options.isLHP;
            }
        }

        /**
         * Returns the active throwing hand socket where the baseball attaches during delivery
         */
        getThrowingSocket() {
            return this.options.isLHP ? this.bones.ballSocketL : this.bones.ballSocketR;
        }

        /**
         * Returns world position of throwing hand
         */
        getThrowingHandWorldPosition(targetVec3) {
            const socket = this.getThrowingSocket();
            if (!socket) return targetVec3.set(0, 0, 0);
            socket.updateWorldMatrix(true, false);
            return socket.getWorldPosition(targetVec3);
        }

        /**
         * Reset all bones to zero rotations (neutral T-pose / rest pose)
         */
        resetPose() {
            for (const name in this.bones) {
                const bone = this.bones[name];
                if (bone && bone.rotation) {
                    bone.rotation.set(0, 0, 0);
                }
            }
        }

        /**
         * Set team primary / secondary colors dynamically
         */
        setTeamColors(primaryHex, secondaryHex) {
            if (primaryHex) {
                this.materials.cap.color.setHex(primaryHex);
                this.materials.undershirt.color.setHex(primaryHex);
                this.materials.socks.color.setHex(primaryHex);
            }
            if (secondaryHex) {
                this.materials.capBill.color.setHex(secondaryHex);
            }
        }
    }

    return PitcherModel;
}));
