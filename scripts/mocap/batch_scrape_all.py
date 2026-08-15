#!/usr/bin/env python3
"""
Multi-Threaded Batch Mocap Scraper & AI Pose Extractor
Scrapes pitch videos from Baseball Savant and extracts 3D skeletal keyframe tracks
for all MLB pitchers in player_metadata.csv.

Usage:
  python3 scripts/mocap/batch_scrape_all.py --limit 30 --workers 4
  python3 scripts/mocap/batch_scrape_all.py --all --workers 4
"""

import os
import sys
import csv
import argparse
import tempfile
import time
from concurrent.futures import ThreadPoolExecutor, as_completed

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
REPO_ROOT = os.path.abspath(os.path.join(SCRIPT_DIR, "../.."))
OUTPUT_DIR = os.path.join(REPO_ROOT, "frontend/animations/pitchers")
METADATA_CSV = os.path.join(REPO_ROOT, "player_metadata.csv")

sys.path.append(SCRIPT_DIR)
from savant_scraper import download_pitcher_video
from pose_extractor import extract_pitcher_mocap_json

def process_single_pitcher(pitcher_info):
    """
    Worker task: downloads video clip and extracts 3D mocap JSON for a pitcher.
    """
    mlb_id = pitcher_info["mlb_id"]
    name = pitcher_info.get("name", str(mlb_id))
    is_lhp = pitcher_info.get("is_lhp", False)
    arm_angle = pitcher_info.get("arm_angle", 45.0)

    json_path = os.path.join(OUTPUT_DIR, f"{mlb_id}.json")
    if os.path.exists(json_path) and os.path.getsize(json_path) > 500:
        return {"mlb_id": mlb_id, "name": name, "status": "skipped", "msg": "already exists"}

    with tempfile.NamedTemporaryFile(suffix=".mp4", delete=False) as tmp_file:
        tmp_video_path = tmp_file.name

    try:
        download_ok = download_pitcher_video(mlb_id, tmp_video_path)
        if not download_ok or not os.path.exists(tmp_video_path) or os.path.getsize(tmp_video_path) < 10000:
            return {"mlb_id": mlb_id, "name": name, "status": "failed", "msg": "video download failed"}

        extract_ok = extract_pitcher_mocap_json(tmp_video_path, json_path, is_lhp=is_lhp, arm_angle=arm_angle)
        if extract_ok and os.path.exists(json_path):
            return {"mlb_id": mlb_id, "name": name, "status": "success", "msg": "mocap extracted"}
        else:
            return {"mlb_id": mlb_id, "name": name, "status": "failed", "msg": "pose tracking failed"}
    except Exception as e:
        return {"mlb_id": mlb_id, "name": name, "status": "error", "msg": str(e)}
    finally:
        if os.path.exists(tmp_video_path):
            try:
                os.remove(tmp_video_path)
            except OSError:
                pass

def main():
    parser = argparse.ArgumentParser(description="Batch Scrape and Extract 3D Pitcher Mocap Animations")
    parser.add_argument("--limit", type=int, default=25, help="Number of pitchers to process (default: 25)")
    parser.add_argument("--all", action="store_true", help="Process all pitchers in player_metadata.csv")
    parser.add_argument("--workers", type=int, default=3, help="Concurrent worker threads (default: 3)")
    args = parser.parse_args()

    os.makedirs(OUTPUT_DIR, exist_ok=True)

    if not os.path.exists(METADATA_CSV):
        print(f"Error: metadata CSV not found at {METADATA_CSV}")
        sys.exit(1)

    pitchers = []
    with open(METADATA_CSV, mode="r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            if row.get("position") == "P" and row.get("player_id"):
                try:
                    pitchers.append({
                        "mlb_id": int(row["player_id"]),
                        "name": row.get("player_name", "").strip(),
                        "is_lhp": False,
                        "arm_angle": 45.0
                    })
                except ValueError:
                    continue

    total_pitchers = len(pitchers)
    target_count = total_pitchers if args.all else min(args.limit, total_pitchers)
    target_pitchers = pitchers[:target_count]

    print(f"============================================================")
    print(f"🚀 Starting Batch Pitcher Mocap Extraction")
    print(f"   Target: {len(target_pitchers)} MLB pitchers | Workers: {args.workers}")
    print(f"============================================================\n")

    start_time = time.time()
    succeeded = 0
    skipped = 0
    failed = 0

    with ThreadPoolExecutor(max_workers=args.workers) as executor:
        futures = {executor.submit(process_single_pitcher, p): p for p in target_pitchers}
        
        for idx, future in enumerate(as_completed(futures), 1):
            p_info = futures[future]
            try:
                res = future.result()
                status = res["status"]
                name = res["name"]
                mlb_id = res["mlb_id"]

                if status == "success":
                    succeeded += 1
                    print(f"[{idx:3d}/{len(target_pitchers)}] ✓ {name} (ID: {mlb_id}) -> Mocap Generated")
                elif status == "skipped":
                    skipped += 1
                    print(f"[{idx:3d}/{len(target_pitchers)}] ⏭ {name} (ID: {mlb_id}) -> Already Cached")
                else:
                    failed += 1
                    print(f"[{idx:3d}/{len(target_pitchers)}] ⚠ {name} (ID: {mlb_id}) -> {res['msg']}")
            except Exception as exc:
                failed += 1
                print(f"[{idx:3d}/{len(target_pitchers)}] ❌ {p_info['name']} -> Exception: {exc}")

    elapsed = time.time() - start_time
    print(f"\n============================================================")
    print(f"🏁 Batch Processing Finished in {elapsed:.1f}s")
    print(f"   ✓ Generated: {succeeded} | ⏭ Cached: {skipped} | ⚠ Failed/Missing: {failed}")
    print(f"   Total Mocap Clips in {OUTPUT_DIR}: {len(os.listdir(OUTPUT_DIR))}")
    print(f"============================================================")

if __name__ == "__main__":
    main()
