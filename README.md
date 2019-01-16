# Sharray
A sharded array, on ipld.

Goals:
- datastructure for a list of CIDs with minimal overhead
- efficient merkle proofs
- fast

Non-goals:
- Easy random insertions
- Supporting removal of items

## Structure

Each sharray has a given width. Each of the leaf nodes contains at most that
many elements, and only the final node may have fewer. When the number of leaf
nodes grows to be more than the given width, an additional layer is added. Each
node has a field specifying which layer the node is in the tree.

```
Single Layer:
[ H:0; 1, 2, 3, 4]


Two Layer:
[ H:1; A, B ] 

A = [ H:0; 1, 2, 3, 4], B = [ H:0; 5, 6, 7, 8]

Three Layer:

[ H:2; X, Y]

X = [ H:1; A, B, C, D], Y = [ H:1, E, F, G, H]

A = [ H:0; 1, 2, 3, 4] .... (and so on)
```

#### License
MIT
