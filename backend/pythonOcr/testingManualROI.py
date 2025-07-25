import cv2
import easyocr

def toGray(mat):
    return cv2.cvtColor(mat, cv2.COLOR_BGR2GRAY)

image = cv2.imread("/home/blanco/Pictures/processed/sharpened.jpg")

reader = easyocr.Reader(["ja"])

roi = cv2.selectROI("", image, True, False)

roiCroppedImage = image[int(roi[1]):int(roi[1]+roi[3]), int(roi[0]):int(roi[0]+roi[2])]

cv2.imshow("", roiCroppedImage)

text = reader.readtext(roiCroppedImage)
print(text)


