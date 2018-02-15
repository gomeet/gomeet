package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/fatih/color"
	//gogoproto "github.com/gogo/protobuf/proto"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	ggdescriptor "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"github.com/huandu/xstrings"
	"github.com/xtgo/set"
	options "google.golang.org/genproto/googleapis/api/annotations"

	"github.com/gomeet/gomeet/utils/project/helpers"
)

type byPathLengthDesc []string

func (s byPathLengthDesc) Len() int      { return len(s) }
func (s byPathLengthDesc) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byPathLengthDesc) Less(i, j int) bool {
	return strings.Count(s[i], "/") > strings.Count(s[j], "/")
}

var jsReservedRe = regexp.MustCompile(`(^|[^A-Za-z])(do|if|in|for|let|new|try|var|case|else|enum|eval|false|null|this|true|void|with|break|catch|class|const|super|throw|while|yield|delete|export|import|public|return|static|switch|typeof|default|extends|finally|package|private|continue|debugger|function|arguments|interface|protected|implements|instanceof)($|[^A-Za-z])`)

var (
	registry *ggdescriptor.Registry // some helpers need access to registry
)

func SetRegistry(reg *ggdescriptor.Registry) {
	registry = reg
}

var protoHelpersFuncMap = template.FuncMap{
	"lowerNospaceCase":               lowerNospaceCase,
	"upperNospaceCase":               upperNospaceCase,
	"upperPascalCase":                upperPascalCase,
	"lowerPascalCase":                lowerPascalCase,
	"lowerKebabCase":                 lowerKebabCase,
	"upperKebabCase":                 upperKebabCase,
	"upperSnakeCase":                 upperSnakeCase,
	"lowerSnakeCase":                 LowerSnakeCase,
	"cliCmdHelpString":               cliCmdHelpString,
	"grpcMethodCliHelp":              grpcMethodCliHelp,
	"curlCmdHelpString":              curlCmdHelpString,
	"runFunctionalTestSession":       runFunctionalTestSession,
	"remoteCliHelp":                  remoteCliHelp,
	"remoteCliGetActionMap":          remoteCliGetActionMap,
	"grpcMethodCmdName":              grpcMethodCmdName,
	"grpcMethodInputGoType":          grpcMethodInputGoType,
	"grpcMethodOutputGoType":         grpcMethodOutputGoType,
	"grpcFunctestHttp":               grpcFunctestHttp,
	"grpcFunctestHttpExtraImport":    grpcFunctestHttpExtraImport,
	"grpcMethodCmdCountArgsValidity": grpcMethodCmdCountArgsValidity,
	"grpcMethodCmdCastArgsToVar":     grpcMethodCmdCastArgsToVar,
	"grpcMethodCmdImports":           grpcMethodCmdImports,
	"protoMessagesNeededImports":     protoMessagesNeededImports,
	"httpMetricsRouteMap":            httpMetricsRouteMap,
	"subSvcPkgString":                subSvcPkgString,
	"string": func(i interface {
		String() string
	}) string {
		return i.String()
	},
	"json": func(v interface{}) string {
		a, err := json.Marshal(v)
		if err != nil {
			return err.Error()
		}
		return string(a)
	},
	"prettyjson": func(v interface{}) string {
		a, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err.Error()
		}
		return string(a)
	},
	"splitArray": func(sep string, s string) []interface{} {
		var r []interface{}
		t := strings.Split(s, sep)
		for i := range t {
			if t[i] != "" {
				r = append(r, t[i])
			}
		}
		return r
	},
	"first": func(a []string) string {
		return a[0]
	},
	"last": func(a []string) string {
		return a[len(a)-1]
	},
	"upperFirst": func(s string) string {
		return strings.ToUpper(s[:1]) + s[1:]
	},
	"lowerFirst": func(s string) string {
		return strings.ToLower(s[:1]) + s[1:]
	},
	"contains": func(sub, s string) bool {
		return strings.Contains(s, sub)
	},
	"trimstr": func(cutset, s string) string {
		return strings.Trim(s, cutset)
	},
	"leftPad":                 leftPad,
	"rightPad":                rightPad,
	"leftPad2Len":             leftPad2Len,
	"rightPad2Len":            rightPad2Len,
	"getProtoFile":            getProtoFile,
	"getMessageType":          getMessageType,
	"getEnumValue":            getEnumValue,
	"isFieldMessage":          isFieldMessage,
	"isFieldMessageTimeStamp": isFieldMessageTimeStamp,
	"isFieldRepeated":         isFieldRepeated,
	"haskellType":             haskellType,
	"goType":                  goType,
	"goTypeWithPackage":       goTypeWithPackage,
	"jsType":                  jsType,
	"jsSuffixReserved":        jsSuffixReservedKeyword,
	"namespacedFlowType":      namespacedFlowType,
	"httpVerb":                httpVerb,
	"httpPath":                httpPath,
	"httpBody":                httpBody,
	"shortType":               shortType,
	"messageGoType":           messageGoType,
	"messageFake":             messageFake,
	"urlHasVarsFromMessage":   urlHasVarsFromMessage,
	"lowerGoNormalize":        lowerGoNormalize,
	"goNormalize":             goNormalize,
}

func init() {
	for k, v := range sprig.TxtFuncMap() {
		protoHelpersFuncMap[k] = v
	}
}

func ProtoHelpersFuncMap() template.FuncMap {
	return protoHelpersFuncMap
}

/*
 * leftPad and rightPad just repoeat the padStr the indicated
 * number of times
 *
 */
func leftPad(s string, padStr string, pLen int) string {
	return strings.Repeat(padStr, pLen) + s
}
func rightPad(s string, padStr string, pLen int) string {
	return s + strings.Repeat(padStr, pLen)
}

