package webcamdetection

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/Blanco0420/Phone-Number-Check/backend/config"
	"github.com/Blanco0420/Phone-Number-Check/backend/logging"
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
			if config.IsDev {
			} else {
				return nil, err
			}
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

func (cs *CameraService) GetFrame(img *gocv.Mat) error {
	if !config.IsDev {
		if cs.webcam == nil || !cs.webcam.IsOpened() {
			return fmt.Errorf("camera could not be found")
		}
		if !cs.webcam.Read(img) || img.Empty() {
			return fmt.Errorf("error getting frame, invalid or empty")
		}
		return nil
	}
	*img = gocv.IMRead("/tmp/Phraud_Example_Image.jpg", gocv.IMReadColor)
	if img.Empty() {
		return fmt.Errorf("dev image not found")
	}
	return nil
}

func matToBytes(mat gocv.Mat) ([]byte, error) {
	img, err := mat.ToImage()
	if err != nil {
		return nil, fmt.Errorf("error converting mat to img: %v", err)
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (cs *CameraService) MonitorCamera(ctx context.Context, roi RoiData) (string, error) {
	attempts := 30 // e.g., try for up to 30 frames (~3 seconds at 100ms)
	var text string

	for range attempts {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			// Read frame
			cameraOutput := gocv.NewMat()
			if err := cs.GetFrame(&cameraOutput); err != nil {
				cameraOutput.Close()
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if ok := gocv.IMWrite("./testImage.jpg", cameraOutput); !ok {
				return "", fmt.Errorf("failed to save test image")
			}

			bounds := cameraOutput.Size()
			x := int(math.Max(0, float64(roi.X)))
			y := int(math.Max(0, float64(roi.Y)))
			w := int(math.Min(float64(roi.Width), float64(bounds[1]-x)))
			h := int(math.Min(float64(roi.Height), float64(bounds[0]-y)))

			if w <= 0 || h <= 0 {
				cameraOutput.Close()
				return "", fmt.Errorf("invalid ROI: width/height zero or negative after clamping")
			}

			rect := image.Rect(x, y, x+w, y+h)
			croppedInput := cameraOutput.Region(rect)
			cameraOutput.Close()

			outputImage, err := processImage(croppedInput)
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			defer outputImage.Close()
			croppedInput.Close()

			bytes, err := matToBytes(outputImage)
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			text, err = ProcessText(bytes)
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			onlyNumbersReg := regexp.MustCompile(`[^\d+]`)
			text = onlyNumbersReg.ReplaceAllString(text, "")

			parsedNumber, err := phonenumbers.Parse(text, "JP")
			if err != nil {
				continue
			}
			if phonenumbers.IsValidNumber(parsedNumber) {
				logging.Info().Msg("Valid number detected")
				return text, nil
			} else {
				logging.Info().Msg("Invalid number")
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	return "", fmt.Errorf("no valid number detected after %d attempts", attempts)
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

	result := newBlurred.Clone()
	return result, nil
}
