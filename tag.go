package echoswagger

import (
	"reflect"
	"strconv"
	"strings"
)

// getTag reports is a tag exists and it's content
// search tagName in all tags when index = -1
func getTag(field reflect.StructField, tagName string, index int) (bool, string) {
	t := field.Tag.Get(tagName)
	s := strings.Split(t, ",")

	if len(s) < index+1 {
		return false, ""
	}

	return true, strings.TrimSpace(s[index])
}

func getSwaggerTags(field reflect.StructField) map[string]string {
	t := field.Tag.Get("swagger")
	r := make(map[string]string)
	for _, v := range strings.Split(t, ",") {
		leftIndex := strings.Index(v, "(")
		rightIndex := strings.LastIndex(v, ")")
		if leftIndex > 0 && rightIndex > leftIndex {
			r[v[:leftIndex]] = v[leftIndex+1 : rightIndex]
		} else {
			r[v] = ""
		}
	}
	return r
}

func getFieldName(f reflect.StructField, in ParamInType) (string, bool) {
	var name string
	switch in {
	case ParamInQuery:
		name = f.Tag.Get("query")
	case paramInForm:
		name = f.Tag.Get("form")
	case paramInBody, ParamInHeader, ParamInPath:
		_, name = getTag(f, "json", 0)
	}
	if name != "" {
		return name, true
	} else {
		return f.Name, false
	}
}

// applySwaggerTags applies the validation related swagger tags to the schema.
// For array types, they are applied to the innermost items.
func (s *JSONSchema) applySwaggerTags(f reflect.StructField, tags map[string]string) {
	s = s.latest()
	if t, ok := tags["min"]; ok {
		if m, err := strconv.ParseFloat(t, 64); err == nil {
			s.Minimum = &m
		}
	}
	if t, ok := tags["max"]; ok {
		if m, err := strconv.ParseFloat(t, 64); err == nil {
			s.Maximum = &m
		}
	}
	if t, ok := tags["minLen"]; ok {
		if m, err := strconv.Atoi(t); err == nil {
			s.MinLength = &m
		}
	}
	if t, ok := tags["maxLen"]; ok {
		if m, err := strconv.Atoi(t); err == nil {
			s.MaxLength = &m
		}
	}

	convert := converter(f.Type)
	if t, ok := tags["enum"]; ok {
		var es []interface{}
		for _, e := range strings.Split(t, "|") {
			v, err := convert(e)
			if err != nil {
				continue
			}
			es = append(es, v)
		}
		s.Enum = es
	}
	if t, ok := tags["default"]; ok {
		v, err := convert(t)
		if err == nil {
			s.DefaultValue = v
		}
	}
}

// handleSwaggerTags applies the swagger tags of field f to the property name of schema s.
func (s *JSONSchema) handleSwaggerTags(f reflect.StructField, name string) {
	propSchema := s.Properties[name]
	tags := getSwaggerTags(f)

	if t, ok := tags["desc"]; ok {
		propSchema.Description = t
	}
	if _, ok := tags["required"]; ok {
		s.Required = append(s.Required, name)
	}
	if _, ok := tags["readOnly"]; ok {
		propSchema.ReadOnly = true
	}
	propSchema.applySwaggerTags(f, tags)
}

func (s *JSONSchema) latest() *JSONSchema {
	if s.Items != nil {
		return s.Items.latest()
	}
	return s
}

// Not support nested elements tag eg:"a>b>c"
// Not support tags: ",chardata", ",cdata", ",comment"
// Not support embedded structure with tag ",innerxml"
// Only support nested elements tag in array type eg:"Name []string `xml:"names>name"`"
func (s *JSONSchema) handleXMLTags(f reflect.StructField) {
	b, a := getTag(f, "xml", 1)
	if b && contains([]string{"chardata", "cdata", "comment"}, a) {
		return
	}

	if b, t := getTag(f, "xml", 0); b {
		if t == "-" || s.Ref != "" {
			return
		} else if t == "" {
			t = f.Name
		}

		if s.XML == nil {
			s.XML = &XMLSchema{}
		}
		s.XML.Name = t
		s.XML.Attribute = a == "attr"
	}
}

func (s *JSONSchema) handleChildXMLTags(rest string, r *RawDefineDic) {
	if rest == "" {
		return
	}

	if s.Items == nil && s.Ref == "" {
		if s.XML == nil {
			s.XML = &XMLSchema{}
		}
		s.XML.Name = rest
	} else if s.Ref != "" {
		key := s.Ref[len(DefPrefix):]
		if sc, ok := (*r)[key]; ok && sc.Schema != nil {
			if sc.Schema.XML == nil {
				sc.Schema.XML = &XMLSchema{}
			}
			sc.Schema.XML.Name = rest
		}
	} else {
		if s.XML == nil {
			s.XML = &XMLSchema{}
		}
		s.XML.Wrapped = true
		i := strings.Index(rest, ">")
		if i <= 0 {
			s.XML.Name = rest
		} else {
			s.XML.Name = rest[:i]
			rest = rest[i+1:]
			s.Items.handleChildXMLTags(rest, r)
		}
	}
}
