package main

import (
	"flag"
	"os"
	"regexp"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type cliArgs struct {
	dbDir         string
	accessLogFile string
	reportTime    string
	reportMail    string

	mailerConfigPath string
	reportHour       int
	reportMinute     int
	reportSecond     int
	printReport      bool
	scheduleInterval time.Duration
}

func main() {
	args := getArgs()
	handleArgs(&args)
	origLogLevel := setupLogger()

	log.Printf("LOG LEVEL: %s, ARGS: %+v", origLogLevel, args)
	log.Infoln("Work is finished")
}

func getArgs() cliArgs {
	var args cliArgs
	flag.StringVar(&args.dbDir, "dbDir", "/tmp/squid-log-monitor-test-db", "DB directory")
	flag.StringVar(&args.accessLogFile, "accessLogFile", "/var/log/squid/access.log", "access.log file")
	flag.StringVar(&args.reportTime, "reportTime", "22:00:00", "Report UTC time in format 22:00:00")
	flag.StringVar(&args.reportMail, "reportMail", "", "Email to send reports")
	flag.DurationVar(&args.scheduleInterval, "scheduleInterval", 2*time.Second, "Interval for scheduler tasks scan")
	flag.StringVar(&args.mailerConfigPath, "mailerConfig", "secrets/mailer.json", "Config for mailer")
	flag.BoolVar(&args.printReport, "printReport", false, "Print report to STDOUT")
	flag.Parse()
	return args
}

func handleArgs(args *cliArgs) {
	if args.dbDir == "" {
		log.Fatalln("-dbDir is required")
	}

	if args.accessLogFile == "" {
		log.Fatalln("-accessLogFile is required")
	}

	matches := regexp.MustCompile("^(-?\\d{1,2}):(-?\\d{1,2}):(-?\\d{1,2})$").FindStringSubmatch(args.reportTime)
	if len(matches) != 4 {
		log.Fatalf("Invalid value for -reportTime: %s", args.reportTime)
	}
	args.reportHour = Must1(strconv.Atoi(matches[1]))
	args.reportMinute = Must1(strconv.Atoi(matches[2]))
	args.reportSecond = Must1(strconv.Atoi(matches[3]))

	if _, err := os.Stat(args.mailerConfigPath); err != nil {
		log.Fatalf("Unable to read -mailerConfig: %s", err)
	}

	scheduleMinInterval := 2 * time.Second
	if args.scheduleInterval < scheduleMinInterval {
		log.Warnf("-scheduleInterval is too small (%s), falling back to %s", args.scheduleInterval, scheduleMinInterval)
		args.scheduleInterval = scheduleMinInterval
	}
}

func setupLogger() log.Level {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05",
		DisableLevelTruncation: true,
		PadLevelText:           true,
		QuoteEmptyFields:       true,
	})

	level, err := log.ParseLevel(StrDef(os.Getenv("LOG_LEVEL"), "INFO"))
	if err != nil {
		log.Warnf("Invalid log level %s, falling back to INFO", err)
		level = log.InfoLevel
	}
	log.SetLevel(level)

	return level
}
