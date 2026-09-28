# Go logging escape lab

本项目观察 `zap`、`log/slog`、标准库 `log` 在真实日志路径上的内存分配与逃逸表现。

演示入口为根目录的 `main.go`。

`go run .` 会让每种日志系统记录 100 万次相同字段，输出总耗时和平均 `ns/op`。计时阶段最终写入 `io.Discard`，不会被终端 I/O 主导。

每个 API 的语义不同，结果应结合调用形式解读。

## 设计边界

- Benchmark 使用三者的原生 API，不使用统一日志门面，避免门面自身改变逃逸结果。
- 三者输出相同的 `msg`、`request_id`、`path`、`status` JSON 字段，并统一关闭时间和级别字段，避免纯文本与 JSON、额外元数据造成不等价比较。
- 标准库 `log` 没有结构化 JSON 能力，因此使用 `encoding/json` 编码后交给 `log.Logger` 输出；这部分成本属于标准库方案完成同等任务所必需的成本。
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
go test -run '^$' -bench . -benchmem -count=5
```

查看本项目代码的编译器逃逸判断：

```powershell
go test -run '^$' -gcflags='go-escape-logging-lab=-m=2' 2>&1 |
  Select-String 'logging_benchmark_test\.go:' |
  Select-String 'request|record|\.\.\. argument|map\[string\]any|slog\.Any' |
  Select-String 'escapes to heap|moved to heap|does not escape'
