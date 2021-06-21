# yomo-app-image-recognition-example



## Getting Started

### 1. Install CLI

```bash
$ go install github.com/yomorun/cli/yomo@latest
```

#### Verify if the CLI was installed successfully

```bash
$ yomo -v

YoMo CLI version: v0.0.1

```

You can also download the yomo file: [yomo](https://github.com/yomorun/yomo-app-image-recognition-example/releases/download/v0.1.0/yomo)



### 2. Zipper

#### Run

```bash
$ yomo serve -c ./zipper/workflow.yaml
```



### 3. Flow

#### Run

```bash
$ cd flow
$ go get -u github.com/second-state/WasmEdge-go/wasmedge
$ go run --tags tensorflow app.go
```

Should be run in LINUX.

### 4.Source

#### Run

```bash
$ go run ./source/main.go ./source/hot-dog.mp4
```

download the mp4 file: [hot-dog.mp4](https://github.com/yomorun/yomo-app-image-recognition-example/releases/download/v0.1.0/hot-dog.mp4) , and save to source directory.

