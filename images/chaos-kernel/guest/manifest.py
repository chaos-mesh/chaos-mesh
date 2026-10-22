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


import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import stat
import subprocess
import sys
import tarfile


BPFKI_COMMIT = "ea3e0906d657ee774c0aaf9b2dd1f30d7197b938"
EXECUTABLES = ["bpfki", "time-chaos", "fail-bio", "fail-mount", "fail-ext4-mount", "fail-alloc-page"]
PROGRAMS = ["clock_gettime", "gettimeofday"]


def require(condition, message):
    if not condition:
        raise ValueError(message)


def digest(path):
    result = hashlib.sha256()
    with path.open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            result.update(block)
    return result.hexdigest()


def command(*args):
    return subprocess.check_output(args, text=True, timeout=60).strip()


def write_json(path, value):
    path.write_text(json.dumps(value, sort_keys=True, indent=2) + "\n")


def resolve(root, name, records):
    require(name.startswith("/"), "expected absolute image path")
    pending = list(PurePosixPath(name).parts[1:])
    resolved = []
    links = 0
    while pending:
        part = pending.pop(0)
        if part == "..":
            require(bool(resolved), "symlink escapes root: " + name)
            resolved.pop()
            continue
        if part == ".":
            continue
        path = root.joinpath(*resolved, part)
        mode = path.lstat().st_mode
        if stat.S_ISLNK(mode):
            links += 1
            require(links <= 40, "symlink loop: " + name)
            target = os.readlink(path)
            records["/" + "/".join(resolved + [part])] = {"symlink": target}
            components = list(PurePosixPath(target).parts)
            if target.startswith("/"):
                resolved = []
                components = components[1:]
            pending = components + pending
        else:
            require(not pending or stat.S_ISDIR(mode), "non-directory in runtime path: " + name)
            resolved.append(part)
    return root.joinpath(*resolved), "/" + "/".join(resolved)


def inventory(root, names):
    records = {}
    visited = set()

    def visit(name):
        path, canonical = resolve(root, name, records)
        if canonical in visited:
            return
        visited.add(canonical)
        mode = path.stat().st_mode
        if stat.S_ISDIR(mode):
            records[canonical] = {"directory": True}
            for child in sorted(path.iterdir()):
                visit(canonical.rstrip("/") + "/" + child.name)
        else:
            require(stat.S_ISREG(mode), "not a regular runtime file: " + name)
            records[canonical] = {"sha256": digest(path), "mode": stat.S_IMODE(mode)}

    for name in names:
        require(name.startswith("/"), "expected absolute image path")
        visit(name)
    return records


def libraries(targets):
    result = {}
    for target in targets:
        output = command("ldd", target)
        require("not found" not in output, "missing library for " + target)
        paths = sorted(set(re.findall(r"(?:=>\s+|^\s*)(/\S+)", output, re.MULTILINE)))
        require(bool(paths), "no dynamic dependency evidence for " + target)
        result[target] = paths
    return result


def capture():
    require(command("git", "-C", "/bpfki", "rev-parse", "HEAD") == BPFKI_COMMIT, "unexpected image bpfki revision")
    bcc_commit = command("git", "-C", "/bcc", "rev-parse", "HEAD")
    require(bcc_commit == command("git", "-C", "/bcc", "rev-parse", "v0.37.0^{commit}"), "unexpected BCC source revision")
    installed = ["/usr/local/bin/" + name for name in EXECUTABLES + PROGRAMS]
    for name in installed[:len(EXECUTABLES)]:
        require(os.access(name, os.X_OK), "missing installed executable: " + name)
    for name in PROGRAMS:
        require(digest(Path("/usr/local/bin") / name) == digest(Path("/bpfki/progs") / name), "installed program differs from source")
    bcc_libraries = sorted({str(path) for pattern in ["lib/libbcc*.so*", "lib/*/libbcc*.so*"]
                            for path in Path("/usr").glob(pattern)})
    require(bool(bcc_libraries), "source-installed BCC library is missing")
    roots = installed + ["/usr/include/bcc", "/usr/lib/llvm-18", "/usr/lib/clang",
                         command("clang-18", "-print-resource-dir")] + bcc_libraries
    targets = installed[:len(EXECUTABLES)] + ["/usr/bin/clang-18"] + bcc_libraries
    closure = libraries(targets)
    roots += targets + [path for paths in closure.values() for path in paths]
    runtime = {"roots": sorted(set(roots)), "libraries": closure}
    runtime["files"] = inventory(Path("/"), runtime["roots"])
    runtime["build"] = {
        "bpfki_commit": BPFKI_COMMIT, "bcc_tag": "v0.37.0", "bcc_commit": bcc_commit,
        "bcc_libbpf_commit": command("git", "-C", "/bcc/src/cc/libbpf", "rev-parse", "HEAD"),
        "bcc_cmake_cache": Path("/bcc/build/CMakeCache.txt").read_text(),
        "bcc_install_manifest_sha256": digest(Path("/bcc/build/install_manifest.txt")),
        "bpfki_cmake_cache": Path("/bpfki/build/CMakeCache.txt").read_text(),
        "bpfki_install_manifest_sha256": digest(Path("/bpfki/build/install_manifest.txt")),
    }
    return runtime


