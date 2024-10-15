package gogogo

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/AIntelligenceGame/clicktail/libclick"
	"github.com/AIntelligenceGame/clicktail/options/globals"
	"github.com/AIntelligenceGame/clicktail/run"
	"github.com/AIntelligenceGame/clicktail/tail"
	"github.com/sirupsen/logrus"
)

/*
globals.GlobalOptions 结构体定义了点击代码中的全局选项参数，用于配置点击工具的行为。以下是 globals.GlobalOptions 结构体中各个参数的含义和用法：

APIHost: 设置 ClickHouse 服务器的 API 地址，点击工具将向该地址发送数据。

TailSample: 标志位，指示是否对日志文件进行采样处理。

SampleRate: 设置事件采样率，控制发送到 ClickHouse 的事件数量。

NumSenders: 设置并发发送事件到 ClickHouse 的发送者数量。

Reqs: 包含必需选项的结构体，包括解析器名称、日志文件路径和数据集名称等信息。

Tail: 包含有关日志文件尾随配置的选项，如从哪里开始读取、是否停止等。

Debug: 标志位，指示是否启用调试模式。

Backfill: 标志位，表示是否启用回填功能。

Localtime: 标志位，指示是否使用本地时间作为时区信息。

Timezone: 设置时区信息，根据需要设置特定时区。

AddFields: 添加字段到事件中以丰富数据内容。

DropFields: 从事件中删除指定字段以过滤数据内容。

ScrubFields: 对指定字段进行脱敏处理（例如：生成 SHA256 哈希值）。

RequestShape: 配置请求形状相关选项，用于解析 HTTP 请求内容并添加额外字段信息。

ShapePrefix: 请求形状前缀，在添加请求形状相关字段时使用的前缀字符串。

RequestPattern: 配置请求模式匹配规则列表，用于解析请求路径和查询字符串等内容。

其他参数如：DynSample,  GoalSampleRate,  MinSampleRate,  PrefixRegex,  RequestParseQuery,  RequestQueryKeys, 等都是根据实际需求设置的参数。这些参数可以影响事件处理、动态采样、正则表达式匹配等功能。

18.StatusInterval : 状态间隔时间

19.BackOff : 是否启用重试机制

20.NumSenders : 发送者数量

21.DebugMode : 调试模式级别

22.ConfigFile : 配置文件路径
*/

func ParseLogFile(filepath string, _APIHost, _Dataset, _database string) {
	options := globals.GlobalOptions{
		APIHost:    _APIHost, // 根据实际情况设置
		TailSample: true,     // 根据实际情况设置
		SampleRate: 1,        // 根据实际情况设置
		NumSenders: 10,       // 根据实际情况设置

		Reqs: globals.RequiredOptions{
			ParserName: _database, // 设置要使用的解析器名称（例如：mysql、postgresql等）
			LogFiles:   []string{filepath},
			Dataset:    _Dataset, // 设置数据集名称
		},

		Tail: tail.TailOptions{
			ReadFrom: "start", // 从文件开始处读取日志文件内容
			Stop:     false,   // 不停止读取日志文件内容
		},

		Debug:     false, // 根据实际情况设置
		Backfill:  false, // 根据实际情况设置
		Localtime: false, // 根据实际情况设置
		Timezone:  "",    // 根据实际情况设置

		AddFields:   []string{}, // 根据实际情况设置
		DropFields:  []string{}, // 根据实际情况设置
		ScrubFields: []string{"normalized_query"},

		RequestShape:   []string{},
		ShapePrefix:    "",
		RequestPattern: []string{},

		DynSample:      []string{},
		GoalSampleRate: 0,
		MinSampleRate:  0,

		PrefixRegex: "",

		RequestParseQuery: "whitelist",
		RequestQueryKeys:  []string{},

		StatusInterval: 60,

		BackOff: false,
	}

	rand.Seed(time.Now().UnixNano())

	if options.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	globals.SetVersionUserAgent(options.Backfill, options.Reqs.ParserName)
	globals.AddParserDefaultOptions(&options)
	globals.SanityCheckOptions(&options)

	if err := libclick.VerifyApiHost(libclick.Config{
		APIHost: options.APIHost,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to ClickHouse server:", err)
		os.Exit(1)
	}

	run.Run(options)
}
