package globals

import (
	"github.com/AIntelligenceGame/clicktail/parsers/arangodb"
	"github.com/AIntelligenceGame/clicktail/parsers/htjson"
	"github.com/AIntelligenceGame/clicktail/parsers/keyval"
	"github.com/AIntelligenceGame/clicktail/parsers/mongodb"
	"github.com/AIntelligenceGame/clicktail/parsers/mysql"
	"github.com/AIntelligenceGame/clicktail/parsers/mysqlaudit"
	"github.com/AIntelligenceGame/clicktail/parsers/nginx"
	"github.com/AIntelligenceGame/clicktail/parsers/postgresql"
	"github.com/AIntelligenceGame/clicktail/parsers/regex"
	"github.com/AIntelligenceGame/clicktail/tail"
)

// GlobalOptions has all the top level CLI flags that clicktail supports
type GlobalOptions struct {
	APIHost    string `long:"api_host" description:"Host of the ClickHouse server" default:"http://localhost:8123/"`
	TailSample bool   `hidden:"true" description:"When true, sample while tailing. When false, sample post-parser events"`

	ConfigFile string `short:"c" long:"config" description:"Config file for clicktail in INI format." no-ini:"true"`

	SampleRate       uint `short:"r" long:"samplerate" description:"Only send 1 / N log lines" default:"1"`
	NumSenders       uint `short:"P" long:"poolsize" description:"Number of concurrent connections to open to ClickHouse" default:"10"`
	BatchFrequencyMs uint `long:"send_frequency_ms" description:"How frequently to flush batches" default:"10000"`
	BatchSize        uint `long:"send_batch_size" description:"Maximum number of messages to put in a batch" default:"1000000"`
	Debug            bool `long:"debug" description:"Print debugging output"`
	StatusInterval   uint `long:"status_interval" description:"How frequently, in seconds, to print out summary info" default:"60"`
	Backfill         bool `long:"backfill" description:"Configure clicktail to ingest old data in order to backfill ClickHouse table. Sets the correct values for --backoff, --tail.read_from, and --tail.stop"`

	Localtime         bool     `long:"localtime" description:"When parsing a timestamp that has no time zone, assume it is in the same timezone as localhost instead of UTC (the default)"`
	Timezone          string   `long:"timezone" description:"When parsing a timestamp use this time zone instead of UTC (the default). Must be specified in TZ format as seen here: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones"`
	ScrubFields       []string `long:"scrub_field" description:"For the field listed, apply a one-way hash to the field content. May be specified multiple times"`
	DropFields        []string `long:"drop_field" description:"Do not send the field to ClickHouse. May be specified multiple times"`
	AddFields         []string `long:"add_field" description:"Add the field to every event. Field should be key=val. May be specified multiple times"`
	RequestShape      []string `long:"request_shape" description:"Identify a field that contains an HTTP request of the form 'METHOD /path HTTP/1.x' or just the request path. Break apart that field into subfields that contain components. May be specified multiple times. Defaults to 'request' when using the nginx parser"`
	ShapePrefix       string   `long:"shape_prefix" description:"Prefix to use on fields generated from request_shape to prevent field collision"`
	RequestPattern    []string `long:"request_pattern" description:"A pattern for the request path on which to base the derived request_shape. May be specified multiple times. Patterns are considered in order; first match wins."`
	RequestParseQuery string   `long:"request_parse_query" description:"How to parse the request query parameters. 'whitelist' means only extract listed query keys. 'all' means to extract all query parameters as individual columns" default:"whitelist"`
	RequestQueryKeys  []string `long:"request_query_keys" description:"Request query parameter key names to extract, when request_parse_query is 'whitelist'. May be specified multiple times."`
	BackOff           bool     `long:"backoff" description:"When rate limited by the API, back off and retry sending failed events. Otherwise failed events are dropped. When --backfill is set, it will override this option=true"`
	PrefixRegex       string   `long:"log_prefix" description:"pass a regex to this flag to strip the matching prefix from the line before handing to the parser. Useful when log aggregation prepends a line header. Use named groups to extract fields into the event."`
	DynSample         []string `long:"dynsampling" description:"enable dynamic sampling using the field listed in this option. May be specified multiple times; fields will be concatenated to form the dynsample key. WARNING increases CPU utilization dramatically over normal sampling"`
	DynWindowSec      int      `long:"dynsample_window" description:"measurement window size for the dynsampler, in seconds" default:"30"`
	GoalSampleRate    int      `hidden:"true" description:"used to hold the desired sample rate and set tailing sample rate to 1"`
	MinSampleRate     int      `long:"dynsample_minimum" description:"if the rate of traffic falls below this, dynsampler won't sample" default:"1"`

	Reqs  RequiredOptions `group:"Required Options"`
	Modes OtherModes      `group:"Other Modes"`

	Tail tail.TailOptions `group:"Tail Options" namespace:"tail"`

	ArangoDB   arangodb.Options   `group:"ArangoDB Parser Options" namespace:"arangodb"`
	JSON       htjson.Options     `group:"JSON Parser Options" namespace:"json"`
	KeyVal     keyval.Options     `group:"KeyVal Parser Options" namespace:"keyval"`
	Mongo      mongodb.Options    `group:"MongoDB Parser Options" namespace:"mongo"`
	MySQL      mysql.Options      `group:"MySQL Parser Options" namespace:"mysql"`
	MySQLAudit mysqlaudit.Options `group:"MySQL Audit Parser Options" namespace:"mysqlaudit"`
	Nginx      nginx.Options      `group:"Nginx Parser Options" namespace:"nginx"`
	PostgreSQL postgresql.Options `group:"PostgreSQL Parser Options" namespace:"postgresql"`
	Regex      regex.Options      `group:"Regex Parser Options" namespace:"regex"`
}
type RequiredOptions struct {
	ParserName string `short:"p" long:"parser" description:"Parser module to use. Use --list to list available options."`
	//WriteKey   string   `short:"k" long:"writekey" description:"Team write key"`
	LogFiles []string `short:"f" long:"file" description:"Log file(s) to parse. Use '-' for STDIN, use this flag multiple times to tail multiple files, or use a glob (/path/to/foo-*.log)"`
	Dataset  string   `short:"d" long:"dataset" description:"Name of the dataset"`
}

type OtherModes struct {
	Help               bool `short:"h" long:"help" description:"Show this help message"`
	ListParsers        bool `short:"l" long:"list" description:"List available parsers"`
	Version            bool `short:"V" long:"version" description:"Show version"`
	WriteDefaultConfig bool `long:"write_default_config" description:"Write a default config file to STDOUT" no-ini:"true"`
	WriteCurrentConfig bool `long:"write_current_config" description:"Write out the current config to STDOUT" no-ini:"true"`

	WriteManPage bool `hidden:"true" long:"write-man-page" description:"Write out a man page"`
}
