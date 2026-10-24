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
#


import http.client
import json
import math
import os
from pathlib import Path
import queue
import re
import socket
import ssl
import threading
from urllib.parse import urlencode

from cases import now, remaining
from checks import API, BUDGET, require, recovered, validate_object
from identity import require_gate


class APIError(RuntimeError):
    def __init__(self, status, value):
        self.status, self.value = status, value
        super().__init__(f'Kubernetes HTTP {status}: {value}')


class Response:
    def __init__(self, connection, method, path, body, headers, deadline):
        self.connection = connection
        self.deadline = deadline
        self.finished = threading.Event()
        self.timed_out = False
        self.socket = None
        self.timer = threading.Thread(target=self.cutoff, daemon=True)
        self.timer.start()
        try:
            connection.connect()
            self.socket = connection.sock
            remaining(deadline)
            connection.request(method, path, body, headers)
            self.response = connection.getresponse()
            remaining(deadline)
        except BaseException:
            self.close()
            raise

    def cutoff(self):
        if not self.finished.wait(max(0, self.deadline - now())):
            self.timed_out = True
            self.abort_socket()

    def abort_socket(self):
        sock = self.socket or self.connection.sock
        if sock is not None:
            try:
                sock.shutdown(socket.SHUT_RDWR)
            except OSError:
                pass
        self.connection.close()

    def read(self):
        remaining(self.deadline)
        require(not self.timed_out, 'absolute HTTP deadline expired')
        if self.connection.sock is not None:
            self.connection.sock.settimeout(remaining(self.deadline))
        data = self.response.read1(65536)
        remaining(self.deadline)
        return data

    def json(self):
        data = bytearray()
        while True:
            part = self.read()
            if not part:
                break
            data.extend(part)
            require(len(data) <= 512 * 1024, 'API response exceeded bound')
        return json.loads(data)

    def close(self):
        self.finished.set()
        self.abort_socket()
        self.timer.join(timeout=1)


class APIClient:
    def __init__(self):
        service = Path('/var/run/secrets/kubernetes.io/serviceaccount')
        self.context = ssl.create_default_context(cafile=str(service / 'ca.crt'))
        self.token = (service / 'token').read_text().strip()
        self.host = os.environ['KUBERNETES_SERVICE_HOST']
        self.port = int(os.environ['KUBERNETES_SERVICE_PORT_HTTPS'])
        self.armed = False

    def arm(self, receipt, guest, bpfki, lease):
        require_gate(receipt, guest, bpfki)
        require(lease['guest'] == guest.__dict__ and 0 < lease['deadline_boottime'] - now() <= BUDGET.active - 5,
                'active lease is expired or belongs to another guest')
        require(not self.armed, 'active lease can only be granted once')
        self.armed = True

    def open(self, method, path, deadline, obj=None):
        end = min(deadline, now() + 15) if 'watch=true' not in path else deadline
        connection = http.client.HTTPSConnection(self.host, self.port, context=self.context,
                                                 timeout=min(3, remaining(end)))
        body = None if obj is None else json.dumps(obj).encode()
        require(body is None or len(body) <= 16384, 'API input exceeded bound')
        return Response(connection, method, path, body,
                        {'Authorization': 'Bearer ' + self.token, 'Content-Type': 'application/json'}, end)

    def request(self, method, path, deadline, obj=None, missing=False):
        response = self.open(method, path, deadline, obj)
        try:
            value = response.json()
            status = response.response.status
            if missing and status == 404:
                require(value.get('kind') == 'Status' and value.get('reason') == 'NotFound'
                        and value.get('code') == 404, 'malformed Kubernetes NotFound response')
                return None
            if not 200 <= status < 300:
                raise APIError(status, value)
            return value
        finally:
            response.close()

    def pods(self, deadline):
        return self.request('GET', '/api/v1/pods', deadline)

    def get(self, name, deadline, missing=False):
        return self.request('GET', API + '/' + name, deadline, missing=missing)

    def create(self, obj, deadline):
        require(self.armed, 'KernelChaos creation requires qualification and active lease')
        return self.request('POST', API + '?fieldValidation=Strict', deadline, obj)

    def delete(self, name, uid, deadline):
        require(re.fullmatch(r'[0-9a-f-]{36}', uid), 'invalid deletion UID')
        return self.request('DELETE', API + '/' + name, deadline,
                            {'apiVersion': 'v1', 'kind': 'DeleteOptions', 'preconditions': {'uid': uid}})

    def watch(self, observer, resource_version, deadline):
        return Watch(self, observer, resource_version, deadline)


class Watch:
    def __init__(self, api, observer, resource_version, deadline):
        require(re.fullmatch(r'[0-9]+', resource_version), 'invalid known resourceVersion')
        self.observer, self.deadline = observer, deadline
        self.buffer = bytearray()
        self.deleted = self.recovered = self.closed = False
        self.queue = queue.Queue(maxsize=16)
        query = urlencode({'watch': 'true', 'fieldSelector': 'metadata.name=delete',
                           'resourceVersion': resource_version, 'allowWatchBookmarks': 'true',
                           'timeoutSeconds': max(1, math.ceil(remaining(deadline)))})
        self.response = api.open('GET', API + '?' + query, deadline)
        require(self.response.response.status == 200, 'watch request failed')
        self.reader = threading.Thread(target=self.receive, daemon=True)
        self.reader.start()
        observer.watches.append(self)
        observer.evidence.add('watch_started', {'resource_version': resource_version}, observer.case_uid)

    def receive(self):
        try:
            while not self.closed:
                data = self.response.read()
                self.queue.put(data, timeout=remaining(self.deadline))
                if not data:
                    return
        except BaseException as error:
            try:
                self.queue.put(error, timeout=0.1)
            except queue.Full:
                self.response.close()

    def poll(self, wait=0):
        require(not self.closed, 'watch is closed')
        remaining(self.deadline)
        try:
            chunk = self.queue.get(timeout=min(wait, remaining(self.deadline)))
        except queue.Empty:
            return
        if isinstance(chunk, BaseException):
            raise chunk
        require(chunk or self.deleted, 'watch ended before terminal deletion evidence')
        self.buffer.extend(chunk)
        require(len(self.buffer) <= 256 * 1024, 'watch object exceeded bound')
        while b'\n' in self.buffer:
            line, _, rest = self.buffer.partition(b'\n')
            self.buffer = bytearray(rest)
            event = json.loads(line)
            self.observer.evidence.add('watch_event', event, self.observer.case_uid)
            kind, obj = event['type'], event['object']
            require(kind in ('ADDED', 'MODIFIED', 'DELETED', 'BOOKMARK'), 'watch error or unknown event')
            if kind == 'BOOKMARK':
                continue
            validate_object(obj, 'delete', self.observer.case_uid)
            if recovered(obj):
                self.recovered = True
            if kind == 'DELETED':
                require(obj['metadata'].get('deletionTimestamp'), 'terminal object lacks deletionTimestamp')
                self.deleted = True
        remaining(self.deadline)

    def close(self):
        if self.closed:
            return
        self.closed = True
        self.response.close()
        self.reader.join(timeout=1)
