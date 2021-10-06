package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/pflag"
)

const (
	defaultNetworkName             = "network1"
	defaultWhereaboutsScaleAppName = "whereabouts-scale-test"
	defaultLowerDevice             = "eth0"
	defaultLeaseDuration           = 1500
	defaultRenewDeadline           = 1000
	defaultRetryPeriod             = 500
	defaultNumberOfReplicas        = 100
)

const (
	netAttachDefPath = "net-attach-def.yaml"
	replicaSetPath   = "replica-set.yaml"
	templateSuffix   = ".in"
)

type templatingDevice struct {
	dumpStdOut       bool
	results          map[string]template.Template
	templatingEngine map[string]interface{}
}

type netAttachDefData struct {
	NetworkName   string
	Namespace     string
	LowerDevice   string
	LeaseDuration int
	RenewDeadline int
	RetryPeriod   int
	SubnetRange   string
}

type replicaSetData struct {
	AppName          string
	Namespace        string
	NumberOfReplicas int32
	NetworkName      string
}

func newTemplatingDevice(
	inputFolder string,
	appName string,
	networkName string,
	ipamRange string,
	leaseDuration int,
	lowerDeviceName string,
	namespace string,
	renewDeadline int,
	numberOfReplicas int,
	retryPeriod int,
	dumpStdout bool) (*templatingDevice, error) {
	templatingEngine := map[string]interface{}{
		netAttachDefPath: netAttachDefData{
			NetworkName:   networkName,
			Namespace:     namespace,
			LowerDevice:   lowerDeviceName,
			LeaseDuration: leaseDuration,
			RenewDeadline: renewDeadline,
			RetryPeriod:   retryPeriod,
			SubnetRange:   ipamRange,
		},
		replicaSetPath: replicaSetData{
			AppName:          appName,
			Namespace:        namespace,
			NumberOfReplicas: int32(numberOfReplicas),
			NetworkName:      networkName,
		},
	}

	td := templatingDevice{templatingEngine: templatingEngine, dumpStdOut: dumpStdout}
	templatingResult := map[string]template.Template{}
	for templatePath := range templatingEngine {
		fileAbsPath := fmt.Sprintf("%s/%s", inputFolder, templatePath+templateSuffix)
		templatingResult[templatePath] = *td.preRenderedTemplate(fileAbsPath)
	}
	td.results = templatingResult
	return &td, nil
}

func (td templatingDevice) preRenderedTemplate(inputFile string) *template.Template {
	return template.Must(template.ParseFiles(inputFile))
}

func (td templatingDevice) spit(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return fmt.Errorf("failed to ensure dir %q exists: %w", outputDir, err)
	}
	for templatePath, templateData := range td.templatingEngine {
		preRenderedTemplate, ok := td.results[templatePath]
		if !ok {
			return fmt.Errorf("could not find the pre-rendered template for %s", templatePath)
		}

		if td.dumpStdOut {
			if err := preRenderedTemplate.Execute(os.Stdout, templateData); err != nil {
				return fmt.Errorf("failed to execute the net-attach-def template: %w", err)
			}
		}

		if err := renderYAML(filepath.Join(outputDir, templatePath), preRenderedTemplate, templateData); err != nil {
			return err
		}
	}

	return nil
}

func renderYAML(outputFile string, preRenderedTemplate template.Template, templateData interface{}) error {
	f, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := preRenderedTemplate.Execute(f, templateData); err != nil {
		return fmt.Errorf("failed to execute the net-attach-def template: %w", err)
	}
	return nil
}

func main() {
	appName := flag.String("app-name", defaultWhereaboutsScaleAppName, "The name of the scale checking application")
	networkName := flag.String("network-name", defaultNetworkName, "The name of the network for which whereabouts will provide IPAM")
	lowerDevice := flag.String("lower-device", defaultLowerDevice, "The name of the lower device on which the macvlan interfaces will be created")
	leaseDuration := flag.Int("lease-duration", defaultLeaseDuration, "How long is an active lease maintained")
	renewDeadline := flag.Int("renew-deadline", defaultRenewDeadline, "Time after which the lease is forcefully re-acquired")
	retryPeriod := flag.Int("retry-period", defaultRetryPeriod, "Period upon which the acquiring the lease is retried")
	ipamRange := flag.String("ipam-range", "", "The CIDR range to assign addresses from")
	numberOfReplicas := flag.Int("number-of-replicas", defaultNumberOfReplicas, "How many replicas in the replica-set")
	namespace := flag.String("namespace", "default", "The namespace to use")
	inputDir := flag.String("input-dir", "", "Directory with the templates")
	outputDir := flag.String("output-dir", "", "Output file dir")
	printHelp := flag.Bool("help", false, "Print help and quit")
	dumpStdout := flag.Bool("dump-stdout", false, "Also print the templates to stdout")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = false
	pflag.Parse()

	if *printHelp {
		pflag.PrintDefaults()
		os.Exit(0)
	}
	if *inputDir == "" {
		panic("Must specify input file")
	}
	if *outputDir == "" {
		panic("Must specify output directory")
	}
	if ipamRange == nil || *ipamRange == "" {
		panic("Must specify an IPAM range")
	}

	templator, err := newTemplatingDevice(
		*inputDir,
		*appName,
		*networkName,
		*ipamRange,
		*leaseDuration,
		*lowerDevice,
		*namespace,
		*renewDeadline,
		*numberOfReplicas,
		*retryPeriod,
		*dumpStdout)
	if err != nil {
		panic(fmt.Errorf("error applying the template: %w", err))
	}

	if err := templator.spit(*outputDir); err != nil {
		panic(fmt.Errorf("error outputting the template: %w", err))
	}
}
