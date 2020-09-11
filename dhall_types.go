package main

import (
	"bytes"
	"io"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer/regex"
)

type RecordType struct {
	Fields []*FieldType `"{" (@@ ("," @@)* )? "}"` //nolint
}

func (rt *RecordType) ToDhall(sb *strings.Builder, indentLevel int) {
	sb.WriteString("{\n")
	first := true
	for _, field := range rt.Fields {
		sb.WriteString(strings.Repeat("\t", indentLevel))
		if first {
			first = false
		} else {
			sb.WriteString(", ")
		}
		field.ToDhall(sb, indentLevel)
	}
	sb.WriteString(strings.Repeat("\t", indentLevel-1))
	sb.WriteString("}")
}

type FieldType struct {
	K string     `(@Ident | @QuotedLabel) ":"` //nolint
	V *ValueType `@@`                          //nolint
}

func (ft *FieldType) ToDhall(sb *strings.Builder, indentLevel int) {
	sb.WriteString(ft.K)
	sb.WriteString(": ")
	ft.V.ToDhall(sb, indentLevel)
	sb.WriteString("\n")
}

type UnionType struct {
	Members []*FieldType `"<" (@@ ("|" @@)* )? ">"` //nolint
}

func (ut *UnionType) ToDhall(sb *strings.Builder, indentLevel int) {
	sb.WriteString("<\n")
	first := true
	for _, member := range ut.Members {
		sb.WriteString(strings.Repeat("\t", indentLevel))
		if first {
			first = false
		} else {
			sb.WriteString(" | ")
		}
		member.ToDhall(sb, indentLevel)
	}
	sb.WriteString(strings.Repeat("\t", indentLevel-1))
	sb.WriteString(">")
}

type LastValueType struct {
	R *RecordType `@@ `  //nolint
	U *UnionType  `| @@` //nolint
}

func (lt *LastValueType) ToDhall(sb *strings.Builder, indentLevel int) {
	if lt.R != nil {
		lt.R.ToDhall(sb, indentLevel+1)
	} else if lt.U != nil {
		lt.U.ToDhall(sb, indentLevel)
	}
}

type ValueType struct {
	S []string       ` @(Ident)*` //nolint
	L *LastValueType ` ( @@ )?`   //nolint
}

func (vt *ValueType) ToDhall(sb *strings.Builder, indentLevel int) {
	n := len(vt.S)
	if vt.L != nil {
		n = n + 1
	}

	numOpenBraces := 0
	if n <= 2 {
		sb.WriteString(strings.Join(vt.S, " "))
	} else {
		sb.WriteString(vt.S[0])

		for _, st := range vt.S[1:] {
			sb.WriteString(" ( ")
			numOpenBraces++
			sb.WriteString(st)
		}
	}

	if vt.L != nil {
		if len(vt.S) > 0 {
			sb.WriteString(" ")
		}
		vt.L.ToDhall(sb, indentLevel+1)
	}

	sb.WriteString(strings.Repeat(" )", numOpenBraces))
}

const lexerSpec = `
whitespace = \s+
roundBraces = [()]
Ident = [a-zA-Z][a-zA-Z_\-\d]*
QuotedLabel = ¬.+¬
Punct = [:{}<>,|]
`

func parseRecordType(src io.Reader) (*RecordType, error) {
	mdreplacer := strings.NewReplacer("¬", "`")
	lspec := mdreplacer.Replace(lexerSpec)
	dhallTypeLexer, err := regex.New(lspec)
	if err != nil {
		return nil, err
	}
	parser, err := participle.Build(&RecordType{}, participle.Lexer(dhallTypeLexer))
	if err != nil {
		return nil, err
	}

	var rt RecordType

	err = parser.Parse(src, &rt)
	if err != nil {
		return nil, err
	}

	return &rt, nil
}

const containerResourcesType = `
	{
         cpu: Optional Text
       , memory: Optional Text
       , ephemeralStorage: Optional Text
    }
`

func improvedContainerResourcesType() *RecordType {
	bf := bytes.NewReader([]byte(containerResourcesType))

	icrt, _ := parseRecordType(bf)
	return icrt
}

const dockerImageType = `
{ image :
    < asText : Text
    | asRecord : { name : Text, registry : Text, sha256 : Text, version : Text }
    >
}
`

func improvedDockerImageType() *RecordType {
	bf := bytes.NewReader([]byte(dockerImageType))

	icrt, _ := parseRecordType(bf)
	return icrt
}

const additionalEnvVarsTypeDhall = `{ additionalEnv : Optional (List { name : Text, value : Text })}`

func additionalEnvVarsType() *RecordType {
	bf := bytes.NewReader([]byte(additionalEnvVarsTypeDhall))

	t, _ := parseRecordType(bf)
	return t
}

func transformRecordType(rt *RecordType) {
	addAdditionalEnvVars := false
	for _, field := range rt.Fields {
		if (field.K == "limits" || field.K == "requests") && field.V.L != nil && field.V.L.R != nil {
			field.V.L.R = improvedContainerResourcesType()
		} else if field.K == "image" && field.V.L != nil && field.V.L.R != nil {
			idr := improvedDockerImageType()
			field.V.L.U = idr.Fields[0].V.L.U
			field.V.L.R = nil
		} else if field.K == "metadata" && field.V.L.R != nil {
			alreadyPresent := false
			for _, f := range field.V.L.R.Fields {
				if f.K == "namespace" {
					alreadyPresent = true
				}
			}
			if !alreadyPresent {
				namespaceValue := &ValueType{
					S: []string{"Optional", "Text"},
				}
				namespaceField := &FieldType{
					K: "namespace",
					V: namespaceValue,
				}
				field.V.L.R.Fields = append(field.V.L.R.Fields, namespaceField)
			}
		} else if field.K == "env" && field.V.L.R != nil {
			for _, env := range field.V.L.R.Fields {
				if len(env.V.S) == 0 {
					env.V.S = []string{"Optional"}
				}
			}
			addAdditionalEnvVars = true
		} else if field.V.L != nil && field.V.L.R != nil {
			transformRecordType(field.V.L.R)
		}
	}

	if addAdditionalEnvVars {
		t := additionalEnvVarsType()
		rt.Fields = append(rt.Fields, t.Fields[0])
	}
}
