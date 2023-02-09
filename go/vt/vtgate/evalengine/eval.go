/*
Copyright 2023 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package evalengine

import (
	"fmt"
	"strconv"

	"vitess.io/vitess/go/mysql/collations"
	"vitess.io/vitess/go/sqltypes"
	vtrpcpb "vitess.io/vitess/go/vt/proto/vtrpc"
	"vitess.io/vitess/go/vt/vterrors"
	"vitess.io/vitess/go/vt/vtgate/evalengine/internal/decimal"
)

type typeFlag uint32

const (
	// flagNull marks that this value is null; implies flagNullable
	flagNull typeFlag = 1 << 0
	// flagNullable marks that this value CAN be null
	flagNullable typeFlag = 1 << 1

	// flagIntegerUdf marks that this value is math.MinInt64, and will underflow if negated
	flagIntegerUdf typeFlag = 1 << 5
	// flagIntegerCap marks that this value is (-math.MinInt64),
	// and should be promoted to flagIntegerUdf if negated
	flagIntegerCap typeFlag = 1 << 6
	// flagIntegerOvf marks that this value will overflow if negated
	flagIntegerOvf typeFlag = 1 << 7

	// flagHex marks that this value originated from a hex literal
	flagHex typeFlag = 1 << 8
	// flagBit marks that this value originated from a bit literal
	flagBit typeFlag = 1 << 9
	// flagExplicitCollation marks that this value has an explicit collation
	flagExplicitCollation typeFlag = 1 << 10

	// flagIntegerRange are the flags that mark overflow/underflow in integers
	flagIntegerRange = flagIntegerOvf | flagIntegerCap | flagIntegerUdf
)

type eval interface {
	toRawBytes() []byte
	sqlType() sqltypes.Type
	hash() (HashCode, error)
}

func evalToSQLValue(e eval) sqltypes.Value {
	if e == nil {
		return sqltypes.NULL
	}
	return sqltypes.MakeTrusted(e.sqlType(), e.toRawBytes())
}

func evalToSQLValueWithType(e eval, resultType sqltypes.Type) sqltypes.Value {
	switch {
	case sqltypes.IsSigned(resultType):
		switch e := e.(type) {
		case *evalInt64:
			return sqltypes.MakeTrusted(resultType, strconv.AppendInt(nil, e.i, 10))
		case *evalUint64:
			return sqltypes.MakeTrusted(resultType, strconv.AppendUint(nil, e.u, 10))
		case *evalFloat:
			return sqltypes.MakeTrusted(resultType, strconv.AppendInt(nil, int64(e.f), 10))
		}
	case sqltypes.IsUnsigned(resultType):
		switch e := e.(type) {
		case *evalInt64:
			return sqltypes.MakeTrusted(resultType, strconv.AppendUint(nil, uint64(e.i), 10))
		case *evalUint64:
			return sqltypes.MakeTrusted(resultType, strconv.AppendUint(nil, e.u, 10))
		case *evalFloat:
			return sqltypes.MakeTrusted(resultType, strconv.AppendUint(nil, uint64(e.f), 10))
		}
	case sqltypes.IsFloat(resultType) || resultType == sqltypes.Decimal:
		switch e := e.(type) {
		case *evalInt64:
			return sqltypes.MakeTrusted(resultType, strconv.AppendInt(nil, e.i, 10))
		case *evalUint64:
			return sqltypes.MakeTrusted(resultType, strconv.AppendUint(nil, e.u, 10))
		case *evalFloat:
			return sqltypes.MakeTrusted(resultType, FormatFloat(resultType, e.f))
		case *evalDecimal:
			return sqltypes.MakeTrusted(resultType, e.dec.FormatMySQL(e.length))
		}
	default:
		return sqltypes.MakeTrusted(resultType, e.toRawBytes())
	}
	return sqltypes.NULL
}

func evalIsTruthy(e eval) boolean {
	if e == nil {
		return boolNULL
	}
	switch e := e.(type) {
	case *evalInt64:
		return makeboolean(e.i != 0)
	case *evalUint64:
		return makeboolean(e.u != 0)
	case *evalFloat:
		return makeboolean(e.f != 0.0)
	case *evalDecimal:
		return makeboolean(!e.dec.IsZero())
	case *evalBytes:
		return makeboolean(parseStringToFloat(e.string()) != 0.0)
	default:
		panic("unhandled case: evalIsTruthy")
	}
}

func evalCoerce(e eval, typ sqltypes.Type, col collations.ID) (eval, error) {
	if e == nil {
		return nil, nil
	}
	if col == collations.Unknown {
		panic("EvalResult.coerce with no collation")
	}
	if typ == sqltypes.VarChar || typ == sqltypes.Char {
		// if we have an explicit VARCHAR coercion, always force it so the collation is replaced in the target
		return evalToText(e, col, false)
	}
	if e.sqlType() == typ {
		// nothing to be done here
		return e, nil
	}
	switch typ {
	case sqltypes.Null:
		return nil, nil
	case sqltypes.Binary, sqltypes.VarBinary:
		return evalToBinary(e), nil
	case sqltypes.Char, sqltypes.VarChar:
		panic("unreacheable")
	case sqltypes.Decimal:
		return evalToNumeric(e).toDecimal(0, 0), nil
	case sqltypes.Float32, sqltypes.Float64:
		f, _ := evalToNumeric(e).toFloat()
		return f, nil
	case sqltypes.Int8, sqltypes.Int16, sqltypes.Int32, sqltypes.Int64:
		return evalToNumeric(e).toInt64(), nil
	case sqltypes.Uint8, sqltypes.Uint16, sqltypes.Uint32, sqltypes.Uint64:
		return evalToNumeric(e).toUint64(), nil
	case sqltypes.Date, sqltypes.Datetime, sqltypes.Year, sqltypes.TypeJSON, sqltypes.Time, sqltypes.Bit:
		return nil, vterrors.Errorf(vtrpcpb.Code_UNIMPLEMENTED, "Unsupported type conversion: %s", typ.String())
	default:
		panic(fmt.Sprintf("BUG: emitted unknown type: %s", typ))
	}
}

func valueToEvalCast(v sqltypes.Value, typ sqltypes.Type) (eval, error) {
	switch {
	case typ == sqltypes.Null:
		return nil, nil

	case sqltypes.IsFloat(typ):
		switch {
		case v.IsSigned():
			ival, err := v.ToInt64()
			return newEvalFloat(float64(ival)), err
		case v.IsUnsigned():
			uval, err := v.ToUint64()
			return newEvalFloat(float64(uval)), err
		case v.IsFloat() || v.Type() == sqltypes.Decimal:
			fval, err := v.ToFloat64()
			return newEvalFloat(fval), err
		case v.IsText() || v.IsBinary():
			return newEvalFloat(parseStringToFloat(v.RawStr())), nil
		default:
			return nil, vterrors.Errorf(vtrpcpb.Code_INTERNAL, "coercion should not try to coerce this value to a float: %v", v)
		}

	case typ == sqltypes.Decimal:
		var dec decimal.Decimal
		switch {
		case v.IsIntegral() || v.Type() == sqltypes.Decimal:
			var err error
			dec, err = decimal.NewFromMySQL(v.Raw())
			if err != nil {
				return nil, err
			}
		case v.IsFloat():
			fval, err := v.ToFloat64()
			if err != nil {
				return nil, err
			}
			dec = decimal.NewFromFloat(fval)
		case v.IsText() || v.IsBinary():
			fval := parseStringToFloat(v.RawStr())
			dec = decimal.NewFromFloat(fval)
		default:
			return nil, vterrors.Errorf(vtrpcpb.Code_INTERNAL, "coercion should not try to coerce this value to a decimal: %v", v)
		}
		return &evalDecimal{dec: dec, length: -dec.Exponent()}, nil

	case sqltypes.IsSigned(typ):
		switch {
		case v.IsSigned():
			ival, err := v.ToInt64()
			return newEvalInt64(ival), err
		case v.IsUnsigned():
			uval, err := v.ToUint64()
			return newEvalInt64(int64(uval)), err
		default:
			return nil, vterrors.Errorf(vtrpcpb.Code_INTERNAL, "coercion should not try to coerce this value to a signed int: %v", v)
		}

	case sqltypes.IsUnsigned(typ):
		switch {
		case v.IsSigned():
			ival, err := v.ToInt64()
			return newEvalUint64(uint64(ival)), err
		case v.IsUnsigned():
			uval, err := v.ToUint64()
			return newEvalUint64(uval), err
		default:
			return nil, vterrors.Errorf(vtrpcpb.Code_INTERNAL, "coercion should not try to coerce this value to a unsigned int: %v", v)
		}

	case sqltypes.IsText(typ) || sqltypes.IsBinary(typ):
		switch {
		case v.IsText() || v.IsBinary():
			// TODO: collation
			return newEvalRaw(v.Type(), v.Raw(), collationBinary), nil
		default:
			return nil, vterrors.Errorf(vtrpcpb.Code_INTERNAL, "coercion should not try to coerce this value to a text: %v", v)
		}
	}
	return nil, vterrors.Errorf(vtrpcpb.Code_INTERNAL, "coercion should not try to coerce this value: %v", v)
}

func valueToEvalNumeric(v sqltypes.Value) (eval, error) {
	switch {
	case v.IsSigned():
		ival, err := v.ToInt64()
		if err != nil {
			return nil, vterrors.Errorf(vtrpcpb.Code_INVALID_ARGUMENT, "%v", err)
		}
		return &evalInt64{ival}, nil
	case v.IsUnsigned():
		var uval uint64
		uval, err := v.ToUint64()
		if err != nil {
			return nil, vterrors.Errorf(vtrpcpb.Code_INVALID_ARGUMENT, "%v", err)
		}
		return newEvalUint64(uval), nil
	default:
		uval, err := strconv.ParseUint(v.RawStr(), 10, 64)
		if err == nil {
			return newEvalUint64(uval), nil
		}
		ival, err := strconv.ParseInt(v.RawStr(), 10, 64)
		if err == nil {
			return &evalInt64{ival}, nil
		}
		return nil, vterrors.Errorf(vtrpcpb.Code_INVALID_ARGUMENT, "could not parse value: '%s'", v.RawStr())
	}
}

func valueToEval(value sqltypes.Value, collation collations.TypedCollation) (eval, error) {
	wrap := func(err error) error {
		if err == nil {
			return nil
		}
		return vterrors.Errorf(vtrpcpb.Code_INVALID_ARGUMENT, "%v", err)
	}

	switch tt := value.Type(); {
	case sqltypes.IsSigned(tt):
		ival, err := value.ToInt64()
		return newEvalInt64(ival), wrap(err)
	case sqltypes.IsUnsigned(tt):
		uval, err := value.ToUint64()
		return newEvalUint64(uval), wrap(err)
	case sqltypes.IsFloat(tt):
		fval, err := value.ToFloat64()
		return newEvalFloat(fval), wrap(err)
	case tt == sqltypes.Decimal:
		dec, err := decimal.NewFromMySQL(value.Raw())
		return newEvalDecimal(dec, 0, 0), wrap(err)
	case sqltypes.IsText(tt):
		if tt == sqltypes.HexNum {
			raw, err := parseHexNumber(value.Raw())
			return newEvalBytesHex(raw), wrap(err)
		} else if tt == sqltypes.HexVal {
			hex := value.Raw()
			raw, err := parseHexLiteral(hex[2 : len(hex)-1])
			return newEvalBytesHex(raw), wrap(err)
		} else {
			return newEvalText(value.Raw(), collation), nil
		}
	case sqltypes.IsBinary(tt):
		return newEvalBinary(value.Raw()), nil
	case sqltypes.IsDate(tt):
		return newEvalRaw(value.Type(), value.Raw(), collationNumeric), nil
	case sqltypes.IsNull(tt):
		return nil, nil
	default:
		return nil, vterrors.Errorf(vtrpcpb.Code_INTERNAL, "Type is not supported: %q %s", value, value.Type())
	}
}
