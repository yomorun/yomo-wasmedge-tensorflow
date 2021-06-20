# Streaming Image Recognition by WebAssembly

该项目是一个Show Case，展示如何借助WebAssembly技术，实时解析视频流，并将每一帧的图片调用深度学习模型，判断该帧中是否存在食物。

项目使用的相关技术：

- 流式计算框架是使用[YoMo Streaming Serverless Framework](https://github.com/yomorun/yomo)构建
- Serverless通过[WasmEdge](https://github.com/WasmEdge/WasmEdge)引入WebAssembly，执行深度学习模型
- **TODO** 深度学习模型来自于 

该Show Case体现的价值

- 低时延传输使得计算机视觉AI可以被推至数据中心处理
- WasmEdge能为ARM架构做深度优化（**TODO** 我们先多想几点）

## 如何运行

### 1. Clone Repository

```bash
$ git clone git@github.com:yomorun/yomo-wasmedge-tensorflow.git
```

### 2. 安装YoMo CLI

```bash
$ go install github.com/yomorun/cli/yomo@latest
```

执行下面的命令，确保yomo已经在环境变量中，有任何问题请参考[YoMo的详细文档](https://github.com/yomorun/yomo)

```bash
$ yomo version
YoMo CLI version: v0.0.4
```

当然也可以直接下载可执行文件: [Linux](https://github.com/yomorun/yomo-app-image-recognition-example/releases/download/v0.1.0/yomo)

### 3. 安装相关依赖

#### 安装WasmEdge

直接下载WasmEdge的共享库如下，或者通过[源码安装](https://github.com/second-state/WasmEdge-go#option-1-build-from-the-source)。

```bash
$ wget https://github.com/WasmEdge/WasmEdge/releases/download/0.8.0/WasmEdge-0.8.0-manylinux2014_x86_64.tar.gz
$ tar -xzf WasmEdge-0.8.0-manylinux2014_x86_64.tar.gz
$ sudo cp WasmEdge-0.8.0-Linux/include/wasmedge.h /usr/local/include
$ sudo cp WasmEdge-0.8.0-Linux/lib64/libwasmedge_c.so /usr/local/lib
$ sudo ldconfig
```

#### 安装WasmEdge-tensorflow

为manylinux2014平台安装预建的tensorflow依赖项：

```bash
$ wget https://github.com/second-state/WasmEdge-tensorflow-deps/releases/download/0.8.0/WasmEdge-tensorflow-deps-TF-0.8.0-manylinux2014_x86_64.tar.gz
$ wget https://github.com/second-state/WasmEdge-tensorflow-deps/releases/download/0.8.0/WasmEdge-tensorflow-deps-TFLite-0.8.0-manylinux2014_x86_64.tar.gz
$ sudo tar -C /usr/local/lib -xzf WasmEdge-tensorflow-deps-TF-0.8.0-manylinux2014_x86_64.tar.gz
$ sudo tar -C /usr/local/lib -xzf WasmEdge-tensorflow-deps-TFLite-0.8.0-manylinux2014_x86_64.tar.gz
$ sudo ln -sf libtensorflow.so.2.4.0 /usr/local/lib/libtensorflow.so.2
$ sudo ln -sf libtensorflow.so.2 /usr/local/lib/libtensorflow.so
$ sudo ln -sf libtensorflow_framework.so.2.4.0 /usr/local/lib/libtensorflow_framework.so.2
$ sudo ln -sf libtensorflow_framework.so.2 /usr/local/lib/libtensorflow_framework.so
$ sudo ldconfig
```

安装WasmEdge-tensorflow：

```bash
$ wget https://github.com/second-state/WasmEdge-tensorflow/releases/download/0.8.0/WasmEdge-tensorflow-0.8.0-manylinux2014_x86_64.tar.gz
$ wget https://github.com/second-state/WasmEdge-tensorflow/releases/download/0.8.0/WasmEdge-tensorflowlite-0.8.0-manylinux2014_x86_64.tar.gz
$ sudo tar -C /usr/local/ -xzf WasmEdge-tensorflow-0.8.0-manylinux2014_x86_64.tar.gz
$ sudo tar -C /usr/local/ -xzf WasmEdge-tensorflowlite-0.8.0-manylinux2014_x86_64.tar.gz
$ sudo ldconfig
```

详细的安装，请参考[官方文档](https://github.com/second-state/WasmEdge-go#wasmedge-tensorflow-extension)，目前该例子仅支持Linux下运行。

#### 安装ffmpeg

YoMo的视频处理依赖[ffmpeg](https://www.ffmpeg.org/)组件，安装方式如下:

```bash
$ sudo apt-get update
$ sudo apt-get install -y ffmpeg
```

### 4. 编写 Streaming Serverless

如何开发一个 serverless app？请参考官方例子：[Create your serverless app](https://github.com/yomorun/yomo#2-create-your-serverless-app)，这里为集成WasmEdge-tensorflow提供了一个例子 [app.go](https://github.com/yomorun/yomo-wasmedge-image-recognition/blob/main/flow/app.go)。简单描述步骤如下：

安装wasmedge-go：

```bash
$ cd flow
$ go get -u github.com/second-state/WasmEdge-go/wasmedge
```

下载训练好的模型文件[mobilenet_v1_192res_1.0_seefood.pb](https://github.com/yomorun/yomo-wasmedge-image-recognition/releases/download/v0.1.0/mobilenet_v1_192res_1.0_seefood.pb)，并放置在目录`rust_mobilenet_foods/src`中：

```bash
$ wget 'https://github.com/yomorun/yomo-wasmedge-image-recognition/releases/download/v0.1.0/mobilenet_v1_192res_1.0_seefood.pb' -o rust_mobilenet_food/src/mobilenet_v1_192res_1.0_seefood.pb
```

编译wasm文件，需要先安装[rustwasmc](https://github.com/second-state/rustwasmc)，首先要确保使用`Rust 1.50.0`版本：

```bash
$ curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
$ export PATH=$PATH:$HOME/.cargo/bin
$ rustc --version
```

设置默认的`rustup`版本为`1.50.0`: 

`$ rustup default 1.50.0`

```bash
$ curl https://raw.githubusercontent.com/second-state/rustwasmc/master/installer/init.sh -sSf | sh
$ cd rust_mobilenet_food
$ rustwasmc build
# The output WASM will be `pkg/rust_mobilenet_food_lib_bg.wasm`.
```

也可以直接下载我们编译好的[rust_mobilenet_food_lib_bg.wasm](https://github.com/yomorun/yomo-wasmedge-image-recognition/releases/download/v0.1.0/rust_mobilenet_food_lib_bg.wasm)文件。

拷贝`pkg/rust_mobilenet_food_lib_bg.wasm`到`flow`目录

```bash
$ cp pkg/rust_mobilenet_food_lib_bg.wasm ../.
```

### 5. 运行YoMo Streaming Orchestrator

```bash
  $ yomo serve -c ./zipper/workflow.yaml
```

### 6. 运行 Streaming Serverless

```bash
$ cd flow
$ go run --tags tensorflow app.go
```

### 7. 模拟视频流并查看运行结果

下载视频文件: [hot-dog.mp4](https://github.com/yomorun/yomo-wasmedge-image-recognition/releases/download/v0.1.0/hot-dog.mp4)，并保存到`source`目录，运行：

```bash
$ wget 'https://github.com/yomorun/yomo-wasmedge-image-recognition/releases/download/v0.1.0/hot-dog.mp4' -o ./source/hot-dog.mp4
$ go run ./source/main.go ./source/hot-dog.mp4
```

### 8. 查看结果

**TODO** 执行结果截图


`$ rustup default 1.50.0`