```

## 解读结果

### `go run .`

```bash
等价 JSON 日志耗时：每种日志系统记录 1000000 次
zap：总耗时 283.8869ms，平均 283.9 ns/op
slog：总耗时 595.0523ms，平均 595.1 ns/op
log+encoding/json：总耗时 396.0738ms，平均 396.1 ns/op
```

### `go test -run '^$' -bench . -benchmem -count=5`

```bash
goos: windows
goarch: amd64
pkg: go-escape-logging-lab
cpu: 12th Gen Intel(R) Core(TM) i9-12900H
BenchmarkPrimitiveFields/zap_typed-20            4015316               302.1 ns/op           192 B/op          1 allocs/op
BenchmarkPrimitiveFields/zap_typed-20            3975543               295.2 ns/op           192 B/op          1 allocs/op
BenchmarkPrimitiveFields/zap_typed-20            3933627               291.9 ns/op           192 B/op          1 allocs/op
BenchmarkPrimitiveFields/zap_typed-20            4231836               293.5 ns/op           192 B/op          1 allocs/op
BenchmarkPrimitiveFields/zap_typed-20            4223890               296.6 ns/op           192 B/op          1 allocs/op
BenchmarkPrimitiveFields/slog_attrs-20           1854670               625.7 ns/op             0 B/op          0 allocs/op
BenchmarkPrimitiveFields/slog_attrs-20           1998180               600.0 ns/op             0 B/op          0 allocs/op
BenchmarkPrimitiveFields/slog_attrs-20           2021539               596.6 ns/op             0 B/op          0 allocs/op
BenchmarkPrimitiveFields/slog_attrs-20           2006443               600.5 ns/op             0 B/op          0 allocs/op
BenchmarkPrimitiveFields/slog_attrs-20           2016734               600.3 ns/op             0 B/op          0 allocs/op
BenchmarkPrimitiveFields/log_json-20             3295974               362.3 ns/op           272 B/op          4 allocs/op
BenchmarkPrimitiveFields/log_json-20             3417949               350.3 ns/op           272 B/op          4 allocs/op
BenchmarkPrimitiveFields/log_json-20             3380252               356.2 ns/op           272 B/op          4 allocs/op
BenchmarkPrimitiveFields/log_json-20             3453411               350.9 ns/op           272 B/op          4 allocs/op
BenchmarkPrimitiveFields/log_json-20             3377373               355.1 ns/op           272 B/op          4 allocs/op
BenchmarkStructuredValue/zap_any-20              2513391               467.2 ns/op           208 B/op          3 allocs/op
BenchmarkStructuredValue/zap_any-20              2534437               469.0 ns/op           208 B/op          3 allocs/op
BenchmarkStructuredValue/zap_any-20              2591865               459.3 ns/op           208 B/op          3 allocs/op
BenchmarkStructuredValue/zap_any-20              2552637               465.2 ns/op           208 B/op          3 allocs/op
BenchmarkStructuredValue/zap_any-20              2622271               461.4 ns/op           208 B/op          3 allocs/op
BenchmarkStructuredValue/slog_any-20             1469733               820.2 ns/op           208 B/op          4 allocs/op
BenchmarkStructuredValue/slog_any-20             1474182               811.1 ns/op           208 B/op          4 allocs/op
BenchmarkStructuredValue/slog_any-20             1466842               837.2 ns/op           208 B/op          4 allocs/op
BenchmarkStructuredValue/slog_any-20             1440254               836.8 ns/op           208 B/op          4 allocs/op
BenchmarkStructuredValue/slog_any-20             1452955               825.5 ns/op           208 B/op          4 allocs/op
BenchmarkStructuredValue/log_json-20             3125858               381.6 ns/op           272 B/op          4 allocs/op
BenchmarkStructuredValue/log_json-20             3113824               385.3 ns/op           272 B/op          4 allocs/op
BenchmarkStructuredValue/log_json-20             3162194               387.7 ns/op           272 B/op          4 allocs/op
BenchmarkStructuredValue/log_json-20             3126900               401.3 ns/op           272 B/op          4 allocs/op
BenchmarkStructuredValue/log_json-20             2638108               428.8 ns/op           272 B/op          4 allocs/op
BenchmarkStructuredPointer/zap_any-20            2388843               494.8 ns/op           208 B/op          3 allocs/op
BenchmarkStructuredPointer/zap_any-20            2422657               502.6 ns/op           208 B/op          3 allocs/op
BenchmarkStructuredPointer/zap_any-20            2436566               490.7 ns/op           208 B/op          3 allocs/op
BenchmarkStructuredPointer/zap_any-20            2156502               539.6 ns/op           208 B/op          3 allocs/op
BenchmarkStructuredPointer/zap_any-20            2372210               502.3 ns/op           208 B/op          3 allocs/op
BenchmarkStructuredPointer/slog_any-20           1406115               861.7 ns/op           208 B/op          4 allocs/op
BenchmarkStructuredPointer/slog_any-20           1353315               876.6 ns/op           208 B/op          4 allocs/op
BenchmarkStructuredPointer/slog_any-20           1416532               846.6 ns/op           208 B/op          4 allocs/op
BenchmarkStructuredPointer/slog_any-20           1401900               860.9 ns/op           208 B/op          4 allocs/op
BenchmarkStructuredPointer/slog_any-20           1396575               860.6 ns/op           208 B/op          4 allocs/op
BenchmarkStructuredPointer/log_json-20           2672620               447.7 ns/op           280 B/op          5 allocs/op
BenchmarkStructuredPointer/log_json-20           2832130               444.9 ns/op           280 B/op          5 allocs/op
BenchmarkStructuredPointer/log_json-20           2698550               443.6 ns/op           280 B/op          5 allocs/op
BenchmarkStructuredPointer/log_json-20           2750589               436.5 ns/op           280 B/op          5 allocs/op
BenchmarkStructuredPointer/log_json-20           2767726               441.6 ns/op           280 B/op          5 allocs/op
BenchmarkErrorField/zap_error-20                 5524418               215.1 ns/op            64 B/op          1 allocs/op
BenchmarkErrorField/zap_error-20                 5593867               213.2 ns/op            64 B/op          1 allocs/op
BenchmarkErrorField/zap_error-20                 5557293               216.3 ns/op            64 B/op          1 allocs/op
BenchmarkErrorField/zap_error-20                 5459619               213.5 ns/op            64 B/op          1 allocs/op
BenchmarkErrorField/zap_error-20                 5707382               213.4 ns/op            64 B/op          1 allocs/op
BenchmarkErrorField/slog_any-20                  2120788               563.5 ns/op            48 B/op          1 allocs/op
BenchmarkErrorField/slog_any-20                  2116220               575.4 ns/op            48 B/op          1 allocs/op
BenchmarkErrorField/slog_any-20                  2148348               561.9 ns/op            48 B/op          1 allocs/op
BenchmarkErrorField/slog_any-20                  2105664               573.9 ns/op            48 B/op          1 allocs/op
BenchmarkErrorField/slog_any-20                  2129208               568.1 ns/op            48 B/op          1 allocs/op
BenchmarkErrorField/log_json-20                  4061157               302.8 ns/op           176 B/op          4 allocs/op
BenchmarkErrorField/log_json-20                  4041626               307.8 ns/op           176 B/op          4 allocs/op
BenchmarkErrorField/log_json-20                  4011074               301.2 ns/op           176 B/op          4 allocs/op
BenchmarkErrorField/log_json-20                  3937196               305.8 ns/op           176 B/op          4 allocs/op
BenchmarkErrorField/log_json-20                  3972204               304.0 ns/op           176 B/op          4 allocs/op
BenchmarkDynamicArguments/zap_sugared-20         2525956               483.4 ns/op           417 B/op          3 allocs/op
BenchmarkDynamicArguments/zap_sugared-20         2577637               477.4 ns/op           417 B/op          3 allocs/op
BenchmarkDynamicArguments/zap_sugared-20         2437898               487.4 ns/op           417 B/op          3 allocs/op
BenchmarkDynamicArguments/zap_sugared-20         2478033               484.7 ns/op           417 B/op          3 allocs/op
BenchmarkDynamicArguments/zap_sugared-20         2488918               495.4 ns/op           417 B/op          3 allocs/op
BenchmarkDynamicArguments/slog_key_values-20     1673284               728.6 ns/op            32 B/op          2 allocs/op
BenchmarkDynamicArguments/slog_key_values-20     1671285               725.1 ns/op            32 B/op          2 allocs/op
BenchmarkDynamicArguments/slog_key_values-20     1683010               716.3 ns/op            32 B/op          2 allocs/op
BenchmarkDynamicArguments/slog_key_values-20     1651778               730.1 ns/op            32 B/op          2 allocs/op
BenchmarkDynamicArguments/slog_key_values-20     1664054               729.3 ns/op            32 B/op          2 allocs/op
BenchmarkDynamicArguments/log_json_map-20        1000000              1091 ns/op             866 B/op         16 allocs/op
BenchmarkDynamicArguments/log_json_map-20        1000000              1087 ns/op             866 B/op         16 allocs/op
BenchmarkDynamicArguments/log_json_map-20        1000000              1098 ns/op             866 B/op         16 allocs/op
BenchmarkDynamicArguments/log_json_map-20        1000000              1088 ns/op             866 B/op         16 allocs/op
BenchmarkDynamicArguments/log_json_map-20        1000000              1083 ns/op             866 B/op         16 allocs/op
PASS
ok      go-escape-logging-lab   89.793s
```
> zap 整体耗时较低，slog 堆分配最少，log 在动态参数场景分配最多。

**怎么看：**
- `ns/op`：单次耗时，越低越快。
- `B/op`：单次堆分配字节数，越低越好。
- `allocs/op`：单次堆分配次数，越低越好。
- `-count=5`：每项重复五次，应看五次的中间趋势。
- 名称后的 `-20` 表示使用了 20 个逻辑处理器，不是运行次数。


**具体观察：**
- 标量字段：zap 约 295 ns/op，最快；slog 为 0 B/op、0 allocs/op，完全没有堆分配，但耗时约 600 ns/op。
- 结构体值：log 约 385 ns/op，略快于 zap 的 465 ns/op；slog 约 825 ns/op。
- 结构体指针：log 从值类型的 272 B、4 次分配 上升到 280 B、5 次分配，说明指针路径增加了一次堆分配。
- 错误字段：zap 约 214 ns/op，最快；log 约 304 ns/op；slog 约 568 ns/op。
- 动态参数：log 使用 map + encoding/json 后达到 866 B/op、16 allocs/op，分配最多；zap 约 484 ns/op，速度最快。

### 编译器逃逸判断

```bash
go test -run '^$' -gcflags='go-escape-logging-lab=-m=2' 2>&1 |
  Select-String 'logging_benchmark_test\.go:' |
  Select-String 'request|record|\.\.\. argument|map\[string\]any|slog\.Any' |
  Select-String 'escapes to heap|moved to heap|does not escape'
