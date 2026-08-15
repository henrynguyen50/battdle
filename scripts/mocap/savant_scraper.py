#!/usr/bin/env python3
"""
Baseball Savant & MLB FilmRoom Video Scraper
Fetches representative pitch MP4 video clips for any MLB pitcher.
"""

import os
import re
import csv
import json
import html
import urllib.request
import urllib.error
USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

def get_pitcher_video_url(mlb_id):
    """
    Finds a representative pitch video for a given pitcher MLB ID
    by querying Baseball Savant statcast search and game feed APIs.
    """
    search_url = (
        f"https://baseballsavant.mlb.com/statcast_search/csv?all=true"
        f"&hfSea=2024%7C2023%7C&player_type=pitcher&pitchers_lookup%5B%5D={mlb_id}"
        f"&type=details&min_pitches=0"
    )
    
    req = urllib.request.Request(search_url, headers={"User-Agent": USER_AGENT})
    game_pk = None
    
    try:
        with urllib.request.urlopen(req, timeout=12) as resp:
            lines = [line.decode("utf-8", errors="ignore") for line in resp.readlines()]
            if len(lines) < 2:
                return None
            
            reader = csv.reader(lines)
            header = next(reader)
            col_map = {name: idx for idx, name in enumerate(header)}
            
            pk_col = col_map.get("game_pk")
            if pk_col is None:
                return None
            
            for row in reader:
                if len(row) > pk_col and row[pk_col].strip():
                    game_pk = row[pk_col].strip()
                    break
    except Exception as e:
        print(f"Error querying Savant search for MLB ID {mlb_id}: {e}")
        return None

    if not game_pk:
        return None

    # Query game feed for the pitch play_id
    gf_url = f"https://baseballsavant.mlb.com/gf?game_pk={game_pk}"
    gf_req = urllib.request.Request(gf_url, headers={"User-Agent": USER_AGENT})
    play_id = None
    
    try:
        with urllib.request.urlopen(gf_req, timeout=12) as resp:
            data = json.loads(resp.read().decode("utf-8", errors="ignore"))
            all_plays = data.get("team_away", []) + data.get("team_home", []) + data.get("all_plays", [])
            for play in all_plays:
                if play.get("pitcher") == mlb_id and play.get("play_id"):
                    play_id = play.get("play_id")
                    break
    except Exception as e:
        print(f"Error querying game feed {game_pk}: {e}")
        return None
    if not play_id:
        return None

    # Query sporty-videos page to get direct MP4 URL
    sporty_url = f"https://baseballsavant.mlb.com/sporty-videos?playId={play_id}"
    sporty_req = urllib.request.Request(sporty_url, headers={"User-Agent": USER_AGENT})
    
    try:
        with urllib.request.urlopen(sporty_req, timeout=10) as resp:
            raw_html = resp.read().decode("utf-8", errors="ignore")
            unescaped_html = html.unescape(raw_html)
            mp4_matches = re.findall(r'https://sporty-clips\.mlb\.com/[^\s\"\'<>]+\.mp4', unescaped_html)
            if mp4_matches:
                return mp4_matches[0]
            # Fallback pattern
            mp4_matches = re.findall(r'https://[^\s\"\'<>]+\.mp4', unescaped_html)
            if mp4_matches:
                return mp4_matches[0]
    except Exception as e:
        print(f"Error querying sporty-videos for play {play_id}: {e}")
        return None
    return None

def download_pitcher_video(mlb_id, output_path):
    """
    Downloads the sample pitch MP4 to output_path.
    """
    mp4_url = get_pitcher_video_url(mlb_id)
    if not mp4_url:
        return False

    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    req = urllib.request.Request(
        mp4_url, 
        headers={
            "User-Agent": USER_AGENT,
            "Referer": "https://baseballsavant.mlb.com/"
        }
    )
    
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            with open(output_path, "wb") as out_f:
                out_f.write(resp.read())
        return True
    except Exception as e:
        print(f"Error downloading MP4 from {mp4_url}: {e}")
        return False

if __name__ == "__main__":
    import sys
    test_id = int(sys.argv[1]) if len(sys.argv) > 1 else 694973 # Paul Skenes
    print(f"Scraping video for MLB ID {test_id}...")
    url = get_pitcher_video_url(test_id)
    print("Found video URL:", url)
