package main

/*
#cgo CFLAGS: -I/usr/include/flint
#cgo LDFLAGS: -L/usr/local/lib -lflint -lmpfr -lgmp
#include <arb.h>
#include <acb.h>
#include <stdlib.h>

const arb_struct* my_acb_realref(const acb_t z) {
    return &(z[0].real);
}

const arb_struct* my_acb_imagref(const acb_t z) {
    return &(z[0].imag);
}

int acb_set_from_str(acb_t z, const char* real, const char* imag, slong prec) {
    arb_t r, i;
    arb_init(r);
    arb_init(i);

    int ok = 1;
    if (arb_set_str(r, real, prec)) {
        printf("arb_set_str(real=%s, %ld) FAILED\n", real, prec);
        ok = 0;
    }

    if (arb_set_str(i, imag, prec)) {
        printf("arb_set_str(imag=%s, %ld) FAILED\n", imag, prec);
        ok = 0;
    }

    if (ok) {
        acb_set_arb_arb(z, r, i);
    }

    arb_clear(r);
    arb_clear(i);
    return ok;
}
*/
import "C"

import (
	"flag"
	"fmt"
	"math"
	"math/big"
	"strings"
	"unsafe"
)

type ArbComplex struct {
	z    C.acb_t
	prec C.slong
}

func NewArbComplex(realStr, imagStr string, prec uint) (*ArbComplex, error) {
	ac := &ArbComplex{prec: C.slong(prec)}
	C.acb_init(&ac.z[0])

	r := C.CString(realStr)
	i := C.CString(imagStr)
	defer C.free(unsafe.Pointer(r))
	defer C.free(unsafe.Pointer(i))

	ok := C.acb_set_from_str(&ac.z[0], r, i, ac.prec)
	if ok == 0 {
		err := fmt.Errorf("failed to parse input strings: (real=%s, imag=%s, prec=%d)", realStr, imagStr, prec)
		return nil, err
	}
	return ac, nil
}

func NewArbComplexFromFloat64(real, imag float64, prec uint) *ArbComplex {
	ac := &ArbComplex{prec: C.slong(prec)}
	C.acb_init(&ac.z[0])
	C.acb_set_d_d(&ac.z[0], C.double(real), C.double(imag))
	return ac
}

func NewArbComplexFromBigFloat(real, imag *big.Float, prec uint) (*ArbComplex, error) {
	ac := &ArbComplex{prec: C.slong(prec)}
	C.acb_init(&ac.z[0])

	realStr := C.CString(real.Text('e', -1))
	imagStr := C.CString(imag.Text('e', -1))
	defer C.free(unsafe.Pointer(realStr))
	defer C.free(unsafe.Pointer(imagStr))

	ok := C.acb_set_from_str(&ac.z[0], realStr, imagStr, ac.prec)
	if ok == 0 {
		err := fmt.Errorf("failed to parse big.Float values as arb strings: (real=%s, imag=%s, prec=%d)", real.Text('e', -1), imag.Text('e', -1), prec)
		return nil, err
	}

	return ac, nil
}

func trimArbStr(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}
	if s[0] == '[' && s[len(s)-1] == ']' {
		s = s[1 : len(s)-1]
	}
	if idx := strings.Index(s, " +/-"); idx != -1 {
		return s[:idx]
	}
	return s
}

func (ac *ArbComplex) String(digits int, rng bool) string {
	realPtr := C.my_acb_realref(&ac.z[0])
	imagPtr := C.my_acb_imagref(&ac.z[0])

	realStr := C.arb_get_str(realPtr, C.slong(digits), 0)
	imagStr := C.arb_get_str(imagPtr, C.slong(digits), 0)

	defer C.free(unsafe.Pointer(realStr))
	defer C.free(unsafe.Pointer(imagStr))

	if rng {
		return fmt.Sprintf("(%s,%si)", C.GoString(realStr), C.GoString(imagStr))
	}
	realGo := trimArbStr(C.GoString(realStr))
	imagGo := trimArbStr(C.GoString(imagStr))

	return fmt.Sprintf("(%s,%si)", realGo, imagGo)
}

