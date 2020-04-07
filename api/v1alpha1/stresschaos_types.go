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

	// Stressors defines plenty of stressors supported to stress system components out.
	// You can use one or more of them to make up various kinds of stresses. At least
	// one of the stressors should be specified.
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

// Stressors defines plenty of stressors supported to stress system components out.
// You can use one or more of them to make up various kinds of stresses
type Stressors struct {
	// VMStressor stresses virtual memory out
	// +optional
	VmStressor *VMStressor `json:"vm,omitempty"`
	// CPUStressor stresses CPU out
	// +optional
	CPUStressor *CPUStressor `json:"cpu,omitempty"`
}

// Normalize the stressors to comply with stress-ng
func (in *Stressors) Normalize() string {
	stressors := ""
	if in.VmStressor != nil {
		stressors += fmt.Sprintf("--vm %d --vm-bytes %s",
			in.VmStressor.Workers, in.VmStressor.Bytes)
	}
	if in.CPUStressor != nil {
		stressors += fmt.Sprintf("--cpu %d --cpu-load %d --cpu-method %s",
			in.CPUStressor.Workers, in.CPUStressor.Load, in.CPUStressor.Method)
	}
	return stressors
}

// Stressor defines common configurations of a stressor
type Stressor struct {
	// Workers specifies N workers to apply the stressor.
	Workers int `json:"workers"`
}

// VMStressor defines how to stress memory out
type VMStressor struct {
	Stressor `json:",inline"`

	// Bytes specifies N bytes consumed per vm worker, default is the total available memory.
	// One can specify the size as % of total available memory or in units of B, KB/KiB,
	// MB/MiB, GB/GiB, TB/TiB.
	// +optional
	Bytes string `json:"bytes,omitempty"`
}

// CPUMethod defines the method to be executed by a CPUStressor
type CPUMethod string

