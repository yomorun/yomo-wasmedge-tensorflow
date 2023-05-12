# Streaming Image Recognition by WebAssembly

[![Youtube: YoMo x WasmEdge](youtube.png)](https://youtu.be/E0ltsn6cLIU)

This project demonstrates how to process a video stream in real-time using
WebAssembly and apply a pre-trained
[food classification model](https://tfhub.dev/google/lite-model/aiy/vision/classifier/food_V1/1)
to each frame of the video in order to determine if food is present in that
frame, all by integrating [WasmEdge](https://github.com/WasmEdge/WasmEdge) into
[YoMo](https://github.com/yomorun/yomo) serverless.

Open-source projects that we used:

- Serverless stream processing framework [YoMo](https://github.com/yomorun/yomo)
- Integrate with [WasmEdge](https://github.com/WasmEdge/WasmEdge) to introduce
  WebAssembly, interop TensorFlow Lite model
- A deep learning model found on
  [TensorFlow Hub](https://tfhub.dev/google/lite-model/aiy/vision/classifier/food_V1/1);
  make sure to download `TFLite (aiy/vision/classifier/food_V1)`, which was
  created by Google

**Advantages:**

- ⚡️ **Low-latency**: Streaming data processing applications can now be done in
  far edge data centers thanks to YoMo's highly efficient network services
- 🔐 **Security**: WasmEdge runs the data processing function in a WebAssembly
  sandbox for isolation, safety, and hot deployment
- 🚀 **High Performance**: Compared with popular containers, such as Docker,
  WasmEdge can be up to 100x faster at startup and have a much smaller footprint
- 🎯 **Edge Devices**: As WasmEdge consumes much less resources than Docker, it
  is now possible to run data processing applications on edge devices

## Steps to run

### 1. Clone This Repository

```bash
$ git clone https://github.com/yomorun/yomo-wasmedge-tensorflow.git
```

### 2. Install YoMo CLI

```bash
$ curl -fsSL "https://get.yomo.run" | sh
$ yomo version
YoMo CLI version: v1.12.2
```

Details about `YoMo CLI` installation can be found
[here](https://github.com/yomorun/yomo).

### 3. Install WasmEdge Dependencies

#### Install WasmEdge with its [Tensorflow and image processing extensions](https://www.secondstate.io/articles/wasi-tensorflow/)

```bash
wget -qO- https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install.sh | bash -s -- -e all -p /usr/local
```

If you have any questions about installation, please refer to
[the official documentation](https://github.com/WasmEdge/WasmEdge/blob/master/docs/install.md).
Currently, this project works on Linux machines only.

#### Install video and image processing dependencies

```bash
$ sudo apt-get update
$ sudo apt-get install -y ffmpeg libjpeg-dev libpng-dev
```

### 4. Write your Streaming Serverless function

Write
[app.go](https://github.com/yomorun/yomo-wasmedge-tensorflow/blob/main/flow/app.go)
to integrate `WasmEdge-tensorflow`:

- Get `WasmEdge-go`:

  ```bash
  $ cd flow
  $ go get -u github.com/second-state/WasmEdge-go/wasmedge
  ```

- download the pretrained model file
  [lite-model_aiy_vision_classifier_food_V1_1.tflite](https://storage.googleapis.com/tfhub-lite-models/google/lite-model/aiy/vision/classifier/food_V1/1.tflite)，and
  store to `rust_mobilenet_food/src` directory:

  ```bash
  $ wget 'https://storage.googleapis.com/tfhub-lite-models/google/lite-model/aiy/vision/classifier/food_V1/1.tflite' -O ./rust_mobilenet_food/src/lite-model_aiy_vision_classifier_food_V1_1.tflite
  ```

- install [rustc and cargo](https://www.rust-lang.org/tools/install)

  ```bash
  $ curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
  $ export PATH=$PATH:$HOME/.cargo/bin
  $ rustc --version
  ```

- install wasm32-wasi target

  ```bash
  $ rustup target add wasm32-wasi
  ```

- compile wasm file

  ```bash
  $ cd rust_mobilenet_food
  $ cargo build --release --target wasm32-wasi
  # The output WASM will be `target/wasm32-wasi/release/rust_mobilenet_food_lib.wasm`.
  $ cd ..
  ```

- copy the compiled wasm file

  ```bash
  $ cp rust_mobilenet_food/target/wasm32-wasi/release/rust_mobilenet_food_lib.wasm .
  ```

- _optional_: compile wasm file to
  [AOT mode](https://wasmedge.org/book/en/quick_start/run_in_aot_mode.html)

  ```bash
  wasmedgec rust_mobilenet_food_lib.wasm rust_mobilenet_food_lib.so
  ```

### 5. Run YoMo Zipper Server

```bash
$ yomo serve -c ./zipper/config.yaml
```

### 6. Run Streaming Serverless function

```bash
$ cd flow
$ go run -tags tensorflow,image app.go
```

### 7. Demonstrate video stream

Download
[this demo video: hot-dog.mp4](https://github.com/yomorun/yomo-wasmedge-tensorflow/releases/download/v0.2.0/hot-dog.mp4),
and store to `source` directory：

```bash
$ wget -P source 'https://github.com/yomorun/yomo-wasmedge-tensorflow/releases/download/v0.2.0/hot-dog.mp4'
```

then run：

```bash
$ go run ./source/main.go ./source/hot-dog.mp4
```

### 8. Result

![YoMo-WasmEdge](result.png)
