package crm

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type blankOptions map[string]string

func (o blankOptions) encodeURL(u *url.URL) error {
	val := url.Values{}
	for k, v := range o {
		val.Add(k, v)
	}
	u.RawQuery = val.Encode()
	return nil
}

type optionEncoder interface {
	encodeURL(*url.URL) error
}

func encodeOptionsToURL(o optionEncoder, u *url.URL) error {
	vals := url.Values{}
	v := reflect.ValueOf(o)

	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	//PRIME:
	for i := 0; i < v.NumField(); i++ {
		name := v.Type().Field(i).Name
		value := v.Field(i).Interface()
		tag := v.Type().Field(i).Tag.Get("zoho")

		//split the tag on comma
		tags := strings.Split(tag, ",")
		//	TAGS:
		for _, a := range tags {
			if a == "required" {
				//if the underlying value of the field is empty
				if reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface()) {
					return fmt.Errorf("Error field '%s' is required for this request", name)
				}
			} else if strings.HasPrefix(a, "default>") {
				//if the underlying value of the field is empty
				if reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface()) {
					//Set the 'default' value for the URL item
					value = strings.TrimPrefix(a, "default>")
				}
			} else {
				//if the underlying value of the field is NOT empty
				if !reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface()) {
					//Set the 'default' value for the URL item
					name = a
				}
			}
		}

		//If the field is set, if field is false check default tag
		switch v := value.(type) {
		case bool:
			//encode to the URL
			vals.Set(name, fmt.Sprintf("%t", v))
		case int:
			vals.Set(name, fmt.Sprintf("%d", v))
		case float64:
			vals.Set(name, fmt.Sprintf("%f", v))
		case string:
			vals.Set(name, fmt.Sprintf("%s", v))
		case time.Time:
			tm := time.Time(v)
			if !tm.IsZero() {
				vals.Set(name, fmt.Sprintf("%d-%d-%d %d:%d:%d",
					v.Year(), v.Month(), v.Day(),
					v.Hour(), v.Minute(), v.Second(),
				))
			}
		case crmData:
			//get the items XML
			vals.Set(name, v.writeXML())
		}
	}

	vals.Set("version", "2")
	vals.Set("newFormat", "1")

	u.RawQuery = vals.Encode()
	return nil
}