const (
	// CPUMethodAll implies all of the CPUMethod should be applied
	CPUMethodAll CPUMethod = "all"
	// CPUMethodAckermann computes Ackermann function
	CPUMethodAckermann CPUMethod = "ackermann"
	// CPUMethodBitops  runs various bit operations from bithack
	CPUMethodBitops CPUMethod = "bitops"
	// CPUMethodCallfunc recursively calls 8 argument C function to a depth of 1024 calls and unwind
	CPUMethodCallfunc CPUMethod = "callfunc"
	// CPUMethodCfloat runs 1000 iterations of a mix of floating point complex operations
	CPUMethodCfloat CPUMethod = "cfloat"
	// CPUMethodCdouble runs 1000 iterations of a mix of double floating point complex operations
	CPUMethodCdouble CPUMethod = "cdouble"
	// CPUMethodClongdouble runs 1000 iterations of a mix of long double floating point complex operations
	CPUMethodClongdouble CPUMethod = "clongdouble"
	// CPUMethodCorrelate performs a 16384 × 1024 correlation of random doubles
	CPUMethodCorrelate CPUMethod = "correlate"
	// CPUMethodCrc16 computes 1024 rounds of CCITT CRC16 on random data
	CPUMethodCrc16 CPUMethod = "crc16"
	// CPUMethodDecimal32 runs 1000 iterations of a mix of 32 bit decimal floating point operations (GCC only)
	CPUMethodDecimal32 CPUMethod = "decimal32"
	// CPUMethodDecimal64 runs 1000 iterations of a mix of 64 bit decimal floating point operations (GCC only)
	CPUMethodDecimal64 CPUMethod = "decimal64"
	// CPUMethodDecimal12 runs 1000 iterations of a mix of 128 bit decimal floating point operations (GCC only)
	CPUMethodDecimal128 CPUMethod = "decimal128"
	// CPUMethodDither runs Floyd–Steinberg dithering of a 1024 × 768 random image from 8 bits down to 1 bit of depth.
	CPUMethodDither CPUMethod = "dither"
	// CPUMethodDjb2a runs 128 rounds of hash DJB2a (Dan Bernstein hash using the xor variant) on 128 to 1 bytes of random strings
	CPUMethodDjb2a CPUMethod = "djb2a"
	// CPUMethodDouble runs 1000 iterations of a mix of double precision floating point operations
	CPUMethodDouble CPUMethod = "double"
	// CPUMethodEuler computes e using n = (1 + (1 ÷ n)) ↑ n
	CPUMethodEuler CPUMethod = "euler"
	// CPUMethodExplog iterates on n = exp(log(n) ÷ 1.00002)
	CPUMethodExplog CPUMethod = "explog"
	// CPUMethodFactorial finds factorials from 1..150 using Stirling's and Ramanujan's approximations
	CPUMethodFactorial CPUMethod = "factorial"
	// CPUMethodFibonacci computes Fibonacci sequence of 0, 1, 1, 2, 5, 8...
	CPUMethodFibonacci CPUMethod = "fibonacci"
	// CPUMethodFft computes 4096 sample Fast Fourier Transform
	CPUMethodFft CPUMethod = "fft"
	// CPUMethodFloat runs 1000 iterations of a mix of floating point operations
	CPUMethodFloat CPUMethod = "float"
	// CPUMethodFloat16 runs 1000 iterations of a mix of 16 bit floating point operations
	CPUMethodFloat16 CPUMethod = "float16"
	// CPUMethodFloat3 runs 1000 iterations of a mix of 32 bit floating point operations
	CPUMethodFloat32 CPUMethod = "float32"
	// CPUMethodFloat80 runs 1000 iterations of a mix of 80 bit floating point operations
	CPUMethodFloat80 CPUMethod = "float80"
	// CPUMethodFloat128 runs 1000 iterations of a mix of 128 bit floating point operations
	CPUMethodFloat128 CPUMethod = "float128"
	// CPUMethodFnv1a runs 128 rounds of hash FNV-1a (Fowler–Noll–Vo hash using the xor then multiply variant) on 128 to 1 bytes of random strings
	CPUMethodFnv1a CPUMethod = "fnv1a"
	// CPUMethodGamma calculates the Euler-Mascheroni constant γ using the limiting difference between the harmonic series (1 + 1/2 + 1/3 + 1/4 + 1/5 ... + 1/n) and the natural
	CPUMethodGamma CPUMethod = "gamma"
	// CPUMethodGcd computes GCD of integers
	CPUMethodGcd CPUMethod = "gcd"
	// CPUMethodGray calculates binary to gray code and gray code back to binary for integers from 0 to 65535
	CPUMethodGray CPUMethod = "gray"
	// CPUMethodHamming computes Hamming H(8,4) codes on 262144 lots of 4 bit data. This turns 4 bit data into 8 bit Hamming code containing 4 parity bits. For data  bits  d1..d4,
	CPUMethodHamming CPUMethod = "hamming"
	// CPUMethodHanoi solves a 21 disc Towers of Hanoi stack using the recursive solution
	CPUMethodHanoi CPUMethod = "hanoi"
	// CPUMethodHyperboli computes sinh(θ) × cosh(θ) + sinh(2θ) + cosh(3θ) for float, double and long double hyperbolic sine and cosine functions where θ = 0 to 2π in 1500 steps
	CPUMethodHyperbolic CPUMethod = "hyperbolic"
	// CPUMethodIdct computes 8 × 8 IDCT (Inverse Discrete Cosine Transform)
	CPUMethodIdct CPUMethod = "idct"
	// CPUMethodInt8 runs 1000 iterations of a mix of 8 bit integer operations
	CPUMethodInt8 CPUMethod = "int8"
	// CPUMethodInt16 runs 1000 iterations of a mix of 16 bit integer operations
	CPUMethodInt16 CPUMethod = "int16"
	// CPUMethodInt32 runs 1000 iterations of a mix of 32 bit integer operations
	CPUMethodInt32 CPUMethod = "int32"
	// CPUMethodInt64 runs 1000 iterations of a mix of 64 bit integer operations
	CPUMethodInt64 CPUMethod = "int64"
	// CPUMethodInt128 runs 1000 iterations of a mix of 128 bit integer operations (GCC only)
	CPUMethodInt128 CPUMethod = "int128"
	// CPUMethodInt32floa runs 1000 iterations of a mix of 32 bit integer and floating point operations
	CPUMethodInt32float CPUMethod = "int32float"
	// CPUMethodInt32doub runs 1000 iterations of a mix of 32 bit integer and double precision floating point operations
	CPUMethodInt32double CPUMethod = "int32double"
	// CPUMethodInt32long runs 1000 iterations of a mix of 32 bit integer and long double precision floating point operations
	CPUMethodInt32longdouble CPUMethod = "int32longdouble"
	// CPUMethodInt64floa runs 1000 iterations of a mix of 64 bit integer and floating point operations
	CPUMethodInt64float CPUMethod = "int64float"
	// CPUMethodInt64doub runs 1000 iterations of a mix of 64 bit integer and double precision floating point operations
	CPUMethodInt64double CPUMethod = "int64double"
	// CPUMethodInt64long runs 1000 iterations of a mix of 64 bit integer and long double precision floating point operations
	CPUMethodInt64longdouble CPUMethod = "int64longdouble"
	// CPUMethodInt128flo runs 1000 iterations of a mix of 128 bit integer and floating point operations (GCC only)
	CPUMethodInt128float CPUMethod = "int128float"
	// CPUMethodInt128dou runs 1000 iterations of a mix of 128 bit integer and double precision floating point operations (GCC only)
	CPUMethodInt128double CPUMethod = "int128double"
	// CPUMethodInt128lon runs 1000 iterations of a mix of 128 bit integer and long double precision floating point operations (GCC only)
	CPUMethodInt128longdouble CPUMethod = "int128longdouble"
	// CPUMethodInt128dec runs 1000 iterations of a mix of 128 bit integer and 32 bit decimal floating point operations (GCC only)
	CPUMethodInt128decimal32 CPUMethod = "int128decimal32"
	// CPUMethodInt128dec runs 1000 iterations of a mix of 128 bit integer and 64 bit decimal floating point operations (GCC only)
	CPUMethodInt128decimal64 CPUMethod = "int128decimal64"
	// CPUMethodInt128dec runs 1000 iterations of a mix of 128 bit integer and 128 bit decimal floating point operations (GCC only)
	CPUMethodInt128decimal128 CPUMethod = "int128decimal128"
	// CPUMethodJenkin computes Jenkin's integer hash on 128 rounds of 128..1 bytes of random data
	CPUMethodJenkin CPUMethod = "jenkin"
	// CPUMethodJmp simulates unoptimised compare >, <, == and jmp branching
	CPUMethodJmp CPUMethod = "jmp"
	// CPUMethodLn2 compute ln(2) based on series:
	CPUMethodLn2 CPUMethod = "ln2"
	// CPUMethodLongdoubl runs 1000 iterations of a mix of long double precision floating point operations
	CPUMethodLongdouble CPUMethod = "longdouble"
	// CPUMethodLoop simulates simple empty loop
	CPUMethodLoop CPUMethod = "loop"
	// CPUMethodMatrixpro computes matrix product of two 128×128 matrices of double floats. Testing on 64 bit x86 hardware shows that this is provides a good mix of memory, cache and
	CPUMethodMatrixprod CPUMethod = "matrixprod"
	// CPUMethodNsqrt computes sqrt() of long doubles using Newton-Raphson
	CPUMethodNsqrt CPUMethod = "nsqrt"
	// CPUMethodOmega computes the omega constant defined by Ωe↑Ω = 1 using efficient iteration of Ωn+1 = (1 + Ωn) / (1 + e↑Ωn)
	CPUMethodOmega CPUMethod = "omega"
	// CPUMethodParity computes parity using various methods from the Standford Bit Twiddling Hacks
	CPUMethodParity CPUMethod = "parity"
	// CPUMethodPhi computes the Golden Ratio ϕ using series
	CPUMethodPhi CPUMethod = "phi"
	// CPUMethodPi computes π using the Srinivasa Ramanujan fast convergence algorithm
	CPUMethodPi CPUMethod = "pi"
	// CPUMethodPjw runs 128 rounds of hash pjw function on 128 to 1 bytes of random strings
	CPUMethodPjw CPUMethod = "pjw"
	// CPUMethodPrime finds all the primes in the range  1..1000000 using a slightly optimised brute force naïve trial division search
	CPUMethodPrime CPUMethod = "prime"
	// CPUMethodPsi computes ψ (the reciprocal Fibonacci constant) using the sum of the reciprocals of the Fibonacci numbers
	CPUMethodPsi CPUMethod = "psi"
	// CPUMethodQueens computes all the solutions of the classic 8 queens problem for board sizes 1..12
	CPUMethodQueens CPUMethod = "queens"
	// CPUMethodRand runs 16384  iterations  of  rand(),  where rand is the MWC pseudo random number generator.  The MWC random function concatenates two 16 bit multiply-with-carry
	CPUMethodRand CPUMethod = "rand"
	// CPUMethodRand48 runs 16384 iterations of drand48(3) and lrand48(3)
	CPUMethodRand48 CPUMethod = "rand48"
	// CPUMethodRgb converts RGB to YUV and back to RGB (CCIR 601)
	CPUMethodRgb CPUMethod = "rgb"
	// CPUMethodSdbm runs 128 rounds of hash sdbm (as used in the SDBM database and GNU awk) on 128 to 1 bytes of random strings
	CPUMethodSdbm CPUMethod = "sdbm"
	// CPUMethodSieve finds the primes in the range 1..10000000 using the sieve of Eratosthenes
	CPUMethodSieve CPUMethod = "sieve"
	// CPUMethodStats calculates minimum, maximum, arithmetic mean, geometric mean, harmoninc mean and standard deviation on 250 randomly  generated  positive  double  precision
	CPUMethodStats CPUMethod = "stats"
	// CPUMethodSqrt computes sqrt of long doubles using Newton-Raphson
	CPUMethodSqrt CPUMethod = "sqrt"
	// CPUMethodTrig computes sin(θ) × cos(θ) + sin(2θ) + cos(3θ) for float, double and long double sine and cosine functions where θ = 0 to 2π in 1500 steps
	CPUMethodTrig CPUMethod = "trig"
	// CPUMethodUnion performs  integer  arithmetic  on  a  mix of bit fields in a C union.  This exercises how well the compiler and CPU can perform integer bit field loads and
	CPUMethodUnion CPUMethod = "union"
	// CPUMethodZeta computes the Riemann Zeta function ζ(s) for s = 2.0..10.0
	CPUMethodZeta CPUMethod = "zeta"
)

// CPUStressor defines how to stress CPU out
type CPUStressor struct {
	Stressor `json:",inline"`
	// Load specifies P percent loading per CPU worker. 0 is effectively a sleep (no load) and 100
	// is full loading.
	// +optional
	Load *int `json:"load,omitempty"`
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
	// +optional
	Method CPUMethod `json:"method,omitempty"`
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
