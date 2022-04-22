# Streaming Image Recognition by WebAssembly

该项目是一个 Show Case，展示如何借助 WebAssembly 技术，实时解析视频流，并将每一帧的图片调用深度学习模型，判断该帧中是否存在食物。

项目使用的相关技术：

- 流式计算框架是使用[YoMo Streaming Serverless Framework](https://github.com/yomorun/yomo)构建
- Streaming Serverless Function 中通过 [WasmEdge](https://github.com/WasmEdge/WasmEdge)引入 WebAssembly，运行深度学习模型
- 深度学习模型来自于[TensorFlow Hub 上 Google 训练的aiy/vision/classifier/food_V1](https://tfhub.dev/google/lite-model/aiy/vision/classifier/food_V1/1) 

该Show Case的价值:

- 更快的网络：YoMo 的低时延传输使得计算机视觉AI可以被推至就近数据中心处理
- 更好的安全：在软件沙箱 WasmEdge 中处理外部用户提交的代码，保障安全
- 更低的Overhead：与 Docker 等流行的应用程序容器相比，WebAssembly 提供了更高级别的抽象，可以即时启动
- 为边缘计算优化：YoMo 配合高性能、轻量级的 Wasm 虚拟机，适合资源受限的边缘设备

## 如何运行

### 1. Clone Repository

```bash
$ git clone https://github.com/yomorun/yomo-wasmedge-tensorflow.git
```

### 2. 安装YoMo CLI

```bash
$ go install github.com/yomorun/cli/yomo@latest
```

执行下面的命令，确保yomo已经在环境变量中，有任何问题请参考 [YoMo 的详细文档](https://github.com/yomorun/yomo)

```bash
$ yomo version
YoMo CLI version: v1.0.2
```

### 3. 安装 WasmEdge 及相关依赖

```bash
wget -qO- https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install.sh | bash -s -- -e all -p /usr/local
```

安装视频和图片处理依赖组件：

```bash
$ sudo apt-get update
$ sudo apt-get install -y ffmpeg libjpeg-dev libpng-dev
```

### 4. 编写 Streaming Serverless

如何开发一个 serverless app？请参考官方例子：[Create your serverless app](https://github.com/yomorun/yomo#2-create-your-serverless-app)，这里为集成 WasmEdge-tensorflow提供了一个例子 [app.go](https://github.com/yomorun/yomo-wasmedge-image-recognition/blob/main/flow/app.go)。简单描述步骤如下：

- 拉取依赖包：

```bash
$ cd flow
$ go get -u github.com/second-state/WasmEdge-go/wasmedge
```

- 下载训练好的模型文件 [lite-model_aiy_vision_classifier_food_V1_1.tflite](https://storage.googleapis.com/tfhub-lite-models/google/lite-model/aiy/vision/classifier/food_V1/1.tflite)，并放置在目录 `rust_mobilenet_foods/src` 中：

```bash
$ wget 'https://storage.googleapis.com/tfhub-lite-models/google/lite-model/aiy/vision/classifier/food_V1/1.tflite' -O ./rust_mobilenet_food/src/lite-model_aiy_vision_classifier_food_V1_1.tflite
```
- 编译 wasm 文件：

安装 [rustc and cargo](https://www.rust-lang.org/tools/install)

```bash
$ curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
$ export PATH=$PATH:$HOME/.cargo/bin
$ rustc --version
```

设置默认的`rust`版本为`1.50.0`: `$ rustup default 1.50.0`

安装 [rustwasmc](https://github.com/second-state/rustwasmc)

```bash
$ curl https://raw.githubusercontent.com/second-state/rustwasmc/master/installer/init.sh -sSf | sh
$ cd rust_mobilenet_food
$ rustwasmc build
# The output WASM will be `pkg/rust_mobilenet_food_lib_bg.wasm`.
```

拷贝`pkg/rust_mobilenet_food_lib_bg.wasm`到`flow`目录：

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
$ go run -tags tensorflow,image app.go
```

### 7. 模拟视频流并查看运行结果

下载视频文件: [hot-dog.mp4](https://github.com/yomorun/yomo-wasmedge-image-recognition/releases/download/v0.1.0/hot-dog.mp4)，并保存到`source`目录，运行：

```bash
$ wget -P source 'https://github.com/yomorun/yomo-wasmedge-tensorflow/releases/download/v0.1.0/hot-dog.mp4'
$ go run ./source/main.go ./source/hot-dog.mp4
```

### 8. 查看结果

![YoMo-WasmEdge](result.png)
