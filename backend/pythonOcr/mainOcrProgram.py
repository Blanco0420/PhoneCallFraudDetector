import cv2
from cv2.gapi import erode
import numpy as np
import easyocr
import re
from matplotlib import pyplot as plt
import os

reader = easyocr.Reader(["ja", "en"])

# --- 1. Load Image and Define ROI ---
image = cv2.imread("/home/blanco/Pictures/phone pictures/WIN_20250630_16_05_20_Pro.jpg")

if image is None:
    print("Error: Could not load the original image 'phone.jpg'. Please check the path.")
    exit()

original_display_image = image.copy()


# roiStartY, roiEndY = 80, 460
# roiStartX, roiEndX = 520, 1250
#
# roiCroppedImage = image[roiStartY:roiEndY, roiStartX:roiEndX].copy()

# --- 2. Preprocessing for Contour Detection ---

# hsv = cv2.cvtColor(image, cv2.COLOR_BGR2HSV)

# lowerBlue = np.array([0,0,73])
# upperBlue = np.array([179,255,255])
#
# mask = cv2.inRange(hsv, lowerBlue, upperBlue)
# # blurred = cv2.GaussianBlur(mask, (3,3), 3)
#
# # kernelErode = np.ones((3, 3), np.uint8)
# # mask = cv2.erode(mask, kernelErode, iterations=1)
# # kernelDilate = np.ones((7, 7), np.uint8)
# # mask = cv2.dilate(mask, kernelDilate, iterations=2)
#
# # edges = cv2.Canny(image, 50,50)
# # contours, _= cv2.findContours(edges, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
#
#
# contours, _ = cv2.findContours(mask.copy(), cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
#
# roiStartX, roiStartY, roiEndX, roiEndY = 0,0, image.shape[1], image.shape[0]
# foundRoi = False
#
#
# imageWithContours = image.copy()
# if len(contours) > 0:
#     # print(len(contours))
#     largestContour = max(contours, key=cv2.contourArea)
#
#     # cv2.drawContours(imageWithContours, contours, -1, (0,255,0), 2)
#     # cv2.imshow("", imageWithContours)
#     # cv2.waitKey(0)
#     # os._exit(0)
#     for contour in contours:
#         perimeter = cv2.arcLength(contour, True)
#
#         epsilon = 0.02 * perimeter
#         approx =cv2.approxPolyDP(contour, epsilon, True)
#
#         if len(approx) == 4 and cv2.isContourConvex(approx):
#             x,y,w,h = cv2.boundingRect(approx)
#
#             aspectRatio = w/float(h)
#
#             minAspectRatio = 1.5
#             maxAspectRatio = 3.0
#
#             minArea = image.shape[0] * image.shape[1] * 0.05
#
#             if (w*h > minArea) and (minAspectRatio < aspectRatio < maxAspectRatio):
#                 roiStartX, roiStartY = x,y
#                 roiEndX, roiEndY = x+w, y+h
#                 foundRoi=True
#                 break
#
# if not foundRoi:
#     print("Could not find ROI...")
#
# roiStartX = max(0, roiStartX)
# roiStartY = max(0, roiStartY)
# roiEndX = min(image.shape[1], roiEndX)
# roiEndY = min(image.shape[0], roiEndY)
#
# # Add a small buffer around the detected ROI to ensure no text is cut off
# # This helps if the contour approximation is slightly too tight
# buffer = 10# Pixels to add around the detected bounding box
# roiStartX = max(0, roiStartX - buffer)
# roiStartY = max(0, roiStartY - buffer)
# roiEndX = min(image.shape[1], roiEndX + buffer)
# roiEndY = min(image.shape[0], roiEndY + buffer)
# roiCroppedImage = image.copy()
# if roiEndX > roiStartX and roiEndY > roiStartY:
#     roiCroppedImage = roiCroppedImage[roiStartY: roiEndY, roiStartX:roiEndX]

