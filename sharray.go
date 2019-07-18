package sharray

import (
	"context"
	"fmt"
	"math"

	cid "github.com/ipfs/go-cid"
	hamt "github.com/ipfs/go-hamt-ipld"
	cbor "github.com/ipfs/go-ipld-cbor"
)

func init() {
	cbor.RegisterCborType(node{})

}

var ErrNotCid = fmt.Errorf("item in intermediate sharray node was not a Cid")

type Sharray struct {
	node  *node
	cst   *hamt.CborIpldStore
	width int
}

type node struct {
	Height int
	Items  []interface{}
}

func Build(ctx context.Context, width int, items []interface{}, cst *hamt.CborIpldStore) (cid.Cid, error) {
	var next []interface{}

	if items == nil {
		items = []interface{}{}
	}

	for height := 0; len(items) > 1 || height < 1; height++ {
		numSlices := (len(items) + width - 1) / width
		// the i == 0 condition allows us to create an 'empty' sharray
		for i := 0; i < numSlices || i == 0; i++ {
			beg := width * i
			end := width * (i + 1)
			if end > len(items) {
				end = len(items)
			}

			nd := &node{
				Height: height,
				Items:  items[beg:end],
			}

			c, err := cst.Put(ctx, nd)
			if err != nil {
				return cid.Undef, err
			}

			next = append(next, c)
		}
		items = next
		next = nil
	}

	return items[0].(cid.Cid), nil
}

func Load(ctx context.Context, c cid.Cid, width int, cst *hamt.CborIpldStore) (*Sharray, error) {
	var nd node
	if err := cst.Get(ctx, c, &nd); err != nil {
		return nil, err
	}

	return &Sharray{
		node:  &nd,
		cst:   cst,
		width: width,
	}, nil
}

func nodesForHeight(width, height int) int {
	return int(math.Pow(float64(width), float64(height)))
}

func (s *Sharray) Len(ctx context.Context) (int, error) {
	if s.node.Height == 0 {
		return len(s.node.Items), nil
	}

	countForFullNodes := (len(s.node.Items) - 1) * nodesForHeight(s.width, s.node.Height-1)

	icid, ok := s.node.Items[len(s.node.Items)-1].(cid.Cid)
	if !ok {
		return 0, ErrNotCid
	}

	last, err := Load(ctx, icid, s.width, s.cst)
	if err != nil {
		return 0, err
	}

	lastCount, err := last.Len(ctx)
	if err != nil {
		return 0, err
	}

	return countForFullNodes + lastCount, nil
}

func (s *Sharray) ForEach(ctx context.Context, f func(interface{}) error) error {
	for _, c := range s.node.Items {
		if s.node.Height == 0 {
			if err := f(c); err != nil {
				return err
			}
		} else {
			icid, ok := c.(cid.Cid)
			if !ok {
				return ErrNotCid
			}

			sub, err := Load(ctx, icid, s.width, s.cst)
			if err != nil {
				return err
			}

			if err := sub.ForEach(ctx, f); err != nil {
				return err
			}
		}
	}

	return nil
}

var ErrOutOfRange = fmt.Errorf("out of range")

func (s *Sharray) Get(ctx context.Context, i int) (interface{}, error) {
	if s.node.Height == 0 {
		if len(s.node.Items) <= i {
			return nil, ErrOutOfRange
		}
		return s.node.Items[i], nil
	}

	nfh := nodesForHeight(s.width, s.node.Height)
	subi := i / nfh
	if subi >= len(s.node.Items) {
		return nil, ErrOutOfRange
	}

	icid, ok := s.node.Items[subi].(cid.Cid)
	if !ok {
		return nil, ErrNotCid
	}

	sub, err := Load(ctx, icid, s.width, s.cst)
	if err != nil {
		return nil, err
	}

	return sub.Get(ctx, i%nfh)
}
