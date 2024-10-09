package gogogo

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/AIntelligenceGame/clicktail/libclick"
	"github.com/AIntelligenceGame/clicktail/options/globals"
	"github.com/AIntelligenceGame/clicktail/run"
	"github.com/honeycombio/honeytail/httime"

	// "github.com/honeycombio/libhoney-go"
	//libclick "github.com/Altinity/libclick-go"
	flag "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

func Gogogo() {
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
		globals.Usage()
		os.Exit(1)
	}
	// read the config file if present
	if options.ConfigFile != "" {
		ini := flag.NewIniParser(flagParser)
		ini.ParseAsDefaults = true
		if err := ini.ParseFile(options.ConfigFile); err != nil {
			fmt.Printf("Error: failed to parse the config file %s\n", options.ConfigFile)
			fmt.Printf("\t%s\n", err)
			globals.Usage()
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

	globals.SetVersionUserAgent(options.Backfill, options.Reqs.ParserName)
	globals.HandleOtherModes(flagParser, options.Modes)
	globals.AddParserDefaultOptions(&options)
	globals.SanityCheckOptions(&options)

	if err := libclick.VerifyApiHost(libclick.Config{
		APIHost: options.APIHost,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to ClickHouse server: ", err)
		os.Exit(1)
	}

	run.Run(options) // 将通道传递给 Run
}
