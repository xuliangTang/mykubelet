package pod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	clientset "k8s.io/client-go/kubernetes"
)

// PatchPodStatus patches pod status. It returns true and avoids an update if the patch contains no changes.
func PatchPodStatus(c clientset.Interface, namespace, name string, uid types.UID, oldPodStatus, newPodStatus v1.PodStatus) (*v1.Pod, []byte, bool, error) {
	patchBytes, unchanged, err := preparePatchBytesForPodStatus(namespace, name, uid, oldPodStatus, newPodStatus)
	if err != nil {
		return nil, nil, false, err
	}
	if unchanged {
		return nil, patchBytes, true, nil
	}

	updatedPod, err := c.CoreV1().Pods(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}, "status")
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to patch status %q for pod %q/%q: %v", patchBytes, namespace, name, err)
	}
	return updatedPod, patchBytes, false, nil
}

func preparePatchBytesForPodStatus(namespace, name string, uid types.UID, oldPodStatus, newPodStatus v1.PodStatus) ([]byte, bool, error) {
	oldData, err := json.Marshal(v1.Pod{
		Status: oldPodStatus,
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to Marshal oldData for pod %q/%q: %v", namespace, name, err)
	}

	newData, err := json.Marshal(v1.Pod{
		ObjectMeta: metav1.ObjectMeta{UID: uid}, // only put the uid in the new object to ensure it appears in the patch as a precondition
		Status:     newPodStatus,
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to Marshal newData for pod %q/%q: %v", namespace, name, err)
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, v1.Pod{})
	if err != nil {
		return nil, false, fmt.Errorf("failed to CreateTwoWayMergePatch for pod %q/%q: %v", namespace, name, err)
	}
	return patchBytes, bytes.Equal(patchBytes, []byte(fmt.Sprintf(`{"metadata":{"uid":%q}}`, uid))), nil
}
