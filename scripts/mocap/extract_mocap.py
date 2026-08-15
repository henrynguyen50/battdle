#!/usr/bin/env python3
"""
Master 3D Pitcher Mocap Extraction CLI
Scrapes video from Baseball Savant and extracts 3D skeletal animation JSON for Three.js.

Usage:
  python3 scripts/mocap/extract_mocap.py --mlb-id 694973
  python3 scripts/mocap/extract_mocap.py --batch 10
  python3 scripts/mocap/extract_mocap.py --all
"""

import os
import sys
import argparse
import tempfile
import csv

from savant_scraper import download_pitcher_video
from pose_extractor import extract_pitcher_mocap_json

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
REPO_ROOT = os.path.abspath(os.path.join(SCRIPT_DIR, "../.."))
OUTPUT_DIR = os.path.join(REPO_ROOT, "frontend/animations/pitchers")
METADATA_CSV = os.path.join(REPO_ROOT, "player_metadata.csv")

def process_pitcher_mocap(mlb_id, is_lhp=False, arm_angle=45.0):
    """
    Downloads video and extracts 3D mocap JSON for a single pitcher.
    """
    json_path = os.path.join(OUTPUT_DIR, f"{mlb_id}.json")
    if os.path.exists(json_path):
        print(f"✓ Animation already exists for MLB ID {mlb_id}: {json_path}")
        return True

    print(f"\n▶ Processing MLB ID {mlb_id} (Arm Angle: {arm_angle}°, LHP: {is_lhp})...")
    with tempfile.NamedTemporaryFile(suffix=".mp4", delete=False) as tmp_video:
        tmp_video_path = tmp_video.name

    try:
        print("  1. Scraping video clip from Baseball Savant / MLB FilmRoom...")
        success = download_pitcher_video(mlb_id, tmp_video_path)
        if not success:
            print(f"  ❌ Failed to download video clip for MLB ID {mlb_id}.")
            return False

        print("  2. Running AI 3D Pose Estimation & Keyframe Interpolation...")
        success = extract_pitcher_mocap_json(tmp_video_path, json_path, is_lhp=is_lhp, arm_angle=arm_angle)
        if success:
            print(f"  🎉 Successfully generated 3D mocap animation: {json_path}")
            return True
        else:
            print(f"  ❌ Failed to extract pose tracks from video.")
            return False
    finally:
        if os.path.exists(tmp_video_path):
            os.remove(tmp_video_path)

def main():
    parser = argparse.ArgumentParser(description="Extract 3D Pitcher Mocap Animations from Baseball Savant Video")
    parser.add_argument("--mlb-id", type=int, help="Single pitcher MLB ID (e.g. 694973 for Paul Skenes)")
    parser.add_argument("--batch", type=int, default=0, help="Number of pitchers to process in batch")
    parser.add_argument("--all", action="store_true", help="Process all MLB pitchers from metadata")
    args = parser.parse_args()

    os.makedirs(OUTPUT_DIR, exist_ok=True)

    if args.mlb_id:
        process_pitcher_mocap(args.mlb_id, is_lhp=False, arm_angle=25.7)
        return

    # Load pitcher metadata
    if not os.path.exists(METADATA_CSV):
        print(f"Metadata file not found: {METADATA_CSV}")
        return

    pitchers = []
    with open(METADATA_CSV, mode="r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            if row.get("position") == "P" and row.get("player_id"):
                try:
                    pitchers.append({
                        "mlb_id": int(row["player_id"]),
                        "name": row.get("player_name", "")
                    })
                except ValueError:
                    continue

    print(f"Loaded {len(pitchers)} MLB pitchers from metadata.")

    limit = len(pitchers) if args.all else (args.batch if args.batch > 0 else 5)
    processed = 0
    succeeded = 0

    for p in pitchers[:limit]:
        res = process_pitcher_mocap(p["mlb_id"])
        processed += 1
        if res:
            succeeded += 1

    print(f"\n==========================================")
    print(f"Mocap Pipeline Complete: {succeeded}/{processed} succeeded.")
    print(f"==========================================")

if __name__ == "__main__":
    main()
