package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

const usageText = `Usage: boron [options]
Options:
  -tags                    Tags to determine correct measurement
  -measurement             Single Telegraf timemeasurement to check against
  -template-file           A provided and working telegraf configuration with only one plugin
  -plugin                  Telegraf Plugin name, use when boron is generating the configuration
  -plugin-parameters       Parameters for Telegraf Plugin, separator '|', use when boron is generating the configuration
  -telegraf-location       Absolute path of telegraf binary
  -working-location        Working location to generate temporary plugin files, use when boron is generating the configuration
  -lessthan                Warning and Critical values will notify if less than
  -warning=                Exits with code 1 if exceeded, Optional
  -critical=               Exits with code 2 if exceeded, Required

Examples:
	./boron -plugin mem --working-location . -telegraf-location ./telegraf -measurement 'mem_used_percent' -critical 0
	./boron -plugin cpu --working-location . -telegraf-location ./telegraf -plugin-parameters 'percpu = true|totalcpu = true|drop = ["cpu_time"]' -measurement 'cpu_usage_idle' -critical 0 -tags 'cpu=cpu0'
	./boron -template-file mem.toml -telegraf-location ./telegraf -measurement 'mem_used_percent' -critical 0
    ./boron -template-file cpu.toml -telegraf-location ./telegraf -measurement 'cpu_usage_idle' -critical 0 -tags 'cpu=cpu0'
`

func main() {
	os.Exit(boronMain())
}

func boronMain() int {

	var (
		Measurement      string
		TemplateFile     string
		Plugin           string
		PluginParameters string
		Tags             string
		TelegrafLocation string
		WorkingLocation  string
		LessThan         bool
		Warning          float64
		Critical         float64
	)

	randomFloat := rand.Float64()

	boronFlags := flag.NewFlagSet("boron", flag.ExitOnError)
	boronFlags.Usage = func() { printUsage() }

	//
	boronFlags.StringVar(&Tags, "tags", "", "tags to determine correct measurement")
	boronFlags.StringVar(&Measurement, "measurement", "", "Single Telegraf timemeasurement to check against")
	boronFlags.StringVar(&TemplateFile, "template-file", "", "A valid telegraf configuration file with one plugin specified")
	boronFlags.StringVar(&Plugin, "plugin", "", "Telegraf Plugin name")
	boronFlags.StringVar(&PluginParameters, "plugin-parameters", "", "Parameters for Telegraf Plugin")
	boronFlags.StringVar(&TelegrafLocation, "telegraf-location", "/usr/bin/telegraf", "Absolute path of telegraf binary")
	boronFlags.StringVar(&WorkingLocation, "working-location", "/tmp/", "Working location to generate temporary plugin files")

	boronFlags.BoolVar(&LessThan, "lessthan", false, "Warning and Critical values will notify if less than")
	boronFlags.Float64Var(&Warning, "warning", randomFloat, "Exits with code 1")
	boronFlags.Float64Var(&Critical, "critical", randomFloat, "Exits with code 2")

	if err := boronFlags.Parse(os.Args[1:]); err != nil {
		boronFlags.Usage()
		return 1
	}

	var configurationLocation string
	var measurementResult float64

	if TemplateFile != "" {
		configurationLocation = TemplateFile

		result, err := executeProvidedTelegraf(TelegrafLocation, configurationLocation, Measurement, Tags)
		if err != nil || measurementResult == math.MaxFloat64 {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		measurementResult = result

	} else {

		configurationLocation, err := writeTemplate(Plugin, PluginParameters, WorkingLocation, TelegrafTemplate)
		if err != nil {
			return 2
		}

		defer os.Remove(configurationLocation)

		result, err := executeGeneratedTelegraf(TelegrafLocation, configurationLocation, Plugin, Measurement, Tags)
		if err != nil || measurementResult == math.MaxFloat64 {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		measurementResult = result
	}

	exitCode := thresholdChecker(measurementResult, Measurement, LessThan, Warning, Critical, randomFloat)
	return exitCode
}

//Splits Plugin parameters
func splitParameters(PluginParameters string) []string {
	return strings.Split(PluginParameters, "|")
}

//Generates template  file for given plugin
func writeTemplate(Plugin string, PluginParameters string, WorkingLocation, Template string) (string, error) {

	genUUID := uuid.NewV4()

	t, err := template.New("plugin").Parse(Template)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", err
	}

	temporaryFileLocation := fmt.Sprintf("%s%s-%s.conf", WorkingLocation, Plugin, genUUID)
	fh, err := os.Create(temporaryFileLocation)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", err
	}
	defer fh.Close()

	wr := bufio.NewWriter(fh)
	defer wr.Flush()

	tp := TelegrafPlugin{
		Plugin: Plugin,
		KV:     splitParameters(PluginParameters),
	}

	err = t.Execute(wr, tp)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", err
	}

	return temporaryFileLocation, nil
}