/* the Pad2Len functions are generally assumed to be padded with short sequences of strings
 * in many cases with a single character sequence
 *
 * so we assume we can build the string out as if the char seq is 1 char and then
 * just substr the string if it is longer than needed
 *
 * this means we are wasting some cpu and memory work
 * but this always get us to want we want it to be
 *
 * in short not optimized to for massive string work
 *
 * If the overallLen is shorter than the original string length
 * the string will be shortened to this length (substr)
 *
 */
func rightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}
func leftPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

func getProtoFile(name string) *ggdescriptor.File {
	if registry == nil {
		return nil
	}
	file, err := registry.LookupFile(name)
	if err != nil {
		panic(err)
	}
	return file
}

func getMessageTypeRegistry(name string) (*ggdescriptor.Message, error) {
	if registry != nil {
		msg, err := registry.LookupMsg(".", name)
		if err != nil {
			return nil, err
		}
		return msg, nil
	}

	return nil, nil
}

func getMessageType(f *descriptor.FileDescriptorProto, name string) *ggdescriptor.Message {
	msg, err := getMessageTypeRegistry(name)
	if err != nil {
		panic(err)
	}
	if msg != nil {
		return msg
	}

	// name is in the form .packageName.MessageTypeName.InnerMessageTypeName...
	// e.g. .article.ProductTag
	splits := strings.Split(name, ".")
	target := splits[len(splits)-1]
	for _, m := range f.MessageType {
		if target == *m.Name {
			return &ggdescriptor.Message{
				DescriptorProto: m,
			}
		}
	}
	return nil
}

func getEnumTypeRegistry(name string) (*ggdescriptor.Enum, error) {
	if registry != nil {
		enum, err := registry.LookupEnum(".", name)
		if err != nil {
			return nil, err
		}
		return enum, nil
	}

	return nil, nil
}

func getEnumType(f *descriptor.FileDescriptorProto, name string) *ggdescriptor.Enum {
	enum, err := getEnumTypeRegistry(name)
	if err != nil {
		panic(err)
	}
	if enum != nil {
		return enum
	}

	// name is in the form .packageName.MessageTypeName.InnerMessageTypeName...
	// e.g. .article.ProductTag
	splits := strings.Split(name, ".")
	target := splits[len(splits)-1]
	for _, m := range f.EnumType {
		if target == *m.Name {
			return &ggdescriptor.Enum{
				EnumDescriptorProto: m,
			}
		}
	}
	return nil
}

func getEnumValue(f []*descriptor.EnumDescriptorProto, name string) []*descriptor.EnumValueDescriptorProto {
	for _, item := range f {
		if strings.EqualFold(*item.Name, name) {
			return item.GetValue()
		}
	}

	return nil
}

func isFieldMessageTimeStamp(f *descriptor.FieldDescriptorProto) bool {
	if f.Type != nil && *f.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		if strings.Compare(*f.TypeName, ".google.protobuf.Timestamp") == 0 {
			return true
		}
	}
	return false
}

func isFieldMessage(f *descriptor.FieldDescriptorProto) bool {
	if f.Type != nil && *f.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		return true
	}

	return false
}

func isFieldRepeated(f *descriptor.FieldDescriptorProto) bool {
	if f.Type != nil && f.Label != nil && *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		return true
	}

	return false
}

func goTypeWithPackage(f *descriptor.FieldDescriptorProto) string {
	pkg := ""
	if *f.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE || *f.Type == descriptor.FieldDescriptorProto_TYPE_ENUM {
		pkg = getPackageTypeName(*f.TypeName)
	}
	return goType(pkg, f)
}

func haskellType(pkg string, f *descriptor.FieldDescriptorProto) string {
	switch *f.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[Float]"
		}
		return "Float"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[Float]"
		}
		return "Float"
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[Int64]"
		}
		return "Int64"
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[Word]"
		}
		return "Word"
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[Int]"
		}
		return "Int"
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[Word]"
		}
		return "Word"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[Bool]"
		}
		return "Bool"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[Text]"
		}
		return "Text"
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if pkg != "" {
			pkg = pkg + "."
		}
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Sprintf("[%s%s]", pkg, shortType(*f.TypeName))
		}
		return fmt.Sprintf("%s%s", pkg, shortType(*f.TypeName))
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[Word8]"
		}
		return "Word8"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return fmt.Sprintf("%s%s", pkg, shortType(*f.TypeName))
	default:
		return "Generic"
	}
}

func goType(pkg string, f *descriptor.FieldDescriptorProto) string {
	if pkg != "" {
		pkg = pkg + "."
	}
	switch *f.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[]float64"
		}
		return "float64"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[]float32"
		}
		return "float32"
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[]int64"
		}
		return "int64"
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[]uint64"
		}
		return "uint64"
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[]int32"
		}
		return "int32"
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[]uint32"
		}
		return "uint32"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[]bool"
		}
		return "bool"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[]string"
		}
		return "string"
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Sprintf("[]*%s%s", pkg, shortType(*f.TypeName))
		}
		return fmt.Sprintf("*%s%s", pkg, shortType(*f.TypeName))
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		if *f.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return "[]byte"
		}
		return "byte"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return fmt.Sprintf("*%s%s", pkg, shortType(*f.TypeName))
	default:
		return "interface{}"
	}
}

func jsType(f *descriptor.FieldDescriptorProto) string {
	template := "%s"
	if isFieldRepeated(f) {
		template = "Array<%s>"
	}

	switch *f.Type {
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		return fmt.Sprintf(template, namespacedFlowType(*f.TypeName))
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
		descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_SINT64:
		return fmt.Sprintf(template, "number")
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return fmt.Sprintf(template, "boolean")
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return fmt.Sprintf(template, "Uint8Array")
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return fmt.Sprintf(template, "string")
	default:
		return fmt.Sprintf(template, "any")
	}
}

func jsSuffixReservedKeyword(s string) string {
	return jsReservedRe.ReplaceAllString(s, "${1}${2}_${3}")
}

func getPackageTypeName(s string) string {
	if strings.Compare(s, ".google.protobuf.Timestamp") == 0 {
		return "timestamp"
	}
	if strings.Contains(s, ".") {
		return strings.Split(s, ".")[1]
	}
	return ""
}

