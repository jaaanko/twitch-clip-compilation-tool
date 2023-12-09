package httpext_test

import (
	"net/http"
	"testing"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/httpext"
)

type testClient struct {
	called bool
}

func (tc *testClient) Do(r *http.Request) (*http.Response, error) {
	tc.called = true
	return nil, nil
}

func TestDecorator(t *testing.T) {
	type logger struct {
		called bool
	}

	wrapper := func(l *logger) httpext.Decorator {
		return func(c httpext.Client) httpext.Client {
			return httpext.ClientFunc(func(req *http.Request) (*http.Response, error) {
				l.called = true
				return c.Do(req)
			})
		}
	}

	l := &logger{}
	tc := &testClient{}
	client := httpext.Decorate(tc, wrapper(l))
	client.Do(nil)

	if !l.called {
		t.Fatal("decorator not called")
	}

	if !tc.called {
		t.Fatal("client Do() method not called")
	}
}
