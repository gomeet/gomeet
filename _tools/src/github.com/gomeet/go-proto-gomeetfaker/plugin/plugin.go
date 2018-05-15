package plugin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/gogo/protobuf/vanity"
	"github.com/gomeet/go-proto-gomeetfaker"
)

var gRCount = uint32(3)

type plugin struct {
	*generator.Generator
	generator.PluginImports
	fmtPkg          generator.Single
	randPkg         generator.Single
	protoPkg        generator.Single
	gomeetfakerPkg  generator.Single
	fakerPkg        generator.Single
	fakerLocalesPkg generator.Single
	stringsPkg      generator.Single
	strconvPkg      generator.Single
	timePkg         generator.Single
	pbTypesPkg      generator.Single
	uuidPkg         generator.Single
	useGogoImport   bool
}

func NewPlugin(useGogoImport bool) generator.Plugin {
	return &plugin{useGogoImport: useGogoImport}
}

func (p *plugin) Name() string {
	return "gomeetfaker"
}

func (p *plugin) Init(g *generator.Generator) {
	p.Generator = g
}

func (p *plugin) Generate(file *generator.FileDescriptor) {
	if !p.useGogoImport {
		vanity.TurnOffGogoImport(file.FileDescriptorProto)
	}
	p.PluginImports = generator.NewPluginImports(p.Generator)
	//p.gomeetfakerPkg = p.NewImport("github.com/gomeet/go-proto-gomeetfaker")
	p.fakerPkg = p.NewImport("github.com/dmgk/faker")
	p.fakerLocalesPkg = p.NewImport("github.com/dmgk/faker/locales")
	p.randPkg = p.NewImport("math/rand")
	p.fmtPkg = p.NewImport("fmt")
	p.stringsPkg = p.NewImport("strings")
	p.strconvPkg = p.NewImport("strconv")
	p.timePkg = p.NewImport("time")
	p.pbTypesPkg = p.NewImport("github.com/golang/protobuf/ptypes")
	p.uuidPkg = p.NewImport("github.com/google/uuid")

	p.generateRandFunc()
	p.generateInitFunc(file)
	for _, msg := range file.Messages() {
		if msg.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		if gogoproto.IsProto3(file.FileDescriptorProto) {
			p.generateProto3Message(file, msg)
		} else {
			p.generateProto2Message(file, msg)
		}

	}
}

func (p *plugin) generateInitFunc(file *generator.FileDescriptor) {
	locales := map[string]string{}
	for _, l := range []string{"de-at", "de-ch", "de", "en-au", "en-bork", "en-ca", "en-gb",
		"en-ind", "en-nep", "en-us", "en-au-ocker", "en", "es", "fa",
		"fr", "it", "ja", "ko", "nb-no", "nl", "pl", "pt-br", "ru",
		"sk", "sv", "vi", "zh-cn", "zh-tw"} {
		parts := strings.Split(l, "-")
		parts[0] = strings.Title(parts[0])
		for i, p := range parts[1:] {
			parts[i+1] = strings.ToUpper(p)
		}
		locales[l] = strings.Join(parts, "_")
	}

	locale := "en"
	if l := getGomeetFakerLocale(file); l != nil {
		lg := strings.ToLower(*l)
		if _, ok := locales[lg]; ok {
			locale = lg
		}
	}
	p.P("func init() {")
	p.In()
	p.P(`GomeetFakerSetLocale("`, locale, `")`)
	//p.P(p.fakerPkg.Use(), ".Locale = ", p.fakerLocalesPkg.Use(), ".", locale)
	p.Out()
	p.P("}")
	p.P("")
	p.P("func GomeetFakerSetLocale(l string) {")
	p.In()
	p.P("switch l {")
	for k, l := range locales {
		p.P(`case "`, k, `":`)
		p.In()
		p.P(p.fakerPkg.Use(), ".Locale = ", p.fakerLocalesPkg.Use(), ".", l)
		p.Out()
	}
	p.P("default:")
	p.In()
	p.P(p.fakerPkg.Use(), ".Locale = ", p.fakerLocalesPkg.Use(), ".En")
	p.Out()
	p.P("}")
	p.Out()
	p.P("}")
}

func (p *plugin) generateRandFunc() {
	p.P("func GomeetFakerRand() *", p.randPkg.Use(), ".Rand {")
	p.In()
	p.P("seed := ", p.timePkg.Use(), ".Now().UnixNano()")
	p.P("return ", p.randPkg.Use(), ".New(", p.randPkg.Use(), ".NewSource(seed))")
	p.Out()
	p.P("}")
}

func getGomeetFakerLocale(file *generator.FileDescriptor) *string {
	if file.Options != nil {
		v, err := proto.GetExtension(file.Options, gomeetfaker.E_Locale)
		if err == nil && v.(*string) != nil {
			return (v.(*string))
		}
	}
	return nil
}

func getFieldFakerRulesIfAny(field *descriptor.FieldDescriptorProto) *gomeetfaker.FieldFakerRules {
	if field.Options != nil {
		v, err := proto.GetExtension(field.Options, gomeetfaker.E_Field)
		if err == nil && v.(*gomeetfaker.FieldFakerRules) != nil {
			return (v.(*gomeetfaker.FieldFakerRules))
		}
	}
	return nil
}

func (p *plugin) isSupportedInt(field *descriptor.FieldDescriptorProto) bool {
	switch *(field.Type) {
	case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_INT64:
		return true
	case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_UINT64:
		return true
	case descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SINT64:
		return true
	}
	return false
}

func (p *plugin) isSupportedFloat(field *descriptor.FieldDescriptorProto) bool {
	switch *(field.Type) {
	case descriptor.FieldDescriptorProto_TYPE_FLOAT, descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return true
	case descriptor.FieldDescriptorProto_TYPE_FIXED32, descriptor.FieldDescriptorProto_TYPE_FIXED64:
		return true
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return true
	}
	return false
}

func (p *plugin) generateProto2Message(file *generator.FileDescriptor, message *generator.Descriptor) {
	ccTypeName := generator.CamelCaseSlice(message.TypeName())

	p.P(`func New`, ccTypeName, `GomeetFaker() *`, ccTypeName, ` {`)
	p.In()
	p.P(`// Gomeetfaker of proto2 fields is unsupported.`)
	p.P(`return nil`)
	p.Out()
	p.P(`}`)
	return
}

func (p *plugin) generateProto3PrintFunc(file *generator.FileDescriptor, message *generator.Descriptor) {
	ccTypeName := generator.CamelCaseSlice(message.TypeName())
	p.P(`func Print`, ccTypeName, `GomeetFaker() {`)
	p.In()
	p.P("a", ccTypeName, ` := New`, ccTypeName, `GomeetFaker()`)
	p.P(`fmt.Println("`, ccTypeName, ` gomeetfaker : ")`)
	p.P(`fmt.Println(a`, ccTypeName, `)`)
	p.Out()
	p.P("}")
}

func (p *plugin) getFuncName(ccTypeName string) string {
	split := strings.Split(ccTypeName, ".")
	if len(split) > 1 {
		return ""
	}
	funcName := "New" + ccTypeName + "GomeetFaker()"
	return funcName
}

func (p *plugin) generateProto3Message(file *generator.FileDescriptor, message *generator.Descriptor) {
	ccTypeName := generator.CamelCaseSlice(message.TypeName())
	//p.P(`func New`, ccTypeName, `GomeetFaker() *`, ccTypeName, ` {`)
	funcName := p.getFuncName(generator.CamelCase(ccTypeName))
	if funcName == "" {
		return
	}
	p.P("func ", funcName, " *", ccTypeName, ` {`)
	p.In()
	p.P(`this := &`, ccTypeName, `{}`)

	oneofs := make(map[string]struct{})
	for _, field := range message.Field {
		fieldFaker := getFieldFakerRulesIfAny(field)
		isOneOf := field.OneofIndex != nil

		fieldName := p.GetOneOfFieldName(message, field)
		fieldVariableName := "this." + fieldName
		variableName := fieldVariableName

		// Golang's proto3 has no concept of unset primitive fields
		if p.fieldIsProto3Map(file, message, field) {
			p.P(`// Gomeetfaker of proto3 map<> fields is unsupported.`)
			continue
		}

		if isOneOf {
			fieldname := p.GetFieldName(message, field)
			if _, ok := oneofs[fieldname]; ok {
				continue
			} else {
				oneofs[fieldname] = struct{}{}
			}
			fieldNumbers := []int32{}
			for _, f := range message.Field {
				fname := p.GetFieldName(message, f)
				if fname == fieldname {
					fieldNumbers = append(fieldNumbers, f.GetNumber())
				}
			}

			p.P(`oneofNumber_`, fieldname, ` := `, fmt.Sprintf("%#v", fieldNumbers), `[GomeetFakerRand().Intn(`, strconv.Itoa(len(fieldNumbers)), `)]`)
			p.P(`switch oneofNumber_`, fieldname, ` {`)
			for _, f := range message.Field {
				fname := p.GetFieldName(message, f)
				if fname != fieldname {
					continue
				}
				ff := getFieldFakerRulesIfAny(f)
				p.P(`case `, strconv.Itoa(int(f.GetNumber())), `:`)
				p.In()
				vname := "aOneOf_" + fname
				skipped := p.generateProto3FieldGomeetFaker("var "+vname, fname, message, f, ff)
				tname := p.OneOfTypeName(message, f)
				p.P("// ", tname)
				if skipped {
					p.P("// this.", fname, " = &", tname, "{aOneOf_", fname, "}")
				} else {
					p.P("this.", fname, " = &", tname, "{aOneOf_", fname, "}")
				}
				p.Out()
			}
			p.P(`}`)
		} else {
			p.generateProto3FieldGomeetFaker(variableName, fieldName, message, field, fieldFaker)
		}
	}
	p.P("return this")
	p.Out()
	p.P("}")
	p.P("")
}

func (p *plugin) generateProto3FieldMessage(variableName, fieldName string, rCount uint32, message *generator.Descriptor, field *descriptor.FieldDescriptorProto) (skipped bool) {
	skipped = false
	desc := p.ObjectNamed(field.GetTypeName())
	typ := p.TypeName(desc)
	funcName := p.getFuncName(generator.CamelCase(typ))
	if funcName != "" {
		if field.IsRepeated() {
			if rCount == 0 {
				rCount = gRCount
			}
			p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
			p.In()
			p.P("aCurrent", fieldName, " := ", funcName)
			p.P(variableName, " = append(", variableName, ", aCurrent", fieldName, ")")
			p.Out()
			p.P("}")
		} else {
			p.P(variableName, " = ", funcName)
		}
	} else {
		goTyp, _ := p.GoType(message, field)
		p.P("// ", variableName, " = ", goTyp, "{} // unknow gomeetfaker function name for type : ", field.GetTypeName())
		skipped = true
	}
	return
}

