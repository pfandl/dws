package validation

import (
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/error"
	"math"
	"reflect"
	"regexp"
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
	DataUnvailable    = "data unavailable"
)

func Validate(v interface{}, ig string, s string) error {
	debug.Ver("Validate: \"%v\", ignore %s, validate %s", v, ig, s)
	to := reflect.TypeOf(v)
	switch to.Kind() {
	case reflect.Bool:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.String:
		debug.Ver("Validate: \"%v\" against %s", v, s)
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
				if to.Kind() != reflect.String {
					return err.New(Invalid, vs, "for", to.Kind().String())
				}
				res = (v == "")

			case "ipv4":
				// ipv4 address
				if to.Kind() != reflect.String {
					return err.New(Invalid, vs, "for", to.Kind().String())
				}
				r := regexp.MustCompile("^(\\d{1,3}\\.){3}\\d{1,3}$")
				res = r.MatchString(v.(string))

			case "ipv4mac":
				// mac address
				if to.Kind() != reflect.String {
					return err.New(Invalid, vs, "for", to.Kind().String())
				}
				r := regexp.MustCompile("^([a-fA-F0-9]{2}:){5}[a-fA-F0-9]{2}$")
				res = r.MatchString(v.(string))

			case "port":
				// port
				if to.Kind() != reflect.String {
					return err.New(Invalid, vs, "for", to.Kind().String())
				}
				if r, e := strconv.ParseUint(v.(string), 10, 64); e == nil {
					res = r < math.MaxUint16
				} else {
					return err.New(Invalid, vs, "for", to.Kind().String())
				}

			case "uts":
				// uts name
				if to.Kind() != reflect.String {
					return err.New(Invalid, vs, "for", to.Kind().String())
				}
				r := regexp.MustCompile("^(([a-zA-Z0-9\\-_])+\\.)*([a-zA-Z0-9\\-_])+\\.([a-zA-Z])+$")
				res = r.MatchString(v.(string))

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
				switch to.Kind() {
				case reflect.Int:
					fallthrough
				case reflect.Int8:
					fallthrough
				case reflect.Int16:
					fallthrough
				case reflect.Int32:
					fallthrough
				case reflect.Int64:
					if ei != nil {
						return err.New(InvalidValue, val, "for", vs)
					}
				case reflect.Uint:
					fallthrough
				case reflect.Uint8:
					fallthrough
				case reflect.Uint16:
					fallthrough
				case reflect.Uint32:
					fallthrough
				case reflect.Uint64:
					fallthrough
				case reflect.String:
					if eu != nil {
						return err.New(InvalidValue, val, "for", vs)
					}
				case reflect.Float32:
					fallthrough
				case reflect.Float64:
					if ef != nil {
						return err.New(InvalidValue, val, "for", vs)
					}
				}
				// parsing ok, check value
				switch to.Kind() {
				case reflect.Bool:
					return err.New(Invalid, vs, "for", to.Kind().String())
				case reflect.Int:
					res = (v.(int) <= int(maxi))
				case reflect.Int8:
					res = (v.(int8) <= int8(maxi))
				case reflect.Int16:
					res = (v.(int16) <= int16(maxi))
				case reflect.Int32:
					res = (v.(int32) <= int32(maxi))
				case reflect.Int64:
					res = (v.(int64) <= maxi)
				case reflect.Uint:
					res = (v.(uint) <= uint(maxu))
				case reflect.Uint8:
					res = (v.(uint8) <= uint8(maxu))
				case reflect.Uint16:
					res = (v.(uint16) <= uint16(maxu))
				case reflect.Uint32:
					res = (v.(uint32) <= uint32(maxu))
				case reflect.Uint64:
					res = (v.(uint64) <= uint64(maxu))
				case reflect.Float32:
					res = (v.(float32) <= float32(maxf))
				case reflect.Float64:
					res = (v.(float64) <= maxf)
				case reflect.String:
					res = (len(v.(string)) <= int(maxu))
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
	case reflect.Slice:
		fallthrough
	case reflect.Struct:
		vo := reflect.ValueOf(v)

		var l int
		if to.Kind() == reflect.Slice {
			l = vo.Len()
		} else {
			l = vo.NumField()
		}
		// for all fields or slices
		vs := ""
		is := ""
		var igs []string
		n := ""

		var in interface{}
		for i := 0; i < l; i++ {
			if to.Kind() == reflect.Struct {
				// get the validation tag
				f := to.Field(i)
				debug.Ver("%d %v", i, f)
				debug.Ver("%v", reflect.Zero(to.Field(i).Type) == vo.Field(i))
				n = f.Name
				vs = f.Tag.Get("validation")
				if vs == "" {
					continue
				}
				// get the validation ignore tag
				is = f.Tag.Get("validation-ignore")
				// joing parent and our ignore tags
				igs = strings.Split(ig, ",")
				for _, igns := range strings.Split(is, ",") {
					igs = append(igs, igns)
				}

				b := false
				// check whether we should ignore some validations
				for _, ign := range igs {
					if ign == "" {
						continue
					}
					// just get first value of splitting by ",""
					// (for getting correct name of "XMLNAME,attr" and such)
					if strings.Split(f.Tag.Get("xml"), ",")[0] == ign {
						b = true
						break
					}
				}
				if b == true {
					// ignore this validation
					debug.Ver("Validate: %s for %s ignored", vs, n)
					continue
				}
			}

			// joing parent and our validation ignore tags
			is = strings.Join(igs, ",")

			debug.Info("Validate: validating %s for %s (%s)", vs, n, to.String())
			if to.Kind() == reflect.Slice {
				in = vo.Index(i).Interface()
			} else {
				in = vo.Field(i).Interface()
			}
			if err := Validate(in, is, vs); err != nil {
				return err
			}
		}
	default:
		return err.New(UnsupportedType, to.Kind().String())
	}
	return nil
}
