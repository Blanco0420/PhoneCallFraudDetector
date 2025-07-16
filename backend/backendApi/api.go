package backendapi

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Blanco0420/Phone-Number-Check/backend/logging"
	webcamdetection "github.com/Blanco0420/Phone-Number-Check/backend/webcamDetection"
	"github.com/gin-gonic/gin"
	"gocv.io/x/gocv"
)

func setupRoutes(ROIChannel chan webcamdetection.RoiData, webcam *webcamdetection.CameraService) (*gin.Engine, error) {

	r := gin.Default()
	r.GET("/getCurrentImage", func(ctx *gin.Context) {
		imgPath := "/home/blanco/Pictures/phone pictures/WIN_20250630_16_05_20_Pro.jpg"
		if info, err := os.Stat(imgPath); err != nil {
			fmt.Println(info)
			fmt.Println("ERROR: ", err)
		}
		img := gocv.IMRead("/home/blanco/Pictures/phone pictures/WIN_20250630_16_05_20_Pro.jpg", gocv.IMReadColor)
		// img, err := webcam.GetFrame()
		// if err != nil {
		// 	ctx.JSON(500, err)
		// }
		if img.Empty() {
			err := errors.New("Could not read image")
			logging.Error().Err(err)
			ctx.JSON(http.StatusNotFound, gin.H{"message": err.Error(), "error": err})
			return
		}
		imgBuf, err := gocv.IMEncode(".jpg", img)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error(), "error": err})
			logging.Error().Err(err).Msg(err.Error())
			return
		}
		defer imgBuf.Close()
		imageBase64 := base64.StdEncoding.EncodeToString(imgBuf.GetBytes())
		ctx.JSON(http.StatusOK, gin.H{
			"image": imageBase64,
		})
	})

	r.POST("/setROIData", func(ctx *gin.Context) {
		payload := webcamdetection.RoiData{}
		if err := ctx.ShouldBindJSON(&payload); err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
		}
		ROIChannel <- payload
		ctx.JSON(200, gin.H{"data": payload})
	})

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "pong!")
	})

	return r, nil
}

func StartBackendApi(ROIChannel chan webcamdetection.RoiData, webcam *webcamdetection.CameraService) error {
	r, err := setupRoutes(ROIChannel, webcam)
	if err != nil {
		return err
	}
	return r.Run(":8080")
}
