package sharray

import (
	"context"
	"math"

	cid "github.com/ipfs/go-cid"
	hamt "github.com/ipfs/go-hamt-ipld"
)

type Sharray struct {
	node *node
	cst  hamt.CborIpldStore
}

type node struct {
	width  int
	Height int
	Cids   []cid.Cid
}

func Build(ctx context.Context, width int, cids []cid.Cid, cst hamt.CborIpldStore) (cid.Cid, error) {
	var next []cid.Cid

	for height := 0; len(cids) > 1 || height < 1; height++ {
		numSlices := (len(cids) + width - 1) / width
		for i := 0; i < numSlices; i++ {
			beg := width * i
			end := width * (i + 1)
			if end > len(cids) {
				end = len(cids)
			}

			nd := &Node{
				width:  width,
				Height: height,
				Cids:   cids[beg:end],
			}

			c, err := cst.Put(ctx, nd)
			if err != nil {
				return cid.Undef, err
			}

			next = append(next, c)
		}
		cids = next
		next = nil
	}

	return cids[0], nil
}

func Load(ctx context.Context, c cid.Cid, width int, cst hamt.CborIpldStore) (*Sharray, error) {
	var nd node
	if err := cst.Get(ctx, c, &nd); err != nil {
		return nil, err
	}

	return &Sharray{
		node:  *nd,
		cst:   cst,
		width: width,
	}
}

func nodesForHeight(width, height int) int {
	return int(math.Pow(float64(width), float64(height)))
}

func (s *Sharray) Len(ctx context.Context) (int, error) {
	if s.node.Height == 0 {
		return len(s.node.Cids), nil
	}

	countForFullNodes := (len(s.node.Cids) - 1) * nodesForHeight(s.node.Height-1)

	last, err := Load(ctx, s.node.Cids[len(s.node.Cids)-1], s.width, s.cst)
	if err != nil {
		return 0, err
	}

	lastCount, err := last.Len(ctx)
	if err != nil {
		return 0, err
	}

	return countForFullNodes + lastCount, nil
}

func (s *Sharray) ForEach(ctx context.Context, f func(cid.Cid)) error {
	for _, c := range s.node.Cids {
		if s.node.Height == 0 {
			f(c)
		} else {
			sub, err := Load(ctx, c, s.width, s.cst)
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
