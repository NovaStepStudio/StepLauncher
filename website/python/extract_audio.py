import os
from moviepy import VideoFileClip

mp4_path = os.path.join(os.path.dirname(__file__), "Bad Apple.mp4")
out_dir = os.path.join(os.path.dirname(__file__), "..", "web", "public")
os.makedirs(out_dir, exist_ok=True)
out_path = os.path.join(out_dir, "bad_apple.mp3")

print("Extracting audio...")
clip = VideoFileClip(mp4_path)
clip.audio.write_audiofile(out_path)
clip.close()
size_mb = os.path.getsize(out_path) / 1024 / 1024
print(f"Done. Audio saved to {out_path} ({size_mb:.1f} MB)")
