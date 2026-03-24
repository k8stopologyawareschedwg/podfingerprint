// SPDX-License-Identifier: Apache-2.0

package podfingerprint

import (
	"bufio"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompatGoldenSign(t *testing.T) {
	for _, tc := range discoverCompatCases(t) {
		name := filepath.Base(tc.podsFile)
		t.Run(name, func(t *testing.T) {
			pods, expectedSign := mustLoadCompatCase(t, tc)

			fp := NewFingerprint(len(pods))
			for _, pod := range pods {
				fp.Add(pod.Namespace, pod.Name)
			}
			got := fp.Sign()
			if got != expectedSign {
				t.Errorf("Sign mismatch: got %q expected %q", got, expectedSign)
			}
		})
	}
}

func TestCompatGoldenCheck(t *testing.T) {
	for _, tc := range discoverCompatCases(t) {
		name := filepath.Base(tc.podsFile)
		t.Run(name, func(t *testing.T) {
			pods, expectedSign := mustLoadCompatCase(t, tc)

			fp := NewFingerprint(len(pods))
			for _, pod := range pods {
				fp.Add(pod.Namespace, pod.Name)
			}
			if err := fp.Check(expectedSign); err != nil {
				t.Errorf("Check failed against golden sign: %v", err)
			}
		})
	}
}

func TestCompatGoldenOrderIndependent(t *testing.T) {
	for _, tc := range discoverCompatCases(t) {
		name := filepath.Base(tc.podsFile)
		t.Run(name, func(t *testing.T) {
			pods, expectedSign := mustLoadCompatCase(t, tc)
			if len(pods) < 2 {
				t.Skip("need at least 2 pods for shuffle test")
			}

			shuffled := make([]NamespacedName, len(pods))
			copy(shuffled, pods)
			rng := rand.New(rand.NewSource(12345))
			rng.Shuffle(len(shuffled), func(i, j int) {
				shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
			})

			fp := NewFingerprint(len(shuffled))
			for _, pod := range shuffled {
				fp.Add(pod.Namespace, pod.Name)
			}
			got := fp.Sign()
			if got != expectedSign {
				t.Errorf("Sign after shuffle: got %q expected %q", got, expectedSign)
			}
		})
	}
}

func TestCompatGoldenTracing(t *testing.T) {
	for _, tc := range discoverCompatCases(t) {
		name := filepath.Base(tc.podsFile)
		t.Run(name, func(t *testing.T) {
			pods, expectedSign := mustLoadCompatCase(t, tc)

			fp := NewTracingFingerprint(len(pods), NullTracer{})
			for _, pod := range pods {
				fp.Add(pod.Namespace, pod.Name)
			}
			got := fp.Sign()
			if got != expectedSign {
				t.Errorf("TracingFingerprint Sign mismatch: got %q expected %q", got, expectedSign)
			}
			if err := fp.Check(expectedSign); err != nil {
				t.Errorf("TracingFingerprint Check failed: %v", err)
			}
		})
	}
}

func mustLoadCompatCase(t *testing.T, tc compatTestCase) ([]NamespacedName, string) {
	t.Helper()
	pods, err := parsePodsTxt(tc.podsFile)
	if err != nil {
		t.Fatalf("reading pods from %s: %v", tc.podsFile, err)
	}
	expectedSign, err := readGoldenSign(tc.signFile)
	if err != nil {
		t.Fatalf("reading golden sign from %s: %v", tc.signFile, err)
	}
	return pods, expectedSign
}

// parsePodsTxt reads a kubectl-style pod listing (as produced by tools/pfp -P)
// and returns the namespace/name pairs. Lines starting with "#" are skipped.
// The first two whitespace-separated fields on each line are namespace and name.
func parsePodsTxt(path string) ([]NamespacedName, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pods []NamespacedName
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pods = append(pods, NamespacedName{
			Namespace: fields[0],
			Name:      fields[1],
		})
	}
	return pods, scanner.Err()
}

func readGoldenSign(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

type compatTestCase struct {
	podsFile string
	signFile string
}

func discoverCompatCases(t *testing.T) []compatTestCase {
	t.Helper()
	txtFiles, err := filepath.Glob(filepath.Join("testdata", "*.txt"))
	if err != nil {
		t.Fatalf("glob testdata/*.txt: %v", err)
	}
	var cases []compatTestCase
	for _, txtFile := range txtFiles {
		signFile := strings.TrimSuffix(txtFile, ".txt") + ".sign"
		if _, err := os.Stat(signFile); err != nil {
			continue
		}
		cases = append(cases, compatTestCase{
			podsFile: txtFile,
			signFile: signFile,
		})
	}
	if len(cases) == 0 {
		t.Fatal("no compat test cases found (need testdata/*.txt with matching .sign files)")
	}
	return cases
}
