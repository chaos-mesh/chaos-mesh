#!/usr/bin/env python3
# Copyright 2026 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

def image_values(record):
    registry, name = record['tag'].split('/', 1)
    repository, tag = name.rsplit(':', 1)
    return {'registry': registry, 'repository': repository, 'tag': tag}


def chart_values(images):
    values = {
        'controllerManager': {'replicaCount': 1, 'hostNetwork': False, 'imagePullPolicy': 'Never',
                              'image': image_values(images['controller']),
                              'resources': {'requests': {'cpu': '25m', 'memory': '256Mi'},
                                            'limits': {'cpu': '1', 'memory': '1Gi'}}},
        'chaosDaemon': {'hostNetwork': False, 'runtime': 'containerd',
                        'socketPath': '/run/containerd/containerd.sock', 'imagePullPolicy': 'Never',
                        'mtls': {'enabled': True}, 'image': image_values(images['daemon']),
                        'resources': {'requests': {'cpu': '100m', 'memory': '256Mi'},
                                      'limits': {'cpu': '500m', 'memory': '512Mi'}}},
        'bpfki': {'create': True, 'imagePullPolicy': 'Never', 'image': image_values(images['kernel']),
                  'resources': {'requests': {'cpu': '250m', 'memory': '512Mi'},
                                'limits': {'cpu': '1500m', 'memory': '2Gi'}}},
        'dashboard': {'create': False}, 'dnsServer': {'create': False}, 'prometheus': {'create': False},
        'chaosDlv': {'enable': False}, 'enableProfiling': False,
        'webhook': {'certManager': {'enabled': False}, 'FailurePolicy': 'Fail'},
    }
    return values


def worker_pods(worker_image):
    pods = []
    for role in ['target', 'control']:
        pods.append({'apiVersion': 'v1', 'kind': 'Pod',
                     'metadata': {'name': role, 'namespace': 'cm-kernel-ci', 'labels': {'cm-lifecycle-role': role}},
                     'spec': {'restartPolicy': 'Never', 'automountServiceAccountToken': False,
                              'terminationGracePeriodSeconds': 1,
                              'containers': [{'name': 'worker', 'image': worker_image,
                                              'imagePullPolicy': 'Never', 'args': ['worker', role],
                                              'securityContext': {'runAsUser': 10001, 'runAsGroup': 10001,
                                                                  'runAsNonRoot': True, 'privileged': False,
                                                                  'allowPrivilegeEscalation': False,
                                                                  'readOnlyRootFilesystem': True,
                                                                  'capabilities': {'drop': ['ALL']}},
                                              'resources': {'requests': {'cpu': '10m', 'memory': '16Mi'},
                                                            'limits': {'cpu': '100m', 'memory': '32Mi'}},
                                              'volumeMounts': [{'name': 'state', 'mountPath': '/var/lib/cm-lifecycle/' + role}]}],
                              'volumes': [{'name': 'state', 'hostPath': {'path': '/var/lib/cm-lifecycle/' + role,
                                                                       'type': 'Directory'}}]}})
    return pods