roi = cv2.selectROI("select roi", image, True, False)
# roi = [570,104,796,421]
print(roi)
roiCroppedImage = image[int(roi[1]):int(roi[1]+roi[3]), int(roi[0]):int(roi[0]+roi[2])]
gray = cv2.cvtColor(roiCroppedImage, cv2.COLOR_BGR2GRAY)

blurred = cv2.medianBlur(gray, 5)

# --- APPLY OPTIMIZED ADAPTIVE THRESHOLDING HERE ---
# Use your optimized values: blockSize = 49, C = 6
# _,thresh = cv2.threshold(blurred, 155, 200, cv2.THRESH_BINARY)
thresh = cv2.adaptiveThreshold(blurred, 200, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, cv2.THRESH_BINARY_INV,255, 9)
# _, thresh = cv2.threshold(blurred, 36,255,cv2.THRESH_BINARY_INV + cv2.THRESH_OTSU)

# Morphological closing to connect broken digits
kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (5,5))
# thresh = cv2.morphologyEx(thresh, cv2.MORPH_CLOSE, kernel)

thresh = cv2.dilate(thresh, kernel, iterations=1)
thresh = cv2.erode(thresh, kernel, iterations=1)
thresh = cv2.GaussianBlur(thresh, (3,3), 0)
# kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (3,3))
# # thresh = cv2.erode(thresh, kernel, iterations=1)
# thresh = cv2.dilate(thresh, kernel, iterations=2)
# thresh = cv2.erode(thresh, kernel)
# thresh = cv2.GaussianBlur(thresh, (3,3), 23)

print(reader.readtext(thresh))
cv2.imshow("", thresh)
cv2.waitKey(0)
os._exit(0)

# thresh = cv2.erode(thresh, kernel, iterations=1)
# thresh = cv2.dilate(thresh, kernel, iterations=1)
# --- 3. Find Contours ---
cnts, _ = cv2.findContours(
    thresh.copy(), cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)

print(f"Number of raw contours found: {len(cnts)}")

# --- 4. Filter and Sort Character Bounding Boxes ---
characterBoxes = []
for c in cnts:
    x, y, w, h = cv2.boundingRect(c)

    area = w * h

    # You might need to re-evaluate these filter values
    # now that the thresholding is better. Start with permissive, then tighten.
    # Keep these as a starting point.
    if area > 50 and area < 3500 and h > 10 and w > 12:
        characterBoxes.append((x, y, w, h))

print(f"Number of character boxes after filtering: {len(characterBoxes)}")

characterBoxes.sort(key=lambda b: b[1])

# --- 5. Group Character Boxes into Lines ---
lines = []
if characterBoxes:
    currentLine = [characterBoxes[0]]

    lineToleranceY = 10 # Try 10-15

    for i in range(1, len(characterBoxes)):
        x1, y1, w1, h1 = currentLine[-1]
        x2, y2, w2, h2 = characterBoxes[i]

        overlap = (min(y1 + h1, y2 + h2) - max(y1, y2)) / \
            min(h1, h2) if min(h1, h2) > 0 else 0

        if abs(y2 - y1) < lineToleranceY or overlap > 0.6:
            currentLine.append(characterBoxes[i])
        else:
            lines.append(currentLine)
            currentLine = [characterBoxes[i]]
    lines.append(currentLine)
else:
    print("No character boxes found after filtering, so no lines will be formed.")

print(f"Number of lines detected: {len(lines)}")

