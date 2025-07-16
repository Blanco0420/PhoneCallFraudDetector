import cv2
import numpy as np
import tkinter as tk
from PIL import Image, ImageTk # Import Pillow for image conversion

# --- 1. Load Image and Define ROI ---
# Ensure phone.jpg is in the parent directory relative to where you run this script.
# If not, adjust the path: e.g., 'phone.jpg' if in the same directory.
try:
    original_image = cv2.imread("/home/blanco/Pictures/phone pictures/WIN_20250630_16_05_20_Pro.jpg")
    if original_image is None:
        raise FileNotFoundError("Image not found at '../phone.jpg'")
except FileNotFoundError as e:
    print(f"Error loading image: {e}")
    print("Please make sure 'phone.jpg' is in the parent directory or update the path.")
    exit()

# Define ROI coordinates (from your previous code)

roi = [570,104,796,422]
roiStartY, roiEndY = 104, 526
roiStartX, roiEndX = 570, 1366

# Crop the ROI from the original image
roiCroppedImage = original_image[roiStartY:roiEndY, roiStartX:roiEndX].copy()

# Convert the cropped ROI to grayscale for thresholding
gray_roi = cv2.cvtColor(roiCroppedImage, cv2.COLOR_BGR2GRAY)

# Global variable to hold the PhotoImage, preventing it from being garbage collected
tk_image = None
image_label = None # To display the image

# --- 2. Function to apply Adaptive Threshold and update the display ---
def update_threshold_image(val=None):
    global tk_image, image_label

    # Get current slider values
    # blockSize must be odd and >= 3
    block_size = int(slider_block_size.get())
    if block_size % 2 == 0: # Ensure it's odd
        block_size += 1
    if block_size < 3: # Ensure minimum
        block_size = 3
    slider_block_size.set(block_size) # Update slider if value changed

    C_value = int(slider_C.get())

    # Apply Adaptive Thresholding
    # We use ADAPTIVE_THRESH_GAUSSIAN_C as it's often good for text.
    # THRESH_BINARY_INV: white text on black background.
    try:
        thresh_img = cv2.adaptiveThreshold(gray_roi, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
                                           cv2.THRESH_BINARY_INV, block_size, C_value)
    except cv2.error as e:
        print(f"OpenCV Error: {e}")
        print(f" blockSize: {block_size}, C: {C_value}")
        # If an error occurs (e.g., block_size too small/large for image),
        # return without updating to prevent crash.
        return

    # Convert OpenCV image (NumPy array) to PIL Image
    pil_img = Image.fromarray(thresh_img)

    # Convert PIL Image to Tkinter PhotoImage
    # Resize for better viewing if desired, but for tuning, showing actual size is good.
    # If the image is too large, you might want to scale it down.
    display_width = 800 # Max width for display
    display_height = int(pil_img.height * (display_width / pil_img.width)) if pil_img.width > 0 else pil_img.height

    # Only resize if the image is larger than desired display size
    if pil_img.width > display_width:
        pil_img = pil_img.resize((display_width, display_height), Image.LANCZOS)
    
    tk_image = ImageTk.PhotoImage(image=pil_img)

    # Update the image displayed in the Tkinter window
    if image_label: # Check if label exists before updating
        image_label.config(image=tk_image)
    else: # First time setup
        image_label = tk.Label(root, image=tk_image)
        image_label.pack(padx=10, pady=10)


# --- 3. Tkinter Window Setup ---
root = tk.Tk()
root.title("Adaptive Threshold Tester")

# Create sliders
# blockSize: Must be odd. Range from 3 to 255 (or higher if needed)
slider_block_size = tk.Scale(root, from_=3, to=255, resolution=2, # resolution 2 to force odd numbers
                             orient=tk.HORIZONTAL, label="Block Size (odd)",
                             command=update_threshold_image)
slider_block_size.set(15) # Default starting value
slider_block_size.pack(fill=tk.X, padx=10, pady=5)

# C: Can be positive or negative. Range from -50 to 50
slider_C = tk.Scale(root, from_=-50, to=50, resolution=1, # Can be any integer
                   orient=tk.HORIZONTAL, label="C Value",
                   command=update_threshold_image)
slider_C.set(4) # Default starting value
slider_C.pack(fill=tk.X, padx=10, pady=5)

# Initial call to display the image with default values
update_threshold_image()

# Run the Tkinter event loop
root.mainloop()
