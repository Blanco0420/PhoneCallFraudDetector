package webcamdetection

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"strings"

	"github.com/nyaruka/phonenumbers"
	"gocv.io/x/gocv"
)

type CameraService struct {
	webcam *gocv.VideoCapture
	// mu     sync.Mutex
}

func NewCameraService(cameraIndex int) (*CameraService, error) {
	webcam, err := gocv.VideoCaptureDevice(cameraIndex)
	if err != nil {
		if strings.Contains(err.Error(), "opening device") {
		} else {
			return nil, err
		}
	}
	webcam.Set(gocv.VideoCaptureFrameWidth, 1920)
	webcam.Set(gocv.VideoCaptureFrameHeight, 1080)

	return &CameraService{webcam: webcam}, nil
}

func (cs *CameraService) Close() error {
	if cs.webcam != nil {
		return cs.webcam.Close()
	}
	return nil
}

func (cs *CameraService) GetFrame() (gocv.Mat, error) {
	img := gocv.NewMat()
	if cs.webcam == nil || !cs.webcam.IsOpened() {
		return img, fmt.Errorf("Camera could not be found")
	}
	if ok := cs.webcam.Read(&img); !ok || img.Empty() {
		img.Close()
		return img, os.ErrInvalid
	}
	return img, nil
}

