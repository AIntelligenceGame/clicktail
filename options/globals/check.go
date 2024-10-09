package globals

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AIntelligenceGame/clicktail/libclick"

	// "github.com/honeycombio/libhoney-go"
	//libclick "github.com/Altinity/libclick-go"
	flag "github.com/jessevdk/go-flags"
)

// BuildID is set by Travis CI
var BuildID string

// internal version identifier
var version string

var ValidParsers = []string{
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

// setVersion sets the internal version ID and updates libclick's user-agent
func SetVersionUserAgent(backfill bool, parserName string) {
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
func HandleOtherModes(fp *flag.Parser, modes OtherModes) {
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
		fmt.Println("Available parsers:", strings.Join(ValidParsers, ", "))
		os.Exit(0)
	}
}

func AddParserDefaultOptions(options *GlobalOptions) {
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

func SanityCheckOptions(options *GlobalOptions) {
	switch {
	case options.Reqs.ParserName == "":
		fmt.Println("Parser required to be specified with the --parser flag.")
		Usage()
		os.Exit(1)
	/*case options.Reqs.WriteKey == "" || options.Reqs.WriteKey == "NULL":
	fmt.Println("Write key required to be specified with the --writekey flag.")
	Usage()
	os.Exit(1)*/
	case len(options.Reqs.LogFiles) == 0:
		fmt.Println("Log file name or '-' required to be specified with the --file flag.")
		Usage()
		os.Exit(1)
	case options.Reqs.Dataset == "":
		fmt.Println("Dataset name required with the --dataset flag.")
		Usage()
		os.Exit(1)
	case options.SampleRate == 0:
		fmt.Println("Sample rate must be an integer >= 1")
		Usage()
		os.Exit(1)
	case options.Tail.ReadFrom == "end" && options.Tail.Stop:
		fmt.Println("Reading from the end and stopping when we get there. Zero lines to process. Ok, all done! ;)")
		Usage()
		os.Exit(1)
	case options.RequestParseQuery != "whitelist" && options.RequestParseQuery != "all":
		fmt.Println("request_parse_query flag must be either 'whitelist' or 'all'.")
		Usage()
		os.Exit(1)
	case len(options.DynSample) != 0 && options.SampleRate <= 1 && options.GoalSampleRate <= 1:
		fmt.Println("sample rate flag must be set >= 2 when dynamic sampling is enabled")
		Usage()
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
			Usage()
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
		Usage()
		os.Exit(1)
	}
}

func Usage() {
	fmt.Print(`
Usage: clicktail -p <parser> -f </path/to/logfile> -d <mydata> [optional arguments]

For even more detail on required and optional parameters, run
clicktail --help
`)
}
