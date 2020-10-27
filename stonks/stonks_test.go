package stonks

import (
	"fmt"
	"io/ioutil"
	"testing"

	"encoding/json"
)

func TestStonks(t *testing.T) {

	b, err := ioutil.ReadFile("/Users/mrod/go/src/github.com/next-alpha/stonks/ha16p4.json")

	if err != nil {
		fmt.Println(err)
		return
	}

	var obj JSONResponse

	err = json.Unmarshal(b, &obj)
	if err != nil {
		t.Fatal(err.Error())
	}
	b, _ = json.Marshal(obj)

	for i, c := range obj[1].GetMore() {

		fmt.Println(i, c.Children)

	}
}