func matToBytes(mat gocv.Mat) ([]byte, error) {
	img, err := mat.ToImage()
	if err != nil {
		return nil, fmt.Errorf("Error converting mat to img: %v", err)
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (cs *CameraService) MonitorCamera(roi RoiData) (string, error) {
	var num string
	// cameraOutput := gocv.NewMat()
	// if ok := cs.webcam.Read(&cameraOutput); !ok || cameraOutput.Empty() {
	// 	return text, fmt.Errorf("error getting frame from camera")
	// }

	cameraOutput := gocv.IMRead("/home/blanco/Pictures/phone pictures/WIN_20250630_16_05_20_Pro.jpg", gocv.IMReadColor)
	bounds := cameraOutput.Size()

	x := int(math.Max(0, float64(roi.X)))
	y := int(math.Max(0, float64(roi.Y)))
	w := int(math.Min(float64(roi.Width), float64(bounds[1]-x)))
	h := int(math.Min(float64(roi.Height), float64(bounds[0]-y)))

	if w <= 0 || h <= 0 {
		return num, fmt.Errorf("invalid ROI: width/height zero or negative after clamping")
	}

	rect := image.Rect(
		x,
		y,
		x+w,
		y+h,
	)
	croppedInput := cameraOutput.Region(rect)
	defer croppedInput.Close()
	outputImage, err := processImage(croppedInput)
	if err != nil {
		return num, err
	}
	bytes, err := matToBytes(outputImage)
	if err != nil {
		return num, err
	}
	for {
		text, err := ProcessText(bytes)
		if err != nil {
			return num, err
		}
		parsedNumber, err := phonenumbers.Parse(text, "JP")
		if err != nil {
			return num, err
		}
		if phonenumbers.IsValidNumber(parsedNumber) {
			num = text
			return num, nil
		}
		// num, err =
		break
	}
	return num, nil
}
func processImage(inputImage gocv.Mat) (gocv.Mat, error) {
	gray := gocv.NewMat()
	defer gray.Close()
	if err := gocv.CvtColor(inputImage, &gray, gocv.ColorBGRToGray); err != nil {
		return gocv.NewMat(), err
	}

	blurred := gocv.NewMat()
	defer blurred.Close()

	if err := gocv.MedianBlur(gray, &blurred, 5); err != nil {
		return gocv.NewMat(), err
	}

	thresh := gocv.NewMat()
	defer thresh.Close()
	if err := gocv.AdaptiveThreshold(blurred, &thresh, 200, gocv.AdaptiveThresholdGaussian, gocv.ThresholdBinaryInv, 255, 9); err != nil {
		return gocv.NewMat(), err
	}

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(5, 5))
	defer kernel.Close()

	if err := gocv.Dilate(thresh, &thresh, kernel); err != nil {
		return gocv.NewMat(), err
	}

	if err := gocv.Erode(thresh, &thresh, kernel); err != nil {
		return gocv.NewMat(), err
	}

	newBlurred := gocv.NewMat()
	defer newBlurred.Close()

	if err := gocv.GaussianBlur(thresh, &newBlurred, image.Pt(3, 3), 0, 0, gocv.BorderConstant); err != nil {
		return gocv.NewMat(), err
	}

	win := gocv.NewWindow("test")
	win.IMShow(newBlurred)
	win.WaitKey(0)
	win.Close()
	result := newBlurred.Clone()
	return result, nil
}

// edged := gocv.NewMat()
// gocv.Canny(blurred, &edged, 50, 200)
// size := image.Point{
// 	X: int(float64(edged.Cols()) * zoom),
// 	Y: int(float64(edged.Rows()) * zoom),
// }
// resized := gocv.NewMat()
// defer resized.Close()
// gocv.Resize(img, &resized, size, 0, 0, gocv.InterpolationArea)
//
// // Clamp view size if resized image is smaller than viewport
// if resized.Cols() < viewWidth || resized.Rows() < viewHeight {
// 	gocv.Resize(edged, &resized, image.Pt(viewWidth, viewHeight), 0, 0, gocv.InterpolationLinear)
// 	zoom = float64(viewWidth) / float64(edged.Cols())
// 	pan.X = 0
// 	pan.Y = 0
// }
//
// // Clamp pan to avoid going outside bounds
// if pan.X < 0 {
// 	pan.X = 0
// }
// if pan.Y < 0 {
// 	pan.Y = 0
// }
// if pan.X+viewWidth > resized.Cols() {
// 	pan.X = resized.Cols() - viewWidth
// 	if pan.X < 0 {
// 		pan.X = 0
// 	}
// }
// if pan.Y+viewHeight > resized.Rows() {
// 	pan.Y = resized.Rows() - viewHeight
// 	if pan.Y < 0 {
// 		pan.Y = 0
// 	}
// }
//
// tempRoi := resized.Region(image.Rect(pan.X, pan.Y, pan.X+viewWidth, pan.Y+viewHeight))
// roi := gocv.NewMat()
// tempRoi.CopyTo(&roi)

// switch key {
// case 'q':
// 	return
// case '+', '=':
// 	zoom *= 1.1
// case '-':
// 	zoom /= 1.1
// case 'w':
// 	pan.Y -= 20
// case 's':
// 	pan.Y += 20
// case 'a':
// 	pan.X -= 20
// case 'd':
// 	pan.X += 20
// case ']': // increase width
// 	viewWidth += 20
// 	if viewWidth > resized.Cols() {
// 		viewWidth = resized.Cols()
// 	}
// case '[': // decrease width
// 	if viewWidth > 100 {
// 		viewWidth -= 20
// 	}
// case '\'': // increase height (single quote key)
// 	viewHeight += 20
// 	if viewHeight > resized.Rows() {
// 		viewHeight = resized.Rows()
// 	}
// case ';': // decrease height
// 	if viewHeight > 100 {
// 		viewHeight -= 20
// 	}
// }
//
// // Clamp pan again because crop size may have changed
// if pan.X+viewWidth > resized.Cols() {
// 	pan.X = resized.Cols() - viewWidth
// 	if pan.X < 0 {
// 		pan.X = 0
// 	}
// }
// if pan.Y+viewHeight > resized.Rows() {
// 	pan.Y = resized.Rows() - viewHeight
// 	if pan.Y < 0 {
// 		pan.Y = 0
// 	}
// }
// window.IMShow(final)
//
// window.WaitKey(0) //
// bytes, err := matToBytes(final)
// if err != nil {
// 	fmt.Println("Error converting mat to bytes: ", err)
// 	continue
// }
// ProcessText(bytes)

type CameraConfig struct {
	Pan  image.Point
	Zoom float64
}

// StartCameraWithControls opens the webcam, displays a movable/zoomable ROI, and supports keybindings for pan, zoom, and viewport size.
// func StartCameraWithControls() (output CameraConfig) {
// 	// webcam, err := gocv.OpenVideoCapture(0)
// 	// if err != nil {
// 	// 	fmt.Println("Error opening webcam:", err)
// 	// 	return
// 	// }
// 	// defer webcam.Close()
//
// 	window := gocv.NewWindow("Camera Controls")
// 	defer window.Close()
//
// 	zoom := 1.0
// 	pan := image.Point{X: 0, Y: 0}
// 	viewWidth, viewHeight := 1920, 1080
//
// 	// img := gocv.NewMat()
// 	img := gocv.IMRead("phone.jpg", gocv.IMReadColor)
// 	defer img.Close()
//
// 	useAdaptive := true // thresholding toggle variable
// 	for {
// 		if img.Empty() {
// 			break
// 		}
// 		// if ok := webcam.Read(&img); !ok || img.Empty() {
// 		// 	fmt.Println("Cannot read frame from webcam")
// 		// 	continue
// 		// }
//
// 		gray := gocv.NewMat()
// 		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
//
// 		// blurred := gocv.NewMat()
// 		// gocv.GaussianBlur(gray, &blurred, image.Pt(5, 5), 0, 0, gocv.BorderDefault)
//
// 		resized := gocv.NewMat()
// 		size := image.Point{
// 			X: int(float64(gray.Cols()) * zoom),
// 			Y: int(float64(gray.Rows()) * zoom),
// 		}
// 		gocv.Resize(gray, &resized, size, 0, 0, gocv.InterpolationArea)
//
// 		// Clamp view size if resized image is smaller than viewport
// 		if resized.Cols() < viewWidth || resized.Rows() < viewHeight {
// 			gocv.Resize(gray, &resized, image.Pt(viewWidth, viewHeight), 0, 0, gocv.InterpolationLinear)
// 			zoom = float64(viewWidth) / float64(gray.Cols())
// 			pan.X = 0
// 			pan.Y = 0
// 		}
// 		if pan.X < 0 {
// 			pan.X = 0
// 		}
// 		if pan.Y < 0 {
// 			pan.Y = 0
// 		}
// 		if pan.X+viewWidth > resized.Cols() {
// 			pan.X = resized.Cols() - viewWidth
// 			if pan.X < 0 {
// 				pan.X = 0
// 			}
// 		}
// 		if pan.Y+viewHeight > resized.Rows() {
// 			pan.Y = resized.Rows() - viewHeight
// 			if pan.Y < 0 {
// 				pan.Y = 0
// 			}
// 		}
//
// 		roi := resized.Region(image.Rect(pan.X, pan.Y, pan.X+viewWidth, pan.Y+viewHeight))
//
// 		// 1. (Optional) Upscale ROI
// 		roi = ResizeROI(roi, 2.0)
//
// 		// 2. Convert to grayscale (already done)
// 		// 3. Apply CLAHE or EqualizeHist for contrast
// 		enhanced := EqualizeHist(roi) // or use CLAHE if available
//
// 		// 4. (Optional) Light Gaussian blur (3,3)
// 		blurred := gocv.NewMat()
// 		gocv.GaussianBlur(enhanced, &blurred, image.Pt(3, 3), 0, 0, gocv.BorderDefault)
//
// 		// 5. Save for OCR
// 		gocv.IMWrite("ocr_input.png", blurred)
//
// 		// 2. Threshold (adaptive or Otsu) for display only
// 		var binary gocv.Mat
// 		if useAdaptive {
// 			binary = AdaptiveThreshold(blurred, 255, 21, 2)
// 		} else {
// 			binary = gocv.NewMat()
// 			gocv.Threshold(blurred, &binary, 0, 255, gocv.ThresholdBinaryInv+gocv.ThresholdOtsu)
// 		}
//
// 		// 3. Morphological closing for display only
// 		kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
// 		closed := gocv.NewMat()
// 		gocv.MorphologyEx(binary, &closed, gocv.MorphClose, kernel)
//
// 		// Show only the final processed image for display
// 		window.IMShow(closed)
// 		key := window.WaitKey(30)
//
// 		// Toggle thresholding method
// 		if key == 't' {
// 			useAdaptive = !useAdaptive
// 			if useAdaptive {
// 				fmt.Println("Switched to Adaptive Thresholding")
// 			} else {
// 				fmt.Println("Switched to Otsu Thresholding")
// 			}
// 		}
//
// 		// Cleanup
// 		closed.Close()
// 		kernel.Close()
// 		binary.Close()
// 		blurred.Close()
// 		roi.Close()
// 		gray.Close()
// 		resized.Close()
//
// 		switch key {
// 		case 'q':
// 			output.Pan = pan
// 			output.Zoom = zoom
// 			return
// 		case '+', '=':
// 			zoom *= 1.1
// 		case '-':
// 			zoom /= 1.1
// 		case 'w':
// 			pan.Y -= 20
// 		case 's':
// 			pan.Y += 20
// 		case 'a':
// 			pan.X -= 20
// 		case 'd':
// 			pan.X += 20
// 		case ']':
// 			viewWidth += 20
// 			if viewWidth > img.Cols() {
// 				viewWidth = img.Cols()
// 			}
// 		case '[':
// 			if viewWidth > 100 {
// 				viewWidth -= 20
// 			}
// 		case '\'': // increase height (single quote key)
// 			viewHeight += 20
// 			if viewHeight > img.Rows() {
// 				viewHeight = img.Rows()
// 			}
// 		case ';': // decrease height
// 			if viewHeight > 100 {
// 				viewHeight -= 20
// 			}
// 		case 't':
// 			useAdaptive = !useAdaptive
// 			if useAdaptive {
// 				fmt.Println("Switched to Adaptive Thresholding")
// 			} else {
// 				fmt.Println("Switched to Otsu Thresholding")
// 			}
// 		}
//
// 		fmt.Printf("Zoom: %.2f, Pan: (%d, %d), Viewport: %dx%d\n", zoom, pan.X, pan.Y, viewWidth, viewHeight)
// 	}
// 	return
// }
//
// // func StartOCRScanner(output chan<- string, stop <-chan struct{}) {
// // 	// webcam, _ := gocv.VideoCaptureDevice(0)
// // 	// defer webcam.Close()
// // 	window := gocv.NewWindow("Test")
// // 	defer window.Close()
// // 	//TODO: Move to separate function
// // 	client := gosseract.NewClient()
// // 	client.SetLanguage("jpn", "eng")
// // 	defer client.Close()
// // 	/////
// // 	zoom := 1.0
// // 	pan := image.Point{X: 0, Y: 0}
// // 	viewWidth, viewHeight := 800, 600
// // 	for {
// // 		select {
// // 		case <-stop:
// // 			return
// // 		default:
// // 			// if ok := webcam.Read(&img); !ok || img.Empty() {
// // 			// 	continue
// // 			// }
// //
// // 			// img := gocv.NewMat()
// // 			img := gocv.IMRead("/home/blanco/Pictures/picture_2025-05-30_16-11-40.jpg", gocv.IMReadColor)
// // 			if img.Empty() {
// // 				fmt.Println("Image is empty")
// // 			}
// // 			defer img.Close()
// // 			gray := gocv.NewMat()
// // 			defer gray.Close()
// // 			gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
// //
// // 			blurred := gocv.NewMat()
// // 			gocv.GaussianBlur(gray, &blurred, image.Pt(5, 5), 0, 0, gocv.BorderDefault)
// //
// // 			resized := gocv.NewMat()
// // 			size := image.Point{
// // 				X: int(float64(blurred.Cols()) * zoom),
// // 				Y: int(float64(blurred.Rows()) * zoom),
// // 			}
// // 			gocv.Resize(blurred, &resized, size, 0, 0, gocv.InterpolationArea)
// //
// // 			// Avoid negative dimensions (can happen if zoomed image is too small)
// // 			if resized.Cols() < viewWidth || resized.Rows() < viewHeight {
// // 				gocv.Resize(blurred, &resized, image.Pt(viewWidth, viewHeight), 0, 0, gocv.InterpolationLinear)
// // 				zoom = float64(viewWidth) / float64(blurred.Cols())
// // 				pan.X = 0
// // 				pan.Y = 0
// // 			}
// // 			if pan.X < 0 {
// // 				pan.X = 0
// // 			}
// // 			if pan.Y < 0 {
// // 				pan.Y = 0
// // 			}
// // 			if pan.X+viewWidth > resized.Cols() {
// // 				pan.X = resized.Cols() - viewWidth
// // 				if pan.X < 0 {
// // 					pan.X = 0
// // 				}
// // 			}
// // 			if pan.Y+viewHeight > resized.Rows() {
// // 				pan.Y = resized.Rows() - viewHeight
// // 				if pan.Y < 0 {
// // 					pan.Y = 0
// // 				}
// // 			}
// //
// // 			roi := resized.Region(image.Rect(pan.X, pan.Y, pan.X+viewWidth, pan.Y+viewHeight))
// // 			roiCopy := gocv.NewMat()
// // 			roi.CopyTo(&roiCopy)
// // 			roi.Close()
// // 			window.IMShow(roiCopy)
// // 			key := window.WaitKey(30)
// // 			switch key {
// // 			case 'q':
// // 				return
// // 			case '+', '=':
// // 				zoom *= 1.1
// // 			case '-':
// // 				zoom /= 1.1
// // 			case 'w':
// // 				pan.Y -= 20
// // 			case 's':
// // 				pan.Y += 20
// // 			case 'a':
// // 				pan.X -= 20
// // 			case 'd':
// // 				pan.X += 20
// // 			case ']':
// // 				viewWidth += 20
// // 				if viewWidth > resized.Cols() {
// // 					viewWidth = resized.Cols()
// // 				}
// // 			case '[':
// // 				if viewWidth > 100 {
// // 					viewWidth -= 20
// // 				}
// // 			}
// // 			if pan.X+viewWidth > resized.Cols() {
// // 				pan.X = resized.Cols() - viewWidth
// // 				if pan.X < 0 {
// // 					pan.X = 0
// // 				}
// // 			}
// // 			if pan.Y+viewHeight > resized.Rows() {
// // 				pan.Y = resized.Rows() - viewHeight
// // 				if pan.Y < 0 {
// // 					pan.Y = 0
// // 				}
// // 			}
// // 			fmt.Printf("WIDTH: %i\nHEIGHT: %i", viewWidth, viewHeight)
// // 			// window.IMShow(blurred)
// // 			// window.WaitKey(1)
// // 			// hsv := gocv.NewMat()
// // 			// gocv.CvtColor(img, &hsv, gocv.ColorBGRToHSV)
// // 			//
// // 			// lowerBlue := gocv.NewScalar(100, 100, 100, 0)
// // 			// upperBlue := gocv.NewScalar(140, 255, 255, 0)
// // 			// mask := gocv.NewMat()
// // 			// gocv.InRangeWithScalar(hsv, lowerBlue, upperBlue, &mask)
// // 			// gocv.CvtColor(gray, &gray, gocv.ColorGrayToBGR)
// // 			// grayMasked := gocv.NewMat()
// // 			// defer grayMasked.Close()
// // 			// gray.CopyToWithMask(&grayMasked, mask)
// // 			//
// // 			// invMask := gocv.NewMat()
// // 			// defer invMask.Close()
// // 			// gocv.BitwiseNot(mask, &invMask)
// // 			//
// // 			// nonBlue := gocv.NewMat()
// // 			// defer nonBlue.Close()
// // 			// img.CopyToWithMask(&nonBlue, invMask)
// //
// // 			// binary := gocv.NewMat()
// // 			// gocv.Threshold(gray, &binary, 0, 255, gocv.ThresholdBinaryInv+gocv.ThresholdOtsu)
// // 			//
// // 			// denoised := gocv.NewMat()
// // 			// gocv.FastNlMeansDenoising(binary, &denoised)
// // 			// if nonBlue.Empty() {
// // 			// 	fmt.Println("nonBlue is empty")
// // 			// }
// // 			// if grayMasked.Empty() {
// // 			// 	fmt.Println("grayMasked is empty")
// // 			// }
// // 			// result := gocv.NewMat()
// // 			// gocv.Add(nonBlue, grayMasked, &result)
// // 			// if result.Empty() {
// // 			// 	fmt.Println("processed image is empty")
// // 			// }
// // 			// defer result.Close()
// // 			// 	buf, err := gocv.IMEncode(".png", result)
// // 			// 	if err != nil {
// // 			// 		fmt.Println("IMEncode error: ", err)
// // 			// 		continue
// // 			// 	}
// // 			// defer buf.Close()
// //
// // 			// window.IMShow(gray)
// // 			// window.WaitKey(1)
// //
// // 			bytes, err := matToBytes(roiCopy)
// // 			if err != nil {
// // 				fmt.Println("Error converting mat to bytes: ", err)
// // 			}
// // 			client.SetImageFromBytes(bytes)
// // 			text, err := client.Text()
// // 			if err != nil {
// // 				fmt.Println("error getting text: ", err)
// // 			}
// // 			fmt.Println(text)
// // 			fmt.Println("####################################")
// // 			fmt.Print("\n\n\n\n\n\n")
// // 			// webcam.Read(&img)
// // 		}
// // 	}
// // }
//
// func extractPhoneNumbers(ocrResults map[string]interface{}) []string {
// 	numbers := []string{}
// 	results := ocrResults["results"].([]interface{})
// 	for _, result := range results {
// 		text := result.(map[string]interface{})["text"].(string)
// 		found := regexp.MustCompile(`\d{10,11}`).FindAllString(text, -1)
// 		numbers = append(numbers, found...)
// 	}
// 	fixed := []string{}
// 	for _, number := range numbers {
// 		if number[0] == '1' {
// 			fixed = append(fixed, "0"+number[1:])
// 		} else if number[0] == '7' {
// 			fixed = append(fixed, "0"+number[1:])
// 		} else {
// 			fixed = append(fixed, number)
// 		}
// 	}
// 	return fixed
// }