func (ac *ArbComplex) BigFloats() (*big.Float, *big.Float, error) {
	realPtr := C.my_acb_realref(&ac.z[0])
	imagPtr := C.my_acb_imagref(&ac.z[0])

	realStr := C.arb_get_str(realPtr, ac.prec, 0)
	imagStr := C.arb_get_str(imagPtr, ac.prec, 0)
	defer C.free(unsafe.Pointer(realStr))
	defer C.free(unsafe.Pointer(imagStr))

	realGo := trimArbStr(C.GoString(realStr))
	imagGo := trimArbStr(C.GoString(imagStr))

	r, _, err := big.ParseFloat(realGo, 10, uint(ac.prec), big.ToNearestEven)
	if err != nil {
		return nil, nil, err
	}
	i, _, err := big.ParseFloat(imagGo, 10, uint(ac.prec), big.ToNearestEven)
	if err != nil {
		return nil, nil, err
	}

	return r, i, nil
}

func (ac *ArbComplex) Clear() {
	C.acb_clear(&ac.z[0])
}

func (ac *ArbComplex) Add(other *ArbComplex) *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_add(&result.z[0], &ac.z[0], &other.z[0], ac.prec)
	return result
}

func (ac *ArbComplex) Sub(other *ArbComplex) *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_sub(&result.z[0], &ac.z[0], &other.z[0], ac.prec)
	return result
}

func (ac *ArbComplex) Mul(other *ArbComplex) *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_mul(&result.z[0], &ac.z[0], &other.z[0], ac.prec)
	return result
}

func (ac *ArbComplex) Div(other *ArbComplex) *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_div(&result.z[0], &ac.z[0], &other.z[0], ac.prec)
	return result
}

func (ac *ArbComplex) Exp() *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_exp(&result.z[0], &ac.z[0], ac.prec)
	return result
}

func (ac *ArbComplex) Ln() *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_log(&result.z[0], &ac.z[0], ac.prec)
	return result
}

func (ac *ArbComplex) Pow(exponent *ArbComplex) *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_pow(&result.z[0], &ac.z[0], &exponent.z[0], ac.prec)
	return result
}

func (ac *ArbComplex) Log(base *ArbComplex) *ArbComplex {
	lnZ := ac.Ln()
	lnBase := base.Ln()
	result := lnZ.Div(lnBase)

	lnZ.Clear()
	lnBase.Clear()
	return result
}

func (ac *ArbComplex) Sqrt() *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_sqrt(&result.z[0], &ac.z[0], ac.prec)
	return result
}

func (ac *ArbComplex) Root(n int) *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_root_ui(&result.z[0], &ac.z[0], C.ulong(n), ac.prec)
	return result
}

func (ac *ArbComplex) Sin() *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_sin(&result.z[0], &ac.z[0], ac.prec)
	return result
}

func (ac *ArbComplex) Cos() *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_cos(&result.z[0], &ac.z[0], ac.prec)
	return result
}

func (ac *ArbComplex) Tan() *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])
	C.acb_tan(&result.z[0], &ac.z[0], ac.prec)
	return result
}

func (ac *ArbComplex) Ctan() *ArbComplex {
	result := ac.Tan()
	// one, _ := NewArbComplex("1.0", "0.0", uint(ac.prec))
	one := NewArbComplexFromFloat64(1.0, 0.0, uint(ac.prec))
	defer one.Clear()

	cot := one.Div(result)
	result.Clear()
	return cot
}

func (ac *ArbComplex) Abs() *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])

	tmp := C.malloc(C.size_t(C.sizeof_arb_struct)) // allocate arb_t manually
	real := (*C.arb_t)(tmp)
	C.arb_init(&(*real)[0])

	C.acb_abs(&(*real)[0], &ac.z[0], ac.prec)
	C.acb_set_arb(&result.z[0], &(*real)[0])

	C.arb_clear(&(*real)[0])
	C.free(tmp)

	return result
}

