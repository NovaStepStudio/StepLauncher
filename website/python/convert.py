import cv2
import os
import zipfile

def main():
  assets_dir = os.path.join(os.path.dirname(__file__), "..", "web", "assets", "Media")
  public_dir = os.path.join(os.path.dirname(__file__), "..", "web", "public")
  os.makedirs(assets_dir, exist_ok=True)
  os.makedirs(public_dir, exist_ok=True)

  cap = cv2.VideoCapture("output.mp4")
  txt_path = os.path.join(assets_dir, "frames.txt")

  with open(txt_path, 'w', newline='\n') as file:
    total = 0
    while cap.isOpened():
        ret, frame = cap.read()
        if not ret:
            break

        output = ""
        gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)

        for i in range(72):
            row = ""
            for j in range(128):
                if gray[i][j] > 153:
                    row += ' '
                else:
                    row += '*'
            output += row

        output += '\n'
        file.write(output)
        total += 1
        print(f"\rFrames processed: {total}", end='')

  cap.release()
  print(f"\nDone. {total} frames saved to {txt_path}")

  zip_path = os.path.join(public_dir, "frames.zip")
  with zipfile.ZipFile(zip_path, 'w', zipfile.ZIP_DEFLATED) as zf:
    zf.write(txt_path, "frames.txt")
  print(f"Compressed to {zip_path} ({os.path.getsize(zip_path) / 1024 / 1024:.1f} MB)")

if __name__ == "__main__":
  main()