func shortType(s string) string {
	t := strings.Split(s, ".")
	return t[len(t)-1]
}

func namespacedFlowType(s string) string {
	trimmed := strings.TrimLeft(s, ".")
	splitted := strings.Split(trimmed, ".")
	return strings.Join(splitted, "$")
}

func httpPathFields(m *descriptor.MethodDescriptorProto) (fields []*descriptor.FieldDescriptorProto) {
	_, fields = httpPathAndFields(m)

	return fields
}

func httpPathAndFields(m *descriptor.MethodDescriptorProto) (path string, fields []*descriptor.FieldDescriptorProto) {
	path = httpPath(m)

	re := regexp.MustCompile(`\{([^\{\}]*)\}`)
	submatchall := re.FindAllString(path, -1)

	for _, param := range submatchall {
		param = strings.Trim(param, "{")
		param = strings.Trim(param, "}")
		in := m.GetInputType()
		allF, err := getMessageTypeRegistry(in)
		if err != nil || allF == nil {
			return path, fields
		}
		for _, f := range allF.GetField() {
			if param == f.GetName() {
				fields = append(fields, f)
			}
		}
	}

	return path, fields
}

func httpPath(m *descriptor.MethodDescriptorProto) string {
	if m == nil {
		return ""
	}
	ext, err := proto.GetExtension(m.Options, options.E_Http)
	if err != nil {
		return err.Error()
	}
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		return fmt.Sprintf("extension is %T; want an HttpRule", ext)
	}

	switch t := opts.Pattern.(type) {
	default:
		return ""
	case *options.HttpRule_Get:
		return t.Get
	case *options.HttpRule_Post:
		return t.Post
	case *options.HttpRule_Put:
		return t.Put
	case *options.HttpRule_Delete:
		return t.Delete
	case *options.HttpRule_Patch:
		return t.Patch
	case *options.HttpRule_Custom:
		return t.Custom.Path
	}
}

func httpVerb(m *descriptor.MethodDescriptorProto) string {
	if m == nil {
		return ""
	}
	ext, err := proto.GetExtension(m.Options, options.E_Http)
	if err != nil {
		return err.Error()
	}
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		return fmt.Sprintf("extension is %T; want an HttpRule", ext)
	}

	switch t := opts.Pattern.(type) {
	default:
		return ""
	case *options.HttpRule_Get:
		return "GET"
	case *options.HttpRule_Post:
		return "POST"
	case *options.HttpRule_Put:
		return "PUT"
	case *options.HttpRule_Delete:
		return "DELETE"
	case *options.HttpRule_Patch:
		return "PATCH"
	case *options.HttpRule_Custom:
		return t.Custom.Kind
	}
}

func httpBody(m *descriptor.MethodDescriptorProto) string {
	if m == nil {
		return ""
	}
	ext, err := proto.GetExtension(m.Options, options.E_Http)
	if err != nil {
		return err.Error()
	}
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		return fmt.Sprintf("extension is %T; want an HttpRule", ext)
	}
	return opts.Body
}

func urlHasVarsFromMessage(path string, d *ggdescriptor.Message) bool {
	if d == nil {
		return false
	}
	for _, field := range d.Field {
		if !isFieldMessage(field) {
			if strings.Contains(path, fmt.Sprintf("{%s}", *field.Name)) {
				return true
			}
		}
	}

	return false
}

// lowerGoNormalize takes a string and applies formatting
// rules to conform to Golang convention. It applies a camel
// case filter, lowers the first character and formats fields
// with `id` to `ID`.
func lowerGoNormalize(s string) string {
	fmtd := xstrings.ToCamelCase(s)
	fmtd = xstrings.FirstRuneToLower(fmtd)
	return formatID(s, fmtd)
}

// goNormalize takes a string and applies formatting rules
// to conform to Golang convention. It applies a camel case
// filter and formats fields with `id` to `ID`.
func goNormalize(s string) string {
	fmtd := xstrings.ToCamelCase(s)
	return formatID(s, fmtd)
}

// formatID takes a base string alonsgide a formatted string.
// It acts as a transformation filter for fields containing
// `id` in order to conform to Golang convention.
func formatID(base string, formatted string) string {
	if formatted == "" {
		return formatted
	}
	switch {
	case base == "id":
		// id -> ID
		return "ID"
	case strings.HasPrefix(base, "id_"):
		// id_some -> IDSome
		return "ID" + formatted[2:]
	case strings.HasSuffix(base, "_id"):
		// some_id -> SomeID
		return formatted[:len(formatted)-2] + "ID"
	case strings.HasSuffix(base, "_ids"):
		// some_ids -> SomeIDs
		return formatted[:len(formatted)-3] + "IDs"
	}
	return formatted
}

// string in lowercasewithoutspace
func lowerNospaceCase(s string) string {
	s = lowerKebabCase(s)
	return strings.Replace(s, "-", "", -1)
}

// string in UPPERCASEWITHOUTSPACE
func upperNospaceCase(s string) string {
	s = upperKebabCase(s)
	return strings.Replace(s, "-", "", -1)
}

// the service name UpperPascalCase
func upperPascalCase(s string) string {
	s = strings.Replace(xstrings.ToSnakeCase(s), "-", "_", -1)
	if len(s) > 1 {
		return xstrings.ToCamelCase(s)
	}

	return strings.ToUpper(s[:1])
}

// the service name lowerPascalCase
func lowerPascalCase(s string) string {
	s = strings.Replace(s, "-", "_", -1)
	if len(s) > 1 {
		s = xstrings.ToCamelCase(s)
	}

	return strings.ToLower(s[:1]) + s[1:]
}

// the service name lower-kebab-case
func lowerKebabCase(s string) string {
	s = xstrings.ToSnakeCase(s)
	s = strings.Replace(s, "_", "-", -1)
	return strings.ToLower(s)
}

