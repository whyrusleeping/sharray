package sharray

import (
	"context"
	"testing"

	hamt "github.com/ipfs/go-hamt-ipld"
	cbor "github.com/ipfs/go-ipld-cbor"
)

type thing struct {
	Foo int
	Bar string
}

func TestSharray(t *testing.T) {
	cbor.RegisterCborType(thing{})
	ctx := context.Background()
	cst := hamt.NewCborStore()

	var items []interface{}
	for i := 0; i < 20; i++ {
		items = append(items, &thing{
			Foo: i,
			Bar: "catdog",
		})
	}

	root, err := Build(ctx, 2, items, cst)
	if err != nil {
		t.Fatal(err)
	}

	sh, err := Load(ctx, root, 2, cst)
	if err != nil {
		t.Fatal(err)
	}

	if err := sh.ForEach(ctx, func(i interface{}) error {
		t.Log("item: ", i)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
