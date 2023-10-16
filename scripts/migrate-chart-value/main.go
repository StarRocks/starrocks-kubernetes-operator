package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	inputFilePath      string
	targetChartVersion string
	outputFilePath     string
)

const (
	NEW_VERSION = "v1.8.0"
	OPERATOR    = "operator"
	STARROCKS   = "starrocks"
)

var _starrocksKeys = []string{"nameOverride", "initPassword", "timeZone", "datadog", "starrocksCluster",
	"starrocksFESpec", "starrocksCnSpec", "starrocksBeSpec", "secrets", "configMaps", "feProxy"}

var _operatorKeys = []string{"global", "timeZone", "nameOverride", "starrocksOperator"}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
This tool is used to upgrade or downgrade the version of values.yaml for kube-starrocks chart.

If the chart version is less than v1.8.0, the values.yaml is in the following format:
key1: value1
key2: value2

If the chart version is greater than or equal to v1.8.0, the values.yaml will be changed to the following format:
operator:
  key1: value1
starrocks:
  key2: value2

If you want to upgrade the version of values.yaml, you can run the following command:
./migrate-chart-value --input values.yaml --target-version v1.8.0 --output values_v1.8.0.yaml

If you want to downgrade the version of values.yaml, you can run the following command:
./migrate-chart-value --input values.yaml --target-version v1.7.1 --output values_v1.7.1.yaml

[Options]
`)
		flag.PrintDefaults()
	}
	flag.StringVar(&inputFilePath, "input", "",
		"the input path of values.yaml for kube-starrocks chart, if not specified, it will read from stdin")
	flag.StringVar(&targetChartVersion, "target-version", "",
		"the chart version, which this tool will change the values.yaml to")
	flag.StringVar(&outputFilePath, "output", "",
		"the output path of values.yaml for kube-starrocks chart, if not specified, it will write to stdout")
	flag.Parse()

	log.SetOutput(os.Stderr)

	if targetChartVersion == "" {
		log.Println("target-version option is required")
		flag.Usage()
		return
	} else if targetChartVersion[0] != 'v' {
		log.Println("version must start with v")
		flag.Usage()
		return
	}
	var reader io.ReadCloser
	var writer io.WriteCloser
	var err error
	if inputFilePath == "" {
		reader = os.Stdin
	} else {
		reader, err = os.Open(inputFilePath)
		if err != nil {
			panic(err)
		}
		defer func() { _ = reader.Close() }()
	}

	if outputFilePath == "" {
		writer = os.Stdout
	} else {
		writer, err = os.Create(outputFilePath)
		if err != nil {
			panic(err)
		}
		defer func() { _ = writer.Close() }()
	}

	err = Do(reader, targetChartVersion, writer)
	if err != nil {
		panic(err)
	}
	log.Println("success")
}

// Do is the main function of this tool. It will read the values.yaml from the reader, and write the new values.yaml to the writer.
func Do(reader io.Reader, targetChartVersion string, writer io.Writer) error {
	input, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	s := map[string]interface{}{}
	if err := yaml.Unmarshal(input, &s); err != nil {
		return err
	}

	// find the value of the field "operator" or "starrocks"
	operator := s[OPERATOR]   // the type of operator is interface{}
	starrocks := s[STARROCKS] // the type of starrocks is interface{}
	if operator != nil || starrocks != nil {
		log.Printf("this values.yaml is from new chart version >= %v\n", NEW_VERSION)
		// values.yaml is from new chart version
		if targetChartVersion >= NEW_VERSION {
			log.Printf("no need to change to upgrade to %v\n", targetChartVersion)
			return nil
		}
		// change the new version to old version
		mapper := operator.(map[interface{}]interface{})
		// remove duplicate fields from operator
		delete(mapper, "timeZone")
		delete(mapper, "nameOverride")
		data1, err := yaml.Marshal(operator)
		if err != nil {
			return err
		}
		data2, err := yaml.Marshal(starrocks)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(writer, "%v\n%v", string(data1), string(data2)); err != nil {
			return err
		}
		return nil
	} else {
		log.Printf("this values.yaml is from old chart version < %v,\n", NEW_VERSION)
		// values.yaml is from old chart version
		if targetChartVersion < NEW_VERSION {
			log.Printf("no need to change to downgrade to %v\n", targetChartVersion)
			return nil
		}

		// change the old version to new version

		if err := Write(writer, s, _operatorKeys, OPERATOR); err != nil {
			return err
		}
		_, _ = writer.Write([]byte("\n"))
		if err := Write(writer, s, _starrocksKeys, STARROCKS); err != nil {
			return err
		}
	}
	return nil
}

func Write(w io.Writer, originalFields map[string]interface{}, keys []string, header string) error {
	getValue := func(key string) interface{} {
		if key == "timeZone" {
			return "Asia/Shanghai"
		} else if key == "nameOverride" {
			return "kube-starrocks"
		}
		return nil
	}

	fields := map[string]interface{}{}
	for _, key := range keys {
		value := originalFields[key]
		if value == nil {
			value = getValue(key)
		} else {
			realValue, ok := value.(string)
			if ok && realValue == "" {
				value = getValue(key)
			}
		}
		if value != nil {
			fields[key] = value
		}
	}
	if len(fields) == 0 {
		if header != "" {
			_, err := w.Write([]byte(header + ":"))
			if err != nil {
				return err
			}
		}
		return nil
	}

	if data, err := AddHeader(fields, header); err != nil {
		return err
	} else if _, err = w.Write(data); err != nil {
		return err
	}
	return nil
}

func AddHeader(fields map[string]interface{}, header string) ([]byte, error) {
	output := fields
	if header != "" {
		output = map[string]interface{}{header: fields}
	}
	data, err := yaml.Marshal(output)
	if err != nil {
		return nil, err
	}
	return data, nil
}
