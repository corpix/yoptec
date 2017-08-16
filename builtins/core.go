// Package core implements core interface for anko script.
package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/covrom/gonec/ast"
	"github.com/covrom/gonec/parser"
	"github.com/covrom/gonec/vm"

	gonec_encoding_json "github.com/covrom/gonec/builtins/encoding/json"
	gonec_errors "github.com/covrom/gonec/builtins/errors"
	gonec_flag "github.com/covrom/gonec/builtins/flag"
	gonec_fmt "github.com/covrom/gonec/builtins/fmt"
	gonec_io "github.com/covrom/gonec/builtins/io"
	gonec_io_ioutil "github.com/covrom/gonec/builtins/io/ioutil"
	gonec_math "github.com/covrom/gonec/builtins/math"
	gonec_math_big "github.com/covrom/gonec/builtins/math/big"
	gonec_math_rand "github.com/covrom/gonec/builtins/math/rand"
	gonec_net "github.com/covrom/gonec/builtins/net"
	gonec_net_http "github.com/covrom/gonec/builtins/net/http"
	gonec_net_url "github.com/covrom/gonec/builtins/net/url"
	gonec_os "github.com/covrom/gonec/builtins/os"
	gonec_os_exec "github.com/covrom/gonec/builtins/os/exec"
	gonec_os_signal "github.com/covrom/gonec/builtins/os/signal"
	gonec_path "github.com/covrom/gonec/builtins/path"
	gonec_path_filepath "github.com/covrom/gonec/builtins/path/filepath"
	gonec_regexp "github.com/covrom/gonec/builtins/regexp"
	gonec_runtime "github.com/covrom/gonec/builtins/runtime"
	gonec_sort "github.com/covrom/gonec/builtins/sort"
	gonec_strings "github.com/covrom/gonec/builtins/strings"
	gonec_time "github.com/covrom/gonec/builtins/time"

	gonec_colortext "github.com/covrom/gonec/builtins/github.com/daviddengcn/go-colortext"
)

// LoadAllBuiltins is a convenience function that loads all defineSd builtins.
func LoadAllBuiltins(env *vm.Env) {
	Import(env)

	pkgs := map[string]func(env *vm.Env) *vm.Env{
		"encoding/json": gonec_encoding_json.Import,
		"errors":        gonec_errors.Import,
		"flag":          gonec_flag.Import,
		"fmt":           gonec_fmt.Import,
		"io":            gonec_io.Import,
		"io/ioutil":     gonec_io_ioutil.Import,
		"math":          gonec_math.Import,
		"math/big":      gonec_math_big.Import,
		"math/rand":     gonec_math_rand.Import,
		"net":           gonec_net.Import,
		"net/http":      gonec_net_http.Import,
		"net/url":       gonec_net_url.Import,
		"os":            gonec_os.Import,
		"os/exec":       gonec_os_exec.Import,
		"os/signal":     gonec_os_signal.Import,
		"path":          gonec_path.Import,
		"path/filepath": gonec_path_filepath.Import,
		"regexp":        gonec_regexp.Import,
		"runtime":       gonec_runtime.Import,
		"sort":          gonec_sort.Import,
		"strings":       gonec_strings.Import,
		"time":          gonec_time.Import,
		"github.com/daviddengcn/go-colortext": gonec_colortext.Import,
	}

	env.DefineS("импорт", func(s string) interface{} {
		if loader, ok := pkgs[strings.ToLower(s)]; ok {
			m := loader(env)
			return m
		}
		panic(fmt.Sprintf("Пакет '%s' не найден", s))
	})

}

