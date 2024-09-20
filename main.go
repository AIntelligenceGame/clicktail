package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AIntelligenceGame/clicktail/libclick"
	"github.com/AIntelligenceGame/clicktail/options/globals"
	"github.com/honeycombio/honeytail/httime"

	// "github.com/honeycombio/libhoney-go"
	//libclick "github.com/Altinity/libclick-go"
	flag "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

// BuildID is set by Travis CI
var BuildID string

// internal version identifier
var version string

var validParsers = []string{
	"arangodb",
	"json",
	"keyval",
	"mongo",
	"mysql",
	"mysqlaudit",
	"nginx",
	"postgresql",
	"regex",
}

func main() {
	var options globals.GlobalOptions
	flagParser := flag.NewParser(&options, flag.PrintErrors)
	flagParser.Usage = "-p <parser> -f </path/to/logfile> -d <mydata> [optional arguments]\n"

	if extraArgs, err := flagParser.Parse(); err != nil || len(extraArgs) != 0 {
		fmt.Println("Error: failed to parse the command line.")
		if err != nil {
			fmt.Printf("\t%s\n", err)
		} else {
			fmt.Printf("\tUnexpected extra arguments: %s\n", strings.Join(extraArgs, " "))
		}
		usage()
		os.Exit(1)
	}
	// read the config file if present
	if options.ConfigFile != "" {
		ini := flag.NewIniParser(flagParser)
		ini.ParseAsDefaults = true
		if err := ini.ParseFile(options.ConfigFile); err != nil {
			fmt.Printf("Error: failed to parse the config file %s\n", options.ConfigFile)
			fmt.Printf("\t%s\n", err)
			usage()
			os.Exit(1)
		}
	}

	rand.Seed(time.Now().UnixNano())

	if options.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Support flag alias: --backfill should cover --backoff --tail.read_from=beginning --tail.stop
	if options.Backfill {
		options.BackOff = true
		options.Tail.ReadFrom = "beginning"
		options.Tail.Stop = true
	}

	// set time zone info
	if options.Localtime {
		httime.Location = time.Now().Location()
	}
	if options.Timezone != "" {
		loc, err := time.LoadLocation(options.Timezone)
		if err != nil {
			fmt.Printf("time zone '%s' not successfully parsed.\n", options.Timezone)
			fmt.Printf("see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones for a list of time zones\n")
			fmt.Printf("expected format example: America/Los_Angeles\n")
			fmt.Printf("Specific error: %s\n", err.Error())
			os.Exit(1)
		}
		httime.Location = loc
	}

	setVersionUserAgent(options.Backfill, options.Reqs.ParserName)
	handleOtherModes(flagParser, options.Modes)
	addParserDefaultOptions(&options)
	sanityCheckOptions(&options)

	if err := libclick.VerifyApiHost(libclick.Config{
		APIHost: options.APIHost,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to ClickHouse server: ", err)
		os.Exit(1)
	}

	Run(options)
}

// setVersion sets the internal version ID and updates libclick's user-agent
func setVersionUserAgent(backfill bool, parserName string) {
	if BuildID == "" {
		version = "dev"
	} else {
		version = BuildID
	}
	if backfill {
		parserName += " backfill"
	}
	libclick.UserAgentAddition = fmt.Sprintf("clicktail/%s (%s)", version, parserName)
}

// handleOtherModes takse care of all flags that say we should just do something
// and exit rather than actually parsing logs
func handleOtherModes(fp *flag.Parser, modes globals.OtherModes) {
	if modes.Version {
		fmt.Println("Clicktail version", version)
		os.Exit(0)
	}
	if modes.Help {
		fp.WriteHelp(os.Stdout)
		fmt.Println("")
		os.Exit(0)
	}
	if modes.WriteManPage {
		fp.WriteManPage(os.Stdout)
		os.Exit(0)
	}
	if modes.WriteDefaultConfig {
		ip := flag.NewIniParser(fp)
		ip.Write(os.Stdout, flag.IniIncludeDefaults|flag.IniCommentDefaults|flag.IniIncludeComments)
		os.Exit(0)
	}
	if modes.WriteCurrentConfig {
		ip := flag.NewIniParser(fp)
		ip.Write(os.Stdout, flag.IniIncludeComments)
		os.Exit(0)
	}

	if modes.ListParsers {
		fmt.Println("Available parsers:", strings.Join(validParsers, ", "))
		os.Exit(0)
	}
}

func addParserDefaultOptions(options *globals.GlobalOptions) {
	switch {
	case options.Reqs.ParserName == "nginx":
		// automatically normalize the request when using the nginx parser
		options.RequestShape = append(options.RequestShape, "request")
	}
	if options.Reqs.ParserName != "mysql" {
		// mysql is the only parser that requires in-parser sampling because it has
		// a multi-line log format.
		// Sample all other parser when tailing to conserve CPU
		options.TailSample = true
	} else {
		options.TailSample = false
	}
	if len(options.DynSample) != 0 {
		// when using dynamic sampling, we make the sampling decision after parsing
		// the content, so we must not tailsample.
		options.TailSample = false
		options.GoalSampleRate = int(options.SampleRate)
		options.SampleRate = 1
	}
}

func sanityCheckOptions(options *globals.GlobalOptions) {
	switch {
	case options.Reqs.ParserName == "":
		fmt.Println("Parser required to be specified with the --parser flag.")
		usage()
		os.Exit(1)
	/*case options.Reqs.WriteKey == "" || options.Reqs.WriteKey == "NULL":
	fmt.Println("Write key required to be specified with the --writekey flag.")
	usage()
	os.Exit(1)*/
	case len(options.Reqs.LogFiles) == 0:
		fmt.Println("Log file name or '-' required to be specified with the --file flag.")
		usage()
		os.Exit(1)
	case options.Reqs.Dataset == "":
		fmt.Println("Dataset name required with the --dataset flag.")
		usage()
		os.Exit(1)
	case options.SampleRate == 0:
		fmt.Println("Sample rate must be an integer >= 1")
		usage()
		os.Exit(1)
	case options.Tail.ReadFrom == "end" && options.Tail.Stop:
		fmt.Println("Reading from the end and stopping when we get there. Zero lines to process. Ok, all done! ;)")
		usage()
		os.Exit(1)
	case options.RequestParseQuery != "whitelist" && options.RequestParseQuery != "all":
		fmt.Println("request_parse_query flag must be either 'whitelist' or 'all'.")
		usage()
		os.Exit(1)
	case len(options.DynSample) != 0 && options.SampleRate <= 1 && options.GoalSampleRate <= 1:
		fmt.Println("sample rate flag must be set >= 2 when dynamic sampling is enabled")
		usage()
		os.Exit(1)
	}

	// check the prefix regex for validity
	if options.PrefixRegex != "" {
		// make sure the regex is anchored against the start of the string
		if options.PrefixRegex[0] != '^' {
			options.PrefixRegex = "^" + options.PrefixRegex
		}
		// make sure it's valid
		_, err := regexp.Compile(options.PrefixRegex)
		if err != nil {
			fmt.Printf("Prefix regex %s doesn't compile: error %s\n", options.PrefixRegex, err)
			usage()
			os.Exit(1)
		}
	}

	// Make sure input files exist
	shouldExit := false
	for _, f := range options.Reqs.LogFiles {
		if f == "-" {
			continue
		}
		if files, err := filepath.Glob(f); err != nil || files == nil {
			fmt.Printf("Log file specified by --file=%s not found!\n", f)
			shouldExit = true
		}
	}
	if shouldExit {
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`
Usage: clicktail -p <parser> -f </path/to/logfile> -d <mydata> [optional arguments]

For even more detail on required and optional parameters, run
clicktail --help
`)
}