// the service name UPPER-KEBAB-CASE
func upperKebabCase(s string) string {
	s = xstrings.ToSnakeCase(s)
	s = strings.Replace(s, "_", "-", -1)
	return strings.ToUpper(s)
}

// string in UPPER_SNAKE_CASE
func upperSnakeCase(s string) string {
	return strings.ToUpper(xstrings.ToSnakeCase(s))
}

// string in lower_snake_case
func LowerSnakeCase(s string) string {
	return strings.ToLower(xstrings.ToSnakeCase(s))
}

func remoteCliGetActionMap(name string, protoFiles []*descriptor.FileDescriptorProto) string {
	var ret []string
	for _, file := range protoFiles {
		for _, svc := range file.GetService() {
			for _, method := range svc.GetMethod() {
				ret = append(
					ret,
					fmt.Sprintf(
						"\"%s\": %s",
						LowerSnakeCase(method.GetName()),
						fmt.Sprintf("c.%s", grpcMethodCmdName(method.GetName())),
					),
				)
			}
		}
	}

	return fmt.Sprintf("\n\t\t%s,\n", strings.Join(ret, ",\n\t\t"))
}

func grpcMethodCmdName(name string) string {
	return fmt.Sprintf("cmd%s", upperPascalCase(name))
}

func countFieldsInMessage(name string) (count int) {
	msg, err := getMessageTypeRegistry(name)
	if err != nil {
		return 0
	}

	if len(msg.Fields) > 0 {
		for _, field := range msg.Fields {
			if isFieldMessage(field.FieldDescriptorProto) {
				count = count + countFieldsInMessage(field.GetTypeName())
			} else {
				count = count + 1
			}
		}
	}

	return count
}

func grpcMethodCmdCountArgsValidity(m *descriptor.MethodDescriptorProto) int {
	return countFieldsInMessage(m.GetInputType())
}

func fieldNfo(prefix string, field *ggdescriptor.Field, opts ...string) (fieldName, varName, cmdHelpName string) {
	sanitizeSplit := []string{}
	split := strings.Split(prefix, ".")
	split = append(split, field.GetName())
	for _, s := range split {
		tmp := LowerSnakeCase(s)
		if tmp != "" {
			sanitizeSplit = append(sanitizeSplit, tmp)
		}
	}
	fieldName = prefix + upperPascalCase(field.GetName())
	varName = lowerPascalCase(strings.Join(sanitizeSplit, "_"))
	cmdHelpName = strings.Join(sanitizeSplit, ".")
	if len(opts) > 0 {
		if opts[0] != "" {
			cmdHelpName = strings.TrimPrefix(cmdHelpName, opts[0]+".")
		}
	}

	return fieldName, varName, cmdHelpName
}