# --- 6. Process Each Line: Create Merged Bounding Boxes and Extract Line Images ---
extracted_line_images = []
for line_chars in lines:
    line_chars.sort(key=lambda b: b[0])

    min_x_roi = min(b[0] for b in line_chars)
    min_y_roi = min(b[1] for b in line_chars)
    max_x_roi = max(b[0] + b[2] for b in line_chars)
    max_y_roi = max(b[1] + b[3] for b in line_chars)

    padding_x_start = 3
    padding_y_start = 2
    buffer_end = 10

    final_min_x_roi = max(0, min_x_roi - padding_x_start)
    final_min_y_roi = max(0, min_y_roi - padding_y_start)

    final_max_x_roi = min(roiCroppedImage.shape[1], max_x_roi + buffer_end)
    final_max_y_roi = min(roiCroppedImage.shape[0], max_y_roi + buffer_end)

    line_image_data = roiCroppedImage[final_min_y_roi:final_max_y_roi,
                                      final_min_x_roi:final_max_x_roi]

    if line_image_data.size > 0:

        line_gray = cv2.cvtColor(line_image_data, cv2.COLOR_BGR2GRAY)
        line_processed_for_ocr = cv2.adaptiveThreshold(
            line_gray, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, cv2.THRESH_BINARY_INV, 49, 6)
        full_line_text = ""
        try:
            # First, try with allowlist for digits/symbols, as phone numbers are critical.
            # This is a stronger hint to EasyOCR to look for numbers.
            ocr_result_with_allowlist = reader.readtext(line_processed_for_ocr)

            # If the allowlist found something, use that.
            # Otherwise, try without allowlist for general text (Japanese/English).
            if ocr_result_with_allowlist:
                ocr_result = ocr_result_with_allowlist
            else:
                # Fallback to default recognition
                ocr_result = reader.readtext(line_processed_for_ocr)

            extracted_texts = []
            for res in ocr_result:
                if len(res) >= 2 and isinstance(res[1], str):
                    extracted_texts.append(res[1])
                else:
                    print(
                        f"WARNING: Unexpected format for EasyOCR result item: {res}")
            full_line_text = " ".join(extracted_texts).strip()

            # print(ocr_result) # Remove this print for cleaner output, the full_line_text will be printed later

        except Exception as e:
            print(f"Error during EasyOCR for line (bbox_roi: {final_min_x_roi},{
                  final_min_y_roi},{final_max_x_roi},{final_max_y_roi}): {e}")
            full_line_text = ""  # Default to empty string on error

        extracted_line_images.append({
            'bbox_roi': (final_min_x_roi, final_min_y_roi,
                         final_max_x_roi - final_min_x_roi, final_max_y_roi - final_min_y_roi),
            'text': full_line_text,
            'image_data': line_image_data
        })

# Optionally, run OCR on the whole ROI for comparison
result = reader.readtext(thresh, allowlist='0123456789+-() ')
print("Full ROI OCR result:", result)

callerId = "Not Found"
phoneNumber = "Not Found"

# Regex for robust phone number matching
phone_number_pattern = r'(\+?\d{1,4}[-.\s]?)?(\(?\d{1,4}\)?[-.\s]?){2,}\d{1,4}'
japanese_caller_keywords = ["通話元", "着信中", "着信"]  # Added "着信" for robustness


# Iterate through the OCR results for each line to identify the specific lines
# Based on your screenshot and typical phone display:
# Line 1: Date/Time (not needed)
# Line 2: "若信中" (Incoming call/status)
# Line 3: "通話元: 07091762683" (Caller ID line)
# Line 4: "07091762683" (Actual phone number line, often a cleaner version)
# Line 5: Bottom buttons (not needed)

# Let's re-sort by Y-coordinate just to be safe, although 'lines' should already be sorted.
extracted_line_images.sort(key=lambda item: item['bbox_roi'][1])

# Attempt to extract based on content and relative position
found_caller_id_line_index = -1
found_phone_number_line_index = -1