```

示例输出已删除编译器重复提示，`#` 后是对应解释：

```bash
# 标量字段
./logging_benchmark_test.go:52:15: ... argument escapes to heap          # zap 的 []zap.Field 可变参数底层数组逃逸到堆
./logging_benchmark_test.go:65:19: ... argument does not escape          # slog 的 []slog.Attr 可变参数切片留在栈上
./logging_benchmark_test.go:84:25: record escapes to heap                # log 的 JSON 记录装入 any 接口后逃逸到堆

# 结构体值
./logging_benchmark_test.go:96:15: ... argument escapes to heap          # zap 的 []zap.Field 可变参数底层数组逃逸到堆
./logging_benchmark_test.go:96:56: request escapes to heap               # zap.Any 接收的 Request 结构体值逃逸到堆
./logging_benchmark_test.go:106:15: ... argument does not escape         # slog 的 []slog.Attr 可变参数切片本身没有逃逸
./logging_benchmark_test.go:106:45: slog.Any("request", request) escapes to heap # slog.Any 生成的属性值逃逸到堆
./logging_benchmark_test.go:106:57: request escapes to heap              # slog.Any 接收的 Request 结构体值逃逸到堆
./logging_benchmark_test.go:116:46: structuredValueRecord{...} escapes to heap # log 的结构体值记录装入 any 接口后逃逸到堆

# 结构体指针
./logging_benchmark_test.go:130:4: moved to heap: request                # zap 接收 &request，局部 request 被移动到堆
./logging_benchmark_test.go:131:15: ... argument escapes to heap         # zap 的 []zap.Field 可变参数底层数组逃逸到堆
./logging_benchmark_test.go:140:4: moved to heap: request                # slog 接收 &request，局部 request 被移动到堆
./logging_benchmark_test.go:141:15: ... argument does not escape         # slog 的 []slog.Attr 可变参数切片本身没有逃逸
./logging_benchmark_test.go:141:45: slog.Any("request", &request) escapes to heap # slog.Any 生成的指针属性值逃逸到堆
./logging_benchmark_test.go:150:4: moved to heap: request                # log 记录保存 &request，局部 request 被移动到堆
./logging_benchmark_test.go:151:48: structuredPointerRecord{...} escapes to heap # log 的指针记录装入 any 接口后逃逸到堆

# 错误字段
./logging_benchmark_test.go:165:15: ... argument escapes to heap         # zap 的 []zap.Field 可变参数底层数组逃逸到堆
./logging_benchmark_test.go:174:15: ... argument does not escape         # slog 的 []slog.Attr 可变参数切片本身没有逃逸
./logging_benchmark_test.go:174:42: slog.Any("error", benchmarkError) escapes to heap # slog.Any 生成的错误属性值逃逸到堆
./logging_benchmark_test.go:183:36: errorRecord{...} escapes to heap     # log 的错误记录装入 any 接口后逃逸到堆

# 动态参数
./logging_benchmark_test.go:199:16: ... argument does not escape         # zap SugaredLogger 的可变参数切片本身没有逃逸
./logging_benchmark_test.go:200:5: "request_id" escapes to heap          # zap 动态键装入 any 接口后逃逸到堆
./logging_benchmark_test.go:200:26: request.ID escapes to heap           # zap 动态字段值装入 any 接口后逃逸到堆
./logging_benchmark_test.go:201:20: request.Path escapes to heap         # zap 动态字段值装入 any 接口后逃逸到堆
./logging_benchmark_test.go:202:22: request.Status escapes to heap       # zap 动态字段值装入 any 接口后逃逸到堆
./logging_benchmark_test.go:212:15: ... argument does not escape         # slog 的可变参数切片本身没有逃逸
./logging_benchmark_test.go:213:5: "request_id" escapes to heap          # slog 动态键装入 any 接口后逃逸到堆
./logging_benchmark_test.go:213:26: request.ID escapes to heap           # slog 动态字段值装入 any 接口后逃逸到堆
./logging_benchmark_test.go:214:20: request.Path escapes to heap         # slog 动态字段值装入 any 接口后逃逸到堆
./logging_benchmark_test.go:215:22: request.Status escapes to heap       # slog 动态字段值装入 any 接口后逃逸到堆
./logging_benchmark_test.go:225:39: map[string]any{...} escapes to heap   # log 的动态 map 本身逃逸到堆
./logging_benchmark_test.go:226:19: "request completed" escapes to heap   # log 的消息值装入 map 后逃逸到堆
./logging_benchmark_test.go:227:26: request.ID escapes to heap           # log 的字段值装入 map 后逃逸到堆
./logging_benchmark_test.go:228:26: request.Path escapes to heap         # log 的字段值装入 map 后逃逸到堆
./logging_benchmark_test.go:229:26: request.Status escapes to heap       # log 的字段值装入 map 后逃逸到堆
```
