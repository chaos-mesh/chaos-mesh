// Copyright 2020 PingCAP, Inc.
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

package v1alpha1

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Stress chaos is a chaos to generate plenty of stresses over a collection of pods.
// A sidecar will be injected along with the target pod during creating. It's the
// sidecar which generates stresses or cancels them. For now, we use stress-ng as
// the stress generator for the chaos.

// +kubebuilder:object:root=true

// StressChaos is the Schema for the stresschaos API
type StressChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a time chaos experiment
	Spec StressChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the time chaos experiment
	Status StressChaosStatus `json:"status"`
}

// StressChaosSpec defines the desired state of StressChaos
type StressChaosSpec struct {
	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the max % of pods the server can do chaos action.
	// If `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the % of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Stressors defines plenty of stressors
	Stressors Stressors `json:"stressors"`

	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	// +optional
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Next time when this action will be applied again
	// +optional
	NextStart *metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered
	// +optional
	NextRecover *metav1.Time `json:"nextRecover,omitempty"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *StressChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *StressChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *StressChaosSpec) GetValue() string {
	return in.Value
}

// StressChaosStatus defines the observed state of StressChaos
type StressChaosStatus struct {
	ChaosStatus `json:",inline"`
	// Instance always specifies a stressing instance
	Instance string `json:"instance"`
}

// GetDuration gets the duration of StressChaos
func (in *StressChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetNextStart gets NextStart field of StressChaos
func (in *StressChaos) GetNextStart() time.Time {
	if in.Spec.NextStart == nil {
		return time.Time{}
	}
	return in.Spec.NextStart.Time
}

// SetNextStart sets NextStart field of StressChaos
func (in *StressChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Spec.NextStart = nil
		return
	}

	if in.Spec.NextStart == nil {
		in.Spec.NextStart = &metav1.Time{}
	}
	in.Spec.NextStart.Time = t
}

// GetNextRecover get NextRecover field of StressChaos
func (in *StressChaos) GetNextRecover() time.Time {
	if in.Spec.NextRecover == nil {
		return time.Time{}
	}
	return in.Spec.NextRecover.Time
}

// SetNextRecover sets NextRecover field of StressChaos
func (in *StressChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Spec.NextRecover = nil
		return
	}

	if in.Spec.NextRecover == nil {
		in.Spec.NextRecover = &metav1.Time{}
	}
	in.Spec.NextRecover.Time = t
}

// GetScheduler returns the scheduler of StressChaos
func (in *StressChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetStatus returns the status of StressChaos
func (in *StressChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// IsDeleted returns whether this resource has been deleted
func (in *StressChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

type Stressors struct {
	// VmStressor stresses virtual memory out
	// +optional
	VmStressor *VmStressor `json:"vm,omitempty"`
	// CpuStressor stresses CPU out
	// +optional
	CpuStressor *CpuStressor `json:"cpu,omitempty"`
}

// Normalize the stressors to comply with stress-ng
func (in *Stressors) Normalize() string {
	stressors := ""
	if in.VmStressor != nil {
		stressors += fmt.Sprintf("--vm %d --vm-bytes %s",
			in.VmStressor.Workers, in.VmStressor.Bytes)
	}
	if in.CpuStressor != nil {
		stressors += fmt.Sprintf("--cpu %d --cpu-load %d --cpu-method %s",
			in.CpuStressor.Workers, in.CpuStressor.Load, in.CpuStressor.Method)
	}
	return stressors
}

type Stressor struct {
	// Workers specifies N workers to apply the stressor.
	Workers int `json:"workers"`
}

type VmStressor struct {
	Stressor `json:",inline"`

	// Bytes specifies N bytes consumed per vm worker, default is the total available memory.
	// One can specify the size as % of total available memory or in units of B, KB/KiB,
	// MB/MiB, GB/GiB, TB/TiB.
	// +kubebuilder:default="100%"
	// +optional
	Bytes string `json:"bytes,omitempty"`
}

type CpuMethod string

const (
	CpuMethodAll              CpuMethod = "all"
	CpuMethodAckermann        CpuMethod = "ackermann"
	CpuMethodBitops           CpuMethod = "bitops"
	CpuMethodCallfunc         CpuMethod = "callfunc"
	CpuMethodCfloat           CpuMethod = "cfloat"
	CpuMethodCdouble          CpuMethod = "cdouble"
	CpuMethodClongdouble      CpuMethod = "clongdouble"
	CpuMethodCorrelate        CpuMethod = "correlate"
	CpuMethodCrc16            CpuMethod = "crc16"
	CpuMethodDecimal32        CpuMethod = "decimal32"
	CpuMethodDecimal64        CpuMethod = "decimal64"
	CpuMethodDecimal128       CpuMethod = "decimal128"
	CpuMethodDither           CpuMethod = "dither"
	CpuMethodDjb2a            CpuMethod = "djb2a"
	CpuMethodDouble           CpuMethod = "double"
	CpuMethodEuler            CpuMethod = "euler"
	CpuMethodExplog           CpuMethod = "explog"
	CpuMethodFactorial        CpuMethod = "factorial"
	CpuMethodFibonacci        CpuMethod = "fibonacci"
	CpuMethodFft              CpuMethod = "fft"
	CpuMethodFloat            CpuMethod = "float"
	CpuMethodFloat16          CpuMethod = "float16"
	CpuMethodFloat32          CpuMethod = "float32"
	CpuMethodFloat80          CpuMethod = "float80"
	CpuMethodFloat128         CpuMethod = "float128"
	CpuMethodFnv1a            CpuMethod = "fnv1a"
	CpuMethodGamma            CpuMethod = "gamma"
	CpuMethodGcd              CpuMethod = "gcd"
	CpuMethodGray             CpuMethod = "gray"
	CpuMethodHamming          CpuMethod = "hamming"
	CpuMethodHanoi            CpuMethod = "hanoi"
	CpuMethodHyperbolic       CpuMethod = "hyperbolic"
	CpuMethodIdct             CpuMethod = "idct"
	CpuMethodInt8             CpuMethod = "int8"
	CpuMethodInt16            CpuMethod = "int16"
	CpuMethodInt32            CpuMethod = "int32"
	CpuMethodInt64            CpuMethod = "int64"
	CpuMethodInt128           CpuMethod = "int128"
	CpuMethodInt32float       CpuMethod = "int32float"
	CpuMethodInt32double      CpuMethod = "int32double"
	CpuMethodInt32longdouble  CpuMethod = "int32longdouble"
	CpuMethodInt64float       CpuMethod = "int64float"
	CpuMethodInt64double      CpuMethod = "int64double"
	CpuMethodInt64longdouble  CpuMethod = "int64longdouble"
	CpuMethodInt128float      CpuMethod = "int128float"
	CpuMethodInt128double     CpuMethod = "int128double"
	CpuMethodInt128longdouble CpuMethod = "int128longdouble"
	CpuMethodInt128decimal32  CpuMethod = "int128decimal32"
	CpuMethodInt128decimal64  CpuMethod = "int128decimal64"
	CpuMethodInt128decimal128 CpuMethod = "int128decimal128"
	CpuMethodJenkin           CpuMethod = "jenkin"
	CpuMethodJmp              CpuMethod = "jmp"
	CpuMethodLn2              CpuMethod = "ln2"
	CpuMethodLongdouble       CpuMethod = "longdouble"
	CpuMethodLoop             CpuMethod = "loop"
	CpuMethodMatrixprod       CpuMethod = "matrixprod"
	CpuMethodNsqrt            CpuMethod = "nsqrt"
	CpuMethodOmega            CpuMethod = "omega"
	CpuMethodParity           CpuMethod = "parity"
	CpuMethodPhi              CpuMethod = "phi"
	CpuMethodPi               CpuMethod = "pi"
	CpuMethodPjw              CpuMethod = "pjw"
	CpuMethodPrime            CpuMethod = "prime"
	CpuMethodPsi              CpuMethod = "psi"
	CpuMethodQueens           CpuMethod = "queens"
	CpuMethodRand             CpuMethod = "rand"
	CpuMethodRand48           CpuMethod = "rand48"
	CpuMethodRgb              CpuMethod = "rgb"
	CpuMethodSdbm             CpuMethod = "sdbm"
	CpuMethodSieve            CpuMethod = "sieve"
	CpuMethodStats            CpuMethod = "stats"
	CpuMethodSqrt             CpuMethod = "sqrt"
	CpuMethodTrig             CpuMethod = "trig"
	CpuMethodUnion            CpuMethod = "union"
	CpuMethodZeta             CpuMethod = "zeta"
)

type CpuStressor struct {
	Stressor `json:",inline"`
	// Load specifies P percent loading per CPU worker. 0 is effectively a sleep (no load) and 100
	// is full loading.
	// +kubebuilder:default=100
	// +optional
	Load int `json:"load,omitempty"`
	// Method specify a cpu stress method. By default, all the stress methods are exercised
	// sequentially, however one can specify just one method to be used if required. Available cpu
	// stress methods are described as follows:
	// Method 			Description
	// all				iterate over all the below cpu stress methods
	// ackermann    	Ackermann function: compute A(3, 10), where:
	//              	  A(m, n) = n + 1 if m = 0;
	//              	  A(m - 1, 1) if m > 0 and n = 0;
	//              	  A(m - 1, A(m, n - 1)) if m > 0 and n > 0
	// bitops       	various bit operations from bithack, namely: reverse bits, parity check, bit count, round to nearest power of 2
	// callfunc     	recursively call 8 argument C function to a depth of 1024 calls and unwind
	// cfloat       	1000 iterations of a mix of floating point complex operations
	// cdouble      	1000 iterations of a mix of double floating point complex operations
	// clongdouble  	1000 iterations of a mix of long double floating point complex operations
	// correlate    	perform a 16384 × 1024 correlation of random doubles
	// crc16        	compute 1024 rounds of CCITT CRC16 on random data
	// decimal32    	1000 iterations of a mix of 32 bit decimal floating point operations (GCC only)
	//
	// decimal64    	1000 iterations of a mix of 64 bit decimal floating point operations (GCC only)
	// decimal128   	1000 iterations of a mix of 128 bit decimal floating point operations (GCC only)
	// dither       	Floyd–Steinberg dithering of a 1024 × 768 random image from 8 bits down to 1 bit of depth.
	// djb2a        	128 rounds of hash DJB2a (Dan Bernstein hash using the xor variant) on 128 to 1 bytes of random strings
	// double       	1000 iterations of a mix of double precision floating point operations
	// euler        	compute e using n = (1 + (1 ÷ n)) ↑ n
	// explog       	iterate on n = exp(log(n) ÷ 1.00002)
	// factorial    	find factorials from 1..150 using Stirling's and Ramanujan's approximations
	// fibonacci    	compute Fibonacci sequence of 0, 1, 1, 2, 5, 8...
	// fft          	4096 sample Fast Fourier Transform
	// float        	1000 iterations of a mix of floating point operations
	// float16      	1000 iterations of a mix of 16 bit floating point operations
	// float32      	1000 iterations of a mix of 32 bit floating point operations
	// float80      	1000 iterations of a mix of 80 bit floating point operations
	// float128     	1000 iterations of a mix of 128 bit floating point operations
	// fnv1a        	128 rounds of hash FNV-1a (Fowler–Noll–Vo hash using the xor then multiply variant) on 128 to 1 bytes of random strings
	// gamma        	calculate  the Euler-Mascheroni constant γ using the limiting difference between the harmonic series (1 + 1/2 + 1/3 + 1/4 + 1/5 ... + 1/n) and the natural
	//              	logarithm ln(n), for n = 80000.
	// gcd              compute GCD of integers
	// gray             calculate binary to gray code and gray code back to binary for integers from 0 to 65535
	// hamming          compute Hamming H(8,4) codes on 262144 lots of 4 bit data. This turns 4 bit data into 8 bit Hamming code containing 4 parity bits. For data  bits  d1..d4,
	//                  parity bits are computed as:
	//                    p1 = d2 + d3 + d4
	//                    p2 = d1 + d3 + d4
	//                    p3 = d1 + d2 + d4
	//                    p4 = d1 + d2 + d3
	// hanoi            solve a 21 disc Towers of Hanoi stack using the recursive solution
	// hyperbolic       compute sinh(θ) × cosh(θ) + sinh(2θ) + cosh(3θ) for float, double and long double hyperbolic sine and cosine functions where θ = 0 to 2π in 1500 steps
	// idct             8 × 8 IDCT (Inverse Discrete Cosine Transform)
	// int8             1000 iterations of a mix of 8 bit integer operations
	// int16            1000 iterations of a mix of 16 bit integer operations
	// int32            1000 iterations of a mix of 32 bit integer operations
	// int64            1000 iterations of a mix of 64 bit integer operations
	// int128           1000 iterations of a mix of 128 bit integer operations (GCC only)
	// int32float       1000 iterations of a mix of 32 bit integer and floating point operations
	// int32double      1000 iterations of a mix of 32 bit integer and double precision floating point operations
	// int32longdouble  1000 iterations of a mix of 32 bit integer and long double precision floating point operations
	// int64float       1000 iterations of a mix of 64 bit integer and floating point operations
	// int64double      1000 iterations of a mix of 64 bit integer and double precision floating point operations
	// int64longdouble  1000 iterations of a mix of 64 bit integer and long double precision floating point operations
	// int128float      1000 iterations of a mix of 128 bit integer and floating point operations (GCC only)
	// int128double     1000 iterations of a mix of 128 bit integer and double precision floating point operations (GCC only)
	// int128longdouble 1000 iterations of a mix of 128 bit integer and long double precision floating point operations (GCC only)
	// int128decimal32  1000 iterations of a mix of 128 bit integer and 32 bit decimal floating point operations (GCC only)
	// int128decimal64  1000 iterations of a mix of 128 bit integer and 64 bit decimal floating point operations (GCC only)
	// int128decimal128 1000 iterations of a mix of 128 bit integer and 128 bit decimal floating point operations (GCC only)
	// jenkin           Jenkin's integer hash on 128 rounds of 128..1 bytes of random data
	// jmp              Simple unoptimised compare >, <, == and jmp branching
	// ln2              compute ln(2) based on series:
	//                    1 - 1/2 + 1/3 - 1/4 + 1/5 - 1/6 ...
	// longdouble       1000 iterations of a mix of long double precision floating point operations
	// loop             simple empty loop
	// matrixprod       matrix  product  of  two  128  × 128 matrices of double floats. Testing on 64 bit x86 hardware shows that this is provides a good mix of memory, cache and
	//                  floating point operations and is probably the best CPU method to use to make a CPU run hot.
	// nsqrt            compute sqrt() of long doubles using Newton-Raphson
	// omega            compute the omega constant defined by Ωe↑Ω = 1 using efficient iteration of Ωn+1 = (1 + Ωn) / (1 + e↑Ωn)
	// parity           compute parity using various methods from the Standford Bit Twiddling Hacks.  Methods employed are: the naïve way, the naïve way with the  Brian  Kernigan
	//                  bit counting optimisation, the multiply way, the parallel way, and the lookup table ways (2 variations).
	// phi              compute the Golden Ratio ϕ using series
	// pi               compute π using the Srinivasa Ramanujan fast convergence algorithm
	// pjw              128 rounds of hash pjw function on 128 to 1 bytes of random strings
	// prime            find all the primes in the range  1..1000000 using a slightly optimised brute force naïve trial division search
	// psi              compute ψ (the reciprocal Fibonacci constant) using the sum of the reciprocals of the Fibonacci numbers
	// queens           compute all the solutions of the classic 8 queens problem for board sizes 1..12
	//
	// rand             16384  iterations  of  rand(),  where rand is the MWC pseudo random number generator.  The MWC random function concatenates two 16 bit multiply-with-carry
	//                  generators:
	//                   x(n) = 36969 × x(n - 1) + carry,
	//                   y(n) = 18000 × y(n - 1) + carry mod 2 ↑ 16
	//
	//                  and has period of around 2 ↑ 60
	// rand48           16384 iterations of drand48(3) and lrand48(3)
	// rgb              convert RGB to YUV and back to RGB (CCIR 601)
	// sdbm             128 rounds of hash sdbm (as used in the SDBM database and GNU awk) on 128 to 1 bytes of random strings
	// sieve            find the primes in the range 1..10000000 using the sieve of Eratosthenes
	// stats            calculate minimum, maximum, arithmetic mean, geometric mean, harmoninc mean and standard deviation on 250 randomly  generated  positive  double  precision
	//                  value.
	// sqrt             compute sqrt(rand()), where rand is the MWC pseudo random number generator
	// trig             compute sin(θ) × cos(θ) + sin(2θ) + cos(3θ) for float, double and long double sine and cosine functions where θ = 0 to 2π in 1500 steps
	// union            perform  integer  arithmetic  on  a  mix of bit fields in a C union.  This exercises how well the compiler and CPU can perform integer bit field loads and
	//                  stores.
	// zeta             compute the Riemann Zeta function ζ(s) for s = 2.0..10.0
	// +kubebuilder:default=all
	// +optional
	Method CpuMethod `json:"method,omitempty"`
}

// +kubebuilder:object:root=true

// StressChaosList contains a list of StressChaos
type StressChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StressChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StressChaos{}, &StressChaosList{})
}