// Import defineSs core language builtins - len, range, println, int64, etc.
func Import(env *vm.Env) *vm.Env {
	env.DefineS("длина", func(v interface{}) int64 {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Interface {
			rv = rv.Elem()
		}
		if rv.Kind() == reflect.String {
			return int64(len([]byte(rv.String())))
		}
		if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
			panic("Аргумент должен быть строкой или массивом")
		}
		return int64(rv.Len())
	})

	env.DefineS("ключи", func(v interface{}) []string {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Interface {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Map {
			panic("Аргумент должен быть структурой")
		}
		keys := []string{}
		mk := rv.MapKeys()
		for _, key := range mk {
			keys = append(keys, key.String())
		}
		// ключи потом обходим в порядке сортировки по алфавиту
		sort.Strings(keys)
		return keys
	})

	env.DefineS("диапазон", func(args ...int64) []int64 {
		if len(args) < 1 {
			panic("Отсутствуют аргументы")
		}
		if len(args) > 2 {
			panic("Должен быть только один аргумент")
		}
		var min, max int64
		if len(args) == 1 {
			min = 0
			max = args[0] - 1
		} else {
			min = args[0]
			max = args[1]
		}
		arr := []int64{}
		for i := min; i <= max; i++ {
			arr = append(arr, i)
		}
		return arr
	})

	env.DefineS("дата", func(v interface{}) time.Time {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.String {
			tt, err := time.Parse(time.RFC3339, v.(string))
			if err == nil {
				return tt
			} else {
				panic(err)
			}
		}
		panic("Дата может быть представлена только строкой в формате RFC3339")
	})

	env.DefineS("нрег", func(v interface{}) string {
		if b, ok := v.([]byte); ok {
			return strings.ToLower(string(b))
		}
		return strings.ToLower(fmt.Sprint(v))
	})

	env.DefineS("врег", func(v interface{}) string {
		if b, ok := v.([]byte); ok {
			return strings.ToUpper(string(b))
		}
		return strings.ToUpper(fmt.Sprint(v))
	})

	env.DefineS("формат", env.Sprintf)
	
	env.DefineS("кодсимвола", func(s string) rune {
		if len(s) == 0 {
			return 0
		}
		return []rune(s)[0]
	})

	// TODO: !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	env.DefineS("вбайтслайс", func(s string) []byte {
		return []byte(s)
	})

	env.DefineS("вслайсрун", func(s string) []rune {
		return []rune(s)
	})

	env.DefineS("вбулевслайс", func(v []interface{}) []bool {
		var result []bool
		toSlice(v, &result)
		return result
	})

	env.DefineS("вслайсчисел", func(v []interface{}) []float64 {
		var result []float64
		toSlice(v, &result)
		return result
	})

	env.DefineS("вслайсцелыхчисел", func(v []interface{}) []int64 {
		var result []int64
		toSlice(v, &result)
		return result
	})

	env.DefineS("вслайсстрок", func(v []interface{}) []string {
		var result []string
		toSlice(v, &result)
		return result
	})

	env.DefineS("вдлительность", func(v int64) time.Duration {
		return time.Duration(v)
	})

	//////////////////////////////////////////////////////////

	

	env.DefineS("типзнч", func(v interface{}) string {
		return ast.UniqueNames.Get(env.TypeName(reflect.TypeOf(v)))
	})

	env.DefineS("присвоенозначение", func(s string) bool {
		_, err := env.Get(ast.UniqueNames.Set(s))
		return err == nil
	})

	env.DefineS("загрузитьивыполнить", func(s string) interface{} {
		body, err := ioutil.ReadFile(s)
		if err != nil {
			panic(err)
		}
		scanner := new(parser.Scanner)
		scanner.Init(string(body))
		stmts, err := parser.Parse(scanner)
		if err != nil {
			if pe, ok := err.(*parser.Error); ok {
				pe.Filename = s
				panic(pe)
			}
			panic(err)
		}
		rv, err := vm.Run(stmts, env)
		if err != nil {
			panic(err)
		}
		if rv.IsValid() && rv.CanInterface() {
			return rv.Interface()
		}
		return nil
	})

	env.DefineS("паника", func(e interface{}) {
		os.Setenv("GONEC_DEBUG", "1")
		panic(e)
	})

	env.DefineS("вывести", env.Print)
	env.DefineS("сообщить", env.Println)
	env.DefineS("сообщитьф", env.Printf)
	env.DefineS("stdout", env.StdOut())
	env.DefineS("закрыть", func(e interface{}) {
		reflect.ValueOf(e).Close()
	})

	env.DefineTypeS("целоечисло", int64(0))
	env.DefineTypeS("число", float64(0.0))
	env.DefineTypeS("булево", true)
	env.DefineTypeS("строка", "")
	env.DefineTypeS("массив", []interface{}{})
	env.DefineTypeS("структура", map[string]interface{}{})
	return env
}

// toSlice takes in a "generic" slice and converts and copies
// it's elements into the typed slice pointed at by ptr.
// Note that this is a costly operation.
func toSlice(from []interface{}, ptr interface{}) {
	// Value of the pointer to the target
	obj := reflect.Indirect(reflect.ValueOf(ptr))
	// We can't just convert from interface{} to whatever the target is (diff memory layout),
	// so we need to create a New slice of the proper type and copy the values individually
	t := reflect.TypeOf(ptr).Elem()
	slice := reflect.MakeSlice(t, len(from), len(from))
	// Copying the data, val is an adressable Pointer of the actual target type
	val := reflect.Indirect(reflect.New(t.Elem()))
	for i := 0; i < len(from); i++ {
		v := reflect.ValueOf(from[i])
		val.Set(v)
		slice.Index(i).Set(v)
	}
	// Ok now assign our slice to the target pointer
	obj.Set(slice)
}
