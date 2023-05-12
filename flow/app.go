package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"github.com/second-state/WasmEdge-go/wasmedge"
	bindgen "github.com/second-state/wasmedge-bindgen/host/go"
	"github.com/yomorun/yomo"
	"github.com/yomorun/yomo/core/frame"
)

var (
	counter uint64
)

const (
	ImageDataKey       = 0x10
	InferenceResultKey = 0x11
)

func main() {
	// Connect to Zipper service
	sfn := yomo.NewStreamFunction(
		"image-recognition",
		"localhost:9900",
		yomo.WithObserveDataTags(ImageDataKey),
	)
	defer sfn.Close()

	// set handler
	sfn.SetHandler(Handler)

	// start
	err := sfn.Connect()
	if err != nil {
		log.Print("❌ Connect to zipper failure: ", err)
		os.Exit(1)
	}

	select {}
}

type WasmObj struct {
	conf      *wasmedge.Configure
	vm        *wasmedge.VM
	tfobj     *wasmedge.Module
	tfliteobj *wasmedge.Module
}

func (w *WasmObj) Release() {
	w.tfobj.Release()
	w.tfliteobj.Release()
	w.conf.Release()
	w.vm.Release()
}

// Handler process the data in the stream
func Handler(img []byte) (frame.Tag, []byte) {
	// Initialize WasmEdge's VM
	w, err := initVM()
	if err != nil {
		fmt.Printf("wasmedge init failed: %v\n", err)
		return 0, nil
	}
	defer w.Release()

	bg := bindgen.New(w.vm)

	// recognize the image
	res, _, err := bg.Execute("infer", img)
	if err == nil {
		fmt.Println("GO: Run bindgen -- infer:", string(res[0].([]byte)))
	} else {
		fmt.Println("GO: Run bindgen -- infer FAILED")
	}

	// print logs
	hash := genSha1(img)
	log.Printf("✅ received image-%d hash %v, img_size=%d \n", atomic.AddUint64(&counter, 1), hash, len(img))

	return InferenceResultKey, []byte(hash)
}

// genSha1 generate the hash value of the image
func genSha1(buf []byte) string {
	h := sha1.New()
	h.Write(buf)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// initVM initialize WasmEdge's VM
func initVM() (*WasmObj, error) {
	wasmedge.SetLogErrorLevel()
	/// Set Tensorflow not to print debug info
	os.Setenv("TF_CPP_MIN_LOG_LEVEL", "3")
	os.Setenv("TF_CPP_MIN_VLOG_LEVEL", "3")

	/// Create configure
	conf := wasmedge.NewConfigure(wasmedge.WASI)

	/// Create VM with configure
	vm := wasmedge.NewVMWithConfig(conf)

	/// Init WASI
	var wasi = vm.GetImportModule(wasmedge.WASI)
	wasi.InitWasi(
		os.Args[1:],     /// The args
		os.Environ(),    /// The envs
		[]string{".:."}, /// The mapping directories
	)

	/// Register WasmEdge-tensorflow and WasmEdge-image
	var tfobj = wasmedge.NewTensorflowModule()
	var tfliteobj = wasmedge.NewTensorflowLiteModule()
	err := vm.RegisterModule(tfobj)
	if err != nil {
		return nil, err
	}
	err = vm.RegisterModule(tfliteobj)
	if err != nil {
		return nil, err
	}
	var imgobj = wasmedge.NewImageModule()
	err = vm.RegisterModule(imgobj)
	if err != nil {
		return nil, err
	}

	filename := "rust_mobilenet_food_lib"
	ok := false
	for _, suffix := range []string{".so", ".wasm"} {
		if _, err := os.Stat(filename + suffix); err == nil {
			filename += suffix
			ok = true
			break
		}
	}
	if !ok {
		return nil, fmt.Errorf("cannot find model file %s[.so|.wasm]", filename)
	}

	/// Instantiate wasm
	err = vm.LoadWasmFile(filename)
	if err != nil {
		return nil, err
	}
	err = vm.Validate()
	if err != nil {
		return nil, err
	}
	err = vm.Instantiate()
	if err != nil {
		return nil, err
	}

	return &WasmObj{conf, vm, tfobj, tfliteobj}, nil
}
