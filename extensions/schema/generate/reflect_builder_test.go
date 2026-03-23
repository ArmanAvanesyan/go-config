package generate

import (
	"reflect"
	"testing"
	"time"
)

type embedded struct {
	Hidden string `json:"hidden"`
}

type EmbeddedPublic struct {
	Public string `json:"public"`
}

type rsCase struct {
	embedded
	EmbeddedPublic
	Name    string            `json:"name"`
	Skip    string            `json:"-"`
	Opt     int               `json:"opt,omitempty"`
	MapVals map[string]uint32 `json:"mapvals"`
	AnyMap  map[int]string    `json:"anymap"`
	Any     any               `json:"any"`
}

type withAnonTime struct {
	time.Time
}

func TestTypeSchemaAndStructSchema_Branches(t *testing.T) {
	t.Parallel()

	if got := typeSchema(nil, map[reflect.Type]string{}, false); got.Type != "" {
		t.Fatalf("nil type expected empty schema, got %#v", got)
	}

	ptrSchema := typeSchema(reflect.TypeOf(&rsCase{}), map[reflect.Type]string{}, false)
	if ptrSchema.Type != "object" {
		t.Fatalf("pointer unwrap expected object, got %#v", ptrSchema)
	}
	if _, ok := ptrSchema.Properties["public"]; !ok {
		t.Fatalf("exported embedded field not flattened")
	}
	if _, ok := ptrSchema.Properties["Skip"]; ok {
		t.Fatalf("json:- field should be skipped")
	}
	if contains(ptrSchema.Required, "opt") {
		t.Fatalf("omitempty field should not be required")
	}
	if ptrSchema.Properties["mapvals"].AdditionalProperties.Type != "integer" {
		t.Fatalf("map[string]T additionalProperties expected integer")
	}
	if ptrSchema.Properties["anymap"].Type != "object" || ptrSchema.Properties["anymap"].AdditionalProperties != nil {
		t.Fatalf("map[int]T should fallback to object")
	}
	if ptrSchema.Properties["any"].Type != "" {
		t.Fatalf("any should be unconstrained")
	}

	anonTime := typeSchema(reflect.TypeOf(withAnonTime{}), map[reflect.Type]string{}, false)
	// Embedded time.Time should not flatten because time has special handling.
	if _, ok := anonTime.Properties["Time"]; !ok {
		t.Fatalf("embedded time.Time should map to field property")
	}

	arrSchema := typeSchema(reflect.TypeOf([2]string{}), map[reflect.Type]string{}, false)
	if arrSchema.Type != "array" || arrSchema.Items == nil || arrSchema.Items.Type != "string" {
		t.Fatalf("array schema unexpected %#v", arrSchema)
	}
	boolSchema := typeSchema(reflect.TypeOf(true), map[reflect.Type]string{}, false)
	if boolSchema.Type != "boolean" {
		t.Fatalf("bool schema unexpected %#v", boolSchema)
	}
	floatSchema := typeSchema(reflect.TypeOf(float64(1)), map[reflect.Type]string{}, false)
	if floatSchema.Type != "number" {
		t.Fatalf("float schema unexpected %#v", floatSchema)
	}

	defs := map[reflect.Type]string{}
	s1 := structSchema(reflect.TypeOf(rsCase{}), defs, true)
	if s1.Type != "object" {
		t.Fatalf("first refs struct should still return inline schema")
	}
	s2 := structSchema(reflect.TypeOf(rsCase{}), defs, true)
	if s2.Ref == "" {
		t.Fatalf("second refs struct should return ref")
	}

	anon := reflect.StructOf([]reflect.StructField{
		{Name: "A", Type: reflect.TypeOf(""), Tag: `json:"a"`},
	})
	defs2 := map[reflect.Type]string{}
	anon2 := reflect.StructOf([]reflect.StructField{
		{Name: "B", Type: reflect.TypeOf(""), Tag: `json:"b"`},
	})
	_ = structSchema(anon, defs2, true)
	_ = structSchema(anon2, defs2, true)
	if defs2[anon] == "" {
		t.Fatalf("anonymous type should receive deterministic def name")
	}
	if defs2[anon] != "AnonStruct" && defs2[anon] != "generate_AnonStruct" {
		t.Fatalf("anonymous type unexpected def name: %q", defs2[anon])
	}
	if defs2[anon2] == defs2[anon] {
		t.Fatalf("definition names should avoid collisions: %q", defs2[anon])
	}
}

func TestIsTimeType(t *testing.T) {
	t.Parallel()
	rt := reflect.TypeOf(struct{}{})
	if isTimeType(rt) {
		t.Fatal("struct{} should not be time type")
	}
	if !isTimeType(reflect.TypeOf((*time.Time)(nil))) {
		t.Fatal("pointer time.Time should be recognized as time type")
	}
}