func castCmdField(idx int, prefix string, field *ggdescriptor.Field) string {
	var ret bytes.Buffer
	gT := goTypeWithPackage(field.FieldDescriptorProto)

	msgFieldName, varName, cmdHelpName := fieldNfo(prefix, field, "req")

	ret.WriteString(
		fmt.Sprintf(
			"// cast args[%d] in %s - type %s to go type %s\n",
			idx,
			msgFieldName,
			field.GetType(),
			gT,
		),
	)

	syntax := field.Message.File.GetSyntax()
	if syntax == "" {
		syntax = "proto2"
	}

	writeAffectation := func(b *bytes.Buffer, syn, msg, v string) {
		b.WriteString(msg)
		b.WriteString(" = ")
		switch syn {
		case "proto2":
			b.WriteString("&")
			b.WriteString(v)
		case "proto3":
			b.WriteString(v)
		}
		b.WriteString("\n")
	}

	switch gT {
	case "[]float64":
	case "float64":
		ret.WriteString(fmt.Sprintf("%s, err := strconv.ParseFloat(args[%d], 64)\n", varName, idx))
		ret.WriteString("if err != nil {\n")
		ret.WriteString(fmt.Sprintf("\treturn \"\", fmt.Errorf(\"Bad arguments : %s is not %s\")\n", cmdHelpName, gT))
		ret.WriteString("}\n")
		writeAffectation(&ret, syntax, msgFieldName, varName)
		//ret.WriteString(fmt.Sprintf("%s = %s\n", msgFieldName, varName))
	case "[]float32":
	case "float32":
		ret.WriteString(fmt.Sprintf("%s, err := strconv.ParseFloat(args[%d], 32)\n", varName, idx))
		ret.WriteString("if err != nil {\n")
		ret.WriteString(fmt.Sprintf("\treturn \"\", fmt.Errorf(\"Bad arguments : %s is not %s\")\n", cmdHelpName, gT))
		ret.WriteString("}\n")
		writeAffectation(&ret, syntax, msgFieldName, varName)
		//ret.WriteString(fmt.Sprintf("%s = %s\n", msgFieldName, varName))
	case "[]uint32":
	case "uint32":
		ret.WriteString(fmt.Sprintf("%s, err := strconv.ParseUint(args[%d], 10, 32)\n", varName, idx))
		ret.WriteString("if err != nil {\n")
		ret.WriteString(fmt.Sprintf("\treturn \"\", fmt.Errorf(\"Bad arguments : %s is not %s\")\n", cmdHelpName, gT))
		ret.WriteString("}\n")
		varNameCast := varName + "Cast"
		ret.WriteString(fmt.Sprintf("%s := uint32(%s)\n", varNameCast, varName))
		writeAffectation(&ret, syntax, msgFieldName, varNameCast)
		//ret.WriteString(fmt.Sprintf("%s = uint32(%s)\n", msgFieldName, varName))
	case "[]int32":
	case "int32":
		ret.WriteString(fmt.Sprintf("%s, err := strconv.ParseInt(args[%d], 10, 32)\n", varName, idx))
		ret.WriteString("if err != nil {\n")
		ret.WriteString(fmt.Sprintf("\treturn \"\", fmt.Errorf(\"Bad arguments : %s is not %s\")\n", cmdHelpName, gT))
		ret.WriteString("}\n")
		varNameCast := varName + "Cast"
		ret.WriteString(fmt.Sprintf("%s := int32(%s)\n", varNameCast, varName))
		writeAffectation(&ret, syntax, msgFieldName, varNameCast)
		//ret.WriteString(fmt.Sprintf("%s = int32(%s)\n", msgFieldName, varName))
	case "[]uint64":
	case "uint64":
		ret.WriteString(fmt.Sprintf("%s, err := strconv.ParseUint(args[%d], 10, 64)\n", varName, idx))
		ret.WriteString("if err != nil {\n")
		ret.WriteString(fmt.Sprintf("\treturn \"\", fmt.Errorf(\"Bad arguments : %s is not %s\")\n", cmdHelpName, gT))
		ret.WriteString("}\n")
		varNameCast := varName + "Cast"
		ret.WriteString(fmt.Sprintf("%s := uint64(%s)\n", varNameCast, varName))
		writeAffectation(&ret, syntax, msgFieldName, varNameCast)
		//ret.WriteString(fmt.Sprintf("%s = uint64(%s)\n", msgFieldName, varName))
	case "[]int64":
	case "int64":
		ret.WriteString(fmt.Sprintf("%s, err := strconv.ParseInt(args[%d], 10, 64)\n", varName, idx))
		ret.WriteString("if err != nil {\n")
		ret.WriteString(fmt.Sprintf("\treturn \"\", fmt.Errorf(\"Bad arguments : %s is not %s\")\n", cmdHelpName, gT))
		ret.WriteString("}\n")
		varNameCast := varName + "Cast"
		ret.WriteString(fmt.Sprintf("%s := int64(%s)\n", varNameCast, varName))
		writeAffectation(&ret, syntax, msgFieldName, varNameCast)
		//ret.WriteString(fmt.Sprintf("%s = int64(%s)\n", msgFieldName, varName))
	case "[]bool":
	case "bool":
		ret.WriteString(fmt.Sprintf("%s, err := strconv.ParseBool(args[%d])\n", varName, idx))
		ret.WriteString("if err != nil {\n")
		ret.WriteString(fmt.Sprintf("\treturn \"\", fmt.Errorf(\"Bad arguments : %s is not %s\")\n", cmdHelpName, gT))
		ret.WriteString("}\n")
		writeAffectation(&ret, syntax, msgFieldName, varName)
		//ret.WriteString(fmt.Sprintf("%s = %s\n", msgFieldName, varName))
	case "[]string":
	case "string":
		writeAffectation(&ret, syntax, msgFieldName, fmt.Sprintf("args[%d]", idx))
		//ret.WriteString(fmt.Sprintf("%s = args[%d]\n", msgFieldName, idx))
	case "[]byte":
		writeAffectation(&ret, syntax, msgFieldName, fmt.Sprintf("[]byte(args[%d])", idx))
		//ret.WriteString(fmt.Sprintf("%s = []byte(args[%d])\n", msgFieldName, idx))
	case "byte":
		ret.WriteString(fmt.Sprintf("%s = []byte(args[%d])\n", varName, idx))
		writeAffectation(&ret, syntax, msgFieldName, fmt.Sprintf("%s[0]", varName))
		//ret.WriteString(fmt.Sprintf("%s = %s[0]\n", msgFieldName, varName))
	default:
		if field.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM {
			if enum, err := getEnumTypeRegistry(field.GetTypeName()); err == nil {
				eName := enum.GetName()
				if len(enum.Outers) > 0 {
					eName = fmt.Sprintf("%s_%s", strings.Join(enum.Outers, "_"), eName)
				}
				ret.WriteString(fmt.Sprintf("%s, ok := %s.%s_value[strings.ToUpper(args[%d])]\n", varName, enum.File.GoPkg.Name, eName, idx))
				ret.WriteString("if !ok {\n")
				ret.WriteString(fmt.Sprintf("\treturn \"\", fmt.Errorf(\"Bad arguments : unknown %s \\\"%%s\\\"\", args[%d])\n", cmdHelpName, idx))
				ret.WriteString("}\n")
				ret.WriteString(fmt.Sprintf("%s = %s.%s(%s)\n", msgFieldName, enum.File.GoPkg.Name, eName, varName))
			}
		}
	}

	return ret.String()
}

func castCmdFields(idxAcc int, varName, name string) (idx int, cast []string) {
	idx = idxAcc
	msg, err := getMessageTypeRegistry(name)
	if err != nil {
		return idx, cast
	}
	cast = append(cast,
		fmt.Sprintf(
			"// decl %s for no nil panic",
			varName,
		),
	)
	cast = append(cast,
		fmt.Sprintf(
			"%s = &%s{}\n",
			varName,
			messageGoType(name),
		),
	)

	if len(msg.Fields) > 0 {
		for _, field := range msg.Fields {
			if isFieldMessage(field.FieldDescriptorProto) {
				newIdx, subCast := castCmdFields(
					idx,
					fmt.Sprintf("%s.%s", varName, upperPascalCase(field.GetName())),
					field.GetTypeName(),
				)
				for _, l := range subCast {
					cast = append(cast, l)
				}
				idx = newIdx
			} else {
				cast = append(cast, castCmdField(idx, varName+".", field))
				idx = idx + 1
			}
		}
	}

	return idx, cast
}

func grpcMethodCmdCastArgsToVar(varName string, m *descriptor.MethodDescriptorProto) string {
	in := m.GetInputType()

	_, err := getMessageTypeRegistry(in)
	if err != nil {
		return ""
	}

	ret := []string{
		"// request message",
		fmt.Sprintf(
			"var %s *%s\n",
			varName,
			messageGoType(in),
		),
	}

	_, cast := castCmdFields(0, varName, in)
	ret = append(ret, cast...)

	return strings.Join(ret, "\n\t")
}