func (ac *ArbComplex) Arg() *ArbComplex {
	result := &ArbComplex{prec: ac.prec}
	C.acb_init(&result.z[0])

	tmp := C.malloc(C.size_t(C.sizeof_arb_struct)) // allocate arb_t manually
	arg := (*C.arb_t)(tmp)
	C.arb_init(&(*arg)[0])

	C.acb_arg(&(*arg)[0], &ac.z[0], ac.prec)
	C.acb_set_arb(&result.z[0], &(*arg)[0])

	C.arb_clear(&(*arg)[0])
	C.free(tmp)

	return result
}

func parseComplexArg(arg string, precision uint) (*ArbComplex, error) {
	parts := strings.Split(arg, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid complex input (expecting real,imag): %q", arg)
	}
	return NewArbComplex(parts[0], parts[1], precision)
}

func BigFloatToInt(f *big.Float) (int, error) {
	i := new(big.Int)
	f.Int(i) // Truncates toward zero

	if !i.IsInt64() {
		return 0, fmt.Errorf("value out of int64 range")
	}

	int64Val := i.Int64()

	// Optional: Handle platforms where int != int64 (e.g., 32-bit)
	if intSize := 32 << (^uint(0) >> 63); intSize == 32 && (int64Val > int64(int(^uint32(0)>>1)) || int64Val < int64(^int32(0))) {
		return 0, fmt.Errorf("value out of int range on 32-bit system")
	}

	return int(int64Val), nil
}

func main() {
	var (
		precision uint
		rng       bool
		operation string
		argA      string
		argB      string
	)

	flag.UintVar(&precision, "prec", 128, "precision in bits")
	flag.BoolVar(&rng, "range", false, "display ranges instead of center values")
	flag.StringVar(&operation, "op", "add", "operation: add, sub, mul, div, exp, ln, pow, log, sqrt, root, sin, cos, tan, ctan, abs, arg")
	flag.StringVar(&argA, "a", "1,0", "first complex number (format: real,imag)")
	flag.StringVar(&argB, "b", "1,0", "second complex number (format: real,imag)")

	flag.Parse()
	displayDigits := int(float64(precision) * math.Log10(2))
	fmt.Printf("Using precision = %d bits (~%d decimal digits)\n", precision, displayDigits)

	a, err := parseComplexArg(argA, precision)
	if err != nil {
		panic("invalid -a input: " + err.Error())
	}
	defer a.Clear()

	b, err := parseComplexArg(argB, precision)
	if err != nil {
		panic("invalid -b input: " + err.Error())
	}
	defer b.Clear()

	var result *ArbComplex
	switch operation {
	case "add":
		result = a.Add(b)
	case "sub":
		result = a.Sub(b)
	case "mul":
		result = a.Mul(b)
	case "div":
		result = a.Div(b)
	case "exp":
		result = a.Exp()
	case "ln":
		result = a.Ln()
	case "pow":
		result = a.Pow(b)
	case "log":
		result = b.Log(a)
	case "sqrt":
		result = a.Sqrt()
	case "root":
		r, _, err := a.BigFloats()
		if err != nil {
			panic(err)
		}
		n, err := BigFloatToInt(r)
		if err != nil {
			panic(err)
		}
		result = b.Root(n)
	case "sin":
		result = a.Sin()
	case "cos":
		result = a.Cos()
	case "tan":
		result = a.Tan()
	case "ctan":
		result = a.Ctan()
	case "abs":
		result = a.Abs()
	case "arg":
		result = a.Arg()
	default:
		panic("unsupported operation: " + operation)
	}
	defer result.Clear()

	fmt.Println("a:", a.String(displayDigits, rng))
	fmt.Println("b:", b.String(displayDigits, rng))
	fmt.Println(operation+":", result.String(displayDigits, rng))
	resR, resI, err := result.BigFloats()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s:(%s,%si)\n", operation, resR.Text('f', displayDigits), resI.Text('f', displayDigits))
}
