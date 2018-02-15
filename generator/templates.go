package generator

var dataTpl = `
package {{.Package}}

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var (
	ErrFailedDecodingInput = errors.New("Failed decoding input")
)

{{range $input := .Structs}}
type {{$input.TitleCamel}} struct {
	{{range $field := $input.Fields}}{{$field.TitleCamel}} {{$field.DataTypeIn}} {{$.Backtick}}{{$field.Tag}}{{$.Backtick}}
	{{end}}
}
{{end}}

{{range $message := .Structs}}
func Get{{MinusP $message.TitleCamel}}(r *http.Request) ({{MinusP $message.Camel}} {{MinusP $message.TitleCamel}}, err error) {
	inputP := &{{$message.TitleCamel}}{}
	if err := Unmarshal(r, inputP); err != nil {
		log.Error("Failed decoding input:", err)
		return {{MinusP $message.Camel}}, ErrFailedDecodingInput
	}

	msg, err := Validate(inputP)
	if err != nil {
		return {{MinusP $message.Camel}}, err
	}

	if msg != "" {
		return {{MinusP $message.Camel}}, errors.New(msg)
	}	

	if err := Transform(inputP); err != nil {
		return {{MinusP $message.Camel}}, err
	}	

	{{MinusP $message.Camel}} = Convert{{$message.TitleCamel}}(inputP)

	return
}
{{end}}

{{range $input := .Structs}}
func Convert{{$input.TitleCamel}}({{$input.Camel}} *{{$input.TitleCamel}}) ({{MinusP $input.Camel}} {{MinusP $input.TitleCamel}}) {
	{{range $field := $input.Fields}}
	{{if $field.IsRepeatedStruct}}

	{{$field.Camel}} := {{MinusStar $field.DataType | MinusP}}{}
	for _, field := range {{$input.Camel}}.{{$field.TitleCamel}} {
		{{$field.Camel}} = append({{$field.Camel}}, Convert{{$field.DataTypeName.TitleCamel}}P(field))
	}
	{{MinusP $input.Camel}}.{{$field.TitleCamel}} = {{$field.Camel}}

	{{else if $field.IsStruct}}

	if {{$input.Camel}}.{{$field.TitleCamel}} != nil {
		{{MinusP $input.Camel}}.{{$field.TitleCamel}} = Convert{{$field.DataTypeName.TitleCamel}}P({{$input.Camel}}.{{$field.TitleCamel}})
	}

	{{else if $field.IsRepeatedBuiltin}}

	{{$field.Camel}} := {{MinusStar $field.DataType}}{}
	for _, field := range {{$input.Camel}}.{{$field.TitleCamel}} {
		{{$field.Camel}} = append({{$field.Camel}}, *field)
	}
	{{MinusP $input.Camel}}.{{$field.TitleCamel}} = {{$field.Camel}}

	{{else}}

	if {{$input.Camel}}.{{$field.TitleCamel}} != nil {
		{{MinusP $input.Camel}}.{{$field.TitleCamel}} = *{{$input.Camel}}.{{$field.TitleCamel}}
	}

	{{end}}{{end}}
	return 
}
{{end}}

`

