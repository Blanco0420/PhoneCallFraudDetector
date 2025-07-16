import cv2
import numpy as np
import tkinter as tk
from tkinter import Scale, HORIZONTAL
from PIL import Image, ImageTk

# --- 1. Load Image ---
image_path = "/home/blanco/Pictures/phone pictures/processed/sharpenedBlurred.jpg" #
try:
    image = cv2.imread(image_path) #
    if image is None:
        raise FileNotFoundError(f"Image not found at '{image_path}'") #
except FileNotFoundError as e:
    print(f"Error loading image: {e}")
    print("Please make sure the image path is correct.")
    exit()

# Convert to HSV once
hsv = cv2.cvtColor(image, cv2.COLOR_BGR2HSV)

# --- 2. Tkinter Window Setup ---
root = tk.Tk()
root.title("HSV Range Selector (Tkinter Unified)")

# Global variables to hold the PhotoImage and image label
tk_image = None
image_label = None

# Sliders (Scale widgets) for HSV
sliders = {}

# --- 3. Function to update the mask image and display ---
def update_mask_image(val=None):
    global tk_image, image_label

    # Get current slider positions
    h_min = sliders["H_min"].get()
    h_max = sliders["H_max"].get()
    s_min = sliders["S_min"].get()
    s_max = sliders["S_max"].get()
    v_min = sliders["V_min"].get()
    v_max = sliders["V_max"].get()

    # Define lower and upper bounds for cv2.inRange
    # Ensure lower bound <= upper bound for all channels
    lower_bound = np.array([h_min, s_min, v_min])
    upper_bound = np.array([h_max, s_max, v_max])

    # Create HSV mask
    mask = cv2.inRange(hsv, lower_bound, upper_bound)

    # Apply morphological operations (from your original code)
    kernel_erode = np.ones((3,3), np.uint8)
    mask_processed = cv2.erode(mask, kernel_erode, iterations=1)
    kernel_dilate = np.ones((7,7), np.uint8)
    mask_processed = cv2.dilate(mask_processed, kernel_dilate, iterations=2)

    # Convert OpenCV mask image (NumPy array) to PIL Image for Tkinter
    pil_img = Image.fromarray(mask_processed)

    # Resize for better viewing (optional, but good for large images)
    display_width = 800
    if pil_img.width > display_width:
        display_height = int(pil_img.height * (display_width / pil_img.width))
        pil_img = pil_img.resize((display_width, display_height), Image.LANCZOS)
    
    tk_image = ImageTk.PhotoImage(image=pil_img)

    # Update the image displayed in the Tkinter window
    if image_label:
        image_label.config(image=tk_image)
    else:
        image_label = tk.Label(root, image=tk_image)
        image_label.pack(side=tk.TOP, padx=10, pady=10)

# --- 4. Create Tkinter Sliders (Scale widgets) ---
control_frame = tk.Frame(root)
control_frame.pack(side=tk.BOTTOM, fill=tk.X, padx=10, pady=10)

# Initial values based on your input (converted to OpenCV's 0-179/0-255 ranges)
initial_h_min, initial_h_max = 90, 118
initial_s_min, initial_s_max = 36, 255
initial_v_min, initial_v_max = 148, 255

# Hue (0-179)
tk.Label(control_frame, text="Hue Min").grid(row=0, column=0, sticky="w")
sliders["H_min"] = Scale(control_frame, from_=0, to=179, orient=HORIZONTAL, command=update_mask_image, length=200)
sliders["H_min"].set(initial_h_min)
sliders["H_min"].grid(row=0, column=1)

tk.Label(control_frame, text="Hue Max").grid(row=0, column=2, sticky="w")
sliders["H_max"] = Scale(control_frame, from_=0, to=179, orient=HORIZONTAL, command=update_mask_image, length=200)
sliders["H_max"].set(initial_h_max)
sliders["H_max"].grid(row=0, column=3)

# Saturation (0-255)
tk.Label(control_frame, text="Saturation Min").grid(row=1, column=0, sticky="w")
sliders["S_min"] = Scale(control_frame, from_=0, to=255, orient=HORIZONTAL, command=update_mask_image, length=200)
sliders["S_min"].set(initial_s_min)
sliders["S_min"].grid(row=1, column=1)

tk.Label(control_frame, text="Saturation Max").grid(row=1, column=2, sticky="w")
sliders["S_max"] = Scale(control_frame, from_=0, to=255, orient=HORIZONTAL, command=update_mask_image, length=200)
sliders["S_max"].set(initial_s_max)
sliders["S_max"].grid(row=1, column=3)

# Value (0-255)
tk.Label(control_frame, text="Value Min").grid(row=2, column=0, sticky="w")
sliders["V_min"] = Scale(control_frame, from_=0, to=255, orient=HORIZONTAL, command=update_mask_image, length=200)
sliders["V_min"].set(initial_v_min)
sliders["V_min"].grid(row=2, column=1)

tk.Label(control_frame, text="Value Max").grid(row=2, column=2, sticky="w")
sliders["V_max"] = Scale(control_frame, from_=0, to=255, orient=HORIZONTAL, command=update_mask_image, length=200)
sliders["V_max"].set(initial_v_max)
sliders["V_max"].grid(row=2, column=3)


# --- 5. Initial Call to Display ---
update_mask_image()

# --- 6. Run Tkinter Event Loop ---
root.mainloop()

# --- 7. Print Final Optimal Values when Tkinter window is closed ---
print(f"\n--- Final Recommended HSV Range ---")
print(f"lowerBlue = np.array([{sliders['H_min'].get()}, {sliders['S_min'].get()}, {sliders['V_min'].get()}])")
print(f"upperBlue = np.array([{sliders['H_max'].get()}, {sliders['S_max'].get()}, {sliders['V_max'].get()}])")

# --- 8. Clean up (not strictly necessary for Tkinter-only but good practice) ---
cv2.destroyAllWindows()
