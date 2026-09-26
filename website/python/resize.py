import cv2

SRC_W, SRC_H = 480, 360
CROP_H = int(SRC_W * 9 / 16)
Y_OFF = (SRC_H - CROP_H) // 2
DST_W, DST_H = 128, 72

def main():
  cap = cv2.VideoCapture("Bad Apple.mp4")
  fourcc = cv2.VideoWriter_fourcc(*"MP4V")
  out = cv2.VideoWriter("output.mp4", fourcc, 30, (DST_W, DST_H))

  if not cap.isOpened():
      print("Error: could not open Bad Apple.mp4")
      return

  total = 0
  while cap.isOpened():
      ret, frame = cap.read()
      if not ret:
          break

      cropped = frame[Y_OFF:Y_OFF+CROP_H, :]
      resized = cv2.resize(cropped, (DST_W, DST_H), interpolation=cv2.INTER_NEAREST)
      out.write(resized)
      total += 1
      print(f"\rFrames processed: {total}", end='')

  print(f"\nDone. {total} frames written to output.mp4 ({DST_W}x{DST_H})")
  cap.release()
  out.release()

if __name__ == "__main__":
  main()