var genniTpl = `
package {{.Package}}

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"syscall"
	
	"github.com/magical/argon2"
	"github.com/spf13/cast"
	log "github.com/sirupsen/logrus"
)

// From templates/helpers.helpers.go.tpl
const (
	TagNameValidate             = "validate"
	TagNameTransform            = "transform"
	TagNameJSON                 = "json"
	TransformStrEncrypt         = "encrypt"
	TransformStrDecrypt         = "decrypt"
	TransformStrHash            = "hash"
	TransformStrPasswordHash    = "passwordhash"
	TransformStrTruncate        = "truncate"
	TransformStrTrimChars       = "trimchars"
	TransformStrTrimSpace       = "trimspace"
	TransformStrDefault         = "default"
	ValidateStrMaxLength        = "maxlength"
	ValidateStrMinLength        = "minlength"
	ValidateStrGreaterThan      = "greaterthan"
	ValidateStrLessThan         = "lessthan"
	ValidateStrRequired         = "required"
	ValidateStrMustHaveChars    = "musthavechars"
	ValidateStrCantHaveChars    = "canthavechars"
	ValidateStrOnlyHaveChars    = "onlyhavechars"
	ValidateStrMaxLengthErr     = "Failed Max Length Validation"
	ValidateStrMinLengthErr     = "Failed Min Length Validation"
	ValidateStrRequiredErr      = "Failed Required Validation"
	ValidateStrMustHaveCharsErr = "Failed Must Have Chars Validation"
	ValidateStrCantHaveCharsErr = "Failed Can't Have Chars Validation"
	ValidateStrOnlyHaveCharsErr = "Failed Only Have Chars Validation"
	ValidateStrGreaterThanErr   = "Failed Greater Than Validation"
	ValidateStrLessThanErr      = "Failed Less Than Validation"
)

var (
	dummyString   string
	dummyInt      int
	dummyInt64    int64
	dummyFloat32  float32
	dummyFloat64  float64
	dummyBool     bool
	dummyStringP  *string
	dummyIntP     *int
	dummyInt64P   *int64
	dummyFloat32P *float32
	dummyFloat64P *float64
	dummyBoolP    *bool

	TypeOfString   = reflect.TypeOf(dummyString)
	TypeOfInt      = reflect.TypeOf(dummyInt)
	TypeOfInt64    = reflect.TypeOf(dummyInt64)
	TypeOfFloat32  = reflect.TypeOf(dummyFloat32)
	TypeOfFloat64  = reflect.TypeOf(dummyFloat64)
	TypeOfBool     = reflect.TypeOf(dummyBool)
	TypeOfStringP  = reflect.TypeOf(dummyStringP)
	TypeOfIntP     = reflect.TypeOf(dummyIntP)
	TypeOfInt64P   = reflect.TypeOf(dummyInt64P)
	TypeOfFloat32P = reflect.TypeOf(dummyFloat32P)
	TypeOfFloat64P = reflect.TypeOf(dummyFloat64P)
	TypeOfBoolP    = reflect.TypeOf(dummyBoolP)

	builtinTypes = []reflect.Type{
		TypeOfString,
		TypeOfInt,
		TypeOfInt64,
		TypeOfFloat32,
		TypeOfFloat64,
		TypeOfBool,
		TypeOfStringP,
		TypeOfIntP,
		TypeOfInt64P,
		TypeOfFloat32P,
		TypeOfFloat64P,
		TypeOfBoolP,
	}
)

func getRandomSalt() (salt []byte, err error) {
	salt = make([]byte, 32)
	_, err = rand.Read(salt)
	return
}

func getTagKV(param string) (k, v string) {
	paramArr := strings.Split(param, "=")

	k = paramArr[0]
	if len(paramArr) == 2 {
		v = paramArr[1]
	}
	k = strings.ToLower(k)
	k = strings.Replace(k, "-", "", -1)
	k = strings.Replace(k, "_", "", -1)
	k = strings.Replace(k, " ", "", -1)
	return
}

func allCharsInStr(allChars, in string) (out bool) {
	for _, char := range allChars {
		if strings.Index(in, string(char)) == -1 {
			return
		}
	}
	return true
}

func onlyCharsInStr(onlyChars, in string) (out bool) {
	for _, char := range onlyChars {
		in = strings.Replace(in, string(char), "", -1)
	}
	return len(in) == 0
}

func dereferenceStringArray(in []*string) (out []string) {
	for _, inP := range in {
		out = append(out, *inP)
	}
	return
}

func dereferenceIntArray(in []*int) (out []int) {
	for _, inP := range in {
		out = append(out, *inP)
	}
	return
}

func dereferenceInt32Array(in []*int32) (out []int32) {
	for _, inP := range in {
		out = append(out, *inP)
	}
	return
}

func dereferenceInt64Array(in []*int64) (out []int64) {
	for _, inP := range in {
		out = append(out, *inP)
	}
	return
}

func dereferenceFloat32Array(in []*float32) (out []float32) {
	for _, inP := range in {
		out = append(out, *inP)
	}
	return
}

func dereferenceFloat64Array(in []*float64) (out []float64) {
	for _, inP := range in {
		out = append(out, *inP)
	}
	return
}

func dereferenceBoolArray(in []*bool) (out []bool) {
	for _, inP := range in {
		out = append(out, *inP)
	}
	return
}

func isBuiltin(fieldType reflect.Type) bool {
	for _, builtinType := range builtinTypes {
		if fieldType == builtinType {
			return true
		}
	}
	
	if strings.Contains(fieldType.String(), "map[") {
		return true
	}

	ft := strings.Replace(fieldType.String(), "[]*", "", -1)
	for _, builtinType := range builtinTypes {
		if ft == builtinType.String() {
			return true
		}
	}

	return false
}

// Courtesy of https://stackoverflow.com/questions/13901819/quick-way-to-detect-empty-values-via-reflection-in-go
func IsZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func ExecDirectory(commandStr string, directory string, envVars ...string) (err error) {
	if err := os.Chdir(directory); err != nil {
		log.Errorf("Failed to cd %s: %s", directory, err)
		return err
	}

	return Exec(commandStr, envVars...)
}

func Exec(commandStr string, envVars ...string) (err error) {
	if strings.TrimSpace(commandStr) == "" {
		return errors.New("No command provided")
	}

	var name string
	var args []string

	cmdArr := strings.Split(commandStr, " ")
	name = cmdArr[0]

	if len(cmdArr) > 1 {
		args = cmdArr[1:]
	}

	command := exec.Command(name, args...)
	command.Env = append(os.Environ(), envVars...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		log.Error("Failed creating command stdoutpipe: ", err)
		return err
	}
	defer stdout.Close()
	stdoutReader := bufio.NewReader(stdout)

	stderr, err := command.StderrPipe()
	if err != nil {
		log.Error("Failed creating command stderrpipe: ", err)
		return err
	}
	defer stderr.Close()
	stderrReader := bufio.NewReader(stderr)

	if err := command.Start(); err != nil {
		log.Error("Failed starting command: ", err)
		return err
	}

	go handleReader(stdoutReader, false)
	go handleReader(stderrReader, true)

	if err := command.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				log.Debug("Exit Status: ", status.ExitStatus())
				return err
			}
		}
		log.Debug("Failed to wait for command: ", err)
		return err
	}

	return
}

func handleReader(reader *bufio.Reader, isStderr bool) {
	printOutput := log.GetLevel() == log.DebugLevel
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if printOutput {
			fmt.Print(str)
		}
	}
}

// From templates/helpers.handler.go.tpl
func WriteJSON(out interface{}, w http.ResponseWriter) {
	jsonBytes, err := json.Marshal(out)
	if err != nil {
			log.Error("Failed marshalling to JSON:", err)
			http.Error(w, ErrorJSON("JSON Marshal Error"), http.StatusInternalServerError)
			return
	}

	if _, err := w.Write(jsonBytes); err != nil {
			log.Error("Failed writing to response writer:", err)
			http.Error(w, ErrorJSON("Failed writing to output"), http.StatusInternalServerError)
			return
	}
}

func ErrorJSON(msg string) (out string) {
	return "{\"error\":\"" + msg + "\"}"
}

// From templates/helpers.transform.go.tpl
func Transform(in interface{}) (err error) {
	t := reflect.TypeOf(in).Elem()
	v := reflect.ValueOf(in).Elem()

	for i := 0; i < t.NumField(); i++ {
		if !isBuiltin(t.Field(i).Type) {
			if IsZeroOfUnderlyingType(v.Field(i).Interface()) {
				continue
			}

			if err := Transform(v.Field(i).Interface()); err != nil {
				log.Debug("Failed field transform:", err)
				return err
			}
		}

		tag := t.Field(i).Tag.Get(TagNameTransform)

		if tag == "" || tag == "-" || tag == "_" || tag == " " {
			continue
		}

		params := strings.Split(tag, ",")
		for _, param := range params {
			log.Debugf("Transforming: %s - %s", v.Type().Field(i).Name, param)

			key, val := getTagKV(param)
			if v.Field(i).Pointer() == 0 && key == TransformStrDefault {
				if err := SetDefaultValue(v.Field(i), val); err != nil {
					return err
				}
				continue
			}

			if v.Field(i).Pointer() == 0 {
				continue
			}

			switch v.Field(i).Elem().Type() {
			case TypeOfString:
				if err := TransformString(param, v.Field(i).Elem()); err != nil {
					return err
				}
			}
		}
	}

	return
}

func SetDefaultValue(value reflect.Value, defaultStr string) (err error) {
	value.Set(reflect.New(value.Type().Elem()))

	switch value.Type() {
	case TypeOfStringP:
		value.Elem().SetString(defaultStr)
	case TypeOfIntP, TypeOfInt64P:
		value.Elem().SetInt(cast.ToInt64(defaultStr))
	case TypeOfFloat32P:
		err = errors.New("Unable to set default: Float32")
	case TypeOfFloat64P:
		value.Elem().SetFloat(cast.ToFloat64(defaultStr))
	case TypeOfBoolP:
		value.Elem().SetBool(cast.ToBool(defaultStr))
	default:
		err = errors.New("Unable to set default: no type defined")
	}
	return
}

func TransformString(param string, value reflect.Value) (err error) {
	k, v := getTagKV(param)

	switch k {
	case TransformStrHash:
		hashBytes32 := sha256.Sum256([]byte(value.String()))
		value.SetString(hex.EncodeToString(hashBytes32[:]))
	case TransformStrEncrypt:
		if value.String() == "" {
			return
		}
		if err := EncryptReflectValue(value); err != nil {
			log.Debug("Failed Encryption...")
			return err
		}
	case TransformStrDecrypt:
		if value.String() == "" {
			return
		}
		if err := DecryptReflectValue(value); err != nil {
			log.Debug("Failed Decryption...")
			return err
		}
	case TransformStrTrimChars:
		value.SetString(strings.Trim(value.String(), v))
	case TransformStrTrimSpace:
		value.SetString(strings.TrimSpace(value.String()))
	case TransformStrTruncate:
		truncateLength := cast.ToInt(v)
		if len(value.String()) < truncateLength {
			return
		}
		value.SetString(value.String()[:truncateLength])
	case TransformStrPasswordHash:
		if value.String() == "" {
			return
		}
		if err := PasswordHashReflectValue(value); err != nil {
			log.Debug("Failed Password Hashing..")
			return err
		}
	}

	return
}

func EncryptReflectValue(value reflect.Value) (err error) {
	log.Warn("DONT USE THIS KEY IN PRODUCTION.. FETCH KEY FROM PKI")
	key := []byte("AES256Key-32Characters1234567890")

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	nonce := []byte("DON'T USE ME")
	log.Warn("DONT USE THIS NONCE IN PRODUCTION.. GENERATE AND STORE RANDOM ONE")
	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	// nonce := make([]byte, 12)
	// if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
	// 	return err
	// }

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	cipherBytes := aesgcm.Seal(nil, nonce, []byte(value.String()), nil)

	value.SetString(hex.EncodeToString(cipherBytes))
	return
}

func DecryptReflectValue(value reflect.Value) (err error) {
	log.Warn("DONT USE THIS KEY IN PRODUCTION.. FETCH KEY FROM PKI")
	key := []byte("AES256Key-32Characters1234567890")
	ciphertext, err := hex.DecodeString(value.String())
	if err != nil {
		return err
	}

	nonce := []byte("DON'T USE ME")
	log.Warn("DONT USE THIS NONCE IN PRODUCTION.. FETCH THE ONE FOR THIS ENTRY")

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	value.SetString(string(plaintext))
	return
}

func PasswordHashReflectValue(value reflect.Value) (err error) {
	salt, err := getRandomSalt()
	if err != nil {
		log.Debug("Failed getting random salt")
		return err
	}
	key, err := argon2.Key([]byte(value.String()), []byte(salt), 2<<14-1, 1, 8, 64)
	if err != nil {
		log.Debug("Failed to get argon2 key")
		return err
	}
	// Store these if you need to verify later
	value.SetString(hex.EncodeToString(key))
	return
}

// From helpers/helpers.unmarshal.go.tpl
func Unmarshal(r *http.Request, dst interface{}) (err error) {
	if r.Method == http.MethodGet {
		t := reflect.TypeOf(dst).Elem()
		v := reflect.ValueOf(dst).Elem()

		if err := r.ParseForm(); err != nil {
			return err
		}

		for i := 0; i < t.NumField(); i++ {
			jsonTag := t.Field(i).Tag.Get(TagNameJSON)
			jsonParams := strings.Split(jsonTag, ",")
			if len(jsonParams) == 0 {
				continue
			}
			jsonName := jsonParams[0]

			validateTag := t.Field(i).Tag.Get(TagNameValidate)
			validateParams := strings.Split(validateTag, ",")
			required := false
			for _, param := range validateParams {
				if param == ValidateStrRequired {
					required = true
				}
			}

			formValue := r.Form.Get(jsonName)
			if formValue == "" && required {
				return errors.New("Empty required field")
			}

			v.Field(i).Set(reflect.New(v.Field(i).Type().Elem()))

			switch v.Field(i).Type() {
			case TypeOfStringP:
				v.Field(i).Elem().SetString(formValue)
			case TypeOfIntP:
				fallthrough
			case TypeOfInt64P:
				v.Field(i).Elem().SetInt(cast.ToInt64(formValue))
			case TypeOfFloat64P:
				v.Field(i).Elem().SetFloat(cast.ToFloat64(formValue))
			case TypeOfFloat32P:
				return errors.New("Float32 not supported")
			default:
				return errors.New(fmt.Sprint("Field not set:", v.Type().Field(i).Name))
			}
		}
		return
	}

	if r.Body == nil {
		return
	}

	err = json.NewDecoder(r.Body).Decode(dst)
	switch {
	case err == io.EOF:
		return nil
	case err != nil:
		log.Debug("Error decoding r.Body")
		return err
	}

	return
}

// From helpers.validate.go.tpl
func Validate(in interface{}) (msg string, err error) {
	t := reflect.TypeOf(in).Elem()
	v := reflect.ValueOf(in).Elem()

	for i := 0; i < t.NumField(); i++ {
		if !isBuiltin(t.Field(i).Type) {
			if IsZeroOfUnderlyingType(v.Field(i).Interface()) {
				continue
			}

			msg, err := Validate(v.Field(i).Interface())
			if err != nil {
				log.Debug("Error field validate:", err)
				return msg, err
			}

			if msg != "" {
				log.Debug("Failed field validate:", msg)
				return msg, err
			}
		}

		tag := t.Field(i).Tag.Get(TagNameValidate)

		if tag == "" || tag == "-" || tag == "_" || tag == " " {
			continue
		}

		fieldPointer := v.Field(i).Pointer()
		if strings.Contains(strings.ToLower(tag), ValidateStrRequired) {
			if fieldPointer == 0 {
				log.Debugf("Required field missing: %s", v.Type().Field(i).Name)
				return ValidateStrRequiredErr, nil
			}
		}

		if fieldPointer == 0 {
			continue
		}

		params := strings.Split(tag, ",")
		for _, param := range params {
			log.Debugf("Validating: %s - %s", v.Type().Field(i).Name, param)

			switch v.Field(i).Elem().Type() {
			case TypeOfString:
				if vMsg := ValidateString(param, v.Field(i).Elem().String()); vMsg != "" {
					return vMsg, nil
				}
			case TypeOfInt:
				if vMsg := ValidateInt(param, int(v.Field(i).Elem().Int())); vMsg != "" {
					return vMsg, nil
				}
			case TypeOfFloat64:
				if vMsg := ValidateFloat64(param, v.Field(i).Elem().Float()); vMsg != "" {
					return vMsg, nil
				}
			}
		}
	}

	return
}

func ValidateString(param, in string) (msg string) {
	k, v := getTagKV(param)

	switch k {
	case ValidateStrMaxLength:
		if len(in) > cast.ToInt(v) {
			return ValidateStrMaxLengthErr
		}
	case ValidateStrMinLength:
		if len(in) < cast.ToInt(v) {
			return ValidateStrMinLengthErr
		}
	case ValidateStrMustHaveChars:
		if !allCharsInStr(v, in) {
			return ValidateStrMustHaveCharsErr
		}
	case ValidateStrCantHaveChars:
		if strings.IndexAny(in, v) > -1 {
			return ValidateStrCantHaveCharsErr
		}
	case ValidateStrOnlyHaveChars:
		if !onlyCharsInStr(v, in) {
			return ValidateStrOnlyHaveCharsErr
		}
	}

	return
}

func ValidateInt(param string, in int) (msg string) {
	k, v := getTagKV(param)

	switch k {
	case ValidateStrGreaterThan:
		if in < cast.ToInt(v) {
			return ValidateStrGreaterThanErr
		}
	case ValidateStrLessThan:
		if in > cast.ToInt(v) {
			return ValidateStrLessThanErr
		}
	}

	return
}

func ValidateFloat64(param string, in float64) (msg string) {
	k, v := getTagKV(param)

	switch k {
	case ValidateStrGreaterThan:
		if in < cast.ToFloat64(v) {
			return ValidateStrGreaterThanErr
		}
	case ValidateStrLessThan:
		if in > cast.ToFloat64(v) {
			return ValidateStrLessThanErr
		}
	}

	return
}

`
