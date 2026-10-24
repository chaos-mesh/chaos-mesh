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


import json
import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import probe


class BoundaryTests(unittest.TestCase):
    def test_other_profile_cannot_reach_cleanup(self):
        with tempfile.TemporaryDirectory() as directory, patch.dict(os.environ, GITHUB_RUN_ID='12', GITHUB_RUN_ATTEMPT='3'):
            with self.assertRaisesRegex(ValueError, 'profile differs'):
                probe.record_for('cm-kvm-12-2', Path(directory))
            self.assertEqual(list(Path(directory).iterdir()), [])

    def test_record_from_another_run_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory, patch.dict(os.environ, GITHUB_RUN_ID='12', GITHUB_RUN_ATTEMPT='3'):
            path = Path(directory)
            (path / 'result.json').write_text(json.dumps(dict(profile='cm-kvm-12-3', run_id='11', run_attempt='3')))
            with self.assertRaisesRegex(ValueError, 'another attempt'):
                probe.record_for('cm-kvm-12-3', path)

    def test_unready_or_wrong_node_cannot_prove_transport(self):
        node = {'metadata': {'name': 'other'}, 'status': {'conditions': [{'type': 'Ready', 'status': 'False'}]}}
        for name in ('other', 'cm-kvm-12-3'):
            node['metadata']['name'] = name
            with self.assertRaises(ValueError):
                probe.node_identity(json.dumps({'items': [node]}), 'cm-kvm-12-3')

    def test_unreadable_and_missing_prerequisites_are_distinct(self):
        base = '\n'.join('config_' + option + '=present' for option in probe.OPTIONS)
        base += '\nkernel=test\nconfig_source=/proc/config.gz\nsymbol=present\nheaders=present\n'
        for evidence, status in (('unavailable', 'unknown'), ('absent', 'unsupported'), ('present', 'observed')):
            self.assertEqual(probe.prerequisites(base + 'allowlist=' + evidence)['status'], status)
        with self.assertRaisesRegex(ValueError, 'incomplete'):
            probe.prerequisites(base)

    def test_skipped_start_does_not_delete_anything(self):
        with tempfile.TemporaryDirectory() as directory, patch.dict(os.environ, START_OUTCOME='skipped'):
            with patch.object(Path, 'home', return_value=Path(directory)), patch.object(probe, 'command') as command:
                record = {'transport': {'status': 'untested'}}
                probe.cleanup('cm-kvm-12-3', Path(directory), record)
                command.assert_not_called()
                self.assertEqual(record['cleanup']['status'], 'not_created')

    def test_failed_delete_is_not_retried_or_reported_removed(self):
        with tempfile.TemporaryDirectory() as directory, patch.dict(os.environ, START_OUTCOME='failure'):
            calls = []
            def run(output, name, args, timeout=30):
                calls.append(args)
                return dict(returncode=1 if name in {'minikube-delete', 'cleanup-unbound-domain'} else 0, stdout='', stderr='')
            with patch.object(Path, 'home', return_value=Path(directory)), patch.object(probe, 'command', side_effect=run):
                record = {'transport': {'status': 'untested'}}
                with self.assertRaisesRegex(ValueError, 'cleanup did not complete'):
                    probe.cleanup('cm-kvm-12-3', Path(directory), record)
                self.assertEqual(sum('delete' in args for args in calls), 1)
                self.assertEqual(record['cleanup']['status'], 'failed')


if __name__ == '__main__':
    unittest.main()