def check(runtime, root):
    actual = inventory(root, runtime["roots"])
    require(actual == runtime["files"], "candidate, library or compiler resource changed during packaging/export")
    if root == Path("/"):
        require(libraries(runtime["libraries"]) == runtime["libraries"], "dynamic dependency resolution changed")


def image_archive(archive, inspection):
    require(len(inspection) == 1, "expected one image")
    image = inspection[0]
    require(image["Architecture"] == "amd64" and image["Os"] == "linux", "expected linux/amd64 image")
    with tarfile.open(archive) as source:
        manifest_bytes = source.extractfile("manifest.json").read()
        manifest = json.loads(manifest_bytes)
        require(len(manifest) == 1, "expected one archived image")
        config_bytes = source.extractfile(manifest[0]["Config"]).read()
        require("sha256:" + hashlib.sha256(config_bytes).hexdigest() == image["Id"], "archive config differs from image ID")
        config = json.loads(config_bytes)
        require(config["architecture"] == "amd64" and config["os"] == "linux", "archive platform differs from image")
        require(config["rootfs"]["diff_ids"] == image["RootFS"]["Layers"], "archive layers differ from image")
        require(len(manifest[0]["Layers"]) == len(image["RootFS"]["Layers"]), "archive layer count differs from image")
    return {"cm_commit": os.environ["GITHUB_SHA"], "image_id": image["Id"],
            "architecture": image["Architecture"], "layers": image["RootFS"]["Layers"],
            "archive_sha256": digest(archive), "manifest_sha256": hashlib.sha256(manifest_bytes).hexdigest()}


def bundle(root, kernel, output):
    runtime = json.loads((output / "runtime.json").read_text())
    check(runtime, root)
    release = (kernel / "bpfki-kernel-release").read_text().strip()
    require(release == "6.8.0-bpfki-ci", "unexpected guest kernel release")
    config = (kernel / ".config").read_text().splitlines()
    requested = (root / "opt/bpfki/tests/kernel/vm/kernel.config").read_text().splitlines()
    for line in requested:
        if line.startswith("CONFIG_") and (line.endswith("=y") or line.startswith("CONFIG_LOCALVERSION=")):
            require(line in config, "missing kernel configuration: " + line)
    require('"' + release + '"' in (kernel / "include/generated/utsrelease.h").read_text(), "kernel header release differs")
    sources = inventory(root / "opt/bpfki", ["/"])
    require(sources == json.loads((output / "source-files.json").read_text()), "pinned source or harness changed")
    require(os.readlink(root / "opt/candidate") == "/usr/local", "candidate prefix does not use the image installation")
    source_control = json.loads((root / "home/runner/work/_temp/bpfki-source-control.json").read_text())
    require(source_control["candidate_revision"] == BPFKI_COMMIT, "source control uses a different candidate revision")
    return {"image": json.loads((output / "image.json").read_text()), "runtime": runtime,
            "source_control": source_control,
            "source_files": sources,
            "kernel": {"release": release, "config_sha256": digest(kernel / ".config"),
                       "bzimage_sha256": digest(kernel / "arch/x86/boot/bzImage"),
                       "build_fingerprint": (kernel / ".bpfki-complete").read_text().strip(),
                       "headers": inventory(kernel, ["/include", "/arch/x86/include"])}}


