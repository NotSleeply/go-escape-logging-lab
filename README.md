# Go logging escape lab

本项目观察 `zap`、`log/slog`、标准库 `log` 在真实日志路径上的内存分配与逃逸表现。演示入口为根目录的 `main.go`。它不是吞吐量排行榜；每个 API 的语义不同，结果应结合调用形式解读。

## 设计边界

- Benchmark 使用三者的原生 API，不使用统一日志门面，避免门面自身改变逃逸结果。
- 三者都处于 `Info` 启用状态，并经不透明写入器统一转发到 `io.Discard`。这样可避免标准库 `log` 识别直接的 `io.Discard` 后跳过格式化；编码、参数构造和格式化仍会发生，终端与磁盘 I/O 不参与测量。
- `zap` 的静态字段场景使用强类型 `zap.Field`；动态键值场景使用原生 `SugaredLogger.Infow`，与 `slog.Info` 的键值参数并列观察。
- 每个场景均测量 `ns/op`、`B/op`、`allocs/op`，并用编译器的 `-m=2` 输出核对「何处逃逸、为什么逃逸」。

## 场景

- `BenchmarkPrimitiveFields`：字符串、整数、`time.Duration` 等标量字段。
- `BenchmarkStructuredValue`：结构体按值传入通用字段。
- `BenchmarkStructuredPointer`：结构体指针传入通用字段。
- `BenchmarkErrorField`：`error` 作为日志字段。
- `BenchmarkDynamicArguments`：动态键值参数与格式化参数。

## 运行

要求 Go 1.24.3 或更高版本。

```powershell
go mod download
go run ./cmd/demo
go test ./...
go test -run '^$' -bench . -benchmem -count=5
```

查看本项目代码的编译器逃逸判断：

```powershell
go test -run '^$' -gcflags='go-escape-logging-lab=-m=2' 2>&1 |
  Select-String 'escapes to heap|moved to heap|does not escape'
```

## 解读结果

`allocs/op` 是 Benchmark 中实际发生的堆分配，最适合比较调用成本。`-m=2` 是编译器的静态判断，用来定位对应变量是否因接口装箱、反射、可变参数或被调用方保留引用而逃逸。

不要把不同场景的绝对数值横向排名。例如 `zap.Any`、`slog.Any` 和 `log.Printf` 对结构体采用了不同序列化或格式化路径；它们说明的是各自原生写法的成本。优先比较同一个 Benchmark 分组内的三项结果，并以同一台机器上的多次运行中位趋势为准。