func (p *plugin) generateProto3FieldGomeetFaker(variableName, fieldName string, message *generator.Descriptor, field *descriptor.FieldDescriptorProto, fieldFaker *gomeetfaker.FieldFakerRules) (skipped bool) {
	skipped = false
	if fieldFaker != nil {
		switch r := fieldFaker.Type.(type) {
		case *gomeetfaker.FieldFakerRules_Skip:
			p.P("// ", variableName, " // skipped by skip rules")
			skipped = true
		case *gomeetfaker.FieldFakerRules_Address:
			p.generateAddressRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_App:
			p.generateAppRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Avatar:
			p.generateAvatarRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Bitcoin:
			p.generateBitcoinRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Business:
			p.generateBusinessRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Code:
			p.generateCodeRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Commerce:
			p.generateCommerceRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Company:
			p.generateCompanyRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Date:
			p.generateDateRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Finance:
			p.generateFinanceRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Hacker:
			p.generateHackerRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Internet:
			p.generateInternetRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Lorem:
			p.generateLoremRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Name:
			p.generateNameRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Number:
			p.generateNumberRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_PhoneNumber:
			p.generatePhoneNumberRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Team:
			p.generateTeamRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Time:
			p.generateTimeRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Value:
			p.generateValueRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Repeated:
			rCount := r.Repeated
			if field.IsMessage() {
				skipped = p.generateProto3FieldMessage(variableName, fieldName, rCount, message, field)
			} else if field.IsString() || field.IsBytes() {
				p.generateStringNoRules(variableName, fieldName, rCount, field)
			} else if p.isSupportedInt(field) || p.isSupportedFloat(field) {
				p.generateNumberNoRules(variableName, fieldName, rCount, field)
			} else if field.IsBool() {
				p.generateBoolNoRules(variableName, fieldName, rCount, field)
			} else {
				p.P("// ", variableName, " // unsupported type")
				skipped = true
			}
		case *gomeetfaker.FieldFakerRules_Uuid:
			p.generateUuidRules(variableName, r, field, fieldName)
		case *gomeetfaker.FieldFakerRules_Enum:
			p.generateEnumRules(variableName, r, message, field, fieldName)
		case nil:
			p.P(`// nil type - ignored`)
		default:
			p.P(fmt.Sprintf("// ", variableName, " unknow gomeetfaker type (%T) (%s)", fieldFaker.Type, r))
		}
	} else if field.IsMessage() {
		skipped = p.generateProto3FieldMessage(variableName, fieldName, gRCount, message, field)
	} else if field.IsString() || field.IsBytes() {
		p.generateStringNoRules(variableName, fieldName, gRCount, field)
	} else if p.isSupportedInt(field) || p.isSupportedFloat(field) {
		p.generateNumberNoRules(variableName, fieldName, gRCount, field)
	} else if field.IsBool() {
		p.generateBoolNoRules(variableName, fieldName, gRCount, field)
	} else if field.IsEnum() {
		p.generateEnumNoRules(variableName, fieldName, gRCount, message, field)
	} else {
		p.P("// ", variableName, " // unsupported type")
		skipped = true
	}
	return
}

func (p *plugin) generateStringNoRules(variableName, fieldName string, rCount uint32, field *descriptor.FieldDescriptorProto) {
	p.P("// ", variableName, " is a string or bytes without gommetfaker rules so faker.Lorem().Lorem() is used")
	loremRules := &gomeetfaker.FieldFakerRules_Lorem{
		Lorem: &gomeetfaker.LoremRules{
			Repeated: &rCount,
			Type: &gomeetfaker.LoremRules_String_{
				String_: true,
			},
		},
	}
	p.generateLoremRules(variableName, loremRules, field, fieldName)
}

func (p *plugin) generateNumberNoRules(variableName, fieldName string, rCount uint32, field *descriptor.FieldDescriptorProto) {
	p.P("// ", variableName, " is a number value without gommetfaker rules so faker.Number().Number(3) is used")
	digits := uint32(3)
	numberRules := &gomeetfaker.FieldFakerRules_Number{
		Number: &gomeetfaker.NumberRules{
			Repeated: &rCount,
			Type: &gomeetfaker.NumberRules_Number{
				Number: &gomeetfaker.NumberRulesDigit{
					Digits: &digits,
				},
			},
		},
	}
	p.generateNumberRules(variableName, numberRules, field, fieldName)
}

