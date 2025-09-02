package controllers

import (
    "context"
    "testing"
    "time"

    "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
    "github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
    "github.com/go-logr/logr"
    "github.com/pkg/errors"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/client-go/tools/record"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/client/fake"
    "sigs.k8s.io/controller-runtime/pkg/log/zap"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// MockChaosRecorder implements ChaosRecorder for testing
type MockChaosRecorder struct {
    Events []recorder.ChaosEvent
}

func (r *MockChaosRecorder) Event(obj runtime.Object, event recorder.ChaosEvent) {
    r.Events = append(r.Events, event)
}

// TestSetCondition_UpdatesObservedGeneration tests that SetCondition updates ObservedGeneration
func TestSetCondition_UpdatesObservedGeneration(t *testing.T) {
    // Create a status
    status := &v1alpha1.WorkflowNodeStatus{}
    
    // Create a condition with a specific generation
    condition := v1alpha1.WorkflowNodeCondition{
        Type:       v1alpha1.ConditionAccomplished,
        Status:     corev1.ConditionTrue,
        Reason:     "TestReason",
        Generation: 42,
    }
    
    // Set the condition
    SetCondition(status, condition)
    
    // Check if ObservedGeneration was updated
    if status.ObservedGeneration != 42 {
        t.Errorf("ObservedGeneration not updated, got: %d, want: %d", 
            status.ObservedGeneration, 42)
    }
}

// TestSetCondition_UpdatesObservedGenerationWithMultipleConditions tests generation tracking with multiple conditions
func TestSetCondition_UpdatesObservedGenerationWithMultipleConditions(t *testing.T) {
    // Create a status with existing conditions
    status := &v1alpha1.WorkflowNodeStatus{
        Conditions: []v1alpha1.WorkflowNodeCondition{
            {
                Type:       v1alpha1.ConditionAccomplished,
                Status:     corev1.ConditionFalse,
                Reason:     "NotYet",
                Generation: 5,
            },
        },
        ObservedGeneration: 5,
    }
    
    // Add a new condition with a newer generation
    condition := v1alpha1.WorkflowNodeCondition{
        Type:       v1alpha1.ConditionAccomplished,
        Status:     corev1.ConditionTrue,
        Reason:     "TestReason",
        Generation: 42,
    }
    
    // Set the condition
    SetCondition(status, condition)
    
    // Check if ObservedGeneration was updated
    if status.ObservedGeneration != 42 {
        t.Errorf("ObservedGeneration not updated, got: %d, want: %d", 
            status.ObservedGeneration, 42)
    }
    
    // Check if the condition was added/updated correctly
    found := false
    for _, cond := range status.Conditions {
        if cond.Type == v1alpha1.ConditionAccomplished {
            found = true
            if cond.Status != corev1.ConditionTrue {
                t.Errorf("Condition status not updated, got: %v, want: %v", 
                    cond.Status, corev1.ConditionTrue)
            }
            if cond.Generation != 42 {
                t.Errorf("Condition generation not updated, got: %d, want: %d", 
                    cond.Generation, 42)
            }
        }
    }
    
    if !found {
        t.Error("Condition not found in status after SetCondition")
    }
}

// TestStatusCheckReconciler_SkipsReconciliationWhenGenerationMatches tests that reconciliation is skipped when generation matches
func TestStatusCheckReconciler_SkipsReconciliationWhenGenerationMatches(t *testing.T) {
    // Create a test scheme
    scheme := runtime.NewScheme()
    _ = v1alpha1.AddToScheme(scheme)
    _ = corev1.AddToScheme(scheme)
    
    // Create a test node with matching generation
    node := &v1alpha1.WorkflowNode{
        ObjectMeta: metav1.ObjectMeta{
            Name:       "test-node",
            Namespace:  "default",
            Generation: 5,
        },
        Spec: v1alpha1.WorkflowNodeSpec{
            Type: v1alpha1.TypeStatusCheck,
        },
        Status: v1alpha1.WorkflowNodeStatus{
            ObservedGeneration: 5, // Same as Generation - should skip
        },
    }
    
    // Create a fake client
    fakeClient := fake.NewClientBuilder().
        WithScheme(scheme).
        WithObjects(node).
        Build()
    
    // Create a mock recorder
    mockRecorder := &MockChaosRecorder{}
    
    // Create the reconciler
    reconciler := &StatusCheckReconciler{
        kubeClient:    fakeClient,
        eventRecorder: mockRecorder,
        logger:        zap.New(zap.UseDevMode(true)),
    }
    
    // Create a spy for syncStatusCheck
    syncStatusCheckCalled := false
    originalSyncStatusCheck := reconciler.syncStatusCheck
    reconciler.syncStatusCheck = func(ctx context.Context, request reconcile.Request, node v1alpha1.WorkflowNode) error {
        syncStatusCheckCalled = true
        return nil
    }
    
    // Run reconciliation
    req := reconcile.Request{
        NamespacedName: types.NamespacedName{
            Name:      "test-node",
            Namespace: "default",
        },
    }
    
    _, err := reconciler.Reconcile(context.Background(), req)
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    
    // Verify syncStatusCheck was not called
    if syncStatusCheckCalled {
        t.Error("syncStatusCheck was called when generation matches - should be skipped")
    }
    
    // Restore original function
    reconciler.syncStatusCheck = originalSyncStatusCheck
}

// TestStatusCheckReconciler_ProcessesReconciliationWhenGenerationDiffers tests that reconciliation is processed when generation differs
func TestStatusCheckReconciler_ProcessesReconciliationWhenGenerationDiffers(t *testing.T) {
    // Create a test scheme
    scheme := runtime.NewScheme()
    _ = v1alpha1.AddToScheme(scheme)
    _ = corev1.AddToScheme(scheme)
    
    // Create a test node with different generation
    node := &v1alpha1.WorkflowNode{
        ObjectMeta: metav1.ObjectMeta{
            Name:       "test-node",
            Namespace:  "default",
            Generation: 6,
        },
        Spec: v1alpha1.WorkflowNodeSpec{
            Type: v1alpha1.TypeStatusCheck,
        },
        Status: v1alpha1.WorkflowNodeStatus{
            ObservedGeneration: 5, // Different from Generation - should process
        },
    }
    
    // Create a fake client
    fakeClient := fake.NewClientBuilder().
        WithScheme(scheme).
        WithObjects(node).
        Build()
    
    // Create a mock recorder
    mockRecorder := &MockChaosRecorder{}
    
    // Create the reconciler
    reconciler := &StatusCheckReconciler{
        kubeClient:    fakeClient,
        eventRecorder: mockRecorder,
        logger:        zap.New(zap.UseDevMode(true)),
    }
    
    // Create a spy for syncStatusCheck
    syncStatusCheckCalled := false
    originalSyncStatusCheck := reconciler.syncStatusCheck
    reconciler.syncStatusCheck = func(ctx context.Context, request reconcile.Request, node v1alpha1.WorkflowNode) error {
        syncStatusCheckCalled = true
        return nil
    }
    
    // Run reconciliation
    req := reconcile.Request{
        NamespacedName: types.NamespacedName{
            Name:      "test-node",
            Namespace: "default",
        },
    }
    
    _, err := reconciler.Reconcile(context.Background(), req)
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    
    // Verify syncStatusCheck was called
    if !syncStatusCheckCalled {
        t.Error("syncStatusCheck was not called when generation differs - should be processed")
    }
    
    // Restore original function
    reconciler.syncStatusCheck = originalSyncStatusCheck
}

// TestUpdateNodeStatus_SetsObservedGeneration tests that updateNodeStatus sets the ObservedGeneration
func TestUpdateNodeStatus_SetsObservedGeneration(t *testing.T) {
    // Create a test scheme
    scheme := runtime.NewScheme()
    _ = v1alpha1.AddToScheme(scheme)
    _ = corev1.AddToScheme(scheme)
    
    // Create a test node
    node := &v1alpha1.WorkflowNode{
        ObjectMeta: metav1.ObjectMeta{
            Name:       "test-node",
            Namespace:  "default",
            Generation: 10,
        },
        Spec: v1alpha1.WorkflowNodeSpec{
            Type: v1alpha1.TypeStatusCheck,
        },
    }
    
    // Create a status check
    statusCheck := &v1alpha1.StatusCheck{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-status-check",
            Namespace: "default",
            Labels: map[string]string{
                v1alpha1.LabelControlledBy: "test-node",
            },
        },
    }
    
    // Create a fake client
    fakeClient := fake.NewClientBuilder().
        WithScheme(scheme).
        WithObjects(node, statusCheck).
        Build()
    
    // Create a mock recorder
    mockRecorder := &MockChaosRecorder{}
    
    // Create the reconciler
    reconciler := &StatusCheckReconciler{
        kubeClient:    fakeClient,
        eventRecorder: mockRecorder,
        logger:        zap.New(zap.UseDevMode(true)),
    }
    
    // Mock fetchChildrenStatusCheck to return our status check
    originalFetch := reconciler.fetchChildrenStatusCheck
    reconciler.fetchChildrenStatusCheck = func(ctx context.Context, node v1alpha1.WorkflowNode) ([]v1alpha1.StatusCheck, error) {
        return []v1alpha1.StatusCheck{*statusCheck}, nil
    }
    
    // Mock IsCompleted to return true
    originalIsCompleted := statusCheck.IsCompleted
    statusCheck.IsCompleted = func() bool {
        return true
    }
    
    // Run updateNodeStatus
    req := reconcile.Request{
        NamespacedName: types.NamespacedName{
            Name:      "test-node",
            Namespace: "default",
        },
    }
    
    err := reconciler.updateNodeStatus(context.Background(), req)()
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    
    // Get the updated node
    updatedNode := &v1alpha1.WorkflowNode{}
    err = fakeClient.Get(context.Background(), req.NamespacedName, updatedNode)
    if err != nil {
        t.Fatalf("Failed to get updated node: %v", err)
    }
    
    // Verify ObservedGeneration was updated
    if updatedNode.Status.ObservedGeneration != 10 {
        t.Errorf("ObservedGeneration not updated, got: %d, want: %d", 
            updatedNode.Status.ObservedGeneration, 10)
    }
    
    // Restore original functions
    reconciler.fetchChildrenStatusCheck = originalFetch
    statusCheck.IsCompleted = originalIsCompleted
}