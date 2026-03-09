// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// WARNING: DO NOT RUN THIS EXAMPLE IN PRODUCTION ENVIRONMENTS!
// This example creates actual PodChaos resources that will kill pods in your cluster.
// It is intended for testing and demonstration purposes only.
// Running this in production could cause service disruptions and data loss.

package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/client/informers/externalversions"
	"github.com/chaos-mesh/chaos-mesh/pkg/client/versioned"
)

func main() {
	// Initialize kubeconfig
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create the Chaos Mesh clientset
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()
	namespace := "default"

	// Step 1: Set up informer factory and start it early
	fmt.Println("========================================")
	fmt.Println("Step 1: Setting up and starting Informer")
	fmt.Println("========================================")

	// Create shared informer factory with 30-second resync period
	informerFactory := externalversions.NewSharedInformerFactory(clientset, time.Second*30)

	// Get PodChaos informer from factory
	podChaosInformer := informerFactory.Api().V1alpha1().Podchaos()

	// Add event handlers to the informer
	podChaosInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pc := obj.(*v1alpha1.PodChaos)
			fmt.Printf("[INFORMER] PodChaos ADDED: %s/%s\n", pc.Namespace, pc.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pc := newObj.(*v1alpha1.PodChaos)
			fmt.Printf("[INFORMER] PodChaos UPDATED: %s/%s\n", pc.Namespace, pc.Name)
		},
		DeleteFunc: func(obj interface{}) {
			pc := obj.(*v1alpha1.PodChaos)
			fmt.Printf("[INFORMER] PodChaos DELETED: %s/%s\n", pc.Namespace, pc.Name)
		},
	})

	// Start the informer factory in a separate goroutine
	stopCh := make(chan struct{})
	defer close(stopCh)

	// Start informers in background
	go informerFactory.Start(stopCh)

	// Wait for caches to sync
	fmt.Println("Waiting for informer caches to sync...")
	if !cache.WaitForCacheSync(stopCh, podChaosInformer.Informer().HasSynced) {
		panic("Failed to sync caches")
	}
	fmt.Println("Informer caches synced successfully")

	// Get the lister from the informer
	podChaosLister := podChaosInformer.Lister()

	// Step 2: List existing PodChaos resources before creation
	fmt.Println("\n========================================")
	fmt.Println("Step 2: Listing existing PodChaos resources")
	fmt.Println("========================================")

	existingList, err := podChaosLister.Podchaos(namespace).List(labels.Everything())
	if err != nil {
		fmt.Printf("Error listing existing PodChaos: %v\n", err)
	} else {
		fmt.Printf("Found %d existing PodChaos resources\n", len(existingList))
		for _, pc := range existingList {
			fmt.Printf("  - %s (Action: %s)\n", pc.Name, pc.Spec.Action)
		}
	}

	// Step 3: Create a PodChaos resource for pod-kill
	fmt.Println("\n========================================")
	fmt.Println("Step 3: Creating PodChaos resource")
	fmt.Println("========================================")

	podChaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-pod-kill",
			Namespace: namespace,
		},
		Spec: v1alpha1.PodChaosSpec{
			Action: v1alpha1.PodKillAction,
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Mode:  v1alpha1.OneMode,
					Value: "1",
					Selector: v1alpha1.PodSelectorSpec{
						GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
							Namespaces: []string{namespace},
							LabelSelectors: map[string]string{
								"app": "test-app-client-go",
							},
						},
					},
				},
			},
		},
	}

	// Create the PodChaos resource
	createdPodChaos, err := clientset.ApiV1alpha1().Podchaos(namespace).Create(ctx, podChaos, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating PodChaos: %v\n", err)
		// If creation fails, still try to clean up existing resource
		defer func() {
			_ = clientset.ApiV1alpha1().Podchaos(namespace).Delete(ctx, "example-pod-kill", metav1.DeleteOptions{})
		}()
	} else {
		fmt.Printf("Successfully created PodChaos: %s\n", createdPodChaos.Name)

		// Set up deferred cleanup
		defer func() {
			fmt.Println("\n========================================")
			fmt.Println("Cleaning up resources (deferred)")
			fmt.Println("========================================")

			err := clientset.ApiV1alpha1().Podchaos(namespace).Delete(ctx, createdPodChaos.Name, metav1.DeleteOptions{})
			if err != nil {
				fmt.Printf("Error deleting PodChaos in defer: %v\n", err)
			} else {
				fmt.Printf("Successfully deleted PodChaos: %s\n", createdPodChaos.Name)
				// Give informer time to process the delete event
				time.Sleep(2 * time.Second)
			}
		}()
	}

	// Give informer time to process the create event
	fmt.Println("\nWaiting for informer to process create event...")
	time.Sleep(2 * time.Second)

	// Step 4: Use the lister to query resources
	fmt.Println("\n========================================")
	fmt.Println("Step 4: Using Lister to query resources")
	fmt.Println("========================================")

	// List all PodChaos resources in the namespace
	podChaosList, err := podChaosLister.Podchaos(namespace).List(labels.Everything())
	if err != nil {
		fmt.Printf("Error listing PodChaos using lister: %v\n", err)
	} else {
		fmt.Printf("Found %d PodChaos resources using lister:\n", len(podChaosList))
		for _, pc := range podChaosList {
			fmt.Printf("  - %s (Action: %s, Mode: %s)\n", pc.Name, pc.Spec.Action, pc.Spec.ContainerSelector.PodSelector.Mode)
		}
	}

	// Get a specific PodChaos using the lister
	if createdPodChaos != nil {
		specificPodChaos, err := podChaosLister.Podchaos(namespace).Get("example-pod-kill")
		if err != nil {
			fmt.Printf("Error getting specific PodChaos using lister: %v\n", err)
		} else {
			fmt.Printf("\nSuccessfully retrieved PodChaos '%s' using lister\n", specificPodChaos.Name)
			fmt.Printf("  Action: %s\n", specificPodChaos.Spec.Action)
			fmt.Printf("  Mode: %s\n", specificPodChaos.Spec.ContainerSelector.PodSelector.Mode)
			if len(specificPodChaos.Spec.ContainerSelector.PodSelector.Selector.LabelSelectors) > 0 {
				fmt.Printf("  Label Selectors: %v\n", specificPodChaos.Spec.ContainerSelector.PodSelector.Selector.LabelSelectors)
			}
		}
	}

	// Step 5: Demonstrate informer watching changes
	if createdPodChaos != nil {
		fmt.Println("\n========================================")
		fmt.Println("Step 5: Demonstrating Informer watching updates")
		fmt.Println("========================================")
		fmt.Println("Note: PodChaos spec updates are restricted by webhook")
		fmt.Println("Attempting to add/update annotations to trigger update event...")

		// Get the latest version before updating to avoid conflicts
		latestPodChaos, err := clientset.ApiV1alpha1().Podchaos(namespace).Get(ctx, createdPodChaos.Name, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Error getting latest PodChaos: %v\n", err)
		} else {
			// Update annotations to trigger an update event (spec changes are restricted)
			if latestPodChaos.Annotations == nil {
				latestPodChaos.Annotations = make(map[string]string)
			}
			latestPodChaos.Annotations["example.update"] = time.Now().Format(time.RFC3339)

			updatedPodChaos, err := clientset.ApiV1alpha1().Podchaos(namespace).Update(ctx, latestPodChaos, metav1.UpdateOptions{})
			if err != nil {
				fmt.Printf("Error updating PodChaos: %v\n", err)
			} else {
				fmt.Printf("Successfully updated PodChaos annotations\n")
				if updatedTime, ok := updatedPodChaos.Annotations["example.update"]; ok {
					fmt.Printf("  Updated timestamp: %s\n", updatedTime)
				}
			}
		}

		// Wait for informer to process the update event
		fmt.Println("Waiting for informer to process update event...")
		time.Sleep(2 * time.Second)
	}

	// Step 6: Compare direct API calls vs cached lister
	fmt.Println("\n========================================")
	fmt.Println("Step 6: Comparing API vs Lister (cached)")
	fmt.Println("========================================")

	// List using direct API call
	directList, err := clientset.ApiV1alpha1().Podchaos(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing PodChaos via API: %v\n", err)
	} else {
		fmt.Printf("Direct API call found %d PodChaos resources\n", len(directList.Items))
	}

	// Wait for lister cache to catch up if needed
	wait.Poll(100*time.Millisecond, 2*time.Second, func() (bool, error) {
		listerList, _ := podChaosLister.Podchaos(namespace).List(labels.Everything())
		return len(listerList) == len(directList.Items), nil
	})

	// List using lister (cached)
	listerList, err := podChaosLister.Podchaos(namespace).List(labels.Everything())
	if err != nil {
		fmt.Printf("Error listing PodChaos via lister: %v\n", err)
	} else {
		fmt.Printf("Lister (cached) found %d PodChaos resources\n", len(listerList))
	}

	fmt.Println("\n========================================")
	fmt.Println("Example completed successfully!")
	fmt.Println("Note: Resources will be cleaned up automatically")
	fmt.Println("========================================")
}
