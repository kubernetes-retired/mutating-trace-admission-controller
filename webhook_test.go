package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
	"go.opencensus.io/trace/propagation"

	admv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPodMutation(t *testing.T) {
	for _, test := range []struct {
		desc   string
		review *admv1beta1.AdmissionReview

		expectedAllowed bool
		expectedResult  *metav1.Status
	}{
		{
			desc: "Should inject span context into pod annotations",

			review: makePodAdmissionReview(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "test",
						},
					},
				},
			}),

			expectedAllowed: true,
		},
		{
			desc: "Should not overwrite existing span context in pod annotations",

			review: makePodAdmissionReview(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{TraceAnnotationKey: "AABjROGocPwPcoDVO/6sY/W7AQEeSJ1a4rHPAgA="},
					Name:        "test",
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "test",
						},
					},
				},
			}),

			expectedAllowed: true,
		},
	} {
		t.Log(test.desc)

		actualResponse := mutatePod(test.review)

		if actualResponse.Allowed != test.expectedAllowed {
			t.Errorf("Expected result\n%+v\ngot\n%+v", test.expectedResult, actualResponse.Allowed)
		}

		if actualResponse.Allowed {
			// write original pod into a pod object
			pod := &v1.Pod{}
			if err := json.Unmarshal(test.review.Request.Object.Raw, pod); err != nil {
				t.Fatal(err)
			}

			// write the patched response into a pod object
			patchedPod, err := applyPodPatch(pod, actualResponse.Patch)
			if err != nil {
				t.Fatal(err)
			}

			// extract pod context
			spanContextAnnotation, ok := patchedPod.Annotations[TraceAnnotationKey]
			if !ok {
				t.Errorf("Expected span 'spanContext' annotation but received none")
			}
			t.Logf("Retrieved spanContext %s from pod\n", spanContextAnnotation)

			// if the original pod had a span context embedded, this controller should not have edited it
			if originalSpanContextAnnotation, ok := pod.Annotations[TraceAnnotationKey]; ok {
				if spanContextAnnotation != originalSpanContextAnnotation {
					t.Errorf("Expected original span context %s to remain unchanged, found %s", originalSpanContextAnnotation, spanContextAnnotation)
				}
			}

			// ensure that embedded span context is deserializable
			validSpanContext, err := embeddedSpanContextIsDeserializable(spanContextAnnotation)
			if !validSpanContext {
				t.Errorf("Expected 'spanContext' annotation to be deserializable, but failed with error: %v", err)
			}
		}
	}
}

func applyPodPatch(pod *v1.Pod, patchBytes []byte) (*v1.Pod, error) {
	if len(patchBytes) == 0 {
		return pod, nil
	}
	patch, err := jsonpatch.DecodePatch(patchBytes)
	if err != nil {
		return nil, err
	}
	podBytes, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}
	patchedPodBytes, err := patch.Apply(podBytes)
	if err != nil {
		return nil, err
	}
	var patchedPod v1.Pod
	if err := json.Unmarshal(patchedPodBytes, &patchedPod); err != nil {
		return nil, err
	}
	return &patchedPod, nil
}

func makePodAdmissionReview(pod *v1.Pod) *admv1beta1.AdmissionReview {
	jsonPod, err := json.Marshal(pod)
	if err != nil {
		log.Fatalf("Failed to marshal pod into json: %v", err)
	}

	return &admv1beta1.AdmissionReview{
		Request: &admv1beta1.AdmissionRequest{
			Resource: metav1.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			Object: runtime.RawExtension{
				Raw: jsonPod,
			},
		},
	}
}

func embeddedSpanContextIsDeserializable(embeddedContext string) (bool, error) {
	decodedContextBytes, err := base64.StdEncoding.DecodeString(embeddedContext)
	if err != nil {
		return false, err
	}

	_, ok := propagation.FromBinary(decodedContextBytes)
	if !ok {
		return false, fmt.Errorf("could not convert raw bytes to span context")
	}

	return true, nil
}