for i, line_data in enumerate(extracted_line_images):
    line_text = line_data['text'].strip()

    # Look for the caller ID line (containing "通話元")
    if "通話元" in line_text and found_caller_id_line_index == -1:  # Only take the first instance
        caller_id = line_text
        found_caller_id_line_index = i
        # Also extract number from this line as a fallback/initial phone_number
        match = re.search(phone_number_pattern, line_text)
        if match:
            # Clean extracted number (remove spaces, hyphens, parens)
            phone_number = "".join(filter(str.isdigit, match.group(0)))

    # Look for a line that is primarily a phone number and potentially below the caller ID line
    # Using a stricter fullmatch for just digits after cleaning
    cleaned_line_for_number = "".join(
        filter(str.isdigit, line_text))  # Extract only digits
    if re.fullmatch(r'\d{7,}', cleaned_line_for_number):  # Checks for 7 or more digits
        # This is a candidate for the standalone phone number line
        if found_phone_number_line_index == -1:  # Take the first one as primary
            phone_number = cleaned_line_for_number
            found_phone_number_line_index = i
        # If a caller ID line was found before this, and this is below it, it's a good candidate for the "pure" number
        elif found_caller_id_line_index != -1 and i > found_caller_id_line_index:
            phone_number = cleaned_line_for_number
            found_phone_number_line_index = i
            break  # Found the most likely candidates, can stop searching

# Fallback: If only caller_id was found and it contains a number, use that for phone_number
if callerId != "Not Found" and phoneNumber == "Not Found":
    match = re.search(phone_number_pattern, callerId)
    if match:
        phone_number = "".join(filter(str.isdigit, match.group(0)))

# --- 7. Visualization ---

# Draw ROI on original image
cv2.rectangle(original_display_image, (roiStartX, roiStartY),
              (roiEndX, roiEndY), (0, 255, 255), 2)
# cv2.imshow("Original Image with ROI", original_display_image) # Keep commented for now if too many windows

# Display the thresholded image (full ROI) for debugging contour detection
cv2.imshow("Thresholded ROI (for Contours)", thresh)

# Display the cropped ROI itself
cv2.imshow("Cropped ROI Image", roiCroppedImage)

# Draw line bounding boxes AND recognized text on a copy of the cropped ROI
cropped_roi_with_lines = roiCroppedImage.copy()
print(f"\n--- OCR Results ---")
for i, line_data in enumerate(extracted_line_images):
    x, y, w, h = line_data['bbox_roi']
    text = line_data['text']
    print("Text: ", text)

    # Draw rectangle and put text on the visualization image
    cv2.rectangle(cropped_roi_with_lines, (x, y),
                  (x + w, y + h), (0, 255, 0), 2)  # Green boxes

    if text:  # Only put text if something was recognized
        font_scale = h / 25.0  # Dynamically scale font size
        # Clamp between 0.4 and 1.0
        font_scale = max(0.4, min(1.0, font_scale))
        cv2.putText(cropped_roi_with_lines, text, (x, y - 5), cv2.FONT_HERSHEY_SIMPLEX,
                    font_scale, (255, 0, 0), 1, cv2.LINE_AA)  # Blue text

    # Also display each line image (raw, as extracted from original color ROI) for debugging
    line_img_raw = line_data['image_data']
    if line_img_raw.size > 0:
        display_height = max(50, line_img_raw.shape[0])
        display_width = int(
            line_img_raw.shape[1] * (display_height / line_img_raw.shape[0]))
        if display_width == 0:
            display_width = line_img_raw.shape[1] if line_img_raw.shape[1] > 0 else 100
        if display_height == 0:
            display_height = line_img_raw.shape[0] if line_img_raw.shape[0] > 0 else 50

        try:
            resized_line_img_raw = cv2.resize(
                line_img_raw, (display_width, display_height), interpolation=cv2.INTER_AREA)
            # Show raw image before OCR preprocessing
            cv2.imshow(f"Raw Line Image {
                       i+1} - {text[:15]}", resized_line_img_raw)
        except cv2.error as e:
            print(f"Error displaying raw line image {i+1}: {e}")

cv2.imshow("Cropped ROI with Line Bounding Boxes & OCR Text",
           cropped_roi_with_lines)


print(f"Final Detected Caller ID: {callerId}")
print(f"Final Detected Phone Number: {phoneNumber}")

cv2.waitKey(0)
cv2.destroyAllWindows()
