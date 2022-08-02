package eb

import (
	"alpr/dckr"
	"alpr/models"
	"alpr/reps"
	"alpr/utils"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path"
	"time"
)

type FFmpegReaderResponseEvent struct {
	Counter   *utils.Counter
	Scheduler dckr.AlprContainerScheduler
	Rb        *reps.RepoBucket
}

func createImage(b64Img *string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(*b64Img)
	if err != nil {
		log.Println("an error occurred while decoding base64 string, err: ", err.Error())
		return "", err
	}
	img, err := jpeg.Decode(bytes.NewReader(b))
	if err != nil {
		log.Println("an error occurred while creating an image, err: ", err.Error())
		return "", err
	}
	baseName := uuid.New().String() + ".jpg"
	imageName := path.Join(utils.GetTempDir(), baseName)
	f, err := os.Create(imageName)
	if err != nil {
		log.Println("an error occurred while creating an image file, err: ", err.Error())
		return "", err
	}
	defer f.Close()
	if err = jpeg.Encode(f, img, nil); err != nil {
		log.Println("an error occurred while creating an image, err: ", err.Error())
		return "", err
	}
	return baseName, nil
}

func (f *FFmpegReaderResponseEvent) handleInternal(req *models.FFmpegReaderResponse) error {
	imageName, err := createImage(&req.Img)
	if err != nil {
		return err
	}
	imgPath := path.Join(utils.GetTempDir(), imageName)
	defer func() {
		e := os.Remove(imgPath)
		if e != nil {
			log.Println("an error occurred while deleting a temp image file, err: ", err.Error())
		}
	}()

	if int(*f.Counter) > 1000000000 {
		*f.Counter = 0
	}
	result, err := f.Scheduler.Detect(f.Counter, imageName)
	if err != nil {
		return err
	}
	if result != nil && len(result.Results) > 0 {
		if result.FileName != imageName {
			log.Println("image name mismatch: passed: ", imageName, " got: ", result.FileName)
			return errors.New("process are shuffled")
		}
		var img *image.RGBA = nil
		resp := models.AlprResponse{ImgWidth: result.ImgWidth, ImgHeight: result.ImgHeight, ProcessingTimeMs: float32(result.ProcessingTimeMs)}
		resp.Id, resp.SourceId, resp.AiClipEnabled = result.FileName[:len(result.FileName)-4], req.Source, req.AiClipEnabled
		resp.CreatedAt = utils.TimeToString(time.Now(), true)
		resp.Results = make([]*models.AlprResponseResult, 0)
		for _, r := range result.Results {
			arr := &models.AlprResponseResult{}
			arr.Plate = r.Plate
			arr.Confidence = float32(r.Confidence)
			arr.ProcessingTimeMs = float32(r.ProcessingTimeMs)

			coor1 := r.Coordinates[0]
			coor2 := r.Coordinates[2]
			x0, y0, x1, y1 := coor1.X, coor1.Y, coor2.X, coor2.Y
			arr.Coordinates = models.AlprResponseCoordinate{X0: x0, Y0: y0, X1: x1, Y1: y1}
			arr.Candidates = make([]*models.AlprResponseCandidate, 0)
			for _, c := range r.Candidates {
				rc := &models.AlprResponseCandidate{}
				rc.Plate = c.Plate
				rc.Confidence = float32(c.Confidence)
				arr.Candidates = append(arr.Candidates, rc)
			}
			resp.Results = append(resp.Results, arr)

			img, err = utils.DrawRect(imgPath, x0, y0, x1, y1, r.Plate, r.Confidence)
			if err != nil {
				log.Println("an error occurred while drawing a rectangle, err: ", err.Error())
				return err
			}
			err = utils.OverwriteImage(img, imgPath)
			if err != nil {
				return err
			}
		}

		evt := EventBus{Rb: f.Rb, Channel: "alpr_service"}
		resp.Base64Image, err = utils.ImageToBase64(img)
		if err != nil {
			log.Println("an error occurred while creating base64 from an image, err: ", err.Error())
			return err
		}

		buffer, err := json.Marshal(&resp)
		if err != nil {
			log.Println("an error occurred while creating json of the response object, err: ", err.Error())
			return err
		}
		err = evt.Publish(buffer)
		fmt.Println("published to alpr_service, source: " + req.Source)
		if err != nil {
			log.Println("an error occurred while publishing the event, err: ", err.Error())
			return err
		}
	}

	return err
}

func (f *FFmpegReaderResponseEvent) Handle(event *redis.Message) error {
	req := &models.FFmpegReaderResponse{}
	err := utils.DeserializeJson(event.Payload, req)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	err = f.handleInternal(req)
	if err != nil {
		log.Println("an error occurred while handling the event, err: ", err.Error())
	}
	return err
}
