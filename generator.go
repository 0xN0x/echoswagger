package echoswagger

import (
	"reflect"
)

// paramSchema returns the schema of a non-body parameter type
func paramSchema(t reflect.Type) *JSONSchema {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	st, sf := toSwaggerType(t)
	if st == "array" {
		return &JSONSchema{Type: JSONType(st), Items: paramSchema(t.Elem())}
	}
	return &JSONSchema{Type: JSONType(st), Format: sf}
}

func (r *RawDefineDic) genSchema(v reflect.Value) *JSONSchema {
	if !v.IsValid() {
		return nil
	}
	v = indirect(v)
	st, sf := toSwaggerType(v.Type())
	schema := &JSONSchema{}
	if st == "array" {
		schema.Type = JSONType(st)
		if v.Len() == 0 {
			v = reflect.MakeSlice(v.Type(), 1, 1)
		}
		schema.Items = r.genSchema(v.Index(0))
	} else if st == "object" && sf == "map" {
		schema.Type = JSONType(st)
		if v.Len() == 0 {
			v = reflect.New(v.Type().Elem())
		} else {
			v = v.MapIndex(v.MapKeys()[0])
		}
		schema.AdditionalProperties = r.genSchema(v)
	} else if st == "object" {
		key := r.addDefinition(v)
		schema.Ref = DefPrefix + key
	} else {
		schema.Type = JSONType(st)
		schema.Format = sf
		zv := reflect.Zero(v.Type())
		if v.CanInterface() && zv.CanInterface() && v.Interface() != zv.Interface() {
			schema.Examples = []interface{}{v.Interface()}
		}
	}
	return schema
}

func (api) genHeader(v reflect.Value) map[string]*Header {
	rt := indirect(v).Type()
	if rt.Kind() != reflect.Struct {
		return nil
	}
	mh := make(map[string]*Header)
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		name, _ := getFieldName(f, ParamInHeader)
		if name == "-" {
			continue
		}
		tags := getSwaggerTags(f)
		s := paramSchema(f.Type)
		s.applySwaggerTags(f, tags)
		mh[name] = &Header{Description: tags["desc"], Schema: s}
	}
	return mh
}
