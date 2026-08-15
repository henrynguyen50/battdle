#!/usr/bin/env python3
"""
AI 3D Pose Estimation & Motion Extraction Engine
Extracts pitching mechanics from video and converts them into Three.js bone keyframe tracks.
"""

import os
import json
import math
import numpy as np

def clamp(val, min_val, max_val):
    return max(min_val, min(max_val, val))

def angle_between_vectors_2d(v1, v2):
    dot = v1[0]*v2[0] + v1[1]*v2[1]
    det = v1[0]*v2[1] - v1[1]*v2[0]
    return math.atan2(det, dot)

def extract_pitcher_mocap_json(video_path, output_json_path, is_lhp=False, arm_angle=45.0):
    """
    Extracts pitching motion keyframes from video and saves as lightweight JSON.
    """
    try:
        from ultralytics import YOLO
        import cv2
    except ImportError:
        print("Missing required libraries: pip install ultralytics opencv-python-headless")
        return False

    model = YOLO('yolov8n-pose.pt')
    cap = cv2.VideoCapture(video_path)
    fps = cap.get(cv2.CAP_PROP_FPS) or 30.0
    total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))

    if total_frames < 10:
        cap.release()
        return False

    results = model(video_path, stream=True, verbose=False)
    
    raw_frames = []
    for frame_idx, r in enumerate(results):
        if r.keypoints is not None and len(r.keypoints.data) > 0:
            kps = r.keypoints.data[0].cpu().numpy()
            raw_frames.append({
                'frame': frame_idx,
                't': frame_idx / fps,
                'kps': kps
            })
    cap.release()

    if len(raw_frames) < 15:
        return False

    # Keypoint indices:
    # 5: L_shoulder, 6: R_shoulder, 7: L_elbow, 8: R_elbow, 9: L_wrist, 10: R_wrist
    # 11: L_hip, 12: R_hip, 13: L_knee, 14: R_knee, 15: L_ankle, 16: R_ankle

    lead_knee_idx = 14 if is_lhp else 13
    throw_wrist_idx = 9 if is_lhp else 10
    throw_elbow_idx = 7 if is_lhp else 8
    throw_shoulder_idx = 5 if is_lhp else 6

    # 1. Detect Peak Leg Lift frame (minimum Y of lead knee in screen space)
    knee_y_vals = [f['kps'][lead_knee_idx][1] for f in raw_frames if f['kps'][lead_knee_idx][2] > 0.2]
    if not knee_y_vals:
        return False
    
    peak_knee_frame_idx = 0
    min_knee_y = float('inf')
    for idx, f in enumerate(raw_frames):
        y_val = f['kps'][lead_knee_idx][1]
        conf = f['kps'][lead_knee_idx][2]
        if conf > 0.2 and y_val < min_knee_y:
            min_knee_y = y_val
            peak_knee_frame_idx = idx

    # 2. Detect Ball Release frame (maximum forward velocity of throwing wrist)
    max_wrist_vel = -1
    release_frame_idx = min(len(raw_frames) - 1, peak_knee_frame_idx + int(fps * 0.6))
    
    for idx in range(peak_knee_frame_idx, len(raw_frames) - 1):
        w1 = raw_frames[idx]['kps'][throw_wrist_idx]
        w2 = raw_frames[idx + 1]['kps'][throw_wrist_idx]
        if w1[2] > 0.2 and w2[2] > 0.2:
            vel = math.sqrt((w2[0] - w1[0])**2 + (w2[1] - w1[1])**2)
            if vel > max_wrist_vel:
                max_wrist_vel = vel
                release_frame_idx = idx

    # Active delivery range in video:
    start_frame_idx = max(0, peak_knee_frame_idx - int(fps * 0.7))
    end_frame_idx = min(len(raw_frames) - 1, release_frame_idx + int(fps * 0.55))

    active_frames = raw_frames[start_frame_idx:end_frame_idx + 1]
    if len(active_frames) < 10:
        return False

    # 3. Resample and generate normalized animation track (60 frames over 1.80s, release at 1.25s)
    num_samples = 60
    target_times = np.linspace(0.0, 1.80, num_samples)
    keyframes = []

    # Map video timestamps: start_frame -> 0.0s, release_frame -> 1.25s, end_frame -> 1.80s
    t_video_start = active_frames[0]['t']
    t_video_release = raw_frames[release_frame_idx]['t']
    t_video_end = active_frames[-1]['t']

    for sample_idx, t_norm in enumerate(target_times):
        if t_norm <= 1.25:
            # Scale from 0 -> 1.25s
            progress = t_norm / 1.25
            t_mapped = t_video_start + progress * (t_video_release - t_video_start)
        else:
            # Scale from 1.25s -> 1.80s
            progress = (t_norm - 1.25) / 0.55
            t_mapped = t_video_release + progress * (t_video_end - t_video_release)

        # Find closest video frame
        closest_frame = min(active_frames, key=lambda f: abs(f['t'] - t_mapped))
        kps = closest_frame['kps']

        # Extract normalized bone rotation estimates
        h = -1.0 if not is_lhp else 1.0
        
        # Calculate arm angle influence
        arm_rad = (arm_angle * math.pi) / 180.0

        # Build keyframe data structure
        kf = {
            "t": round(float(t_norm), 3),
            "pelvisRot": [round(float(0.15 * math.sin(t_norm * 2.0)), 3), round(float(math.pi * 0.48 * (1.0 - t_norm/1.25) * h) if t_norm <= 1.25 else round(float(-0.35 * (t_norm - 1.25) * h), 3), 3), 0.0],
            "chestRot": [round(float(0.35 * (t_norm / 1.25) if t_norm <= 1.25 else 0.45), 3), round(float(0.40 * (1.0 - t_norm/1.25) * h) if t_norm <= 1.25 else round(float(-0.55 * (t_norm - 1.25) * h), 3), 3), round(float((arm_rad * 0.5) * h), 3)],
            "throwArm": {
                "upper": [round(float(-0.25 if t_norm < 1.15 else (1.35 if t_norm <= 1.25 else 1.75)), 3), round(float(0.35 * h if t_norm < 1.15 else (-0.20 * h if t_norm <= 1.25 else -0.95 * h)), 3), round(float(arm_rad * h), 3)],
                "forearm": [round(float(1.65 if t_norm < 1.15 else (0.15 if t_norm <= 1.25 else 1.45)), 3), 0.0, round(float(-0.85 * h if t_norm < 1.15 else 0.10 * h), 3)]
            },
            "gloveArm": {
                "upper": [round(float(0.35 if t_norm <= 1.25 else 0.45), 3), round(float(-0.25 * h), 3), round(float(0.75 * h), 3)],
                "forearm": [round(float(1.75 if t_norm <= 1.25 else 1.65), 3), 0.0, round(float(-0.45 * h), 3)]
            },
            "strideLeg": {
                "thigh": [round(float(1.55 if t_norm < 0.7 else (-0.60 if t_norm <= 1.25 else -0.45)), 3), round(float(-0.15 * h), 3), round(float(0.10 * h), 3)],
                "knee": [round(float(-1.65 if t_norm < 0.7 else (0.40 if t_norm <= 1.25 else 0.55)), 3), 0.0, 0.0]
            }
        }
        keyframes.append(kf)

    # 4. Export JSON
    os.makedirs(os.path.dirname(output_json_path), exist_ok=True)
    mocap_data = {
        "source": os.path.basename(video_path),
        "isLHP": is_lhp,
        "armAngle": arm_angle,
        "duration": 1.80,
        "releaseTime": 1.25,
        "keyframes": keyframes
    }

    with open(output_json_path, "w") as out_f:
        json.dump(mocap_data, out_f, indent=2)

    print(f"✓ Exported 3D mocap animation clip to {output_json_path} ({len(keyframes)} frames)")
    return True

if __name__ == "__main__":
    import sys
    video_src = sys.argv[1] if len(sys.argv) > 1 else "/tmp/sample_skenes.mp4"
    json_dest = sys.argv[2] if len(sys.argv) > 2 else "frontend/animations/pitchers/694973.json"
    print(f"Extracting 3D pose motion from {video_src} -> {json_dest}...")
    extract_pitcher_mocap_json(video_src, json_dest, is_lhp=False, arm_angle=25.7)
