// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gurvy DO NOT EDIT

package bn256

import (
	"fmt"
	"math/big"
	"runtime"
	"testing"

	"github.com/consensys/gurvy/bn256/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// utils
func fuzzJacobianG2(p *G2Jac, f *E2) G2Jac {
	var res G2Jac
	res.X.Mul(&p.X, f).Mul(&res.X, f)
	res.Y.Mul(&p.Y, f).Mul(&res.Y, f).Mul(&res.Y, f)
	res.Z.Mul(&p.Z, f)
	return res
}

func fuzzProjectiveG2(p *G2Proj, f *E2) G2Proj {
	var res G2Proj
	res.X.Mul(&p.X, f)
	res.Y.Mul(&p.Y, f)
	res.Z.Mul(&p.Z, f)
	return res
}

func fuzzExtendedJacobianG2(p *g2JacExtended, f *E2) g2JacExtended {
	var res g2JacExtended
	var ff, fff E2
	ff.Square(f)
	fff.Mul(&ff, f)
	res.X.Mul(&p.X, &ff)
	res.Y.Mul(&p.Y, &fff)
	res.ZZ.Mul(&p.ZZ, &ff)
	res.ZZZ.Mul(&p.ZZZ, &fff)
	return res
}

// ------------------------------------------------------------
// tests

func TestG2Conversions(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)
	genFuzz1 := GenE2()
	genFuzz2 := GenE2()

	properties.Property("Affine representation should be independent of the Jacobian representative", prop.ForAll(
		func(a *E2) bool {
			g := fuzzJacobianG2(&g2Gen, a)
			var op1 G2Affine
			op1.FromJacobian(&g)
			return op1.X.Equal(&g2Gen.X) && op1.Y.Equal(&g2Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("Affine representation should be independent of a Extended Jacobian representative", prop.ForAll(
		func(a *E2) bool {
			var g g2JacExtended
			g.X.Set(&g2Gen.X)
			g.Y.Set(&g2Gen.Y)
			g.ZZ.Set(&g2Gen.Z)
			g.ZZZ.Set(&g2Gen.Z)
			gfuzz := fuzzExtendedJacobianG2(&g, a)

			var op1 G2Affine
			gfuzz.ToAffine(&op1)
			return op1.X.Equal(&g2Gen.X) && op1.Y.Equal(&g2Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("Projective representation should be independent of a Jacobian representative", prop.ForAll(
		func(a *E2) bool {

			g := fuzzJacobianG2(&g2Gen, a)

			var op1 G2Proj
			op1.FromJacobian(&g)
			var u, v E2
			u.Mul(&g.X, &g.Z)
			v.Square(&g.Z).Mul(&v, &g.Z)

			return op1.X.Equal(&u) && op1.Y.Equal(&g.Y) && op1.Z.Equal(&v)
		},
		genFuzz1,
	))

	properties.Property("Jacobian representation should be the same as the affine representative", prop.ForAll(
		func(a *E2) bool {
			var g G2Jac
			var op1 G2Affine
			op1.X.Set(&g2Gen.X)
			op1.Y.Set(&g2Gen.Y)

			var one E2
			one.SetOne()

			g.FromAffine(&op1)

			return g.X.Equal(&g2Gen.X) && g.Y.Equal(&g2Gen.Y) && g.Z.Equal(&one)
		},
		genFuzz1,
	))

	properties.Property("Converting affine symbol for infinity to Jacobian should output correct infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G2Affine
			g.X.SetZero()
			g.Y.SetZero()
			var op1 G2Jac
			op1.FromAffine(&g)
			var one, zero E2
			one.SetOne()
			return op1.X.Equal(&one) && op1.Y.Equal(&one) && op1.Z.Equal(&zero)
		},
	))

	properties.Property("Converting infinity in extended Jacobian to affine should output infinity symbol in Affine", prop.ForAll(
		func() bool {
			var g G2Affine
			var op1 g2JacExtended
			var zero E2
			op1.X.Set(&g2Gen.X)
			op1.Y.Set(&g2Gen.Y)
			op1.ToAffine(&g)
			return g.X.Equal(&zero) && g.Y.Equal(&zero)
		},
	))

	properties.Property("Converting infinity in extended Jacobian to Jacobian should output infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G2Jac
			var op1 g2JacExtended
			var zero, one E2
			one.SetOne()
			op1.X.Set(&g2Gen.X)
			op1.Y.Set(&g2Gen.Y)
			op1.ToJac(&g)
			return g.X.Equal(&one) && g.Y.Equal(&one) && g.Z.Equal(&zero)
		},
	))

	properties.Property("[Jacobian] Two representatives of the same class should be equal", prop.ForAll(
		func(a, b *E2) bool {
			op1 := fuzzJacobianG2(&g2Gen, a)
			op2 := fuzzJacobianG2(&g2Gen, b)
			return op1.Equal(&op2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG2Ops(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)
	genFuzz1 := GenE2()
	genFuzz2 := GenE2()

	genScalar := GenFr()

	properties.Property("[Jacobian] Add should call double when having adding the same point", prop.ForAll(
		func(a, b *E2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			fop2 := fuzzJacobianG2(&g2Gen, b)
			var op1, op2 G2Jac
			op1.Set(&fop1).AddAssign(&fop2)
			op2.Double(&fop2)
			return op1.Equal(&op2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[Jacobian] Adding the opposite of a point to itself should output inf", prop.ForAll(
		func(a, b *E2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			fop2 := fuzzJacobianG2(&g2Gen, b)
			fop2.Neg(&fop2)
			fop1.AddAssign(&fop2)
			return fop1.Equal(&g2Infinity)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[Jacobian] Adding the inf to a point should not modify the point", prop.ForAll(
		func(a *E2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			fop1.AddAssign(&g2Infinity)
			var op2 G2Jac
			op2.Set(&g2Infinity)
			op2.AddAssign(&g2Gen)
			return fop1.Equal(&g2Gen) && op2.Equal(&g2Gen)
		},
		genFuzz1,
	))

	properties.Property("[Jacobian Extended] mAdd (-G) should equal mSub(G)", prop.ForAll(
		func(a *E2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			var p1, p1Neg G2Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 g2JacExtended
			o1.mAdd(&p1Neg)
			o2.mSub(&p1)

			return o1.X.Equal(&o2.X) &&
				o1.Y.Equal(&o2.Y) &&
				o1.ZZ.Equal(&o2.ZZ) &&
				o1.ZZZ.Equal(&o2.ZZZ)
		},
		genFuzz1,
	))

	properties.Property("[Jacobian Extended] double (-G) should equal doubleNeg(G)", prop.ForAll(
		func(a *E2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			var p1, p1Neg G2Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 g2JacExtended
			o1.double(&p1Neg)
			o2.doubleNeg(&p1)

			return o1.X.Equal(&o2.X) &&
				o1.Y.Equal(&o2.Y) &&
				o1.ZZ.Equal(&o2.ZZ) &&
				o1.ZZZ.Equal(&o2.ZZZ)
		},
		genFuzz1,
	))

	properties.Property("[Jacobian] Addmix the negation to itself should output 0", prop.ForAll(
		func(a *E2) bool {
			fop1 := fuzzJacobianG2(&g2Gen, a)
			fop1.Neg(&fop1)
			var op2 G2Affine
			op2.FromJacobian(&g2Gen)
			fop1.AddMixed(&op2)
			return fop1.Equal(&g2Infinity)
		},
		genFuzz1,
	))

	properties.Property("scalar multiplication (double and add) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g G2Jac
			var gaff G2Affine
			gaff.FromJacobian(&g2Gen)
			g.ScalarMultiplication(&gaff, r)

			var scalar, blindedScalard, rminusone big.Int
			var op1, op2, op3, gneg G2Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.ScalarMultiplication(&gaff, &rminusone)
			gneg.Neg(&g2Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			op1.ScalarMultiplication(&gaff, &scalar)
			op2.ScalarMultiplication(&gaff, &blindedScalard)

			return op1.Equal(&op2) && g.Equal(&g2Infinity) && !op1.Equal(&g2Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))

	properties.Property("scalar multiplication (GLV) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g G2Jac
			var gaff G2Affine
			gaff.FromJacobian(&g2Gen)
			g.ScalarMulGLV(&gaff, r)

			var scalar, blindedScalard, rminusone big.Int
			var op1, op2, op3, gneg G2Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.ScalarMulGLV(&gaff, &rminusone)
			gneg.Neg(&g2Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			op1.ScalarMulGLV(&gaff, &scalar)
			op2.ScalarMulGLV(&gaff, &blindedScalard)

			return op1.Equal(&op2) && g.Equal(&g2Infinity) && !op1.Equal(&g2Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))

	properties.Property("GLV and Double and Add should output the same result", prop.ForAll(
		func(s fr.Element) bool {

			var r big.Int
			var op1, op2 G2Jac
			var gaff G2Affine
			s.ToBigIntRegular(&r)
			gaff.FromJacobian(&g2Gen)
			op1.ScalarMultiplication(&gaff, &r)
			op2.ScalarMulGLV(&gaff, &r)
			return op1.Equal(&op2) && !op1.Equal(&g2Infinity)

		},
		genScalar,
	))

	// note : this test is here as we expect to have a different multiExp than the above bucket method
	// for small number of points
	properties.Property("Multi exponentation (<50points) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var g G2Jac
			g.Set(&g2Gen)

			// mixer ensures that all the words of a fpElement are set
			samplePoints := make([]G2Affine, 30)
			sampleScalars := make([]fr.Element, 30)

			for i := 1; i <= 30; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
				samplePoints[i-1].FromJacobian(&g)
				g.AddAssign(&g2Gen)
			}

			var op1MultiExp G2Jac
			op1MultiExp.MultiExp(samplePoints, sampleScalars)

			var finalBigScalar fr.Element
			var finalBigScalarBi big.Int
			var op1ScalarMul G2Jac
			var op1Aff G2Affine
			op1Aff.FromJacobian(&g2Gen)
			finalBigScalar.SetString("9455").MulAssign(&mixer)
			finalBigScalar.ToBigIntRegular(&finalBigScalarBi)
			op1ScalarMul.ScalarMultiplication(&op1Aff, &finalBigScalarBi)

			return op1ScalarMul.Equal(&op1MultiExp)
		},
		genScalar,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG2MultiExp(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)

	genScalar := GenFr()

	// size of the multiExps
	const nbSamples = 500

	// multi exp points
	var samplePoints [nbSamples]G2Affine
	var g G2Jac
	g.Set(&g2Gen)
	for i := 1; i <= nbSamples; i++ {
		samplePoints[i-1].FromJacobian(&g)
		g.AddAssign(&g2Gen)
	}

	// final scalar to use in double and add method (without mixer factor)
	// n(n+1)(2n+1)/6  (sum of the squares from 1 to n)
	var scalar big.Int
	scalar.SetInt64(nbSamples)
	scalar.Mul(&scalar, new(big.Int).SetInt64(nbSamples+1))
	scalar.Mul(&scalar, new(big.Int).SetInt64(2*nbSamples+1))
	scalar.Div(&scalar, new(big.Int).SetInt64(6))

	properties.Property("Multi exponentation (c=4) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc4(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=5) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc5(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=6) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc6(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=7) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc7(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=8) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc8(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=9) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc9(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=10) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc10(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=11) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc11(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=12) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc12(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=13) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc13(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=14) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc14(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=15) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc15(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=16) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc16(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=17) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc17(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=18) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc18(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=19) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G2Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			numCpus := runtime.NumCPU()
			chCpus := make(chan struct{}, numCpus)
			for i := 0; i < numCpus; i++ {
				chCpus <- struct{}{}
			}

			result.multiExpc19(samplePoints[:], sampleScalars[:], chCpus)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g2GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkG2ScalarMul(b *testing.B) {

	var scalar big.Int
	scalar.SetString("5243587517512619047944770508185965837690552500527637822603658699938581184513", 10)

	var doubleAndAdd G2Jac

	b.Run("double and add", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			doubleAndAdd.ScalarMultiplication(&g2GenAff, &scalar)
		}
	})

	var glv G2Jac
	b.Run("GLV", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			glv.ScalarMulGLV(&g2GenAff, &scalar)
		}
	})

}

func BenchmarkG2Add(b *testing.B) {
	var a G2Jac
	a.Double(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddAssign(&g2Gen)
	}
}

func BenchmarkG2mAdd(b *testing.B) {
	var a g2JacExtended
	a.double(&g2GenAff)

	var c G2Affine
	c.FromJacobian(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.mAdd(&c)
	}

}

func BenchmarkG2AddMixed(b *testing.B) {
	var a G2Jac
	a.Double(&g2Gen)

	var c G2Affine
	c.FromJacobian(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddMixed(&c)
	}

}

func BenchmarkG2Double(b *testing.B) {
	var a G2Jac
	a.Set(&g2Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.DoubleAssign()
	}

}

func BenchmarkG2MultiExpG2(b *testing.B) {
	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	const pow = 24
	const nbSamples = 1 << pow

	var samplePoints [nbSamples]G2Affine
	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
		samplePoints[i-1] = g2GenAff
	}

	var testPoint G2Jac

	for i := 5; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				testPoint.MultiExp(samplePoints[:using], sampleScalars[:using])
			}
		})
	}
}
