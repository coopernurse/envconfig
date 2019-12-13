// +build go1.8

package envconfig

import (
	"errors"
	"net/url"
	"os"
	"testing"
)

type SpecWithURL struct {
	URLValue   url.URL
	URLPointer *url.URL
}

func TestParseURL(t *testing.T) {
	var s SpecWithURL

	os.Clearenv()
	os.Setenv("ENV_CONFIG_URLVALUE", "https://github.com/kelseyhightower/envconfig")
	os.Setenv("ENV_CONFIG_URLPOINTER", "https://github.com/kelseyhightower/envconfig")

	err := Process("env_config", &s)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	u, err := url.Parse("https://github.com/kelseyhightower/envconfig")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.URLValue != *u {
		t.Errorf("expected %q, got %q", u, s.URLValue.String())
	}

	if *s.URLPointer != *u {
		t.Errorf("expected %q, got %q", u, s.URLPointer)
	}
}

func TestParseURLError(t *testing.T) {
	var s SpecWithURL

	os.Clearenv()
	os.Setenv("ENV_CONFIG_URLPOINTER", "http_://foo")

	err := Process("env_config", &s)

	v, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected ParseError, got %T %v", err, err)
	}
	if v.FieldName != "URLPointer" {
		t.Errorf("expected %s, got %v", "URLPointer", v.FieldName)
	}

	expectedUnerlyingError := url.Error{
		Op:  "parse",
		URL: "http_://foo",
		Err: errors.New("first path segment in URL cannot contain colon"),
	}

	if v.Err.Error() != expectedUnerlyingError.Error() {
		t.Errorf("expected %q, got %q", expectedUnerlyingError, v.Err)
	}
}
