/*
 * Copyright 2023 Red Hat, Inc.
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

package podfingerprint

import (
	"testing"
)

func TestMarkCompleted(t *testing.T) {
	type testCase struct {
		name          string
		sink          chan Status
		status        Status
		expectedError error
	}

	testCases := []testCase{
		{
			name: "no completion sink",
			status: Status{
				NodeName: "test-node-0",
				Pods: []NamespacedName{
					{
						Namespace: "ns-1",
						Name:      "pod-1",
					},
					{
						Namespace: "ns-2",
						Name:      "pod-2",
					},
					{
						Namespace: "ns-2",
						Name:      "pod-3",
					},
				},
				FingerprintExpected: "pfp0v001807d932586d44a8a",
				FingerprintComputed: "pfp0v001807d932586d44a8a",
			},
			sink: nil, // explicit
		},
		{
			name: "with completion sink",
			status: Status{
				NodeName: "test-node-0",
				Pods: []NamespacedName{
					{
						Namespace: "ns-1",
						Name:      "pod-1",
					},
					{
						Namespace: "ns-2",
						Name:      "pod-2",
					},
					{
						Namespace: "ns-2",
						Name:      "pod-3",
					},
				},
				FingerprintExpected: "pfp0v001807d932586d44a8a",
				FingerprintComputed: "pfp0v001807d932586d44a8a",
			},
			sink: make(chan Status, 5), // anything > 1 is fine
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			SetCompletionSink(tc.sink)
			err := MarkCompleted(tc.status)
			if err != tc.expectedError {
				t.Errorf("got error=%v expected=%v", err, tc.expectedError)
			}
			if tc.sink != nil && len(tc.sink) != 1 {
				t.Errorf("unexpected data in sink: %d", len(tc.sink))
			}
		})
	}
}