func (p *plugin) generateBoolNoRules(variableName, fieldName string, rCount uint32, field *descriptor.FieldDescriptorProto) {
	fieldVariableName := variableName
	af := " = "
	if field.IsRepeated() {
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()

	}
	p.P("// ", variableName, " is a number value without gommetfaker rules so a random true/false is used")
	p.P(variableName, " ", af, " []bool{true, false}[GomeetFakerRand().Intn(2)]")
	if field.IsRepeated() {
		p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateEnumNoRules(variableName, fieldName string, rCount uint32, message *generator.Descriptor, field *descriptor.FieldDescriptorProto) {
	p.P("// ", variableName, " is a string or bytes without gommetfaker rules so faker.Lorem().Lorem() is used")
	enumRules := &gomeetfaker.FieldFakerRules_Enum{
		Enum: &gomeetfaker.EnumRules{
			Repeated: &rCount,
			Type: &gomeetfaker.EnumRules_Random{
				Random: true,
			},
		},
	}
	p.generateEnumRules(variableName, enumRules, message, field, fieldName)
}

func (p *plugin) getEnumValNfo(field *descriptor.FieldDescriptorProto) ([]string, int) {
	enum := p.ObjectNamed(field.GetTypeName()).(*generator.EnumDescriptor)
	l := len(enum.Value)
	values := make([]string, l)
	for i := range enum.Value {
		values[i] = strconv.Itoa(int(*enum.Value[i].Number))
	}
	return values, l
}

func (p *plugin) generateEnumRules(variableName string, r *gomeetfaker.FieldFakerRules_Enum, message *generator.Descriptor, field *descriptor.FieldDescriptorProto, fieldName string) {
	enum := r.Enum
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := enum.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	goTyp, _ := p.GoType(message, field)
	values, l := p.getEnumValNfo(field)
	switch enum.Type.(type) {
	case *gomeetfaker.EnumRules_Random:
		p.P(variableName, " ", af, " ", strings.Join([]string{generator.GoTypeToName(goTyp), "([]int32{", strings.Join(values, ","), "}[GomeetFakerRand().Intn(", fmt.Sprintf("%d", l), ")])"}, ""))
	case *gomeetfaker.EnumRules_RandomNoEmpty:
		values2 := []string{}
		for _, v := range values {
			if v != "0" {
				values2 = append(values2, v)
			}
		}
		l2 := len(values2)
		p.P(variableName, " ", af, " ", strings.Join([]string{generator.GoTypeToName(goTyp), "([]int32{", strings.Join(values2, ","), "}[GomeetFakerRand().Intn(", fmt.Sprintf("%d", l2), ")])"}, ""))
	case *gomeetfaker.EnumRules_First:
		p.P(variableName, " ", af, " ", generator.GoTypeToName(goTyp), "(", values[0], ")")
	case *gomeetfaker.EnumRules_Last:
		p.P(variableName, " ", af, " ", generator.GoTypeToName(goTyp), "(", values[l-1], ")")
	case *gomeetfaker.EnumRules_Index:
		idx := enum.GetIndex()
		if int(idx) >= l {
			p.P("// ", variableName, " ", fmt.Sprintf("%d out of range 0-%d", idx, l-1), " // skipped")
			skipAppend = true
		} else {
			p.P(variableName, " ", af, " ", generator.GoTypeToName(goTyp), "(", values[idx], ")")
		}
	case *gomeetfaker.EnumRules_Value:
		val := fmt.Sprintf("%d", enum.GetValue())
		findIt := false
		for _, v := range values {
			if v == val {
				p.P(variableName, " ", af, " ", generator.GoTypeToName(goTyp), "(", v, ")")
				findIt = true
				break
			}
		}
		if !findIt {
			p.P("// ", variableName, " ", af, " ", generator.GoTypeToName(goTyp), "(", val, ") // skipped isn't a value of enum")
			skipAppend = true
		}

	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateAddressRules(variableName string, r *gomeetfaker.FieldFakerRules_Address, field *descriptor.FieldDescriptorProto, fieldName string) {
	addr := r.Address
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := addr.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := addr.Type.(type) {
	case *gomeetfaker.AddressRules_City:
		if addr.GetCity() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().City()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().City())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().City() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().City() // skipped [(gomeetfaker.field).address.city = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_StreetName:
		if addr.GetStreetName() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StreetName()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().StreetName())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StreetName() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StreetName() // skipped [(gomeetfaker.field).address.street_name = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_StreetAddress:
		if addr.GetStreetAddress() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StreetAddress()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().StreetAddress())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StreetAddress() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StreetAddress() // skipped [(gomeetfaker.field).address.street_address = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_SecondaryAddress:
		if addr.GetSecondaryAddress() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().SecondaryAddress()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().SecondaryAddress())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().SecondaryAddress() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().SecondaryAddress() // skipped [(gomeetfaker.field).address.street_address = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_BuildingNumber:
		if addr.GetBuildingNumber() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().BuildingNumber()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().BuildingNumber())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().BuildingNumber() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().BuildingNumber() // skipped [(gomeetfaker.field).address.building_number = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_Postcode:
		if addr.GetPostcode() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Postcode()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().Postcode())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Postcode() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Postcode() // skipped [(gomeetfaker.field).address.postcode = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_ZipCode:
		if addr.GetZipCode() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().ZipCode()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().ZipCode())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().ZipCode() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().ZipCode() // skipped [(gomeetfaker.field).address.zip_code = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_TimeZone:
		if addr.GetTimeZone() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().TimeZone()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().TimeZone())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().TimeZone() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().TimeZone() // skipped [(gomeetfaker.field).address.time_zone = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_CityPrefix:
		if addr.GetCityPrefix() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().CityPrefix()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().CityPrefix())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().CityPrefix() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().CityPrefix() // skipped [(gomeetfaker.field).address.city_prefix = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_CitySuffix:
		if addr.GetCitySuffix() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().CitySuffix()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().CitySuffix())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().CitySuffix() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().CitySuffix() // skipped [(gomeetfaker.field).address.city_suffix = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_StreetSuffix:
		if addr.GetStreetSuffix() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StreetSuffix()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().StreetSuffix())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StreetSuffix() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StreetSuffix() // skipped [(gomeetfaker.field).address.street_suffix = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_State:
		if addr.GetState() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().State()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().State())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().State() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().State() // skipped [(gomeetfaker.field).address.state = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_StateAbbr:
		if addr.GetStateAbbr() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StateAbbr()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().StateAbbr())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StateAbbr() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().StateAbbr() // skipped [(gomeetfaker.field).address.state_abbr = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_Country:
		if addr.GetCountry() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Country()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().Country())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Country() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Country() // skipped [(gomeetfaker.field).address.country = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_CountryCode:
		if addr.GetCountryCode() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().CountryCode()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().CountryCode())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().CountryCode() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().CountryCode() // skipped [(gomeetfaker.field).address.country_code = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_Latitude:
		if addr.GetLatitude() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_FLOAT:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Latitude()")
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
				p.P(variableName, " ", af, " float64(", p.fakerPkg.Use(), ".Address().Latitude())")
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.strconvPkg.Use(), ".FormatFloat(float64(", p.fakerPkg.Use(), ".Address().Latitude()), 'f', 5, 32)")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.strconvPkg.Use(), ".FormatFloat(float64(", p.fakerPkg.Use(), ".Address().Latitude()), 'f', 5, 32))")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Latitude() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Latitude() // skipped [(gomeetfaker.field).address.latitude = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_Longitude:
		if addr.GetLongitude() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_FLOAT:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Longitude()")
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
				p.P(variableName, " ", af, " float64(", p.fakerPkg.Use(), ".Address().Longitude())")
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.strconvPkg.Use(), ".FormatFloat(float64(", p.fakerPkg.Use(), ".Address().Longitude()), 'f', 5, 32)")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.strconvPkg.Use(), ".FormatFloat(float64(", p.fakerPkg.Use(), ".Address().Longitude()), 'f', 5, 32))")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Longitude() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Longitude() // skipped [(gomeetfaker.field).address.longitude = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_String_:
		if addr.GetString_() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().String()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Address().String())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().String() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().String() // skipped [(gomeetfaker.field).address.string = false]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_PostcodeByState:
		params := addr.GetPostcodeByState()
		if params != nil {
			state := params.GetState()
			if state != "" {
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), `.Address().PostcodeByState("`, state, `")`)
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), `.Address().PostcodeByState("`, state, `"))`)
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Address().PostcodeByState("`, state, `") // bad type convertion`, fmt.Sprintf("%T", field))
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Address().PostcodeByState() // skipped [(gomeetfaker.field).address.postcode_by_state = {state: ""}]`)
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().PostcodeByState() // skipped [(gomeetfaker.field).address.postcode_by_state = {}]")
			skipAppend = true
		}
	case *gomeetfaker.AddressRules_ZipCodeByState:
		params := addr.GetZipCodeByState()
		if params != nil {
			state := params.GetState()
			if state != "" {
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), `.Address().ZipCodeByState("`, state, `")`)
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), `.Address().ZipCodeByState("`, state, `"))`)
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Address().ZipCodeByState("`, state, `") // bad type convertion`, fmt.Sprintf("%T", field))
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Address().ZipCodeByState() // skipped [(gomeetfaker.field).address.zip_code_by_state = {state: ""}]`)
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().ZipCodeByState() // skipped [(gomeetfaker.field).address.zip_code_by_state = {}]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Address().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateAppRules(variableName string, r *gomeetfaker.FieldFakerRules_App, field *descriptor.FieldDescriptorProto, fieldName string) {
	app := r.App
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := app.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := app.Type.(type) {
	case *gomeetfaker.AppRules_Name:
		if app.GetName() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".App().Name()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".App().Name())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".App().Name() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".App().Name() // skipped [(gomeetfaker.field).app.name = false]")
			skipAppend = true
		}
	case *gomeetfaker.AppRules_Version:
		if app.GetVersion() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".App().Version()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".App().Version())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".App().Version() // bad type convertion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".App().Version() // skipped [(gomeetfaker.field).app.version = false]")
			skipAppend = true
		}
	case *gomeetfaker.AppRules_Author:
		if app.GetAuthor() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".App().Author()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".App().Author())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".App().Author() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".App().Author() // skipped [(gomeetfaker.field).app.author = false]")
			skipAppend = true
		}
	case *gomeetfaker.AppRules_String_:
		if app.GetString_() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".App().String()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".App().String())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".App().String() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".App().String() // skipped [(gomeetfaker.field).app.string = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".App().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateAvatarRules(variableName string, r *gomeetfaker.FieldFakerRules_Avatar, field *descriptor.FieldDescriptorProto, fieldName string) {
	avatar := r.Avatar
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := avatar.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := avatar.Type.(type) {
	case *gomeetfaker.AvatarRules_String_:
		if avatar.GetString_() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Avatar().String()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Avatar().String())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Avatar().String() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Avatar().String() // skipped [(gomeetfaker.field).avatar.string = false]")
			skipAppend = true
		}
	case *gomeetfaker.AvatarRules_Url:
		params := avatar.GetUrl()
		if params != nil {
			arg := fmt.Sprintf(`"%s", %d, %d`, params.GetFormat(), params.GetWidth(), params.GetHeight())
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Avatar().Url(", arg, ")")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Avatar().Url(", arg, "))")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Avatar().Url(", arg, ") // bad type convertion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Avatar().Url(...) // skipped [(gomeetfaker.field).avatar.url = {}]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Avatar().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateBitcoinRules(variableName string, r *gomeetfaker.FieldFakerRules_Bitcoin, field *descriptor.FieldDescriptorProto, fieldName string) {
	bitcoin := r.Bitcoin
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := bitcoin.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := bitcoin.Type.(type) {
	case *gomeetfaker.BitcoinRules_Address:
		if bitcoin.GetAddress() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Bitcoin().Address()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Bitcoin().Address())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Bitcoin().Address() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Bitcoin().Address() // skipped [(gomeetfaker.field).bitcoin.address = false]")
			skipAppend = true
		}
	case *gomeetfaker.BitcoinRules_String_:
		if bitcoin.GetString_() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Bitcoin().String()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Bitcoin().String())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Bitcoin().String() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Bitcoin().String() // skipped [(gomeetfaker.field).bitcoin.string = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Bitcoin().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateBusinessRules(variableName string, r *gomeetfaker.FieldFakerRules_Business, field *descriptor.FieldDescriptorProto, fieldName string) {
	business := r.Business
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := business.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := business.Type.(type) {
	case *gomeetfaker.BusinessRules_CreditCardNumber:
		if business.GetCreditCardNumber() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Business().CreditCardNumber()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Business().CreditCardNumber())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Business().CreditCardNumber() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Business().CreditCardNumber() // skipped [(gomeetfaker.field).business.credit_card_number = false]")
			skipAppend = true
		}
	case *gomeetfaker.BusinessRules_CreditCardExpiryDate:
		if business.GetCreditCardExpiryDate() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Business().CreditCardExpiryDate()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Business().CreditCardExpiryDate())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Business().CreditCardExpiryDate() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Business().CreditCardExpiryDate() // skipped [(gomeetfaker.field).business.credit_card_expiry_date = false]")
			skipAppend = true
		}
	case *gomeetfaker.BusinessRules_CreditCardType:
		if business.GetCreditCardType() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Business().CreditCardType()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Business().CreditCardType())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Business().CreditCardType() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Business().CreditCardType() // skipped [(gomeetfaker.field).business.credit_card_type = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Business().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateCodeRules(variableName string, r *gomeetfaker.FieldFakerRules_Code, field *descriptor.FieldDescriptorProto, fieldName string) {
	code := r.Code
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := code.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := code.Type.(type) {
	case *gomeetfaker.CodeRules_Isbn10:
		if code.GetIsbn10() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Isbn10()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Code().Isbn10())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Isbn10() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Isbn10() // skipped [(gomeetfaker.field).code.isbn10 = false]")
			skipAppend = true
		}
	case *gomeetfaker.CodeRules_Isbn13:
		if code.GetIsbn13() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Isbn13()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Code().Isbn13())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Isbn13() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Isbn13() // skipped [(gomeetfaker.field).code.isbn13 = false]")
			skipAppend = true
		}
	case *gomeetfaker.CodeRules_Ean13:
		if code.GetEan13() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Ean13()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Code().Ean13())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Ean13() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Ean13() // skipped [(gomeetfaker.field).code.ean13 = false]")
			skipAppend = true
		}
	case *gomeetfaker.CodeRules_Ean8:
		if code.GetEan8() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Ean8()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Code().Ean8())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Ean8() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Ean8() // skipped [(gomeetfaker.field).code.ean8 = false]")
			skipAppend = true
		}
	case *gomeetfaker.CodeRules_Rut:
		if code.GetRut() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Rut()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Code().Rut())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Rut() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Rut() // skipped [(gomeetfaker.field).code.rut = false]")
			skipAppend = true
		}
	case *gomeetfaker.CodeRules_Abn:
		if code.GetAbn() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Abn()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Code().Abn())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Abn() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Abn() // skipped [(gomeetfaker.field).code.abn = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Code().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateCommerceRules(variableName string, r *gomeetfaker.FieldFakerRules_Commerce, field *descriptor.FieldDescriptorProto, fieldName string) {
	commerce := r.Commerce
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := commerce.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := commerce.Type.(type) {
	case *gomeetfaker.CommerceRules_Color:
		if commerce.GetColor() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().Color()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Commerce().Color())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().Color() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().Color() // skipped [(gomeetfaker.field).commerce.color = false]")
			skipAppend = true
		}
	case *gomeetfaker.CommerceRules_Department:
		if commerce.GetDepartment() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().Department()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Commerce().Department())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().Department() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().Department() // skipped [(gomeetfaker.field).commerce.department = false]")
			skipAppend = true
		}
	case *gomeetfaker.CommerceRules_ProductName:
		if commerce.GetProductName() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().ProductName()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Commerce().ProductName())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().ProductName() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().ProductName() // skipped [(gomeetfaker.field).commerce.product_name = false]")
			skipAppend = true
		}
	case *gomeetfaker.CommerceRules_Price:
		if commerce.GetPrice() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_FLOAT:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().Price()")
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
				p.P(variableName, " ", af, " float64(", p.fakerPkg.Use(), ".Commerce().Price())")
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.strconvPkg.Use(), ".FormatFloat(float64(", p.fakerPkg.Use(), ".Commerce().Price()), 'f', 2, 32)")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.strconvPkg.Use(), ".FormatFloat(float64(", p.fakerPkg.Use(), ".Commerce().Price()), 'f', 2, 32))")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().Price() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().Price() // skipped [(gomeetfaker.field).commerce.price = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Commerce().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateCompanyRules(variableName string, r *gomeetfaker.FieldFakerRules_Company, field *descriptor.FieldDescriptorProto, fieldName string) {
	company := r.Company
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := company.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := company.Type.(type) {
	case *gomeetfaker.CompanyRules_Name:
		if company.GetName() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Name()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Company().Name())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Name() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Name() // skipped [(gomeetfaker.field).company.name = false]")
			skipAppend = true
		}
	case *gomeetfaker.CompanyRules_Suffix:
		if company.GetSuffix() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Suffix()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Company().Suffix())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Suffix() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Suffix() // skipped [(gomeetfaker.field).company.suffix = false]")
			skipAppend = true
		}
	case *gomeetfaker.CompanyRules_CatchPhrase:
		if company.GetCatchPhrase() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().CatchPhrase()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Company().CatchPhrase())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().CatchPhrase() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().CatchPhrase() // skipped [(gomeetfaker.field).company.catch_phrase = false]")
			skipAppend = true
		}
	case *gomeetfaker.CompanyRules_Bs:
		if company.GetBs() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Bs()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Company().Bs())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Bs() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Bs() // skipped [(gomeetfaker.field).company.bs = false]")
			skipAppend = true
		}
	case *gomeetfaker.CompanyRules_Ein:
		if company.GetEin() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Ein()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Company().Ein())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Ein() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Ein() // skipped [(gomeetfaker.field).company.ein = false]")
			skipAppend = true
		}
	case *gomeetfaker.CompanyRules_DunsNumber:
		if company.GetDunsNumber() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().DunsNumber()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Company().DunsNumber())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().DunsNumber() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().DunsNumber() // skipped [(gomeetfaker.field).company.duns_number = false]")
			skipAppend = true
		}
	case *gomeetfaker.CompanyRules_Logo:
		if company.GetLogo() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Logo()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Company().Logo())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Logo() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Logo() // skipped [(gomeetfaker.field).company.logo = false]")
			skipAppend = true
		}
	case *gomeetfaker.CompanyRules_String_:
		if company.GetString_() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().String()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Company().String())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().String() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().String() // skipped [(gomeetfaker.field).company.string = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Company().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateHackerRules(variableName string, r *gomeetfaker.FieldFakerRules_Hacker, field *descriptor.FieldDescriptorProto, fieldName string) {
	hacker := r.Hacker
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := hacker.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := hacker.Type.(type) {
	case *gomeetfaker.HackerRules_SaySomethingSmart:
		if hacker.GetSaySomethingSmart() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().SaySomethingSmart()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Hacker().SaySomethingSmart())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().SaySomethingSmart() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().SaySomethingSmart() // skipped [(gomeetfaker.field).hacker.say_something_smart = false]")
			skipAppend = true
		}
	case *gomeetfaker.HackerRules_Abbreviation:
		if hacker.GetAbbreviation() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Abbreviation()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Hacker().Abbreviation())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Abbreviation() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Abbreviation() // skipped [(gomeetfaker.field).hacker.abbreviation = false]")
			skipAppend = true
		}
	case *gomeetfaker.HackerRules_Adjective:
		if hacker.GetAdjective() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Adjective()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Hacker().Adjective())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Adjective() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Adjective() // skipped [(gomeetfaker.field).hacker.adjective = false]")
			skipAppend = true
		}
	case *gomeetfaker.HackerRules_Noun:
		if hacker.GetNoun() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Noun()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Hacker().Noun())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Noun() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Noun() // skipped [(gomeetfaker.field).hacker.noun = false]")
			skipAppend = true
		}
	case *gomeetfaker.HackerRules_Verb:
		if hacker.GetVerb() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Verb()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Hacker().Verb())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Verb() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Verb() // skipped [(gomeetfaker.field).hacker.verb = false]")
			skipAppend = true
		}
	case *gomeetfaker.HackerRules_IngVerb:
		if hacker.GetIngVerb() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().IngVerb()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Hacker().IngVerb())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().IngVerb() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().IngVerb() // skipped [(gomeetfaker.field).hacker.ing_verb = false]")
			skipAppend = true
		}
	case *gomeetfaker.HackerRules_Phrases:
		if hacker.GetPhrases() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), `.Hacker().Phrases(), " ")`)
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), `.Hacker().Phrases(), " "))`)
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Phrases() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Phrases() // skipped [(gomeetfaker.field).hacker.phrases = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Hacker().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateInternetRules(variableName string, r *gomeetfaker.FieldFakerRules_Internet, field *descriptor.FieldDescriptorProto, fieldName string) {
	internet := r.Internet
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := internet.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := internet.Type.(type) {
	case *gomeetfaker.InternetRules_Email:
		if internet.GetEmail() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Email()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().Email())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Email() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Email() // skipped [(gomeetfaker.field).internet.email = false]")
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_FreeEmail:
		if internet.GetFreeEmail() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().FreeEmail()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().FreeEmail())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().FreeEmail() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().FreeEmail() // skipped [(gomeetfaker.field).internet.free_email = false]")
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_SafeEmail:
		if internet.GetSafeEmail() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().SafeEmail()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().SafeEmail())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().SafeEmail() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().SafeEmail() // skipped [(gomeetfaker.field).internet.safe_email = false]")
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_UserName:
		if internet.GetUserName() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().UserName()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().UserName())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().UserName() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().UserName() // skipped [(gomeetfaker.field).internet.user_name = false]")
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_Password:
		params := internet.GetPassword()
		if params != nil {
			min, max := params.GetMin(), params.GetMax()
			if max < min {
				min, max = max, min
			}
			arg := fmt.Sprintf(`%d, %d`, min, max)
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Password(", arg, ")")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().Password(", arg, "))")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Password(", arg, ") // bad type convertion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Internet().Password(min, max) // skipped [(gomeetfaker.field).internet.password = {}]`)
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_DomainName:
		if internet.GetDomainName() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().DomainName()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().DomainName())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().DomainName() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().DomainName() // skipped [(gomeetfaker.field).internet.domain_name = false]")
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_DomainWord:
		if internet.GetDomainWord() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().DomainWord()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().DomainWord())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().DomainWord() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().DomainWord() // skipped [(gomeetfaker.field).internet.domain_word = false]")
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_DomainSuffix:
		if internet.GetDomainSuffix() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().DomainSuffix()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().DomainSuffix())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().DomainSuffix() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().DomainSuffix() // skipped [(gomeetfaker.field).internet.domain_suffix = false]")
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_MacAddress:
		if internet.GetMacAddress() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().MacAddress()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().MacAddress())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().MacAddress() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().MacAddress() // skipped [(gomeetfaker.field).internet.mac_address = false]")
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_Ipv4Address:
		if internet.GetIpv4Address() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().IpV4Address()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().IpV4Address())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().IpV4Address() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().IpV4Address() // skipped [(gomeetfaker.field).internet.ipv4_address = false]")
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_Ipv6Address:
		if internet.GetIpv6Address() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().IpV6Address()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().IpV6Address())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().IpV6Address() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().IpV6Address() // skipped [(gomeetfaker.field).internet.ipv6_address = false]")
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_Url:
		if internet.GetUrl() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Url()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().Url())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Url() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Url() // skipped [(gomeetfaker.field).internet.url = false]")
			skipAppend = true
		}
	case *gomeetfaker.InternetRules_Slug:
		if internet.GetSlug() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Slug()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Internet().Slug())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Slug() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Slug() // skipped [(gomeetfaker.field).internet.slug = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Internet().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateNameRules(variableName string, r *gomeetfaker.FieldFakerRules_Name, field *descriptor.FieldDescriptorProto, fieldName string) {
	name := r.Name
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := name.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := name.Type.(type) {
	case *gomeetfaker.NameRules_Name:
		if name.GetName() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Name()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Name().Name())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Name() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Name() // skipped [(gomeetfaker.field).name.name = false]")
			skipAppend = true
		}
	case *gomeetfaker.NameRules_FirstName:
		if name.GetFirstName() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().FirstName()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Name().FirstName())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().FirstName() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().FirstName() // skipped [(gomeetfaker.field).name.first_name = false]")
			skipAppend = true
		}
	case *gomeetfaker.NameRules_LastName:
		if name.GetLastName() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().LastName()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Name().LastName())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().LastName() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().LastName() // skipped [(gomeetfaker.field).name.last_name = false]")
			skipAppend = true
		}
	case *gomeetfaker.NameRules_Prefix:
		if name.GetPrefix() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Prefix()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Name().Prefix())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Prefix() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Prefix() // skipped [(gomeetfaker.field).name.prefix = false]")
			skipAppend = true
		}
	case *gomeetfaker.NameRules_Suffix:
		if name.GetSuffix() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Suffix()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Name().Suffix())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Suffix() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Suffix() // skipped [(gomeetfaker.field).name.suffix = false]")
			skipAppend = true
		}
	case *gomeetfaker.NameRules_Title:
		if name.GetTitle() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Title()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Name().Title())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Title() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Title() // skipped [(gomeetfaker.field).name.title = false]")
			skipAppend = true
		}
	case *gomeetfaker.NameRules_String_:
		if name.GetString_() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().String()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Name().String())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().String() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().String() // skipped [(gomeetfaker.field).name.string = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Name().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateFinanceRules(variableName string, r *gomeetfaker.FieldFakerRules_Finance, field *descriptor.FieldDescriptorProto, fieldName string) {
	finance := r.Finance
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := finance.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := finance.Type.(type) {
	case *gomeetfaker.FinanceRules_CreditCard:
		params := finance.GetCreditCard()
		if params != nil {
			ccType := strings.ToLower(strings.TrimSpace(params.GetType()))
			switch ccType {
			case "visa", "mastercard", "american_express", "diners_club",
				"discover", "maestro", "switch", "solo", "forbrugsforeningen",
				"dankort", "laser":
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), `.Finance().CreditCard("`, ccType, `")`)
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), `.Finance().CreditCard("`, ccType, `"))`)
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Finance().CreditCard("`, ccType, `") // bad type convertion`, fmt.Sprintf("%T", field))
					skipAppend = true
				}
			case "":
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Finance().CreditCard("") // skipped [(gomeetfaker.field).finance.credit_card = {type: ""}]`)
				skipAppend = true
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Finance().CreditCard("`, ccType, `") // skipped unknow credit_card type [(gomeetfaker.field).finance.credit_card = {type: "`, ccType, `"}]`)
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Finance().CreditCard(type) // skipped [(gomeetfaker.field).address.postcode_by_state = {}]`)
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Finance().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateLoremRules(variableName string, r *gomeetfaker.FieldFakerRules_Lorem, field *descriptor.FieldDescriptorProto, fieldName string) {
	lorem := r.Lorem
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := lorem.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := lorem.Type.(type) {
	case *gomeetfaker.LoremRules_Character:
		if lorem.GetCharacter() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Character()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Lorem().Character())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Character() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Character() // skipped [(gomeetfaker.field).lorem.character = false]")
			skipAppend = true
		}
	case *gomeetfaker.LoremRules_Characters:
		params := lorem.GetCharacters()
		if params != nil {
			num := params.GetNum()
			if num > 0 {
				arg := fmt.Sprintf("%d", num)
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Characters(", arg, ")")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Lorem().Characters(", arg, "))")
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Characters(", arg, ") // bad type convertion ", fmt.Sprintf("%T", field))
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Lorem().Characters() // skipped [(gomeetfaker.field).lorem.characters = {num: 0}]`)
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Characters() // skipped [(gomeetfaker.field).address.postcode_by_state = {}]")
			skipAppend = true
		}
	case *gomeetfaker.LoremRules_Word:
		if lorem.GetWord() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Word()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Lorem().Word())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Word() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Word() // skipped [(gomeetfaker.field).lorem.word = false]")
			skipAppend = true
		}
	case *gomeetfaker.LoremRules_Words:
		params := lorem.GetWords()
		if params != nil {
			num := params.GetNum()
			if num > 0 {
				arg := fmt.Sprintf("%d", num)
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), ".Lorem().Words(", arg, `), " ")`)
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), ".Lorem().Words(", arg, `), " "))`)
				default:
					p.P("// ", variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), ".Lorem().Words(", arg, `), " ") // bad type convertion `, fmt.Sprintf("%T", field))
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), `.Lorem().Words(...), " ") // skipped [(gomeetfaker.field).lorem.words = {num: 0}]`)
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), `.Lorem().Words(...), " ") // skipped [(gomeetfaker.field).loream.words = {}]`)
			skipAppend = true
		}
	case *gomeetfaker.LoremRules_Sentence:
		params := lorem.GetSentence()
		if params != nil {
			words := params.GetWords()
			if words > 0 {
				arg := fmt.Sprintf("%d", words)
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Sentence(", arg, ")")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Lorem().Sentence(", arg, "))")
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Sentence(", arg, ") // bad type convertion ", fmt.Sprintf("%T", field))
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Lorem().Sentence() // skipped [(gomeetfaker.field).lorem.sentence = {words: 0}]`)
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Sentence() // skipped [(gomeetfaker.field).loream.sentence = {}]")
			skipAppend = true
		}
	case *gomeetfaker.LoremRules_Sentences:
		params := lorem.GetSentences()
		if params != nil {
			num := params.GetNum()
			if num > 0 {
				arg := fmt.Sprintf("%d", num)
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), ".Lorem().Sentences(", arg, `), " ")`)
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), ".Lorem().Sentences(", arg, `), " "))`)
				default:
					p.P("// ", variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), ".Lorem().Sentences(", arg, `), " ") // bad type convertion `, fmt.Sprintf("%T", field))
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), `.Lorem().Sentences(...), " ") // skipped [(gomeetfaker.field).lorem.sentences = {num: 0}]`)
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), `.Lorem().Sentences(...), " ") // skipped [(gomeetfaker.field).loream.sentences = {}]`)
			skipAppend = true
		}
	case *gomeetfaker.LoremRules_Paragraph:
		params := lorem.GetParagraph()
		if params != nil {
			sentence := params.GetSentence()
			if sentence > 0 {
				arg := fmt.Sprintf("%d", sentence)
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Paragraph(", arg, ")")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Lorem().Paragraph(", arg, "))")
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Paragraph(", arg, ") // bad type convertion ", fmt.Sprintf("%T", field))
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), `.Lorem().Paragraph() // skipped [(gomeetfaker.field).lorem.paragraph = {sentence: 0}]`)
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Paragraph() // skipped [(gomeetfaker.field).loream.paragraph = {}]")
			skipAppend = true
		}
	case *gomeetfaker.LoremRules_Paragraphs:
		params := lorem.GetParagraphs()
		if params != nil {
			num := params.GetNum()
			if num > 0 {
				arg := fmt.Sprintf("%d", num)
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), ".Lorem().Paragraphs(", arg, `), " ")`)
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), ".Lorem().Paragraphs(", arg, `), " "))`)
				default:
					p.P("// ", variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), ".Lorem().Paragraphs(", arg, `), " ") // bad type convertion `, fmt.Sprintf("%T", field))
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), `.Lorem().Paragraphs(...), " ") // skipped [(gomeetfaker.field).lorem.paragraphs = {num: 0}]`)
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.stringsPkg.Use(), ".Join(", p.fakerPkg.Use(), `.Lorem().Paragraphs(...), " ") // skipped [(gomeetfaker.field).loream.paragraphs = {}]`)
			skipAppend = true
		}
	case *gomeetfaker.LoremRules_String_:
		if lorem.GetString_() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().String()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Lorem().String())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().String() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().String() // skipped [(gomeetfaker.field).lorem.string = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Lorem().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateTeamRules(variableName string, r *gomeetfaker.FieldFakerRules_Team, field *descriptor.FieldDescriptorProto, fieldName string) {
	team := r.Team
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := team.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := team.Type.(type) {
	case *gomeetfaker.TeamRules_Name:
		if team.GetName() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().Name()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Team().Name())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().Name() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().Name() // skipped [(gomeetfaker.field).team.name = false]")
			skipAppend = true
		}
	case *gomeetfaker.TeamRules_Creature:
		if team.GetCreature() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().Creature()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Team().Creature())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().Creature() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().Creature() // skipped [(gomeetfaker.field).team.creature = false]")
			skipAppend = true
		}
	case *gomeetfaker.TeamRules_State:
		if team.GetState() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().State()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Team().State())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().State() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().State() // skipped [(gomeetfaker.field).team.state = false]")
			skipAppend = true
		}
	case *gomeetfaker.TeamRules_String_:
		if team.GetString_() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().String()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Team().String())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().String() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().String() // skipped [(gomeetfaker.field).team.string = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Team().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateDateRules(variableName string, r *gomeetfaker.FieldFakerRules_Date, field *descriptor.FieldDescriptorProto, fieldName string) {
	date := r.Date
	fieldVariableName := variableName
	af := " = "
	if field.IsRepeated() {
		rCount := date.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := date.Type.(type) {
	case *gomeetfaker.DateRules_Between:
		params := date.GetBetween()
		format := date.GetFormat()
		if format == "" {
			format = "2006-01-02 15:04:05"
		}
		from, to := params.GetFrom(), params.GetTo()
		if _, err := time.Parse(format, from); err == nil {
			if _, err := time.Parse(format, to); err == nil {
				fT := *(field.Type)
				switch fT {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P("if t1, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, from, `"); err == nil {`)
					p.In()
					p.P("if t2, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, to, `"); err == nil {`)
					p.In()
					p.P("a", fieldName, "Date := ", p.fakerPkg.Use(), ".Date().Between(t1, t2)")
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", a", fieldName, `Date.Format("`, format, `"))`)
					} else {
						p.P(variableName, " ", af, " a", fieldName, `Date.Format("`, format, `")`)
					}
					p.Out()
					p.P("}")
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P("if t1, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, from, `"); err == nil {`)
					p.In()
					p.P("if t2, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, to, `"); err == nil {`)
					p.In()
					p.P("a", fieldName, "Date := ", p.fakerPkg.Use(), ".Date().Between(t1, t2)")
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", []byte(a", fieldName, `Date.Format("`, format, `")))`)
					} else {
						p.P(variableName, " ", af, " []byte(a", fieldName, `Date.Format("`, format, `"))`)
					}
					p.Out()
					p.P("}")
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
					tName := field.GetTypeName()
					switch tName {
					case ".google.protobuf.Timestamp":
						p.P("if t1, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, from, `"); err == nil {`)
						p.In()
						p.P("if t2, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, to, `"); err == nil {`)
						p.In()
						p.P("if t, err := ", p.pbTypesPkg.Use(), ".TimestampProto(", p.fakerPkg.Use(), ".Date().Between(t1, t2)); err == nil {")
						p.In()
						if field.IsRepeated() {
							p.P(fieldVariableName, " = append(", fieldVariableName, ", t)")
						} else {
							p.P(variableName, " ", af, " t")
						}
						p.Out()
						p.P("}")
						p.Out()
						p.P("}")
						p.Out()
						p.P("}")
					default:
						p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Date().Between(...) // bad type conversion unknow message type ", tName)
					}
				default:
					p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Date().Between(...) // bad type conversion", fmt.Sprintf("%s", fT))
				}
			} else {
				p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Date().Between(...) // bad type conversion ", err.Error())
			}
		} else {
			p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Date().Between(...) // bad type conversion ", err.Error())
		}
	case *gomeetfaker.DateRules_Forward:
		params := date.GetForward()
		format := date.GetFormat()
		if format == "" {
			format = "2006-01-02 15:04:05"
		}
		if params != "" {
			if duration, err := time.ParseDuration(params); err == nil {
				if duration < 0 {
					p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), `.Date().Forward(...) // bad duration time.ParseDuration("`, params, `") < 0`)
				} else {
					fT := *(field.Type)
					switch fT {
					case descriptor.FieldDescriptorProto_TYPE_STRING:
						p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
						p.In()
						p.P("a", fieldName, "Date := ", p.fakerPkg.Use(), ".Date().Forward(a", fieldName, "Duration)")
						if field.IsRepeated() {
							p.P(fieldVariableName, " = append(", fieldVariableName, ", a", fieldName, `Date.Format("`, format, `"))`)
						} else {
							p.P(variableName, " ", af, " a", fieldName, `Date.Format("`, format, `")`)
						}
						p.Out()
						p.P("}")
					case descriptor.FieldDescriptorProto_TYPE_BYTES:
						p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
						p.In()
						p.P("a", fieldName, "Date := ", p.fakerPkg.Use(), ".Date().Forward(a", fieldName, "Duration)")
						if field.IsRepeated() {
							p.P(fieldVariableName, " = append(", fieldVariableName, ", []byte(a", fieldName, `Date.Format("`, format, `")))`)
						} else {
							p.P(variableName, " ", af, " []byte(a", fieldName, `Date.Format("`, format, `"))`)
						}
						p.Out()
						p.P("}")
					case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
						tName := field.GetTypeName()
						switch tName {
						case ".google.protobuf.Timestamp":
							p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
							p.In()
							p.P("if t, err := ", p.pbTypesPkg.Use(), ".TimestampProto(", p.fakerPkg.Use(), ".Date().Forward(a", fieldName, "Duration)); err == nil {")
							p.In()
							if field.IsRepeated() {
								p.P(fieldVariableName, " = append(", fieldVariableName, ", t)")
							} else {
								p.P(variableName, " ", af, " t")
							}
							p.Out()
							p.P("}")
							p.Out()
							p.P("}")
						default:
							p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Date().Forward(...) // bad type conversion unknow message type ", tName)
						}
					default:
						p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Date().Forward(...) // bad type conversion", fmt.Sprintf("%s", fT))
					}
				}
			} else {
				p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Date().Forward(...) // bad duration ", err.Error())
			}
		} else {
			p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), `.Date().Forward(...) // skipped [(gomeetfaker.field).date.forward = ""]`)
		}
	case *gomeetfaker.DateRules_Birthday:
		params := date.GetBirthday()
		if params != nil {
			min, max := params.GetMin(), params.GetMax()
			if min > max {
				min, max = max, min
			}
			sMin, sMax := strconv.FormatInt(int64(min), 10), strconv.FormatInt(int64(max), 10)
			format := date.GetFormat()
			if format == "" {
				format = "2006-01-02"
			}
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P("a", fieldName, "Date := ", p.fakerPkg.Use(), ".Date().Birthday(", sMin, ", ", sMax, ")")
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", a", fieldName, `Date.Format("`, format, `"))`)
				} else {
					p.P(variableName, " ", af, " a", fieldName, `Date.Format("`, format, `")`)
				}
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P("a", fieldName, "Date := ", p.fakerPkg.Use(), ".Date().Birthday(", sMin, ", ", sMax, ")")
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", []byte(a", fieldName, `Date.Format("`, format, `")))`)
				} else {
					p.P(variableName, " ", af, " []byte(a", fieldName, `Date.Format("`, format, `"))`)
				}
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				tName := field.GetTypeName()
				switch tName {
				case ".google.protobuf.Timestamp":
					p.P("if t, err := ", p.pbTypesPkg.Use(), ".TimestampProto(", p.fakerPkg.Use(), ".Date().Birthday(", sMin, ", ", sMax, ")); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", t)")
					} else {
						p.P(variableName, " ", af, " t")
					}
					p.Out()
					p.P("}")
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Date().Birthday(", sMin, ", ", sMax, ") // bad type conversion unknow message type ", tName)
				}
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Date().Birthday(", sMin, ", ", sMax, ") // bad type conversion")
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Date().Birthday(...) // skipped [(gomeetfaker.field).date.birthday = {}]")
		}
	case *gomeetfaker.DateRules_Backward:
		params := date.GetBackward()
		format := date.GetFormat()
		if format == "" {
			format = "2006-01-02 15:04:05"
		}
		if params != "" {
			if duration, err := time.ParseDuration(params); err == nil {
				if duration < 0 {
					p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), `.Date().Backward(...) // bad duration time.ParseDuration("`, params, `") < 0`)
				} else {
					fT := *(field.Type)
					switch fT {
					case descriptor.FieldDescriptorProto_TYPE_STRING:
						p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
						p.In()
						p.P("a", fieldName, "Date := ", p.fakerPkg.Use(), ".Date().Backward(a", fieldName, "Duration)")
						if field.IsRepeated() {
							p.P(fieldVariableName, " = append(", fieldVariableName, ", a", fieldName, `Date.Format("`, format, `"))`)
						} else {
							p.P(variableName, " ", af, " a", fieldName, `Date.Format("`, format, `")`)
						}
						p.Out()
						p.P("}")
					case descriptor.FieldDescriptorProto_TYPE_BYTES:
						p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
						p.In()
						p.P("a", fieldName, "Date := ", p.fakerPkg.Use(), ".Date().Backward(a", fieldName, "Duration)")
						if field.IsRepeated() {
							p.P(fieldVariableName, " = append(", fieldVariableName, ", []byte(a", fieldName, `Date.Format("`, format, `")))`)
						} else {
							p.P(variableName, " ", af, " []byte(a", fieldName, `Date.Format("`, format, `"))`)
						}
						p.Out()
						p.P("}")
					case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
						tName := field.GetTypeName()
						switch tName {
						case ".google.protobuf.Timestamp":
							p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
							p.In()
							p.P("if t, err := ", p.pbTypesPkg.Use(), ".TimestampProto(", p.fakerPkg.Use(), ".Date().Backward(a", fieldName, "Duration)); err == nil {")
							p.In()
							if field.IsRepeated() {
								p.P(fieldVariableName, " = append(", fieldVariableName, ", t)")
							} else {
								p.P(variableName, " ", af, " t")
							}
							p.Out()
							p.P("}")
							p.Out()
							p.P("}")
						default:
							p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Date().Backward(...) // bad type conversion unknow message type ", tName)
						}
					default:
						p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Date().Backward(...) // bad type conversion", fmt.Sprintf("%s", fT))
					}
				}
			} else {
				p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Date().Backward(...) // bad duration ", err.Error())
			}
		} else {
			p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), `.Date().Backward(...) // skipped [(gomeetfaker.field).date.backward = ""]`)
		}
	case *gomeetfaker.DateRules_Now:
		params := date.GetNow()
		format := date.GetFormat()
		if format == "" {
			format = "2006-01-02 15:04:05"
		}
		if params {
			fT := *(field.Type)
			switch fT {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", ", p.timePkg.Use(), `.Now().Format("`, format, `"))`)
				} else {
					p.P(variableName, " ", af, " ", p.timePkg.Use(), `.Now().Format("`, format, `")`)
				}
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", []byte(", p.timePkg.Use(), `.Now().Format("`, format, `")))`)
				} else {
					p.P(variableName, " ", af, " []byte(", p.timePkg.Use(), `.Now().Format("`, format, `"))`)
				}
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				tName := field.GetTypeName()
				switch tName {
				case ".google.protobuf.Timestamp":
					p.P("if t, err := ", p.pbTypesPkg.Use(), ".TimestampProto(", p.timePkg.Use(), ".Now()); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", t)")
					} else {
						p.P(variableName, " ", af, " t")
					}
					p.Out()
					p.P("}")
				default:
					p.P("// ", variableName, " ", af, " ", p.timePkg.Use(), `.Now().Format("`, format, `") // bad type conversion unknow message type `, tName)
				}
			default:
				p.P("// ", variableName, " ", af, " ", p.timePkg.Use(), `.Now().Format("`, format, `") // bad type conversion`, fmt.Sprintf("%s", fT))
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.timePkg.Use(), `.Now().Format("`, format, `") // skipped [(gomeetfaker.field).date.now = false]`)
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Date().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
	}
	if field.IsRepeated() {
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateNumberRules(variableName string, r *gomeetfaker.FieldFakerRules_Number, field *descriptor.FieldDescriptorProto, fieldName string) {
	number := r.Number
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := number.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := number.Type.(type) {
	case *gomeetfaker.NumberRules_Digit:
		if number.GetDigit() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Digit()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Number().Digit())")
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Digit(), 64); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", float64(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " float64(v)")
				}
				p.Out()
				p.P("}")
			case descriptor.FieldDescriptorProto_TYPE_FLOAT:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Digit(), 32); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", float32(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " float32(v)")
				}
				p.Out()
				p.P("}")
			case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.Number().Digit(), 10, 32); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", int32(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " int32(v)")
				}
				p.Out()
				p.P("}")
			case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.Number().Digit(), 10, 64); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", int64(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " int64(v)")
				}
				p.Out()
				p.P("}")
			case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_FIXED32:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseUint(faker.Number().Digit(), 10, 32); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", uint32(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " uint32(v)")
				}
				p.Out()
				p.P("}")
			case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseUint(faker.Number().Digit(), 10, 64); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", uint64(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " uint64(v)")
				}
				p.Out()
				p.P("}")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Digit() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Digit() // skipped [(gomeetfaker.field).number.digit = false]")
			skipAppend = true
		}
	case *gomeetfaker.NumberRules_Number:
		params := number.GetNumber()
		if params != nil {
			digits := params.GetDigits()
			arg := strconv.FormatUint(uint64(digits), 10)
			if digits > 0 {
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Number(", arg, ")")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Number().Number(", arg, "))")
				case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Number(", arg, "), 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", float64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " float64(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_FLOAT:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Number(", arg, "), 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", float32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " float32(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.Number().Number(", arg, "), 10, 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", int32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " int32(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.Number().Number(", arg, "), 10, 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", int64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " int64(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_FIXED32:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseUint(faker.Number().Number(", arg, "), 10, 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", uint32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " uint32(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseUint(faker.Number().Number(", arg, "), 10, 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", uint64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " uint64(v)")
					}
					p.Out()
					p.P("}")
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Number(", arg, ") // bad type conversion")
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Number(", arg, ") // skipped ", arg, "< 1")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Number(...) // skipped [(gomeetfaker.field).number.number = {}]")
			skipAppend = true
		}
	case *gomeetfaker.NumberRules_Decimal:
		params := number.GetDecimal()
		if params != nil {
			precision, scale := params.GetPrecision(), params.GetScale()
			sP, sS := strconv.FormatUint(uint64(precision), 10), strconv.FormatUint(uint64(scale), 10)
			if precision > 0 && scale > 0 {
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Decimal(", sP, ", ", sS, ")")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Number().Decimal(", sP, ", ", sS, "))")
				case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Decimal(", sP, ", ", sS, "), 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", float64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " float64(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_FLOAT:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Decimal(", sP, ", ", sS, "), 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", float32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " float32(v)")
					}
					p.Out()
					p.P("}")
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Decimal(", sP, ", ", sS, ") // bad type conversion")
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Decimal(", sP, ", ", sS, ") // skipped ", sP, "< 1 OR ", sS, "< 1")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Digit(...) // skipped [(gomeetfaker.field).number.decimal = {}]")
			skipAppend = true
		}
	case *gomeetfaker.NumberRules_Hexadecimal:
		params := number.GetHexadecimal()
		if params != nil {
			digits := params.GetDigits()
			arg := strconv.FormatUint(uint64(digits), 10)
			if digits > 0 {
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Hexadecimal(", arg, ")")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Number().Hexadecimal(", arg, "))")
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Hexadecimal(", arg, ") // bad type conversion")
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Hexadecimal(", arg, ") // skipped ", arg, "< 1")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Hexadecimal(...) // skipped [(gomeetfaker.field).number.hexadecimal = {}]")
			skipAppend = true
		}
	case *gomeetfaker.NumberRules_Between:
		params := number.GetBetween()
		if params != nil {
			min, max := params.GetMin(), params.GetMax()
			if min > max {
				min, max = max, min
			}
			sMin, sMax := strconv.FormatInt(int64(min), 10), strconv.FormatInt(int64(max), 10)
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Between(", sMin, ", ", sMax, ")")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Number().Between(", sMin, ", ", sMax, "))")
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Between(", sMin, ", ", sMax, "), 64); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", float64(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " float64(v)")
				}
				p.Out()
				p.P("}")
			case descriptor.FieldDescriptorProto_TYPE_FLOAT:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Between(", sMin, ", ", sMax, "), 32); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", float32(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " float32(v)")
				}
				p.Out()
				p.P("}")
			case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.Number().Between(", sMin, ", ", sMax, "), 10, 32); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", int32(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " int32(v)")
				}
				p.Out()
				p.P("}")
			case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.Number().Between(", sMin, ", ", sMax, "), 10, 64); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", int64(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " int64(v)")
				}
				p.Out()
				p.P("}")
			case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_FIXED32:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseUint(faker.Number().Between(", sMin, ", ", sMax, "), 10, 32); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", uint32(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " uint32(v)")
				}
				p.Out()
				p.P("}")
			case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
				p.P("if v, err := ", p.strconvPkg.Use(), ".ParseUint(faker.Number().Between(", sMin, ", ", sMax, "), 10, 64); err == nil {")
				p.In()
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", uint64(v))")
					skipAppend = true
				} else {
					p.P(variableName, " ", af, " uint64(v)")
				}
				p.Out()
				p.P("}")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Between(", sMin, ", ", sMax, ") // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Between(...) // skipped [(gomeetfaker.field).number.number.between = {}]")
			skipAppend = true
		}
	case *gomeetfaker.NumberRules_Positive:
		params := number.GetPositive()
		if params != nil {
			max := params.GetMax()
			sMax := strconv.FormatInt(int64(max), 10)
			if max > 0 {
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Positive(", sMax, ")")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Number().Positive(", sMax, "))")
				case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Positive(", sMax, "), 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", float64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " float64(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_FLOAT:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Positive(", sMax, "), 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", float32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " float32(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.Number().Positive(", sMax, "), 10, 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", int32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " int32(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.Number().Positive(", sMax, "), 10, 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", int64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " int64(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_FIXED32:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseUint(faker.Number().Positive(", sMax, "), 10, 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", uint32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " uint32(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseUint(faker.Number().Positive(", sMax, "), 10, 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", uint64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " uint64(v)")
					}
					p.Out()
					p.P("}")
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Positive(", sMax, ") // bad type conversion")
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Positive(", sMax, ") // skipped ", sMax, "< 1")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Positive(...) // skipped [(gomeetfaker.field).number.positive = {}]")
			skipAppend = true
		}
	case *gomeetfaker.NumberRules_Negative:
		params := number.GetNegative()
		if params != nil {
			min := params.GetMin()
			sMin := strconv.FormatInt(int64(min), 10)
			if min < 0 {
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Negative(", sMin, ")")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".Number().Negative(", sMin, "))")
				case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Negative(", sMin, "), 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", float64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " float64(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_FLOAT:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.Number().Negative(", sMin, "), 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", float32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " float32(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.Number().Negative(", sMin, "), 10, 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", int32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " int32(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.Number().Negative(", sMin, "), 10, 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", int64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " int64(v)")
					}
					p.Out()
					p.P("}")
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Negative(", sMin, ") // bad type conversion")
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Negative(", sMin, ") // skipped ", sMin, "> -1")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Negative(...) // skipped [(gomeetfaker.field).number.positive = {}]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Number().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generatePhoneNumberRules(variableName string, r *gomeetfaker.FieldFakerRules_PhoneNumber, field *descriptor.FieldDescriptorProto, fieldName string) {
	phoneNumber := r.PhoneNumber
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := phoneNumber.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := phoneNumber.Type.(type) {
	case *gomeetfaker.PhoneNumberRules_PhoneNumber:
		if phoneNumber.GetPhoneNumber() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().PhoneNumber()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".PhoneNumber().PhoneNumber())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().PhoneNumber() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().PhoneNumber() // skipped [(gomeetfaker.field).phone_number.phone_number = false]")
			skipAppend = true
		}
	case *gomeetfaker.PhoneNumberRules_CellPhone:
		if phoneNumber.GetCellPhone() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().CellPhone()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".PhoneNumber().CellPhone())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().CellPhone() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().CellPhone() // skipped [(gomeetfaker.field).phone_number.cell_phone = false]")
			skipAppend = true
		}
	case *gomeetfaker.PhoneNumberRules_AreaCode:
		if phoneNumber.GetAreaCode() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().AreaCode()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".PhoneNumber().AreaCode())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().AreaCode() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().AreaCode() // skipped [(gomeetfaker.field).phone_number.area_code = false]")
			skipAppend = true
		}
	case *gomeetfaker.PhoneNumberRules_SubscriberNumber:
		params := phoneNumber.GetSubscriberNumber()
		if params != nil {
			digits := params.GetDigits()
			arg := strconv.FormatUint(uint64(digits), 10)
			if digits > 0 {
				switch *(field.Type) {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().SubscriberNumber(", arg, ")")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".PhoneNumber().SubscriberNumber(", arg, "))")
				case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.PhoneNumber().SubscriberNumber(", arg, "), 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", float64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " float64(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_FLOAT:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseFloat(faker.PhoneNumber().SubscriberNumber(", arg, "), 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", float32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " float32(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.PhoneNumber().SubscriberNumber(", arg, "), 10, 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", int32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " int32(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseInt(faker.PhoneNumber().SubscriberNumber(", arg, "), 10, 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", int64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " int64(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_FIXED32:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseUint(faker.PhoneNumber().SubscriberNumber(", arg, "), 10, 32); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", uint32(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " uint32(v)")
					}
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
					p.P("if v, err := ", p.strconvPkg.Use(), ".ParseUint(faker.PhoneNumber().SubscriberNumber(", arg, "), 10, 64); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", uint64(v))")
						skipAppend = true
					} else {
						p.P(variableName, " ", af, " uint64(v)")
					}
					p.Out()
					p.P("}")
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().SubscriberNumber(", arg, ") // bad type conversion")
					skipAppend = true
				}
			} else {
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().SubscriberNumber(", arg, ") // skipped ", arg, "< 1")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().SubscriberNumber(...) // skipped [(gomeetfaker.field).phone_number.subscriber_number = {}]")
			skipAppend = true
		}
	case *gomeetfaker.PhoneNumberRules_ExchangeCode:
		if phoneNumber.GetExchangeCode() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().ExchangeCode()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".PhoneNumber().ExchangeCode())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().ExchangeCode() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().ExchangeCode() // skipped [(gomeetfaker.field).phone_number.exchange_code = false]")
			skipAppend = true
		}
	case *gomeetfaker.PhoneNumberRules_String_:
		if phoneNumber.GetString_() {
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().String()")
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(variableName, " ", af, " []byte(", p.fakerPkg.Use(), ".PhoneNumber().String())")
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().String() // bad type conversion")
				skipAppend = true
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().String() // skipped [(gomeetfaker.field).phone_number.string = false]")
			skipAppend = true
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".PhoneNumber().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateTimeRules(variableName string, r *gomeetfaker.FieldFakerRules_Time, field *descriptor.FieldDescriptorProto, fieldName string) {
	ti := r.Time
	fieldVariableName := variableName
	af := " = "
	if field.IsRepeated() {
		rCount := ti.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	switch rr := ti.Type.(type) {
	case *gomeetfaker.TimeRules_Between:
		params := ti.GetBetween()
		format := ti.GetFormat()
		if format == "" {
			format = "2006-01-02 15:04:05"
		}
		from, to := params.GetFrom(), params.GetTo()
		if _, err := time.Parse(format, from); err == nil {
			if _, err := time.Parse(format, to); err == nil {
				fT := *(field.Type)
				switch fT {
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					p.P("if t1, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, from, `"); err == nil {`)
					p.In()
					p.P("if t2, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, to, `"); err == nil {`)
					p.In()
					p.P("a", fieldName, "Time := ", p.fakerPkg.Use(), ".Time().Between(t1, t2)")
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", a", fieldName, `Time.Format("`, format, `"))`)
					} else {
						p.P(variableName, " ", af, " a", fieldName, `Time.Format("`, format, `")`)
					}
					p.Out()
					p.P("}")
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					p.P("if t1, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, from, `"); err == nil {`)
					p.In()
					p.P("if t2, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, to, `"); err == nil {`)
					p.In()
					p.P("a", fieldName, "Time := ", p.fakerPkg.Use(), ".Time().Between(t1, t2)")
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", []byte(a", fieldName, `Time.Format("`, format, `")))`)
					} else {
						p.P(variableName, " ", af, " []byte(a", fieldName, `Time.Format("`, format, `"))`)
					}
					p.Out()
					p.P("}")
					p.Out()
					p.P("}")
				case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
					tName := field.GetTypeName()
					switch tName {
					case ".google.protobuf.Timestamp":
						p.P("if t1, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, from, `"); err == nil {`)
						p.In()
						p.P("if t2, err := ", p.timePkg.Use(), `.Parse("`, format, `", "`, to, `"); err == nil {`)
						p.In()
						p.P("if t, err := ", p.pbTypesPkg.Use(), ".TimestampProto(", p.fakerPkg.Use(), ".Time().Between(t1, t2)); err == nil {")
						p.In()
						if field.IsRepeated() {
							p.P(fieldVariableName, " = append(", fieldVariableName, ", t)")
						} else {
							p.P(variableName, " ", af, " t")
						}
						p.Out()
						p.P("}")
						p.Out()
						p.P("}")
						p.Out()
						p.P("}")
					default:
						p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Time().Between(...) // bad type conversion unknow message type ", tName)
					}
				default:
					p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Time().Between(...) // bad type conversion", fmt.Sprintf("%s", fT))
				}
			} else {
				p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Time().Between(...) // bad type conversion ", err.Error())
			}
		} else {
			p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Time().Between(...) // bad type conversion ", err.Error())
		}
	case *gomeetfaker.TimeRules_Forward:
		params := ti.GetForward()
		format := ti.GetFormat()
		if format == "" {
			format = "2006-01-02 15:04:05"
		}
		if params != "" {
			if duration, err := time.ParseDuration(params); err == nil {
				if duration < 0 {
					p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), `.Time().Forward(...) // bad duration time.ParseDuration("`, params, `") < 0`)
				} else {
					fT := *(field.Type)
					switch fT {
					case descriptor.FieldDescriptorProto_TYPE_STRING:
						p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
						p.In()
						p.P("a", fieldName, "Time := ", p.fakerPkg.Use(), ".Time().Forward(a", fieldName, "Duration)")
						if field.IsRepeated() {
							p.P(fieldVariableName, " = append(", fieldVariableName, ", a", fieldName, `Time.Format("`, format, `"))`)
						} else {
							p.P(variableName, " ", af, " a", fieldName, `Time.Format("`, format, `")`)
						}
						p.Out()
						p.P("}")
					case descriptor.FieldDescriptorProto_TYPE_BYTES:
						p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
						p.In()
						p.P("a", fieldName, "Time := ", p.fakerPkg.Use(), ".Time().Forward(a", fieldName, "Duration)")
						if field.IsRepeated() {
							p.P(fieldVariableName, " = append(", fieldVariableName, ", []byte(a", fieldName, `Time.Format("`, format, `")))`)
						} else {
							p.P(variableName, " ", af, " []byte(a", fieldName, `Time.Format("`, format, `"))`)
						}
						p.Out()
						p.P("}")
					case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
						tName := field.GetTypeName()
						switch tName {
						case ".google.protobuf.Timestamp":
							p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
							p.In()
							p.P("if t, err := ", p.pbTypesPkg.Use(), ".TimestampProto(", p.fakerPkg.Use(), ".Time().Forward(a", fieldName, "Duration)); err == nil {")
							p.In()
							if field.IsRepeated() {
								p.P(fieldVariableName, " = append(", fieldVariableName, ", t)")
							} else {
								p.P(variableName, " ", af, " t")
							}
							p.Out()
							p.P("}")
							p.Out()
							p.P("}")
						default:
							p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Time().Forward(...) // bad type conversion unknow message type ", tName)
						}
					default:
						p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Time().Forward(...) // bad type conversion", fmt.Sprintf("%s", fT))
					}
				}
			} else {
				p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Time().Forward(...) // bad duration ", err.Error())
			}
		} else {
			p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), `.Time().Forward(...) // skipped [(gomeetfaker.field).time.forward = ""]`)
		}
	case *gomeetfaker.TimeRules_Birthday:
		params := ti.GetBirthday()
		if params != nil {
			min, max := params.GetMin(), params.GetMax()
			if min > max {
				min, max = max, min
			}
			sMin, sMax := strconv.FormatInt(int64(min), 10), strconv.FormatInt(int64(max), 10)
			format := ti.GetFormat()
			if format == "" {
				format = "2006-01-02"
			}
			switch *(field.Type) {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P("a", fieldName, "Time := ", p.fakerPkg.Use(), ".Time().Birthday(", sMin, ", ", sMax, ")")
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", a", fieldName, `Time.Format("`, format, `"))`)
				} else {
					p.P(variableName, " ", af, " a", fieldName, `Time.Format("`, format, `")`)
				}
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P("a", fieldName, "Time := ", p.fakerPkg.Use(), ".Time().Birthday(", sMin, ", ", sMax, ")")
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", []byte(a", fieldName, `Time.Format("`, format, `")))`)
				} else {
					p.P(variableName, " ", af, " []byte(a", fieldName, `Time.Format("`, format, `"))`)
				}
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				tName := field.GetTypeName()
				switch tName {
				case ".google.protobuf.Timestamp":
					p.P("if t, err := ", p.pbTypesPkg.Use(), ".TimestampProto(", p.fakerPkg.Use(), ".Time().Birthday(", sMin, ", ", sMax, ")); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", t)")
					} else {
						p.P(variableName, " ", af, " t")
					}
					p.Out()
					p.P("}")
				default:
					p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Time().Birthday(", sMin, ", ", sMax, ") // bad type conversion unknow message type ", tName)
				}
			default:
				p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Time().Birthday(", sMin, ", ", sMax, ") // bad type conversion")
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Time().Birthday(...) // skipped [(gomeetfaker.field).time.birthday = {}]")
		}
	case *gomeetfaker.TimeRules_Backward:
		params := ti.GetBackward()
		format := ti.GetFormat()
		if format == "" {
			format = "2006-01-02 15:04:05"
		}
		if params != "" {
			if duration, err := time.ParseDuration(params); err == nil {
				if duration < 0 {
					p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), `.Time().Backward(...) // bad duration time.ParseDuration("`, params, `") < 0`)
				} else {
					fT := *(field.Type)
					switch fT {
					case descriptor.FieldDescriptorProto_TYPE_STRING:
						p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
						p.In()
						p.P("a", fieldName, "Time := ", p.fakerPkg.Use(), ".Time().Backward(a", fieldName, "Duration)")
						if field.IsRepeated() {
							p.P(fieldVariableName, " = append(", fieldVariableName, ", a", fieldName, `Time.Format("`, format, `"))`)
						} else {
							p.P(variableName, " ", af, " a", fieldName, `Time.Format("`, format, `")`)
						}
						p.Out()
						p.P("}")
					case descriptor.FieldDescriptorProto_TYPE_BYTES:
						p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
						p.In()
						p.P("a", fieldName, "Time := ", p.fakerPkg.Use(), ".Time().Backward(a", fieldName, "Duration)")
						if field.IsRepeated() {
							p.P(fieldVariableName, " = append(", fieldVariableName, ", []byte(a", fieldName, `Time.Format("`, format, `")))`)
						} else {
							p.P(variableName, " ", af, " []byte(a", fieldName, `Time.Format("`, format, `"))`)
						}
						p.Out()
						p.P("}")
					case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
						tName := field.GetTypeName()
						switch tName {
						case ".google.protobuf.Timestamp":
							p.P("if a", fieldName, "Duration, err := ", p.timePkg.Use(), `.ParseDuration("`, params, `"); err == nil {`)
							p.In()
							p.P("if t, err := ", p.pbTypesPkg.Use(), ".TimestampProto(", p.fakerPkg.Use(), ".Time().Backward(a", fieldName, "Duration)); err == nil {")
							p.In()
							if field.IsRepeated() {
								p.P(fieldVariableName, " = append(", fieldVariableName, ", t)")
							} else {
								p.P(variableName, " ", af, " t")
							}
							p.Out()
							p.P("}")
							p.Out()
							p.P("}")
						default:
							p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Time().Backward(...) // bad type conversion unknow message type ", tName)
						}
					default:
						p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Time().Backward(...) // bad type conversion", fmt.Sprintf("%s", fT))
					}
				}
			} else {
				p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), ".Time().Backward(...) // bad duration ", err.Error())
			}
		} else {
			p.P("// ", fieldName, " ", af, " ", p.fakerPkg.Use(), `.Time().Backward(...) // skipped [(gomeetfaker.field).time.backward = ""]`)
		}
	case *gomeetfaker.TimeRules_Now:
		params := ti.GetNow()
		format := ti.GetFormat()
		if format == "" {
			format = "2006-01-02 15:04:05"
		}
		if params {
			fT := *(field.Type)
			switch fT {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", ", p.timePkg.Use(), `.Now().Format("`, format, `"))`)
				} else {
					p.P(variableName, " ", af, " ", p.timePkg.Use(), `.Now().Format("`, format, `")`)
				}
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				if field.IsRepeated() {
					p.P(fieldVariableName, " = append(", fieldVariableName, ", []byte(", p.timePkg.Use(), `.Now().Format("`, format, `")))`)
				} else {
					p.P(variableName, " ", af, " []byte(", p.timePkg.Use(), `.Now().Format("`, format, `"))`)
				}
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				tName := field.GetTypeName()
				switch tName {
				case ".google.protobuf.Timestamp":
					p.P("if t, err := ", p.pbTypesPkg.Use(), ".TimestampProto(", p.timePkg.Use(), ".Now()); err == nil {")
					p.In()
					if field.IsRepeated() {
						p.P(fieldVariableName, " = append(", fieldVariableName, ", t)")
					} else {
						p.P(variableName, " ", af, " t")
					}
					p.Out()
					p.P("}")
				default:
					p.P("// ", variableName, " ", af, " ", p.timePkg.Use(), `.Now().Format("`, format, `") // bad type conversion unknow message type `, tName)
				}
			default:
				p.P("// ", variableName, " ", af, " ", p.timePkg.Use(), `.Now().Format("`, format, `") // bad type conversion`, fmt.Sprintf("%s", fT))
			}
		} else {
			p.P("// ", variableName, " ", af, " ", p.timePkg.Use(), `.Now().Format("`, format, `") // skipped [(gomeetfaker.field).date.now = false]`)
		}
	default:
		p.P("// ", variableName, " ", af, " ", p.fakerPkg.Use(), ".Time().Unknow() // bad type conversion", fmt.Sprintf("%T", rr))
	}
	if field.IsRepeated() {
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateUuidRules(variableName string, r *gomeetfaker.FieldFakerRules_Uuid, field *descriptor.FieldDescriptorProto, fieldName string) {
	uuid := r.Uuid
	fieldVariableName := variableName
	af := " = "
	if field.IsRepeated() {
		rCount := uuid.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	fT := *(field.Type)
	version := strings.ToUpper(strings.TrimSpace(uuid.GetVersion()))
	switch version {
	case "V1":
		switch fT {
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			p.P("if aUuid, err := ", p.uuidPkg.Use(), ".NewUUID(); err == nil {")
			p.In()
			p.P(variableName, " ", af, " aUuid.String()")
			if field.IsRepeated() {
				p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
			}
			p.Out()
			p.P("}")
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			p.P("if aUuid, err := ", p.uuidPkg.Use(), ".NewUUID(); err == nil {")
			p.In()
			p.P(variableName, " ", af, " []byte(aUuid.String())")
			if field.IsRepeated() {
				p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
			}
			p.Out()
			p.P("}")
		default:
			p.P("// ", variableName, " ", af, " ", version, " // unknow field type bad type conversion", fmt.Sprintf("%T", fT))
		}
	case "V4":
		switch fT {
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			p.P(variableName, " ", af, " ", p.uuidPkg.Use(), ".New().String()")
			if field.IsRepeated() {
				p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
			}
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			p.P(variableName, " ", af, " []byte(", p.uuidPkg.Use(), ".New().String())")
			if field.IsRepeated() {
				p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
			}
		default:
			p.P("// ", variableName, " ", af, " ", version, " // unknow field type bad type conversion", fmt.Sprintf("%T", fT))
		}
	default:
		p.P("// can't set an uuid version ", version, " in ", variableName, " unknow uuid version - bad uuid version [V1, V4]")
	}

	if field.IsRepeated() {
		p.Out()
		p.P("}")
	}
}

func (p *plugin) generateValueRules(variableName string, r *gomeetfaker.FieldFakerRules_Value, field *descriptor.FieldDescriptorProto, fieldName string) {
	value := r.Value
	fieldVariableName := variableName
	skipAppend := false
	af := " = "
	if field.IsRepeated() {
		rCount := value.GetRepeated()
		if rCount == 0 {
			rCount = gRCount
		}
		variableName = "aCurrent" + fieldName
		af = " := "
		p.P("for i := 0; i < ", fmt.Sprintf("%d", rCount), "; i++ {")
		p.In()
	}
	fT := *(field.Type)
	content := value.GetContent()
	if content != "" {
		switch fT {
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			if _, err := strconv.ParseFloat(content, 64); err == nil {
				p.P(variableName, " ", af, " float64(", content, ")")
			} else {
				p.P("// ", variableName, " ", af, " float64(", content, ") // skipped bad type convertion - ", err.Error())
				skipAppend = true
			}
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			if _, err := strconv.ParseFloat(content, 32); err == nil {
				p.P(variableName, " ", af, " float32(", content, ")")
			} else {
				p.P("// ", variableName, " ", af, " float32(", content, ") // skipped bad type convertion - ", err.Error())
				skipAppend = true
			}
		case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
			if _, err := strconv.ParseInt(content, 10, 32); err == nil {
				p.P(variableName, " ", af, " int32(", content, ")")
			} else {
				p.P("// ", variableName, " ", af, " int32(", content, ") // skipped bad type convertion - ", err.Error())
				skipAppend = true
			}
		case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
			if _, err := strconv.ParseInt(content, 10, 64); err == nil {
				p.P(variableName, " ", af, " int64(", content, ")")
			} else {
				p.P("// ", variableName, " ", af, " int64(", content, ") // skipped bad type convertion - ", err.Error())
				skipAppend = true
			}
		case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_FIXED32:
			if _, err := strconv.ParseUint(content, 10, 32); err == nil {
				p.P(variableName, " ", af, " uint32(", content, ")")
			} else {
				p.P("// ", variableName, " ", af, " uint32(", content, ") // skipped bad type convertion - ", err.Error())
				skipAppend = true
			}
		case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
			if _, err := strconv.ParseUint(content, 10, 64); err == nil {
				p.P(variableName, " ", af, " uint64(", content, ")")
			} else {
				p.P("// ", variableName, " ", af, " uint64(", content, ") // skipped bad type convertion - ", err.Error())
				skipAppend = true
			}
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			if v, err := strconv.ParseBool(content); err == nil {
				if v {
					p.P(variableName, " ", af, " true")
				} else {
					p.P(variableName, " ", af, " false")
				}
			} else {
				p.P("// ", variableName, " ", af, " ", content, " // skipped bad type convertion - ", err.Error())
				skipAppend = true
			}
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			p.P(variableName, " ", af, " ", strconv.Quote(content))
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			p.P(variableName, " ", af, " []byte(", strconv.Quote(content), ")")
		default:
			p.P("// ", variableName, " ", af, " ", content, " // unknow field type bad type conversion", fmt.Sprintf("%T", fT))
			skipAppend = true
		}
	} else {
		p.P("// ", variableName, " ", af, " ", content, ` // skipped empty content [(gomeetfaker.field).value = { content: ""}]  - `)
		skipAppend = true
	}
	if field.IsRepeated() {
		if skipAppend {
			p.P("// ", fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ") // skipped")
		} else {
			p.P(fieldVariableName, " = append(", fieldVariableName, ", ", variableName, ")")
		}
		p.Out()
		p.P("}")
	}
}

func (p *plugin) fieldIsProto3Map(file *generator.FileDescriptor, message *generator.Descriptor, field *descriptor.FieldDescriptorProto) bool {
	// Context from descriptor.proto
	// Whether the message is an automatically generated map entry type for the
	// maps field.
	//
	// For maps fields:
	//     map<KeyType, ValueType> map_field = 1;
	// The parsed descriptor looks like:
	//     message MapFieldEntry {
	//         option map_entry = true;
	//         optional KeyType key = 1;
	//         optional ValueType value = 2;
	//     }
	//     repeated MapFieldEntry map_field = 1;
	//
	// Implementations may choose not to generate the map_entry=true message, but
	// use a native map in the target language to hold the keys and values.
	// The reflection APIs in such implementions still need to work as
	// if the field is a repeated message field.
	//
	// NOTE: Do not set the option in .proto files. Always use the maps syntax
	// instead. The option should only be implicitly set by the proto compiler
	// parser.
	if field.GetType() != descriptor.FieldDescriptorProto_TYPE_MESSAGE || !field.IsRepeated() {
		return false
	}
	typeName := field.GetTypeName()
	var msg *descriptor.DescriptorProto
	if strings.HasPrefix(typeName, ".") {
		// Fully qualified case, look up in global map, must work or fail badly.
		msg = p.ObjectNamed(field.GetTypeName()).(*generator.Descriptor).DescriptorProto
	} else {
		// Nested, relative case.
		msg = file.GetNestedMessage(message.DescriptorProto, field.GetTypeName())
	}
	return msg.GetOptions().GetMapEntry()
}
