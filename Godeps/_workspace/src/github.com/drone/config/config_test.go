package config

import (
	"flag"
	"testing"
)

var config = []byte(`
# Global vars
my_bool = true
my_int = 22
my_bigint = -23
my_uint = 24
my_biguint = 25
my_string = "ok"
my_bigfloat = 26.1

# A String Array
cities = ["erie","dayton","stamford","scottsdale","sf"]

# A config section
[section]
name = "cool dude"

# A deep section
[places.california]
name = "neat dude"

`)

func testParse(t *testing.T, c *ConfigSet) {

	boolSetting := c.Bool("my-bool", false)
	intSetting := c.Int("my-int", 0)
	int64Setting := c.Int64("my-bigint", 0)
	uintSetting := c.Uint("my-uint", 0)
	uint64Setting := c.Uint64("my-biguint", 0)
	stringSetting := c.String("my-string", "nope")
	float64Setting := c.Float64("my-bigfloat", 0)
	nestedSetting := c.String("section-name", "")
	deepNestedSetting := c.String("places-california-name", "")
	cities := c.Strings("cities")

	err := c.ParseBytes(config)
	if err != nil {
		t.Fatal(err)
	}

	if *boolSetting != true {
		t.Error("bool setting should be true, is", *boolSetting)
	}
	if *intSetting != 22 {
		t.Error("int setting should be 22, is", *intSetting)
	}
	if *int64Setting != int64(-23) {
		t.Error("int64 setting should be -23, is", *int64Setting)
	}
	if *uintSetting != 24 {
		t.Error("uint setting should be 24, is", *uintSetting)
	}
	if *uint64Setting != uint64(25) {
		t.Error("uint64 setting should be 25, is", *uint64Setting)
	}
	if *stringSetting != "ok" {
		t.Error("string setting should be \"ok\", is", *stringSetting)
	}
	if *float64Setting != float64(26.1) {
		t.Error("float64 setting should be 26.1, is", *float64Setting)
	}
	if *nestedSetting != "cool dude" {
		t.Error("nested setting should be \"cool dude\", is", *nestedSetting)
	}
	if *deepNestedSetting != "neat dude" {
		t.Error("deepNested setting should be \"neat dude\", is", *deepNestedSetting)
	}
	if len(*cities) != 5 {
		t.Error("string array should have 5 items, is", cities)
	}
}

func TestParse(t *testing.T) {
	testParse(t, globalConfig)
	testParse(t, NewConfigSet("App Config", flag.ExitOnError))
}
