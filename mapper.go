package mapper

import (
	"errors"
	"fmt"
	"reflect"
)

type Configuration struct {
	store map[string]map[string]func(src interface{}) interface{}
}

func NewConfiguration() *Configuration {
	return &Configuration{
		store: make(map[string]map[string]func(src interface{}) interface{}),
	}
}

func Add[SourceType interface{}, TargetType interface{}](config *Configuration, mp map[string]func(src SourceType) any) error {

	var TargetTypeObj TargetType
	var SourceTypeObj SourceType
	target := reflect.TypeOf(TargetTypeObj).Name()
	source := reflect.TypeOf(SourceTypeObj).Name()

	mp1 := make(map[string]func(src interface{}) interface{})

	for k, v := range mp {
		mp1[k] = func(src interface{}) interface{} {
			return v(src.(SourceType))
		}
	}

	config.store[source+" | "+target] = mp1
	return nil
}

func Map[TargetType interface{}](src interface{}, config *Configuration) (TargetType, error) {

	var target TargetType

	mp := config.store[reflect.TypeOf(src).Name()+" | "+reflect.TypeOf(target).Name()]

	target1, err := mapp(reflect.TypeOf(target), src, mp)

	if err != nil {
		return target, err
	}

	return target1.(TargetType), nil
}

func mapp(targetType reflect.Type, src interface{}, mp map[string]func(src interface{}) interface{}) (interface{}, error) {

	givenTargetType := targetType

	srcValueObj := reflect.ValueOf(src)

	if srcValueObj.Kind() == reflect.Ptr {
		if srcValueObj.IsNil() {
			return nil, nil
		}
		srcValueObj = srcValueObj.Elem()
	}

	// srcTypeObj := srcValueObj.Type()

	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	targetObject := reflect.New(targetType).Elem()

	if srcValueObj.Kind() != reflect.Struct {
		return targetObject.Interface(), errors.New("src must be a struct")
	}

	if targetObject.Kind() != reflect.Struct {
		return targetObject.Interface(), errors.New("target type must be a struct")
	}

	for i := 0; i < targetObject.NumField(); i++ {

		targetField := targetObject.Field(i)
		targetFieldName := targetType.Field(i).Name

		if !targetField.IsValid() || !targetField.CanSet() {
			continue
		}

		if v, ok := mp[targetFieldName]; ok {
			targetField.Set(reflect.ValueOf(v(src)))
			continue
		}

		sourceField := srcValueObj.FieldByName(targetFieldName)

		if !sourceField.IsValid() {
			continue
		}

		if sourceField.Kind() == reflect.Ptr {
			if sourceField.IsNil() {
				continue
			}
			sourceField = sourceField.Elem()
		}

		if sourceField.Kind() == reflect.Struct {
			nestedObjectType := targetField.Type()
			nestedObject, _ := mapp(nestedObjectType, srcValueObj.FieldByName(targetFieldName).Interface(), mp)
			targetField.Set(reflect.ValueOf(nestedObject))
			continue
		}

		targetField.Set(sourceField.Convert(targetField.Type()))
	}

	fmt.Println(targetObject)

	if givenTargetType.Kind() == reflect.Ptr {
		value := targetObject.Addr()
		return value.Interface(), nil
	}

	return targetObject.Interface(), nil
}