def result(output):
    expected = output / "image-provenance.json"
    require(digest(expected) == digest(output / "guest/image-provenance.json"), "guest image provenance differs")
    provenance = json.loads(expected.read_text())
    directories = list((output / "guest").glob("bpfki-kernel-*"))
    require(len(directories) == 1, "expected one guest comparison")
    evidence = directories[0]
    environment = json.loads((evidence / "environment.json").read_text())
    summary = json.loads((evidence / "summary.json").read_text())
    vm = json.loads((output / "summary.json").read_text())
    require(vm["passed"] is True and vm["exit_code"] == 0 and vm["qemu_exit_code"] == 0, "VM failed")
    require(vm["kernel"] == environment["kernel"] == provenance["kernel"]["release"] and
            environment["native_architecture"] == "x86_64", "guest kernel or architecture differs")
    require(environment["outcome"] == "qualified" and environment["config_sha256"] == provenance["kernel"]["config_sha256"], "kernel qualification differs")
    require(environment["source_control"] == provenance["source_control"], "source control provenance differs")
    for name in ["bpfki"] + PROGRAMS:
        path = "/usr/local/bin/" + name
        require(environment["installations"]["candidate"][path] == provenance["runtime"]["files"][path]["sha256"], "guest candidate hash differs")
    require(summary["passed"] is True and summary["cleanup"] is True, "comparison or disk cleanup failed")
    cases = summary["cases"]
    require(len(cases) == 6 and {(case["installation"], case["fault"]) for case in cases} ==
            {(installation, fault) for installation in ["source_control", "candidate"] for fault in ["SLAB", "PAGE", "BIO"]}, "incomplete fault comparison")
    for case in cases:
        require(case["communication"] is True and case["process_cleanup"] is True and not case["error"], "case setup/cleanup failed")
        if case["installation"] == "source_control":
            require(case["known_layout_bug_reproduced"] is True and case["functional_pass"] is False and
                    case["passed"] is False, "original defect was not reproduced")
        else:
            require(all(case[key] is True for key in ["passed", "functional_pass", "zero_probability_verified", "finite_limit_verified"]), "candidate profile failed")
            require(case["same_process_recoveries"] == 2 and case["observed_faults"] == 5 and case["unaffected_controls"] == 2, "fault/control/recovery evidence is incomplete")
            if case["fault"] == "SLAB":
                require(case["invalid_update_preserved"] is True, "SLAB invalid update changed the active fault")
    require((evidence / "events.jsonl").stat().st_size > 0, "missing operation evidence")
    for case in cases:
        require((evidence / (case["installation"] + "-" + case["fault"] + ".log")).is_file(), "missing service log")
    events = [json.loads(line) for line in (evidence / "events.jsonl").read_text().splitlines()]
    probes = [event for event in events if event.get("installation") == "candidate" and
              event.get("fault") == "BIO" and event.get("stage") == "owned_probe" and
              event.get("operation") == "finite_fault"]
    require(len(probes) == 2 and [probe["phase"] for probe in probes] == ["before", "after"], "missing BIO probe mode evidence")
    identity = ["sequence", "pid", "starttime", "service_pid", "service_starttime", "map_id",
                "key", "perf_fd", "program_id", "program_fd", "program_tag", "address_hex"]
    require(all(probes[0][key] == probes[1][key] for key in identity), "owned BIO probe changed")
    require(probes[0]["sequence"] == 3 and probes[0]["key"] == probes[0]["pid"], "probe is not the finite target operation")
    for probe in probes:
        require(probe["identity_stable"] is True and probe["gaps"] == [], "incomplete owned probe identity")
        require(probe["kprobes_read"] is True and probe["kprobe_matches"] == len(probe["kprobes"]) > 0, "missing owned kprobe status")
        require(probe["perf_matches"] == probe["program_matches"] == 1, "ambiguous owned probe")
        require(len(probe["queries"]) == 2 and {query["point"] for query in probe["queries"]} == {"enumeration", "recheck"}, "missing probe query recheck")
        for query in probe["queries"]:
            require(query["errno"] == 0 and query["fd_type"] == 2 and
                    query["symbol"] == "should_fail_bio" and query["offset"] == 0, "wrong logical BIO probe")
            require(query["perf_fd"] == probe["perf_fd"] and query["program_id"] == probe["program_id"] and
                    query["address_hex"] == probe["address_hex"] and int(query["address_hex"], 16) != 0, "probe query identity mismatch")
        for status in probe["kprobes"]:
            address, kind, symbol, *_ = status["raw"].split()
            require(int(address, 16) == int(probe["address_hex"], 16) and kind == "k" and
                    symbol == status["symbol"] and symbol in {"should_fail_bio+0x0", "should_fail_bio+0x4"}, "wrong owned kprobe address")
            require(status["FTRACE"] is True and all(status[flag] is False for flag in ["OPTIMIZED", "DISABLED", "GONE"]), "BIO probe did not use the qualified FTRACE mode")
    return {"passed": True, "bio_probe_mode": "FTRACE", "image_id": provenance["image"]["image_id"],
            "provenance_sha256": digest(expected), "events_sha256": digest(evidence / "events.jsonl")}


def main():
    mode, *args = sys.argv[1:]
    if mode == "capture":
        print(json.dumps(capture(), sort_keys=True, indent=2))
    elif mode == "check":
        check(json.loads(Path(args[0]).read_text()), Path(args[1]))
    elif mode == "check-build-paths":
        runtime = json.loads(Path(args[0]).read_text())
        require(not any(path == "/bcc" or path.startswith("/bcc/") or path == "/bpfki/build" or path.startswith("/bpfki/build/")
                        for path in runtime["files"]), "runtime depends on build paths")
    elif mode == "image":
        print(json.dumps(image_archive(Path(args[0]), json.loads(Path(args[1]).read_text())), sort_keys=True, indent=2))
    elif mode == "source":
        print(json.dumps(inventory(Path(args[0]), ["/"]), sort_keys=True, indent=2))
    elif mode == "bundle":
        write_json(Path(args[2]) / "image-provenance.json", bundle(*map(Path, args)))
    elif mode == "result":
        write_json(Path(args[0]) / "image-result.json", result(Path(args[0])))
    else:
        raise ValueError("unknown manifest operation")


if __name__ == "__main__":
    main()
