package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

type Options struct {
	Server        string
	Username      string
	Password      string
	Source        string
	Interval      time.Duration
	AllowInsecure bool
	DryRun        bool
}

var (
	version = "HEAD"
	commit  = "unknown"

	options     = Options{}
	showVersion = flag.Bool("version", false, "Show version information and exit")
)

func init() {
	flag.StringVar(&options.Server, "server", "", "SMTP server address")
	flag.StringVar(&options.Username, "username", "", "Username for login to SMTP server")
	flag.StringVar(&options.Password, "password", "", "Password for login to SMTP server")
	flag.StringVar(&options.Source, "source", "-", "Path of JSON file that including email data (file path or - for stdin)")
	flag.DurationVar(&options.Interval, "interval", 0, "Interval to send each emails (0 means no interval)")
	flag.BoolVar(&options.AllowInsecure, "allow-insecure", false, "Allow connection without encryption (NOT recommended)")
	flag.BoolVar(&options.DryRun, "dry-run", false, "Run json2mail without server connection to testing JSON data")
}

func (opts *Options) ParseEnv() {
	if opts.Server == "" {
		opts.Server = os.Getenv("JSON2MAIL_SERVER")
	}
	if opts.Username == "" {
		opts.Username = os.Getenv("JSON2MAIL_USERNAME")
	}
	if opts.Password == "" {
		opts.Password = os.Getenv("JSON2MAIL_PASSWORD")
	}
}

func (opts *Options) Assert(out io.Writer) (ok bool) {
	var msgs []string
	if opts.Server == "" {
		msgs = append(msgs, "-server is required.")
	}
	if opts.Username == "" {
		msgs = append(msgs, "-username is required.")
	}
	if opts.Password == "" {
		msgs = append(msgs, "-password is required.")
	}

	if msgs != nil {
		fmt.Fprintln(out, "error:")
		for _, m := range msgs {
			fmt.Fprintln(out, " ", m)
		}
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Please see -help to more detail.")
	}

	return msgs == nil
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintf(os.Stdout, "json2mail %s (%s)\n", version, commit)
		os.Exit(0)
	}

	options.ParseEnv()
	if !options.Assert(os.Stderr) {
		os.Exit(2)
	}

	l := NewLogger(os.Stdout)
	s := NewMailScanner(os.Stdin)

	if options.Source != "" && options.Source != "-" {
		f, err := os.Open(options.Source)
		if err != nil {
			l.Error("failed to open source file: "+err.Error(), options.Source)
			os.Exit(2)
		}
		defer f.Close()

		// skip BOM if exists
		var bom [3]byte
		if _, err := f.Read(bom[:]); err != nil {
			l.Error("failed to read source file: "+err.Error(), options.Source)
		} else if bom[0] != 0xEF || bom[1] != 0xBB || bom[2] != 0xBF {
			if _, err := f.Seek(0, 0); err != nil {
				l.Error("failed to read source file: "+err.Error(), options.Source)
			}
		}

		s = NewMailScanner(f)
	}

	m, err := NewMailer(options)
	if err != nil {
		l.Error("failed to connect server: "+err.Error(), options.Server)
		os.Exit(1)
	}

	for s.Scan() {
		err := m.Send(s.Mail())
		if err != nil {
			l.Error("failed to send: "+err.Error(), s.Mail())
		} else {
			l.Mail(s.Mail())
		}
		time.Sleep(options.Interval)
	}

	if s.Err() != nil {
		l.Error(s.Err().Error(), s.CurrentString())
	}
}
