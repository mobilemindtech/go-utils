package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/mobilemindtech/go-utils/v2/optional"
)

// precisa de redis em localhost:6379 (sem sessionproviderconfig no app.conf,
// o client usa o endereço padrão); sem redis o teste é ignorado
func newTestCache(t *testing.T) *CacheService {
	srv := New(60000)
	if err := srv.rdb.Ping().Err(); err != nil {
		t.Skipf("redis indisponível: %v", err)
	}
	return srv
}

func testKey(name string) string {
	return fmt.Sprintf("go_utils_test_%v_%v", name, time.Now().UnixNano())
}

func TestMemoizeValUsaCache(t *testing.T) {
	srv := newTestCache(t)

	for _, tc := range []struct {
		name string
		run  func(key string, calls *int) any
		want any
	}{
		{"bool true", func(key string, calls *int) any {
			return MemoizeVal(srv, key, func() bool { *calls++; return true })
		}, true},
		{"bool false", func(key string, calls *int) any {
			return MemoizeVal(srv, key, func() bool { *calls++; return false })
		}, false},
		{"string", func(key string, calls *int) any {
			return MemoizeStr(srv, key, func() string { *calls++; return "valor com \"aspas\"" })
		}, "valor com \"aspas\""},
		{"int64", func(key string, calls *int) any {
			return MemoizeVal(srv, key, func() int64 { *calls++; return 42 })
		}, int64(42)},
		{"struct", func(key string, calls *int) any {
			type item struct{ Nome string }
			return MemoizeVal(srv, key, func() item { *calls++; return item{"a"} })
		}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key := testKey(tc.name)
			defer srv.Delete(key)

			calls := 0
			first := tc.run(key, &calls)
			second := tc.run(key, &calls)

			if calls != 1 {
				t.Fatalf("cacheable chamado %v vezes, esperado 1 (cache não usado)", calls)
			}
			if fmt.Sprint(first) != fmt.Sprint(second) {
				t.Fatalf("valor do cache %v diferente do original %v", second, first)
			}
			if tc.want != nil && second != tc.want {
				t.Fatalf("valor %v, esperado %v", second, tc.want)
			}
		})
	}
}

func TestGetVal(t *testing.T) {
	srv := newTestCache(t)
	key := testKey("getval")
	defer srv.Delete(key)

	if _, ok := srv.GetVal(key).(*optional.None); !ok {
		t.Fatalf("chave inexistente: esperado None, veio %T", srv.GetVal(key))
	}

	srv.Put(key, "abc")

	some, ok := srv.GetVal(key).(*optional.Some)
	if !ok {
		t.Fatalf("chave existente: esperado Some, veio %T", srv.GetVal(key))
	}
	if some.Item != "abc" {
		t.Fatalf("valor %v, esperado abc", some.Item)
	}
}