func grpcMethodCmdImportsNeedStrings(name string) bool {
	msg, err := getMessageTypeRegistry(name)
	if err != nil {
		return false
	}

	if len(msg.Fields) > 0 {
		for _, field := range msg.Fields {
			if isFieldMessage(field.FieldDescriptorProto) {
				b := grpcMethodCmdImportsNeedStrings(field.GetTypeName())
				if b {
					return true
				}
			} else {
				if field.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM {
					return true
				}
			}
		}
	}

	return false
}

func grpcMethodCmdImportsNeedStrconv(name string) bool {
	msg, err := getMessageTypeRegistry(name)
	if err != nil {
		return false
	}

	if len(msg.Fields) > 0 {
		for _, field := range msg.Fields {
			if isFieldMessage(field.FieldDescriptorProto) {
				b := grpcMethodCmdImportsNeedStrconv(field.GetTypeName())
				if b {
					return true
				}
			} else {
				switch field.GetType() {
				case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
					descriptor.FieldDescriptorProto_TYPE_FLOAT,
					descriptor.FieldDescriptorProto_TYPE_INT64,
					descriptor.FieldDescriptorProto_TYPE_UINT64,
					descriptor.FieldDescriptorProto_TYPE_INT32,
					descriptor.FieldDescriptorProto_TYPE_UINT32,
					descriptor.FieldDescriptorProto_TYPE_BOOL:
					return true
					//case descriptor.FieldDescriptorProto_TYPE_STRING,
					//descriptor.FieldDescriptorProto_TYPE_BYTES,
					//descriptor.FieldDescriptorProto_TYPE_ENUM:
					//default:
					//continue
				}
			}
		}
	}

	return false
}

func uniqStringSlice(ss []string) []string {
	data := sort.StringSlice(ss)
	sort.Sort(data)
	n := set.Uniq(data)

	return data[:n]
}

func dedupImports(imports ...string) string {
	imports = uniqStringSlice(imports)
	return fmt.Sprintf("\t\"%s\"", strings.Join(imports, "\"\n\t\""))
}

func messagesNeededImports(recursive bool, name string) map[string]string {
	ret := map[string]string{}
	msg, err := getMessageTypeRegistry(name)
	if err != nil {
		return ret
	}

	if recursive && len(msg.Fields) > 0 {
		for _, field := range msg.Fields {
			if isFieldMessage(field.FieldDescriptorProto) {
				for path, name := range messagesNeededImports(recursive, field.GetTypeName()) {
					if path == "" {
						continue
					}
					if _, ok := ret[path]; !ok {
						ret[path] = name
					}
				}
			} else {
				goPkg := field.Message.File.GoPkg
				if goPkg.Path == "" {
					continue
				}
				if _, ok := ret[goPkg.Path]; !ok {
					ret[goPkg.Path] = goPkg.Name
				}
			}
		}
	} else {
		goPkg := msg.File.GoPkg
		if goPkg.Path != "" {
			if _, ok := ret[goPkg.Path]; !ok {
				ret[goPkg.Path] = goPkg.Name
			}
		}
	}

	return ret
}

func protoMessagesNeededImports(recursive bool, protoMessages ...string) string {
	if len(protoMessages) < 1 {
		return ""
	}

	imp := map[string]string{}
	for _, name := range protoMessages {
		for path, name := range messagesNeededImports(recursive, name) {
			if path == "" {
				continue
			}
			if _, ok := imp[path]; !ok {
				imp[path] = name
			}
		}
	}

	ret := []string{}
	for path, name := range imp {
		ret = append(ret, fmt.Sprintf("%s \"%s\"", name, path))
	}

	return "\t" + strings.Join(ret, "\t\n")
}

func grpcMethodCmdImports(m *descriptor.MethodDescriptorProto) string {
	ret := []string{}

	if grpcMethodCmdImportsNeedStrconv(m.GetInputType()) {
		ret = append(ret, `"strconv"`)
	}

	if grpcMethodCmdImportsNeedStrings(m.GetInputType()) {
		ret = append(ret, `"strings"`)
	}

	ret = append(ret, "")
	ret = append(ret, protoMessagesNeededImports(true, m.GetInputType()))

	return "\t" + strings.Join(ret, "\t\n")
}

