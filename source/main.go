package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/yomorun/yomo"
)

const ImageDataKey = 0x10

func main() {
	fmt.Println("Go: Args:", os.Args)
	filePath := os.Args[1]

	// connect to yomo-zipper.
	source := yomo.NewSource("image-recognition-source", "localhost:9900")
	defer source.Close()

	err := source.Connect()
	if err != nil {
		fmt.Printf("❌ Emit the data to yomo-zipper failure with err: %v\n", err)
		return
	}

	loadVideoAndSendData(source, filePath)
}

func loadVideoAndSendData(source yomo.Source, filePath string) {
	send := func(id int, img []byte) {
		err := source.Write(ImageDataKey, img)
		if err != nil {
			fmt.Printf("❌ Send image-%v of %s to yomo-zipper failure with err: %v\n", id, filePath, err)
		} else {
			fmt.Printf("✅ Send image-frame-%v of %s to yomo-zipper, hash=%s, img_size=%v\n", id, filePath, genSha1(img), len(img))
		}
		time.Sleep(1 * time.Millisecond)
	}

	// load video and convert to images
	video := VideoImage{}
	num, err := video.GetFrameCount(filePath)
	if err != nil {
		fmt.Printf("❌ get frame count from %s err: %v", filePath, err)
		return
	}
	fmt.Printf("total %d image frames\n", num)
	ffStream := ffmpeg.Input(filePath)
	for i := 0; i < num; i++ {
		if i%24 == 0 {
			img, err := video.ExtractImageBytes(ffStream, i)
			if err != nil {
				fmt.Printf("ExtractImage64 error: %v\n", err)
			}
			send(i, img)
		}
	}

	fmt.Printf("Successfully sent %d images\n", num)
	time.Sleep(5 * time.Second)
}

func genSha1(buf []byte) string {
	h := sha1.New()
	h.Write(buf)
	return fmt.Sprintf("%x", h.Sum(nil))
}

type VideoImage struct {
}

func (v *VideoImage) ExtractImageBytes(stream *ffmpeg.Stream, frameNum int) ([]byte, error) {
	reader := v.extractImage(stream, frameNum)
	img, err := imaging.Decode(reader)
	if err != nil {
		return nil, err
	}
	imgBuf := new(bytes.Buffer)
	err = imaging.Encode(imgBuf, img, imaging.JPEG)
	if err != nil {
		return nil, err
	}
	return imgBuf.Bytes(), nil
}

func (v *VideoImage) extractImage(stream *ffmpeg.Stream, frameNum int) io.Reader {
	buf := bytes.NewBuffer(nil)
	err := stream.
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		//WithOutput(buf, os.Stdout).
		WithOutput(buf, nil).
		Run()
	if err != nil {
		panic(err)
	}
	return buf
}

func (v *VideoImage) GetFrameCount(inFileName string) (int, error) {
	data, err := ffmpeg.Probe(inFileName)
	if err != nil {
		return 0, err
	}
	var m map[string]interface{}
	err = json.Unmarshal([]byte(data), &m)
	if err != nil {
		return 0, err
	}

	var strInt string
	items := m["streams"].([]interface{})
	for _, item := range items {
		v := item.(map[string]interface{})
		if v["profile"] == "Main" || v["profile"] == "High" {
			strInt = v["nb_frames"].(string)
			break
		}
	}

	if len(strInt) == 0 {
		return 0, fmt.Errorf("not find profile(Main).nb_frames")
	}

	num, err := strconv.Atoi(strInt)
	if err != nil {
		return 0, nil
	}

	return num, nil
}
