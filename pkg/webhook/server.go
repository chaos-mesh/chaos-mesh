// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/pingcap/chaos-operator/pkg/util"
	"github.com/pingcap/chaos-operator/pkg/webhook/config"
	"github.com/pingcap/chaos-operator/pkg/webhook/inject"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WebHookServer is a server that handles mutating admission webhooks
type WebHookServer struct {
	server *http.Server
	cfg    *config.Config
}

func NewWebHookServer(param Parameters) *WebHookServer {
	sCert, err := util.ConfigTLS(param.CertFile, param.KeyFile)
	if err != nil {
		glog.Fatalf("failed to create scert file, %v", err)
	}

	cfg, err := config.LoadConfigDirectory(param.ConfigDirectory)
	if err != nil {
		glog.Fatalf("failed to load config, %v", err)
	}

	if param.AnnotationNamespace != "" {
		cfg.AnnotationNamespace = param.AnnotationNamespace
	}

	server := &http.Server{
		Addr:      param.Addr,
		TLSConfig: sCert,
	}

	ws := &WebHookServer{
		server: server,
		cfg:    cfg,
	}

	ws.router()
	return ws
}

func (w *WebHookServer) router() {
	mux := http.NewServeMux()
	mux.HandleFunc("/inject", w.injectHandler)

	w.server.Handler = mux
}

func (w *WebHookServer) Run(stopCh <-chan struct{}) error {
	go func() {
		if err := w.server.ListenAndServeTLS("", ""); err != nil {
			glog.Fatal(err)
		}
	}()

	<-stopCh
	glog.Info("Shutting webHook server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.server.Shutdown(ctx); err != nil {
		glog.Error(err)
		return err
	}

	return nil
}

// admitFunc is the type we use for all of our validators
type admitFunc func(request *v1beta1.AdmissionRequest, cfg *config.Config) *v1beta1.AdmissionResponse

func (w *WebHookServer) serve(res http.ResponseWriter, req *http.Request, admit admitFunc) {
	var (
		data []byte
		err  error
	)

	if req.Body != nil {
		data, err = ioutil.ReadAll(req.Body)
		if err != nil {
			glog.Errorf("failed to read body, %v", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if len(data) == 0 {
		glog.Warning("received empty payload")
		return
	}

	response := w.processReq(data, admit)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		glog.Errorf("failed to marshal response data, %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	glog.V(6).Infof("response: %s", string(responseJSON))

	if _, err := res.Write(responseJSON); err != nil {
		glog.Errorf("failed to write response to res, %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (w *WebHookServer) processReq(data []byte, admit admitFunc) *v1beta1.AdmissionReview {
	admissionReview, err := decode(data)
	if err != nil {
		glog.Errorf("failed to decode data. Reason: %v", err)
		admissionReview.Response = &v1beta1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
		return admissionReview
	}
	glog.Infof("received admission review request %s", admissionReview.Request.UID)
	glog.V(6).Infof("admission request: %+v", admissionReview.Request)

	admissionResponse := admit(admissionReview.Request, w.cfg)

	newAdmissionReview := &v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		newAdmissionReview.Response = admissionResponse
		if admissionReview.Request != nil {
			newAdmissionReview.Response.UID = admissionReview.Request.UID
		}
	}

	return newAdmissionReview
}

func (w *WebHookServer) injectHandler(res http.ResponseWriter, req *http.Request) {
	w.serve(res, req, inject.Inject)
}

func decode(data []byte) (*v1beta1.AdmissionReview, error) {
	var admissionReview v1beta1.AdmissionReview
	_, _, err := deserializer.Decode(data, nil, &admissionReview)
	return &admissionReview, err
}
