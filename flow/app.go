package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"github.com/second-state/WasmEdge-go/wasmedge"
	"github.com/yomorun/yomo"
)

var (
	vm      *wasmedge.VM
	vmConf  *wasmedge.Configure
	counter uint64
)

const ImageDataKey = 0x10

func main() {
	// Initialize WasmEdge's VM
	initVM()
	defer vm.Delete()
	defer vmConf.Delete()

	// Connect to Zipper service
	sfn := yomo.NewStreamFunction("image-recognition", yomo.WithZipperAddr("localhost:9900"))
	defer sfn.Close()

	// set only monitoring data
	sfn.SetObserveDataID(ImageDataKey)

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

// Handler process the data in the stream
func Handler(img []byte) (byte, []byte) {
	// recognize the image
	res, err := vm.ExecuteBindgen("infer", wasmedge.Bindgen_return_array, img)
	if err == nil {
		fmt.Println("GO: Run bindgen -- infer:", string(res.([]byte)))
	} else {
		fmt.Println("GO: Run bindgen -- infer FAILED")
	}

	// print logs
	hash := genSha1(img)
	log.Printf("✅ received image-%d hash %v, img_size=%d \n", atomic.AddUint64(&counter, 1), hash, len(img))

	return 0x11, nil
}

// genSha1 generate the hash value of the image
func genSha1(buf []byte) string {
	h := sha1.New()
	h.Write(buf)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// initVM initialize WasmEdge's VM
func initVM() {
	wasmedge.SetLogErrorLevel()
	/// Set Tensorflow not to print debug info
	os.Setenv("TF_CPP_MIN_LOG_LEVEL", "3")
	os.Setenv("TF_CPP_MIN_VLOG_LEVEL", "3")

	/// Create configure
	vmConf = wasmedge.NewConfigure(wasmedge.WASI)

	/// Create VM with configure
	vm = wasmedge.NewVMWithConfig(vmConf)

	/// Init WASI
	var wasi = vm.GetImportObject(wasmedge.WASI)
	wasi.InitWasi(
		os.Args[1:],     /// The args
		os.Environ(),    /// The envs
		[]string{".:."}, /// The mapping directories
		[]string{},      /// The preopens will be empty
	)

	/// Register WasmEdge-tensorflow and WasmEdge-image
	var tfobj = wasmedge.NewTensorflowImportObject()
	var tfliteobj = wasmedge.NewTensorflowLiteImportObject()
	vm.RegisterImport(tfobj)
	vm.RegisterImport(tfliteobj)
	var imgobj = wasmedge.NewImageImportObject()
	vm.RegisterImport(imgobj)

	/// Instantiate wasm
	vm.LoadWasmFile("rust_mobilenet_food_lib_bg.wasm")
	vm.Validate()
	vm.Instantiate()
}
