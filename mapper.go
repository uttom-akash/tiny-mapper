package mapper

import (
	"errors"
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

func Add[SourceType interface{}, TargetType interface{}](config *Configuration, stronglyTypedMap map[string]func(src SourceType) any) error {

	var TargetTypeObj TargetType
	var SourceTypeObj SourceType
	targetType := reflect.TypeOf(TargetTypeObj).Name()
	sourceType := reflect.TypeOf(SourceTypeObj).Name()

	genericMap := make(map[string]func(src interface{}) interface{})

	for key, mapFunc := range stronglyTypedMap {
		genericMap[key] = func(src interface{}) interface{} {
			return mapFunc(src.(SourceType))
		}
	}

	key:= createKey(sourceType, targetType)

	config.store[key] = genericMap
	return nil
}

func Map[TargetType interface{}](src interface{}, config *Configuration) (TargetType, error) {

	var dummyTarget TargetType

	target, err := internal_map(reflect.TypeOf(dummyTarget), src, config)

	if err != nil {
		return target.Interface().(TargetType), err
	}

	return target.Interface().(TargetType), nil
}

func internal_map(targetType reflect.Type, source interface{}, config *Configuration) (reflect.Value, error) {

	srcReflectValue := reflect.ValueOf(source)

	if srcReflectValue.Kind() == reflect.Ptr {
		if srcReflectValue.IsNil() {
			return reflect.Value{}, nil
		}
		srcReflectValue = srcReflectValue.Elem()

		return internal_map(targetType, srcReflectValue.Interface(), config)
	}

	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()

		target, err := internal_map(targetType, srcReflectValue.Interface(), config)

		if err != nil {
			return target, err
		}

		return target.Addr(), nil
	}

	if srcReflectValue.Kind() == reflect.Slice {

		targetElementType := targetType.Elem()

		targetSlice := reflect.MakeSlice(reflect.SliceOf(targetElementType), srcReflectValue.Len(), srcReflectValue.Len())

		for i := 0; i < srcReflectValue.Len(); i++ {

			target, err := internal_map(targetElementType, srcReflectValue.Index(i).Interface(), config)

			if err != nil {
				return reflect.Zero(targetElementType), err
			}
			targetSlice.Index(i).Set(target)
		}

		return targetSlice, nil
	}

	if srcReflectValue.Kind() == reflect.Struct {

		targetReflectedValue := reflect.New(targetType).Elem()

		key := createKey(srcReflectValue.Type().Name(), targetType.Name())

		mp := config.store[key]

		for index := 0; index < targetReflectedValue.NumField(); index++ {

			targetField := targetReflectedValue.Field(index)
			targetFieldName := targetType.Field(index).Name

			if !targetField.IsValid() || !targetField.CanSet() {
				continue
			}

			if customMapFunc, ok := mp[targetFieldName]; ok {
				targetField.Set(reflect.ValueOf(customMapFunc(source)))
				continue
			}

			sourceField := srcReflectValue.FieldByName(targetFieldName)

			if !sourceField.IsValid() {
				continue
			}

			mappedField, err := internal_map(targetField.Type(), srcReflectValue.FieldByName(targetFieldName).Interface(), config)
			
			if err != nil {
				return reflect.Value{}, err	
			}

			targetField.Set(mappedField)
		}

		return targetReflectedValue, nil
	}
	if srcReflectValue.Kind() == reflect.Map {
		return reflect.Value{}, errors.New("unsupported type")
	}

	sourceField := reflect.ValueOf(source)

	return sourceField.Convert(targetType), nil
}

func createKey(sourceType, targetType string) string {
	return sourceType + " | " + targetType
}