func filterTelegrafResults(Output []byte, Measurement, Tags string) (float64, error) {

	re := regexp.MustCompile(fmt.Sprintf(`(%s),(.*)value=(\d+\.?\d+)\s+(\d*)|(%s)\s+value=(\d+\.?\d+)\s+(\d*)`, Measurement, Measurement))
	regexResult := re.FindAllStringSubmatch(string(Output), -1)

	if len(regexResult) >= 1 {
		for _, group := range regexResult {

			measurementName := group[1]
			measurementTags := group[2]
			measurementValue := group[3]
			measurementNoTagsValue := group[6]

			if measurementName != "" { //measurement has tags

				if strings.Contains(measurementTags, Tags) {
					f, err := strconv.ParseFloat(measurementValue, 64)
					if err != nil {
						return math.MaxFloat64, fmt.Errorf("Measurement: %s not a float64 value", Measurement)
					}
					return f, nil
				}

				continue

			} else { //no tags first measurement name match found will return

				f, err := strconv.ParseFloat(measurementNoTagsValue, 64)
				if err != nil {
					return math.MaxFloat64, fmt.Errorf("Measurement: %s not a float64 value", Measurement)
				}
				return f, nil

			}
		}
	} else {
		return math.MaxFloat64, fmt.Errorf("Measurement: %s not found", Measurement)
	}
	return math.MaxFloat64, fmt.Errorf("Measurement: %s not found", Measurement)

}

//Executes Telegraf in STDOUT mode with the provided config file.
//Parses STDOUT and abstracts value
func executeProvidedTelegraf(TelegrafLocation, Config, Measurement, Tags string) (float64, error) {
	cmd := exec.Command(TelegrafLocation, "-config", Config, "-test", "|", "grep", Measurement)
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return math.MaxFloat64, err
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return math.MaxFloat64, err
	}

	r, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return math.MaxFloat64, err
	}

	return filterTelegrafResults(r, Measurement, Tags)
}

//Executes Telegraf in STDOUT mode with the generated config file.
//Parses STDOUT and abstracts value
func executeGeneratedTelegraf(TelegrafLocation, Config, Plugin, Measurement, Tags string) (float64, error) {
	cmd := exec.Command(TelegrafLocation, "-config", Config, "-filter", Plugin, "-test", "|", "grep", Measurement)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return math.MaxFloat64, err
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return math.MaxFloat64, err
	}

	r, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return math.MaxFloat64, err
	}

	return filterTelegrafResults(r, Measurement, Tags)
}

func messageGenerator(thresholdType string, thresholdValue, curerntValue float64, measurement string) {
	fmt.Fprintf(os.Stdout, "Threshold Event:%s ThresholdValue:%f CurrentValue:%f Measurement:%s\n", thresholdType, thresholdValue, curerntValue, measurement)
}

func exceeded(thresholdValue, currentValue float64, lessthan bool) bool {
	switch lessthan {
	case true:
		if thresholdValue <= currentValue {
			return false
		}
		return true
	case false:
		if thresholdValue >= currentValue {
			return false
		}
		return true
	default:
		return false
	}

}

func thresholdChecker(currentValue float64, measurement string, lessthan bool, warningValue, criticalValue, random float64) int {
	if warningValue == random && criticalValue == random {
		fmt.Fprintln(os.Stderr, "Critical threshold must be specified")
		return 1
	}

	if exceeded(criticalValue, currentValue, lessthan) {
		//ALERT CRITICAL
		messageGenerator("CRITICAL", criticalValue, currentValue, measurement)
		return 2
	} else if warningValue != random && exceeded(criticalValue, currentValue, lessthan) == false && exceeded(warningValue, currentValue, lessthan) == true {
		//ALERT WARNING
		messageGenerator("WARNING", warningValue, currentValue, measurement)
		return 1
	} else {
		messageGenerator("PASSING", warningValue, currentValue, measurement)
		return 0
	}

}

func printUsage() {
	fmt.Fprintln(os.Stderr, usageText)
}
