package main

import (
	"testing"
	"reflect"
	)

func TestRemoveByIndex(t *testing.T) {
	var actualSlice []adresse
	var expectedSlice []adresse
	actualSlice = append(actualSlice, adresse{ip:"191.168.48.1", port:5998})
	actualSlice = append(actualSlice, adresse{ip:"192.167.47.2", port:5997})
	actualSlice = append(actualSlice, adresse{ip:"193.163.43.3", port:5993})
	actualSlice = removeIndex(actualSlice, 2)
	expectedSlice = []adresse{{ip:"191.168.48.1", port:5998}, {ip:"193.163.43.3", port:5997}}
	if reflect.DeepEqual(actualSlice, expectedSlice){
		t.Errorf("Expected Slice %v is not the same as actual slice %v", expectedSlice, actualSlice)
	}
}
