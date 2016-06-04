package validation

import (
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/error"
	"reflect"
	"strconv"
	"strings"
)

var (
	UnsupportedType   = "type not supported for validation"
	Failed            = "validation failed"
	UnknownIdentifier = "validation identifier unknown"
	Invalid           = "invalid validation"
	InvalidSyntax     = "invalid validation syntax"
	InvalidValue      = "invalid validation value"
)

func Validate(v interface{}, s string) error {
	debug.Ver("Validate: %v", v)
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Bool:
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Float32:
	case reflect.Float64:
	case reflect.String:
		debug.Ver("Validate: %v against %s", v, s)
		for _, vs := range strings.Split(s, ",") {
			svs := vs
			// check if we have a negation prefix
			negate := strings.HasPrefix(vs, "!")
			// if so, we have to remove it before we check the validation identifier
			if negate == true {
				svs = svs[1:]
			}
			// some validators need a value, our seperatro is "="
			// val will contain the value a string and vs will
			// contain the validation identifier
			var val string
			_val := strings.Split(vs, "=")
			l := len(_val)
			if l > 2 {
				// too many vals
				return err.New(InvalidSyntax, vs)
			} else if l == 2 {
				svs = _val[0]
				val = _val[1]
			} else {
				// check the identifier if we need a value
				switch svs {
				case "max":
					return err.New(InvalidSyntax, vs, "needs a value")
				}
			}

			var res bool

			// validate
			switch svs {

			case "empty":
				// empty
				if t.Kind() != reflect.String {
					return err.New(Invalid, vs, "for", t.Kind().String())
				}
				res = (v == "")
				break

			case "max":
				// max value or len
				var maxi int64
				var maxu uint64
				var maxf float64
				// try to parse any number
				maxi, ei := strconv.ParseInt(val, 10, 64)
				maxu, eu := strconv.ParseUint(val, 10, 64)
				maxf, ef := strconv.ParseFloat(val, 64)
				// check if parsing was ok
				switch t.Kind() {
				case reflect.Int:
				case reflect.Int8:
				case reflect.Int16:
				case reflect.Int32:
				case reflect.Int64:
					if ei != nil {
						return err.New(InvalidValue, val, "for", vs)
					}
					break
				case reflect.Uint:
				case reflect.Uint8:
				case reflect.Uint16:
				case reflect.Uint32:
				case reflect.Uint64:
				case reflect.String:
					if eu != nil {
						return err.New(InvalidValue, val, "for", vs)
					}
					break
				case reflect.Float32:
				case reflect.Float64:
					if ef != nil {
						return err.New(InvalidValue, val, "for", vs)
					}
					break
				}
				// parsing ok, check value
				switch t.Kind() {
				case reflect.Bool:
					return err.New(Invalid, vs, "for", t.Kind().String())
				case reflect.Int:
					res = (v.(int) <= int(maxi))
					break
				case reflect.Int8:
					res = (v.(int8) <= int8(maxi))
					break
				case reflect.Int16:
					res = (v.(int16) <= int16(maxi))
					break
				case reflect.Int32:
					res = (v.(int32) <= int32(maxi))
					break
				case reflect.Int64:
					res = (v.(int64) <= maxi)
					break
				case reflect.Uint:
					res = (v.(uint) <= uint(maxu))
					break
				case reflect.Uint8:
					res = (v.(uint8) <= uint8(maxu))
					break
				case reflect.Uint16:
					res = (v.(uint16) <= uint16(maxu))
					break
				case reflect.Uint32:
					res = (v.(uint32) <= uint32(maxu))
					break
				case reflect.Uint64:
					res = (v.(uint64) <= uint64(maxu))
					break
				case reflect.Float32:
					res = (v.(float32) <= float32(maxf))
					break
				case reflect.Float64:
					res = (v.(float64) <= maxf)
					break
				case reflect.String:
					res = (len(v.(string)) <= int(maxu))
					break
				}
			default:
				return err.New(UnknownIdentifier, vs)
			}
			// negate if necessary
			res = (res != negate)
			if res != true {
				return err.New(Failed, vs, "for", v.(string))
			}
		}
		return nil
		break
	case reflect.Array:
	case reflect.Interface:
	case reflect.Map:
	case reflect.Slice:
	case reflect.Struct:
		vo := reflect.ValueOf(v)
		values := make([]interface{}, vo.NumField())
		// for all fields
		for i := 0; i < len(values); i++ {
			// get the validation tag
			vs := t.Field(i).Tag.Get("validation")
			if vs != "" {
				if err := Validate(vo.Field(i).Interface(), vs); err != nil {
					return err
				}
			}
		}
		break
	default:
		return err.New(UnsupportedType, t.Name())

	}
	return nil
}
