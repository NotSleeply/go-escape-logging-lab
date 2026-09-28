# Go logging escape lab

本项目观察 `zap`、`log/slog`、标准库 `log` 在真实日志路径上的内存分配与逃逸表现。

演示入口为根目录的 `main.go`。

`go run .` 会先各输出一条日志，再让每种日志器记录 100 万次相同的标量字段，输出总耗时和平均 `ns/op`。计时阶段最终写入 `io.Discard`，不会被终端 I/O 主导。

每个 API 的语义不同，结果应结合调用形式解读。

## 设计边界

- Benchmark 使用三者的原生 API，不使用统一日志门面，避免门面自身改变逃逸结果。
- 三者都处于 `Info` 启用状态，并经不透明写入器统一转发到 `io.Discard`。这样可避免标准库 `log` 识别直接的 `io.Discard` 后跳过格式化；编码、参数构造和格式化仍会发生，终端与磁盘 I/O 不参与测量。
- `zap` 的静态字段场景使用强类型 `zap.Field`；动态键值场景使用原生 `SugaredLogger.Infow`，与 `slog.Info` 的键值参数并列观察。
- 每个场景均测量 `ns/op`、`B/op`、`allocs/op`，并用编译器的 `-m=2` 输出核对「何处逃逸、为什么逃逸」。

## 场景

- `BenchmarkPrimitiveFields`：字符串、整数等标量字段。
- `BenchmarkStructuredValue`：结构体按值传入通用字段。
- `BenchmarkStructuredPointer`：结构体指针传入通用字段。
- `BenchmarkErrorField`：`error` 作为日志字段。
- `BenchmarkDynamicArguments`：动态键值参数与格式化参数。

## 运行

```powershell
go mod download
go run .
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

`go run .` 的手工计时用于先建立直觉，单次结果会受 CPU 调度、系统负载与运行顺序影响。需要下结论时，以 `go test -bench . -benchmem -count=5` 的多次结果为准，再结合 `-m=2` 定位逃逸原因。

不要把不同场景的绝对数值横向排名。例如 `zap.Any`、`slog.Any` 和 `log.Printf` 对结构体采用了不同序列化或格式化路径；它们说明的是各自原生写法的成本。优先比较同一个 Benchmark 分组内的三项结果，并以同一台机器上的多次运行中位趋势为准。
