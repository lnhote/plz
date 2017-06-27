package native

import (
	"github.com/v2pro/plz/accessor"
	"reflect"
	"fmt"
	"github.com/v2pro/plz"
	"unsafe"
)

func init() {
	accessor.Providers = append(accessor.Providers, accessorOf)
}

func accessorOf(typ reflect.Type) accessor.Accessor {
	if typ.Kind() == reflect.Map {
		return &mapAccessor{}
	}
	if typ.Kind() != reflect.Ptr {
		return nil
	}
	typ = typ.Elem()
	switch typ.Kind() {
	case reflect.Int:
		return &intAccessor{}
	case reflect.Struct:
		return &structAccessor{
			typ: typ,
		}
	}
	panic(fmt.Sprintf("do not support: %v", typ.Kind()))
}

type intAccessor struct {
	accessor.NoopAccessor
}

func (acc *intAccessor) Kind() reflect.Kind {
	return reflect.Int
}

func (acc *intAccessor) Int(obj interface{}) int {
	typedObj := obj.(*int)
	return *typedObj
}

func (acc *intAccessor) SetInt(obj interface{}, val int) {
	*(obj.(*int)) = val
}

type structAccessor struct {
	accessor.NoopAccessor
	typ reflect.Type
}

func (acc *structAccessor) Kind() reflect.Kind {
	return reflect.Struct
}

func (acc *structAccessor) NumField() int {
	return acc.typ.NumField()
}

func (acc *structAccessor) Field(index int) accessor.StructField {
	field := acc.typ.Field(index)
	ptrType := reflect.PtrTo(field.Type)
	fieldAcc := plz.AccessorOf(ptrType)
	templateObj := castToEmptyInterface(reflect.New(field.Type).Interface())
	return accessor.StructField{
		Name: field.Name,
		Accessor: &structFieldAccessor{
			field:       field,
			templateObj: templateObj,
			accessor:    fieldAcc,
		},
	}
}

type structFieldAccessor struct {
	accessor.NoopAccessor
	field       reflect.StructField
	templateObj emptyInterface
	accessor    accessor.Accessor
}

func (acc *structFieldAccessor) Kind() reflect.Kind {
	return acc.accessor.Kind()
}

func (acc *structFieldAccessor) Int(obj interface{}) int {
	return acc.accessor.Int(acc.fieldOf(obj))
}

func (acc *structFieldAccessor) SetInt(obj interface{}, val int) {
	acc.accessor.SetInt(acc.fieldOf(obj), val)
}

func (acc *structFieldAccessor) fieldOf(obj interface{}) interface{} {
	structPtr := uintptr(extractPtrFromEmptyInterface(obj))
	structFieldPtr := structPtr + acc.field.Offset
	objEmptyInterface := acc.templateObj
	objEmptyInterface.word = unsafe.Pointer(structFieldPtr)
	return castBackEmptyInterface(objEmptyInterface)
}

type mapAccessor struct {
	accessor.NoopAccessor
}

func (acc *mapAccessor) Kind() reflect.Kind {
	return reflect.Map
}

func (acc *mapAccessor) IterateMap(obj interface{}, cb func(key interface{}, value interface{}) bool) {
	reflectVal := reflect.ValueOf(obj)
	for _, key := range reflectVal.MapKeys() {
		value := reflectVal.MapIndex(key)
		if !cb(key.Interface(), value.Interface()) {
			return
		}
	}
}

func (acc *mapAccessor) SetMapIndex(obj interface{}, key interface{}, value interface{}) {
	reflect.ValueOf(obj).SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
}