func httpMetricsRouteMap(name string, protoFiles []*descriptor.FileDescriptorProto) string {
	m := make(map[string]string)

	for _, file := range protoFiles {
		for _, svc := range file.GetService() {
			for _, method := range svc.GetMethod() {
				// exclude streamed methods
				if method.GetServerStreaming() || method.GetClientStreaming() {
					continue
				}
				httpP := httpPath(method)
				label := fmt.Sprintf("Api.%s", upperPascalCase(method.GetName()))
				if strings.Contains(httpP, "{") {
					// trunc httpP with first "{"
					newHttpP := strings.TrimSuffix(strings.Split(httpP, "{")[0], "/")
					helpers.Log(
						helpers.LogDangerous,
						fmt.Sprintf(
							"Gomeet's HTTP/1.1 metrics methods doesn't yet support restFul declarations path %s will be converted with %s\n",
							color.YellowString(httpP),
							color.CyanString(newHttpP),
						),
					)
					httpP = newHttpP
				}
				if httpP == "" {
					helpers.Log(
						helpers.LogDangerous,
						fmt.Sprintf(
							"empty HTTP/1.1 metrics path %s will be skipped\n",
							color.CyanString(label),
						),
					)
					continue
				}
				if val, present := m[httpP]; present {
					label = fmt.Sprintf(
						"%s|%s",
						val,
						label,
					)
					helpers.Log(
						helpers.LogDangerous,
						fmt.Sprintf(
							"existing HTTP/1.1 metrics path %s with label %s. new label become %s",
							color.YellowString(httpP),
							color.YellowString(val),
							color.CyanString(label),
						),
					)
				}
				m[httpP] = label
			}
		}
	}
	keys := make([]string, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Sort(byPathLengthDesc(keys))

	// sort m keys by count of "/" desc /foo/bar/plop before /foo/bar before /foo
	var ret []string
	for _, k := range keys {
		ret = append(
			ret,
			fmt.Sprintf(
				"\"%s\": \"%s\"",
				k,
				m[k],
			),
		)
	}

	return fmt.Sprintf("\n\t\t%s,\n", strings.Join(ret, ",\n\t\t"))
}
func remoteCliHelp(name string, protoFiles []*descriptor.FileDescriptorProto) string {
	var ret []string
	for _, file := range protoFiles {
		for _, svc := range file.GetService() {
			for _, method := range svc.GetMethod() {
				ret = append(
					ret,
					fmt.Sprintf(
						"\t┌─ %s\n\t└─ %s",
						grpcMethodCliHelp(method),
						fmt.Sprintf("call %s service", LowerSnakeCase(method.GetName())),
					),
				)
			}
		}
	}

	return fmt.Sprintf("\n%s\n", strings.Join(ret, "\n\n"))
}

func runFunctionalTestSession(name string, protoFiles []*descriptor.FileDescriptorProto) string {
	var ret []string

	for _, file := range protoFiles {
		for _, svc := range file.GetService() {
			for _, method := range svc.GetMethod() {
				ret = append(
					ret,
					fmt.Sprintf(
						"failures = appendFailures(failures, functest.Test%s(config))",
						upperPascalCase(method.GetName()),
					),
				)
				ret = append(
					ret,
					fmt.Sprintf(
						"failures = appendFailures(failures, functest.TestHttp%s(config))",
						upperPascalCase(method.GetName()),
					),
				)
			}
		}
	}
	return fmt.Sprintf("\n\t%s\n", strings.Join(ret, "\n\t"))
}

func subSvcPkgString(subSvc []*helpers.PkgNfo) string {
	if len(subSvc) < 1 {
		return ""
	}

	var ret []string
	for _, svc := range subSvc {
		ret = append(ret, svc.GoPkg())
	}
	sort.Strings(ret)

	return strings.Join(ret, ",")
}

func curlCmdHelpString(name string, protoFiles []*descriptor.FileDescriptorProto) string {
	var ret []string

	for _, file := range protoFiles {
		for _, svc := range file.GetService() {
			for _, method := range svc.GetMethod() {
				attr := ""
				httpV := httpVerb(method)
				httpB := httpBody(method)
				switch httpV {
				case "POST", "PUT":
					attr = fmt.Sprintf(" -d '%s'", curlCmdHelpJson(file, method.GetInputType()))
				default:
					if httpB != "" {
						helpers.Log(
							helpers.LogError,
							fmt.Sprintf(
								"not implemented jsonBody on no PUT/POST methods %s %s - body : %s",
								httpV,
								httpPath(method),
								httpB,
							),
						)
					}
				}
				ret = append(
					ret,
					fmt.Sprintf(
						"curl -X %s http://localhost:13000%s%s",
						rightPad2Len(httpV, " ", 6), // 6 - DELETE len + 1
						httpPath(method),
						attr,
					),
				)
			}
		}
	}
	return fmt.Sprintf("  $ %s", strings.Join(ret, "\n  $ "))
}

func curlCmdHelpJsonValue(file *descriptor.FileDescriptorProto, f *descriptor.FieldDescriptorProto) string {
	template := "%s"
	if isFieldRepeated(f) {
		template = "Array<%s>"
	}

	switch *f.Type {
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		enum := getEnumType(file, f.GetTypeName())
		enumVals := []string{}
		for _, v := range enum.GetValue() {
			enumVals = append(enumVals, v.GetName())
		}
		return fmt.Sprintf(template, "\""+strings.Join(enumVals, "|")+"\"")
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
		descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_SINT64:
		return fmt.Sprintf(template, "<number>")
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return fmt.Sprintf(template, "<boolean>")
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return fmt.Sprintf(template, "<Uint8Array>")
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return fmt.Sprintf(template, "\"<string>\"")
	default:
		return fmt.Sprintf(template, "<any>")
	}
}

func curlCmdHelpJson(file *descriptor.FileDescriptorProto, name string) string {
	a := []string{}
	msg := getMessageType(file, name)
	if len(msg.Fields) > 0 {
		for _, field := range msg.Fields {
			var jsonString string
			if isFieldMessage(field.FieldDescriptorProto) {
				jsonString = curlCmdHelpJson(file, field.GetTypeName())
			} else {
				jsonString = curlCmdHelpJsonValue(file, field.FieldDescriptorProto)
			}
			a = append(
				a,
				fmt.Sprintf(
					"\"%s\": %s",
					field.GetName(), //field.GetJsonName(),
					jsonString,
				),
			)
		}
	}

	return fmt.Sprintf("{%s}", strings.Join(a, ", "))
}

func cliHelpField(prefix string, field *ggdescriptor.Field) string {
	_, _, cmdHelpName := fieldNfo(prefix, field)

	helpType := ""
	if field.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM {
		if enum, err := getEnumTypeRegistry(field.GetTypeName()); err == nil {
			enumVals := []string{}
			for _, v := range enum.GetValue() {
				enumVals = append(enumVals, v.GetName())
			}
			helpType = strings.Join(enumVals, "|")
		}
	} else {
		helpType = goTypeWithPackage(field.FieldDescriptorProto)
	}

	return fmt.Sprintf("%s [%s]", cmdHelpName, helpType)
}

func cliHelpFields(prefix string, name string) (ret []string) {
	msg, err := getMessageTypeRegistry(name)
	if err != nil {
		return ret
	}

	if len(msg.Fields) > 0 {
		for _, field := range msg.Fields {
			if isFieldMessage(field.FieldDescriptorProto) {
				subRet := cliHelpFields(
					prefix+upperPascalCase(field.GetName())+".",
					field.GetTypeName(),
				)
				for _, l := range subRet {
					ret = append(ret, l)
				}
			} else {
				ret = append(ret, cliHelpField(prefix, field))
			}
		}
	}

	return ret
}

func grpcMethodCliHelp(m *descriptor.MethodDescriptorProto) string {
	attr := ""
	a := cliHelpFields("", m.GetInputType())
	if len(a) > 0 {
		attr = " <" + strings.Join(a, "> <") + ">"
	}
	return fmt.Sprintf(
		"%s%s",
		LowerSnakeCase(m.GetName()),
		attr,
	)
}

func cliCmdHelpString(name string, protoFiles []*descriptor.FileDescriptorProto) string {
	var ret []string
	for _, file := range protoFiles {
		for _, svc := range file.GetService() {
			for _, method := range svc.GetMethod() {
				ret = append(
					ret,
					fmt.Sprintf(
						"%s cli %s",
						name,
						grpcMethodCliHelp(method),
					),
				)
			}
		}
	}

	return fmt.Sprintf("  $ %s", strings.Join(ret, "\n  $ "))
}

func grpcMethodOutputGoType(m *descriptor.MethodDescriptorProto) string {
	if m == nil {
		return ""
	}
	outputType := m.GetOutputType()

	return shortType(outputType)
}

func grpcMethodInputGoType(m *descriptor.MethodDescriptorProto) string {
	if m == nil {
		return ""
	}
	inputType := m.GetInputType()

	return shortType(inputType)
}

func grpcFunctestHttpExtraImport(m *descriptor.MethodDescriptorProto) string {
	if m == nil {
		return ""
	}

	// exclude streamed methods
	if m.GetServerStreaming() || m.GetClientStreaming() {
		return ""
	}

	var ret bytes.Buffer
	httpV := httpVerb(m)
	httpParams := httpPathFields(m)
	switch httpV {
	case "POST", "PUT":
		ret.WriteString("\t\"bytes\"\n")
		if len(httpParams) > 0 {
			ret.WriteString("\t\"encoding/json\"\n")
		}
	}

	return ret.String()
}

func grpcFunctestHttp(m *descriptor.MethodDescriptorProto) string {
	if m == nil {
		return "// gomeet generator nil method descriptor error"
	}

	tplMain := `url := fmt.Sprintf("%s://%s{{ .NormalizedPath }}", proto, serverAddr{{ .NormalizedPathParam }})
	{{ if (or (eq .HttpVerb "POST") (eq .HttpVerb "PUT")) }}
		// Proto to JSON
		ma := jsonpb.Marshaler{}
		sMsg, err := ma.MarshalToString(req)
		if err != nil {
			testCaseResults = append(
				testCaseResults,
				&TestCaseResult{
					req,
					nil,
					fmt.Errorf("{{ .MethodName }}/HTTP {{ .HttpVerb }} error to marshalling the message with %s (%v) - %v", url, err, req),
				},
			)
			continue
		}

		{{ if .RemoveUrlVarFromJsonMsg }}
			// removing fields in the URL
			var raw map[string]interface{}
			json.Unmarshal([]byte(sMsg), &raw)
			{{ .RemoveUrlVarFromJsonMsg -}}

			sMsg2, err := json.Marshal(raw)
			if err != nil {
				testCaseResults = append(
					testCaseResults,
					&TestCaseResult{
						req,
						nil,
						fmt.Errorf("{{ .MethodName }}/HTTP {{ .HttpVerb }} error to unmarshalling the message with %s (%v) - %v", url, err, req),
					},
				)
				continue
			}
			data := bytes.NewBuffer(sMsg2)
		{{ else }}
			data := bytes.NewBufferString(sMsg)
		{{ end }}
		// construct HTTP request
		httpReq, err := http.NewRequest("{{ .HttpVerb }}", url, data)
	{{ else }}
		// construct HTTP request
		httpReq, err := http.NewRequest("{{ .HttpVerb }}", url, nil)
	{{ end -}}
	if err != nil {
		testCaseResults = append(
			testCaseResults,
			&TestCaseResult{
				req,
				nil,
				fmt.Errorf("{{ .MethodName }}/HTTP {{ .HttpVerb }} error to construct the http request with %s (%v) - %v", url, err, req),
			},
		)
		continue
	}
	`

	t, err := template.
		New("functestHttpMain").
		Parse(string(tplMain))
	if err != nil {
		return fmt.Sprintf("// gomeet generator error - %s", err)
	}

	var (
		bPathParams, bRemoveUrlVarFromMsgBlock bytes.Buffer
	)
	httpPath, httpParams := httpPathAndFields(m)
	for _, field := range httpParams {
		httpPath = strings.Replace(httpPath, fmt.Sprintf("{%s}", field.GetName()), "%s", 1)
		bPathParams.WriteString(
			fmt.Sprintf(
				", req.Get%s()",
				upperPascalCase(field.GetName()),
			),
		)
		bRemoveUrlVarFromMsgBlock.WriteString(
			fmt.Sprintf(
				"\t\tdelete(raw, \"%s\") // json_name: %s, name: %s\n",
				field.GetJsonName(),
				field.GetJsonName(),
				field.GetName(),
			),
		)
	}

	vData := struct {
		NormalizedPath          string
		NormalizedPathParam     string
		MethodName              string
		HttpVerb                string
		RemoveUrlVarFromJsonMsg string
	}{
		httpPath,
		bPathParams.String(),
		upperPascalCase(m.GetName()),
		httpVerb(m),
		bRemoveUrlVarFromMsgBlock.String(),
	}

	var out bytes.Buffer
	err = t.Execute(&out, vData)
	if err != nil {
		return fmt.Sprintf("// gomeet generator error - %s", err)
	}

	return out.String()
}

func messageGoType(name string) string {
	msg, err := getMessageTypeRegistry(name)
	if err != nil {
		return ""
	}

	return msg.GoType(msg.File.GoPkg.Name)
}

func messageFake(name string) string {
	msg, err := getMessageTypeRegistry(name)
	if err != nil {
		return ""
	}

	ret := fmt.Sprintf("%s.New%sGomeetFaker()", msg.File.GoPkg.Name, msg.GetName())
	return ret
}
