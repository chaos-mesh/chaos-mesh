// Copyright 2026 Chaos Mesh Authors.
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

#include "filter_layout.h"

#include <array>
#include <atomic>
#include <cerrno>
#include <climits>
#include <cstddef>
#include <cstdint>
#include <cstring>
#include <dirent.h>
#include <fcntl.h>
#include <iomanip>
#include <iostream>
#include <linux/bpf.h>
#include <new>
#include <poll.h>
#include <sstream>
#include <stdexcept>
#include <string>
#include <sys/mman.h>
#include <sys/stat.h>
#include <sys/syscall.h>
#include <sys/utsname.h>
#include <time.h>
#include <unistd.h>
#include <utility>
#include <vector>

#ifndef CM_FILTER_HEADER_SHA256
#error CM_FILTER_HEADER_SHA256 must identify the included pinned filter_layout.h
#endif
#if !defined(__linux__) || !defined(__x86_64__)
#error This fixture requires Linux amd64 interprocess lock-free atomics
#endif

namespace {
constexpr uint32_t kMagic = 0x434d4c46;
constexpr size_t kStateBytes = 4096;
constexpr uint32_t kLimit = 16;
constexpr uint64_t kTimeout = 3000000000ULL;
constexpr std::array<const char*, 5> kRoles{"target", "control", "bpfki", "daemon", "controller"};

struct alignas(64) SharedState {
  uint32_t magic;
  uint32_t version;
  uint32_t role;
  std::atomic<uint32_t> ready{0};
  std::atomic<uint32_t> requested{0};
  std::atomic<uint32_t> completed{0};
  int32_t result = 0;
  int32_t error = 0;
  int32_t close_error = 0;
};
static_assert(std::atomic<uint32_t>::is_always_lock_free, "shared u32 must be lock-free");
static_assert(sizeof(SharedState) <= kStateBytes, "state exceeds mapping");
static_assert(bpfki::filter_layout::field_count == 6, "unexpected pinned filter layout");

void Require(bool value, const std::string& message) {
  if (!value) throw std::runtime_error(message);
}
uint64_t Now() {
  timespec ts{};
  Require(clock_gettime(CLOCK_BOOTTIME, &ts) == 0, "clock_gettime failed");
  return uint64_t(ts.tv_sec) * 1000000000ULL + ts.tv_nsec;
}
void Pause() {
  const timespec delay{0, 1000000};
  nanosleep(&delay, nullptr);
}
std::string Quote(const std::string& input) {
  std::ostringstream out;
  out << '"';
  for (unsigned char ch : input) {
    if (ch == '"' || ch == '\\') out << '\\' << ch;
    else if (ch >= 32 && ch < 127) out << ch;
    else out << "\\u" << std::hex << std::setw(4) << std::setfill('0') << unsigned(ch);
  }
  out << '"';
  return out.str();
}
struct FD {
  int value = -1;
  FD() = default;
  explicit FD(int fd) : value(fd) {}
  FD(const FD&) = delete;
  FD& operator=(const FD&) = delete;
  FD(FD&& other) noexcept : value(other.value) { other.value = -1; }
  FD& operator=(FD&& other) noexcept {
    if (value >= 0) close(value);
    value = other.value;
    other.value = -1;
    return *this;
  }
  ~FD() { if (value >= 0) close(value); }
};
std::string Read(const std::string& path, size_t limit = 8192) {
  FD fd(open(path.c_str(), O_RDONLY | O_CLOEXEC | O_NOFOLLOW));
  Require(fd.value >= 0, "cannot open " + path);
  std::string data(limit + 1, '\0');
  size_t used = 0;
  while (used < data.size()) {
    ssize_t count = read(fd.value, data.data() + used, data.size() - used);
    if (count < 0 && errno == EINTR) continue;
    Require(count >= 0, "cannot read " + path);
    if (count == 0) break;
    used += size_t(count);
  }
  Require(used <= limit, "oversized " + path);
  data.resize(used);
  return data;
}
void RequireGuest() {
  utsname kernel{};
  Require(geteuid() == 0 && uname(&kernel) == 0 &&
              std::string(kernel.release) == "6.8.0-cm-kernel-ci", "not the owned lifecycle guest");
  auto marker = Read("/etc/cm-lifecycle/guest-id", 65);
  Require(marker.size() == 65 && marker.back() == '\n', "invalid guest marker");
  marker.pop_back();
  Require(marker.find_first_not_of("0123456789abcdef") == std::string::npos, "invalid guest marker");
  auto commandline = " " + Read("/proc/cmdline", 4096);
  if (!commandline.empty() && commandline.back() == '\n') commandline.back() = ' ';
  Require(commandline.find(" cm_lifecycle=" + marker + " ") != std::string::npos &&
              commandline.find(" root=/dev/vda ") != std::string::npos, "guest boot identity differs");
}
uint64_t Number(const std::string& text) {
  Require(!text.empty() && text.find_first_not_of("0123456789") == std::string::npos,
          "invalid unsigned integer");
  size_t used = 0;
  uint64_t value = std::stoull(text, &used);
  Require(used == text.size(), "incomplete integer");
  return value;
}
std::string BootID() {
  auto value = Read("/proc/sys/kernel/random/boot_id", 64);
  if (!value.empty() && value.back() == '\n') value.pop_back();
  return value;
}
uint64_t Namespace(pid_t pid) {
  struct stat st{};
  Require(stat(("/proc/" + std::to_string(pid) + "/ns/pid").c_str(), &st) == 0,
          "cannot inspect PID namespace");
  return st.st_ino;
}
uint64_t Starttime(pid_t pid) {
  auto data = Read("/proc/" + std::to_string(pid) + "/stat");
  auto end = data.rfind(')');
  Require(end != std::string::npos, "invalid process stat");
  std::istringstream fields(data.substr(end + 2));
  std::string word;
  for (int field = 3; field <= 22; ++field) Require(bool(fields >> word), "short process stat");
  return Number(word);
}
std::vector<uint64_t> NsPIDs(pid_t pid) {
  std::istringstream lines(Read("/proc/" + std::to_string(pid) + "/status"));
  std::string line;
  while (std::getline(lines, line)) {
    if (line.rfind("NSpid:", 0) != 0) continue;
    std::istringstream values(line.substr(6));
    std::vector<uint64_t> result;
    std::string word;
    while (values >> word) {
      Require(result.size() < 8, "too many PID namespaces");
      result.push_back(Number(word));
    }
    Require(!result.empty(), "empty NSpid");
    return result;
  }
  throw std::runtime_error("missing NSpid");
}
struct Identity {
  pid_t pid = 0;
  uint64_t start = 0;
  uint64_t ns = 0;
  std::vector<uint64_t> nspid;
  FD pidfd;
  void Check() const {
    pollfd wait{pidfd.value, POLLIN, 0};
    Require(poll(&wait, 1, 0) == 0, "owned process exited");
    Require(Starttime(pid) == start && Namespace(pid) == ns && NsPIDs(pid) == nspid,
            "owned process identity changed");
    Require(poll(&wait, 1, 0) == 0, "owned process exited during identity check");
  }
};

class Manifest {
  std::string data_;
  size_t position_ = 0;
 public:
  explicit Manifest(std::string data) : data_(std::move(data)) {}
  void Token(const std::string& token) {
    while (position_ < data_.size() && (data_[position_] == ' ' || data_[position_] == '\n')) ++position_;
    Require(data_.compare(position_, token.size(), token) == 0, "invalid fixed owned manifest");
    position_ += token.size();
  }
  uint64_t Integer() {
    size_t first = position_;
    while (position_ < data_.size() && data_[position_] >= '0' && data_[position_] <= '9') ++position_;
    return Number(data_.substr(first, position_ - first));
  }
  std::string String() {
    Token("\"");
    size_t end = data_.find('"', position_);
    Require(end != std::string::npos && end - position_ < 128, "invalid manifest string");
    auto value = data_.substr(position_, end - position_);
    Require(value.find_first_not_of("0123456789abcdef-") == std::string::npos,
            "invalid boot ID characters");
    position_ = end + 1;
    return value;
  }
  std::vector<uint64_t> PIDs() {
    Token("[");
    std::vector<uint64_t> result{Integer()};
    while (position_ < data_.size() && data_[position_] == ',') {
      Token(",");
      Require(result.size() < 8, "too many manifest PID namespaces");
      result.push_back(Integer());
    }
    Token("]");
    return result;
  }
  void Finish() {
    while (position_ < data_.size() && data_[position_] == '\n') ++position_;
    Require(position_ == data_.size(), "trailing manifest data");
  }
};

std::string StatePath(unsigned role) {
  Require(role < 2, "invalid worker role");
  return std::string("/var/lib/cm-lifecycle/") + kRoles[role] + "/state";
}
struct Mapping {
  FD fd;
  SharedState* state = nullptr;
  dev_t device = 0;
  ino_t inode = 0;
  unsigned role = 0;
  ~Mapping() { if (state) munmap(state, kStateBytes); }
  void Open(unsigned selected) {
    role = selected;
    fd = FD(open(StatePath(role).c_str(), O_RDWR | O_NOFOLLOW | O_CLOEXEC));
    Require(fd.value >= 0, "state open failed");
    struct stat st{};
    Require(fstat(fd.value, &st) == 0 && S_ISREG(st.st_mode) && st.st_nlink == 1 &&
                st.st_size == kStateBytes && st.st_uid == 10001 && st.st_gid == 10001 &&
                (st.st_mode & 0777) == 0600, "invalid state file");
    device = st.st_dev;
    inode = st.st_ino;
    void* area = mmap(nullptr, kStateBytes, PROT_READ | PROT_WRITE, MAP_SHARED, fd.value, 0);
    Require(area != MAP_FAILED, "state mmap failed");
    state = static_cast<SharedState*>(area);
    Require(state->magic == kMagic && state->version == 1 && state->role == role,
            "state ABI mismatch");
    Require(state->ready.is_lock_free() && state->requested.is_lock_free() &&
                state->completed.is_lock_free(), "shared atomics are not lock-free");
    volatile unsigned char* bytes = static_cast<unsigned char*>(area);
    for (size_t offset = 0; offset < kStateBytes; offset += 64) {
      unsigned char value = bytes[offset];
      (void)value;
    }
  }
  void Check() const {
    struct stat st{};
    Require(lstat(StatePath(role).c_str(), &st) == 0 && S_ISREG(st.st_mode) &&
                st.st_dev == device && st.st_ino == inode && st.st_nlink == 1 &&
                st.st_size == kStateBytes && st.st_uid == 10001 && st.st_gid == 10001 &&
                (st.st_mode & 0777) == 0600, "state file was replaced");
  }
};
void Prepare() {
  Require(geteuid() == 0, "prepare requires guest root");
  Require(mkdir("/var/lib/cm-lifecycle", 0755) == 0 || errno == EEXIST, "state root mkdir failed");
  for (unsigned role = 0; role < 2; ++role) {
    std::string directory = std::string("/var/lib/cm-lifecycle/") + kRoles[role];
    Require(mkdir(directory.c_str(), 0700) == 0, "state directory already exists or cannot be created");
    Require(chown(directory.c_str(), 10001, 10001) == 0, "state directory chown failed");
    FD fd(open(StatePath(role).c_str(), O_CREAT | O_EXCL | O_RDWR | O_NOFOLLOW | O_CLOEXEC, 0600));
    Require(fd.value >= 0 && ftruncate(fd.value, kStateBytes) == 0 &&
                fchown(fd.value, 10001, 10001) == 0, "state create failed");
    void* area = mmap(nullptr, kStateBytes, PROT_READ | PROT_WRITE, MAP_SHARED, fd.value, 0);
    Require(area != MAP_FAILED, "prepare mmap failed");
    auto* state = new (area) SharedState{};
    state->magic = kMagic;
    state->version = 1;
    state->role = role;
    Require(msync(area, kStateBytes, MS_SYNC) == 0, "state sync failed");
    munmap(area, kStateBytes);
  }
}
[[noreturn]] void Worker(unsigned role) {
  Require(getuid() == 10001 && geteuid() == 10001 && getgid() == 10001 && getpid() == 1,
          "worker must be UID/GID 10001 and container PID 1");
  Mapping mapping;
  mapping.Open(role);
  auto* state = mapping.state;
  Require(state->ready.load() == 0 && state->requested.load() == 0 && state->completed.load() == 0,
          "worker state is not fresh");
  volatile unsigned char stack[16384]{};
  for (size_t i = 0; i < sizeof(stack); i += 64) stack[i] = 1;
  uint32_t previous = 0;
  const timespec delay{0, 1000000};
  state->ready.store(1, std::memory_order_release);
  for (;;) {
    uint32_t request = state->requested.load(std::memory_order_acquire);
    if (request == previous) {
      syscall(SYS_nanosleep, &delay, nullptr);
      continue;
    }
    if (request != previous + 1 || request > kLimit) syscall(SYS_exit_group, 71);
    int fd = int(syscall(SYS_openat, AT_FDCWD, "/dev/null", O_RDONLY | O_CLOEXEC, 0));
    int saved = fd < 0 ? errno : 0;
    int close_error = 0;
    if (fd >= 0 && syscall(SYS_close, fd) != 0) close_error = errno;
    state->result = fd < 0 ? -1 : 0;
    state->error = saved;
    state->close_error = close_error;
    state->completed.store(request, std::memory_order_release);
    previous = request;
  }
}
uint32_t MapID(const std::string& path) {
  std::istringstream lines(Read(path, 4096));
  std::string line;
  uint32_t id = 0;
  while (std::getline(lines, line)) {
    if (line.rfind("map_id:", 0) != 0) continue;
    std::istringstream field(line.substr(7));
    std::string value, extra;
    Require(bool(field >> value) && !(field >> extra) && id == 0, "invalid owned map ID");
    auto parsed = Number(value);
    Require(parsed > 0 && parsed <= UINT32_MAX, "invalid map ID range");
    id = uint32_t(parsed);
  }
  return id;
}
struct Map {
  FD fd;
  bpf_map_info info{};
  std::string owner;
};
class Observer {
  std::array<Identity, 5> identities_;
  std::array<Mapping, 2> mappings_;
  std::array<uint32_t, 2> sequences_{};
  std::string boot_;
  uint64_t initial_ns_ = 0;
  Map retained_;
  void Budget(uint64_t deadline) { Require(Now() < deadline, "observation exceeded three seconds"); }
  Map OwnedMap(uint64_t deadline) {
    auto path = "/proc/" + std::to_string(identities_[2].pid) + "/fdinfo";
    DIR* directory = opendir(path.c_str());
    Require(directory != nullptr, "cannot enumerate owned bpfki fdinfo");
    Map selected;
    size_t count = 0;
    try {
      for (;;) {
        errno = 0;
        dirent* entry = readdir(directory);
        if (!entry) { Require(errno == 0, "fdinfo readdir failed"); break; }
        std::string name(entry->d_name);
        if (name == "." || name == "..") continue;
        Budget(deadline);
        Require(++count <= 256, "owned bpfki has more than 256 descriptors");
        Number(name);
        std::string owner = path + "/" + name;
        uint32_t id = MapID(owner);
        if (!id) continue;
        union bpf_attr attr{};
        attr.map_id = id;
        attr.open_flags = BPF_F_RDONLY;
        FD fd(int(syscall(SYS_bpf, BPF_MAP_GET_FD_BY_ID, &attr, sizeof(attr))));
        Require(fd.value >= 0, "read-only owned map open failed");
        bpf_map_info info{};
        attr = {};
        attr.info.bpf_fd = fd.value;
        attr.info.info_len = sizeof(info);
        attr.info.info = reinterpret_cast<uintptr_t>(&info);
        Require(syscall(SYS_bpf, BPF_OBJ_GET_INFO_BY_FD, &attr, sizeof(attr)) == 0 &&
                    attr.info.info_len >= offsetof(bpf_map_info, name) + sizeof(info.name) &&
                    info.id == id, "incomplete owned map metadata");
        Require(MapID(owner) == id, "owned map descriptor changed");
        if (std::string(info.name, strnlen(info.name, sizeof(info.name))) != "ctxs") continue;
        Require(info.type == BPF_MAP_TYPE_HASH && info.key_size == sizeof(uint32_t) &&
                    info.value_size == bpfki::filter_layout::ValueSize(1), "ctxs schema mismatch");
        Require(selected.fd.value < 0 || selected.info.id == id, "multiple owned ctxs maps");
        if (selected.fd.value < 0) {
          selected.fd = std::move(fd);
          selected.info = info;
          selected.owner = owner;
        }
      }
    } catch (...) { closedir(directory); throw; }
    closedir(directory);
    Budget(deadline);
    return selected;
  }
  int Lookup(const Map& map, uint32_t key, std::array<uint64_t, 7>& words) {
    union bpf_attr attr{};
    attr.map_fd = map.fd.value;
    attr.key = reinterpret_cast<uintptr_t>(&key);
    attr.value = reinterpret_cast<uintptr_t>(words.data());
    if (syscall(SYS_bpf, BPF_MAP_LOOKUP_ELEM, &attr, sizeof(attr)) == 0) return 0;
    return errno;
  }
 public:
  Observer() {
    struct stat st{};
    Require(lstat("/run/cm-lifecycle/owned.json", &st) == 0 && S_ISREG(st.st_mode) &&
                st.st_uid == 0 && st.st_nlink == 1 && (st.st_mode & 0777) == 0600,
            "owned manifest must be root-owned mode 0600");
    Manifest manifest(Read("/run/cm-lifecycle/owned.json", 8192));
    manifest.Token("{\"schema_version\":1,\"boot_id\":");
    boot_ = manifest.String();
    manifest.Token(",\"initial_pid_namespace\":");
    initial_ns_ = manifest.Integer();
    manifest.Token(",\"containers\":[");
    for (size_t index = 0; index < identities_.size(); ++index) {
      if (index) manifest.Token(",");
      manifest.Token("{\"role\":\"" + std::string(kRoles[index]) + "\",\"pid\":");
      auto pid = manifest.Integer();
      Require(pid > 1 && pid <= INT_MAX, "invalid owned PID");
      auto& identity = identities_[index];
      identity.pid = pid_t(pid);
      manifest.Token(",\"starttime\":"); identity.start = manifest.Integer();
      manifest.Token(",\"pid_namespace\":"); identity.ns = manifest.Integer();
      manifest.Token(",\"nspid\":"); identity.nspid = manifest.PIDs();
      manifest.Token("}");
      Require(identity.nspid.front() == pid && identity.start > 0, "invalid process identity");
      if (index < 2) Require(identity.ns != initial_ns_ && identity.nspid.size() == 2 &&
                                 identity.nspid.back() == 1, "worker is not container PID 1");
      if (index == 2 || index == 3) Require(identity.ns == initial_ns_ && identity.nspid.size() == 1,
                                            "bpfki/daemon is not in initial PID namespace");
      for (size_t prior = 0; prior < index; ++prior)
        Require(identities_[prior].pid != identity.pid, "duplicate owned process");
      identity.pidfd = FD(int(syscall(SYS_pidfd_open, identity.pid, 0)));
      Require(identity.pidfd.value >= 0, "pidfd_open failed");
      identity.Check();
    }
    manifest.Token("]}"); manifest.Finish();
    Check();
    uint64_t deadline = Now() + kTimeout;
    for (unsigned role = 0; role < 2; ++role) {
      mappings_[role].Open(role);
      while (mappings_[role].state->ready.load(std::memory_order_acquire) != 1) {
        Budget(deadline); Check(); Pause();
      }
      Require(mappings_[role].state->requested.load() == 0 &&
                  mappings_[role].state->completed.load() == 0, "observer state is not fresh");
    }
    Check();
  }
  void Check() {
    Require(BootID() == boot_ && Namespace(1) == initial_ns_, "guest identity changed");
    for (auto& identity : identities_) identity.Check();
  }
  std::string CheckJSON() {
    Check();
    std::ostringstream out;
    out << "\"checked_ns\":" << Now() << ",\"processes\":[";
    for (size_t i = 0; i < identities_.size(); ++i) {
      if (i) out << ',';
      out << "{\"role\":" << Quote(kRoles[i]) << ",\"pid\":" << identities_[i].pid
          << ",\"starttime\":" << identities_[i].start << ",\"pid_namespace\":" << identities_[i].ns << '}';
    }
    out << ']';
    return out.str();
  }
  std::string Operate(unsigned role) {
    uint64_t start = Now(), deadline = start + kTimeout;
    Check();
    auto& mapping = mappings_[role];
    mapping.Check();
    auto* state = mapping.state;
    uint32_t previous = sequences_[role];
    Require(previous < kLimit && state->completed.load(std::memory_order_acquire) == previous &&
                state->requested.load() == previous && state->ready.load() == 1,
            "operation sequence is not ready");
    uint32_t sequence = previous + 1;
    Budget(deadline);
    state->requested.store(sequence, std::memory_order_release);
    while (state->completed.load(std::memory_order_acquire) != sequence) {
      Budget(deadline); identities_[role].Check(); Pause();
    }
    int result = state->result, error = state->error, close_error = state->close_error;
    uint64_t finish = Now();
    sequences_[role] = sequence;
    Budget(deadline); mapping.Check(); Check();
    Require(close_error == 0 && (result == -1 || result == 0) &&
                ((result == 0 && error == 0) || (result == -1 && error > 0)),
            "invalid worker result or close failure");
    std::ostringstream out;
    out << "\"role\":" << Quote(kRoles[role]) << ",\"seq\":" << sequence
        << ",\"started_ns\":" << start << ",\"finished_ns\":" << finish
        << ",\"result\":" << result << ",\"errno\":" << error << ",\"close_errno\":" << close_error;
    return out.str();
  }
  std::string Sample(const std::string& phase) {
    uint64_t deadline = Now() + kTimeout;
    Check();
    Map owner = OwnedMap(deadline);
    if (phase == "absent") {
      Require(retained_.fd.value < 0 && owner.fd.value < 0, "unexpected existing ctxs map");
      Check(); Budget(deadline);
      return "\"phase\":\"absent\",\"owned_ctxs\":0";
    }
    if (phase == "active") {
      Require(owner.fd.value >= 0, "missing active owned ctxs map");
      if (retained_.fd.value < 0) retained_ = std::move(owner);
      else Require(owner.info.id == retained_.info.id && owner.owner == retained_.owner,
                   "active ctxs map or owner descriptor changed");
      Require(MapID(retained_.owner) == retained_.info.id, "ctxs owner changed before lookup");
    } else Require(retained_.fd.value >= 0 && owner.fd.value < 0,
                   "recovery did not release service ctxs map");
    std::array<uint64_t, 7> words{}, control{};
    int target_error = Lookup(retained_, uint32_t(identities_[0].pid), words);
    int control_error = Lookup(retained_, uint32_t(identities_[1].pid), control);
    Require(control_error == ENOENT, "control unexpectedly has a filter or lookup failed");
    if (phase == "active") {
      Require(target_error == 0 && words[bpfki::filter_layout::pct] == bpfki::filter_layout::probability_scale &&
                  words[bpfki::filter_layout::max_inject_times] == 0, "target filter mismatch");
      Require(MapID(retained_.owner) == retained_.info.id, "ctxs owner changed after lookup");
    } else Require(target_error == ENOENT, "recovery did not remove target key");
    Check(); Budget(deadline);
    std::ostringstream out;
    out << "\"phase\":" << Quote(phase) << ",\"map_id\":" << retained_.info.id
        << ",\"owner_fdinfo\":" << Quote(retained_.owner)
        << ",\"key\":" << identities_[0].pid << ",\"type\":" << retained_.info.type
        << ",\"key_size\":" << retained_.info.key_size << ",\"value_size\":" << retained_.info.value_size
        << ",\"max_entries\":" << retained_.info.max_entries << ",\"map_flags\":" << retained_.info.map_flags
        << ",\"map_name\":\"ctxs\",\"header_sha256\":" << Quote(CM_FILTER_HEADER_SHA256)
        << ",\"target_errno\":" << target_error << ",\"control_errno\":" << control_error;
    if (phase == "active") {
      out << ",\"inject_times\":" << words[bpfki::filter_layout::inject_times] << ",\"raw_hex\":\"";
      const auto* bytes = reinterpret_cast<const unsigned char*>(words.data());
      for (size_t i = 0; i < sizeof(words); ++i)
        out << std::hex << std::setw(2) << std::setfill('0') << unsigned(bytes[i]);
      out << '"';
    } else {
      Map after = OwnedMap(deadline);
      Require(after.fd.value < 0, "service reacquired ctxs during recovery sample");
      Check(); Budget(deadline);
      out << ",\"owned_ctxs\":0";
      retained_ = Map{};
    }
    return out.str();
  }
};
void Observe() {
  Require(geteuid() == 0, "observer requires guest root");
  Observer observer;
  std::cout << "{\"ok\":true,\"command\":\"ready\",\"header_sha256\":"
            << Quote(CM_FILTER_HEADER_SHA256) << "}" << std::endl;
  bool failed = false;
  for (size_t count = 0; count < 512; ++count) {
    char line[33]{};
    std::cin.getline(line, sizeof(line));
    if (std::cin.eof()) return;
    Require(bool(std::cin), "observer command exceeds 32 bytes");
    std::string command(line);
    if (command == "quit") {
      std::cout << "{\"ok\":true,\"command\":\"quit\"}" << std::endl;
      return;
    }
    try {
      Require(!failed, "observer is sealed after a failed sample");
      std::string result;
      if (command == "check") result = observer.CheckJSON();
      else if (command == "operate target") result = observer.Operate(0);
      else if (command == "operate control") result = observer.Operate(1);
      else if (command == "map absent") result = observer.Sample("absent");
      else if (command == "map active") result = observer.Sample("active");
      else if (command == "map recovered") result = observer.Sample("recovered");
      else throw std::runtime_error("unknown fixed observer command");
      std::cout << "{\"ok\":true,\"command\":" << Quote(command) << ',' << result << "}" << std::endl;
    } catch (const std::exception& error) {
      failed = true;
      std::cout << "{\"ok\":false,\"command\":" << Quote(command)
                << ",\"error\":" << Quote(error.what()) << "}" << std::endl;
    }
  }
  throw std::runtime_error("observer command limit exceeded");
}
}

int main(int argc, char** argv) {
  try {
    if (argc == 2 && std::string(argv[1]) == "prepare") { RequireGuest(); Prepare(); }
    else if (argc == 2 && std::string(argv[1]) == "observe") { RequireGuest(); Observe(); }
    else if (argc == 3 && std::string(argv[1]) == "worker" && std::string(argv[2]) == "target") Worker(0);
    else if (argc == 3 && std::string(argv[1]) == "worker" && std::string(argv[2]) == "control") Worker(1);
    else throw std::runtime_error("expected prepare, observe, or worker target|control");
    return 0;
  } catch (const std::exception& error) {
    std::cerr << error.what() << '\n';
    return 1;
  }
}
