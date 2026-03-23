/*
 * Copyright 2022 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// to build go build -o pfp tools/pfp/main.go
// to use: kubectl get pods --field-selector spec.nodeName=$NODE -A --no-headers [-o wide] | pfp

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/k8stopologyawareschedwg/podfingerprint"
)

type Fingerprinter interface {
	Add(namespace, name string) error
	Sign() string
}

func main() {
	randSeed := flag.Int64("S", 1, "random seed for syntetic generation")
	genPods := flag.Int("P", 0, "generate N syntetic pods")
	genNS := flag.Int("N", 1, "generate pods in M different namespaces")
	withTrace := flag.Bool("T", false, "enable tracing")
	flag.Parse()

	if *genPods > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		pods := generatePods(*randSeed, *genNS, *genPods)
		for _, pod := range pods {
			fmt.Fprintf(w, "%-31s\t%-17s\t1/1\tRunning\t1\t42d\n", pod.Namespace, pod.Name)
		}
		w.Flush()
		return
	}

	var fp Fingerprinter
	st := podfingerprint.MakeStatus("STDIN")
	if *withTrace {
		fp = podfingerprint.NewTracingFingerprint(0, &st)
	} else {
		fp = podfingerprint.NewFingerprint(0)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		fp.Add(fields[0], fields[1])
	}
	pfp := fp.Sign()
	if *withTrace {
		podfingerprint.MarkCompleted(st)
	}
	fmt.Println(pfp)

	if *withTrace {
		json.NewEncoder(os.Stderr).Encode(st)
	}
}
