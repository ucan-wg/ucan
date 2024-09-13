package selector

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/schema"
)

// Selector describes a UCAN policy selector, as specified here:
// https://github.com/ucan-wg/delegation/blob/4094d5878b58f5d35055a3b93fccda0b8329ebae/README.md#selectors
type Selector []segment

func (s Selector) String() string {
	var res strings.Builder
	for _, seg := range s {
		res.WriteString(seg.String())
	}
	return res.String()
}

var Identity = segment{".", true, false, false, nil, "", 0}

var (
	indexRegex = regexp.MustCompile(`^-?\d+$`)
	sliceRegex = regexp.MustCompile(`^((\-?\d+:\-?\d*)|(\-?\d*:\-?\d+))$`)
	fieldRegex = regexp.MustCompile(`^\.[a-zA-Z_]*?$`)
)

type segment struct {
	str      string
	identity bool
	optional bool
	iterator bool
	slice    []int
	field    string
	index    int
}

// String returns the segment's string representation.
func (s segment) String() string {
	return s.str
}

// Identity flags that this selector is the identity selector.
func (s segment) Identity() bool {
	return s.identity
}

// Optional flags that this selector is optional.
func (s segment) Optional() bool {
	return s.optional
}

// Iterator flags that this selector is an iterator segment.
func (s segment) Iterator() bool {
	return s.iterator
}

// Slice flags that this segment targets a range of a slice.
func (s segment) Slice() []int {
	return s.slice
}

// Field is the name of a field in a struct/map.
func (s segment) Field() string {
	return s.field
}

// Index is an index of a slice.
func (s segment) Index() int {
	return s.index
}

// Select uses a selector to extract an IPLD node or set of nodes from the
// passed subject node.
func Select(sel Selector, subject ipld.Node) (ipld.Node, []ipld.Node, error) {
	return resolve(sel, subject, nil)
}

func resolve(sel Selector, subject ipld.Node, at []string) (ipld.Node, []ipld.Node, error) {
	cur := subject
	for i, seg := range sel {
		if seg.Identity() {
			continue
		} else if seg.Iterator() {
			if cur != nil && cur.Kind() == datamodel.Kind_List {
				var many []ipld.Node
				it := cur.ListIterator()
				for {
					if it.Done() {
						break
					}

					k, v, err := it.Next()
					if err != nil {
						return nil, nil, err
					}

					key := fmt.Sprintf("%d", k)
					o, m, err := resolve(sel[i+1:], v, append(at[:], key))
					if err != nil {
						return nil, nil, err
					}

					if m != nil {
						many = append(many, m...)
					} else {
						many = append(many, o)
					}
				}
				return nil, many, nil
			} else if cur != nil && cur.Kind() == datamodel.Kind_Map {
				var many []ipld.Node
				it := cur.MapIterator()
				for {
					if it.Done() {
						break
					}

					k, v, err := it.Next()
					if err != nil {
						return nil, nil, err
					}

					key, _ := k.AsString()
					o, m, err := resolve(sel[i+1:], v, append(at[:], key))
					if err != nil {
						return nil, nil, err
					}

					if m != nil {
						many = append(many, m...)
					} else {
						many = append(many, o)
					}
				}
				return nil, many, nil
			} else if seg.Optional() {
				cur = nil
			} else {
				return nil, nil, newResolutionError(fmt.Sprintf("can not iterate over kind: %s", kindString(cur)), at)
			}

		} else if seg.Field() != "" {
			at = append(at, seg.Field())
			if cur != nil && cur.Kind() == datamodel.Kind_Map {
				n, err := cur.LookupByString(seg.Field())
				if err != nil {
					if isMissing(err) {
						if seg.Optional() {
							cur = nil
						} else {
							return nil, nil, newResolutionError(fmt.Sprintf("object has no field named: %s", seg.Field()), at)
						}
					} else {
						return nil, nil, err
					}
				}
				cur = n
			} else if seg.Optional() {
				cur = nil
			} else {
				return nil, nil, newResolutionError(fmt.Sprintf("can not access field: %s on kind: %s", seg.Field(), kindString(cur)), at)
			}
		} else if seg.Slice() != nil {
			if cur != nil && cur.Kind() == datamodel.Kind_List {
				slice := seg.Slice()
				start, end := int64(0), cur.Length()

				if len(slice) > 0 {
					start = int64(slice[0])
					if start < 0 {
						start = cur.Length() + start
						if start < 0 {
							start = 0
						}
					}
				}

				if len(slice) > 1 {
					end = int64(slice[1])
					if end <= 0 {
						end = cur.Length() + end
						if end < start {
							end = start
						}
					}
				}

				if start < 0 || start >= cur.Length() || end < start || end > cur.Length() {
					if seg.Optional() {
						cur = nil
					} else {
						return nil, nil, newResolutionError(fmt.Sprintf("slice out of bounds: [%d:%d]", start, end), at)
					}
				} else {
					nb := basicnode.Prototype.List.NewBuilder()
					assembler, err := nb.BeginList(int64(end - start))
					if err != nil {
						return nil, nil, err
					}
					for i := start; i < end; i++ {
						item, err := cur.LookupByIndex(int64(i))
						if err != nil {
							return nil, nil, err
						}
						if err := assembler.AssembleValue().AssignNode(item); err != nil {
							return nil, nil, err
						}
					}
					if err := assembler.Finish(); err != nil {
						return nil, nil, err
					}
					cur = nb.Build()
				}
			} else if cur != nil && cur.Kind() == datamodel.Kind_Bytes {
				return nil, nil, newResolutionError("bytes slice selection not yet implemented", at)
			} else if seg.Optional() {
				cur = nil
			} else {
				return nil, nil, newResolutionError(fmt.Sprintf("can not index: %s on kind: %s", seg.Field(), kindString(cur)), at)
			}
		} else {
			at = append(at, fmt.Sprintf("%d", seg.Index()))
			if cur != nil && cur.Kind() == datamodel.Kind_List {
				idx := int64(seg.Index())
				if idx < 0 {
					idx = cur.Length() + idx
				}
				if idx < 0 {
					// necessary until https://github.com/ipld/go-ipld-prime/pull/571
					// after, isMissing() below will work
					// TODO: remove
					return nil, nil, newResolutionError(fmt.Sprintf("index out of bounds: %d", seg.Index()), at)
				}
				n, err := cur.LookupByIndex(idx)
				if err != nil {
					if isMissing(err) {
						if seg.Optional() {
							cur = nil
						} else {
							return nil, nil, newResolutionError(fmt.Sprintf("index out of bounds: %d", seg.Index()), at)
						}
					} else {
						return nil, nil, err
					}
				}
				cur = n
			} else if cur != nil && cur.Kind() == datamodel.Kind_String {
				str, err := cur.AsString()
				if err != nil {
					return nil, nil, err
				}
				idx := seg.Index()
				// handle negative indices by adjusting them to count from the end of the string
				if idx < 0 {
					idx = len(str) + idx
				}
				if idx < 0 || idx >= len(str) {
					if seg.Optional() {
						cur = nil
					} else {
						return nil, nil, newResolutionError(fmt.Sprintf("index out of bounds: %d", seg.Index()), at)
					}
				} else {
					cur = basicnode.NewString(string(str[idx]))
				}
			} else if cur != nil && cur.Kind() == datamodel.Kind_Bytes {
				b, err := cur.AsBytes()
				if err != nil {
					return nil, nil, err
				}
				idx := seg.Index()
				if idx < 0 {
					idx = len(b) + idx
				}
				if idx < 0 || idx >= len(b) {
					if seg.Optional() {
						cur = nil
					} else {
						return nil, nil, newResolutionError(fmt.Sprintf("index out of bounds: %d", seg.Index()), at)
					}
				} else {
					cur = basicnode.NewInt(int64(b[idx]))
				}
			} else if seg.Optional() {
				cur = nil
			} else {
				return nil, nil, newResolutionError(fmt.Sprintf("can not access field: %s on kind: %s", seg.Field(), kindString(cur)), at)
			}
		}
	}

	return cur, nil, nil
}

func kindString(n datamodel.Node) string {
	if n == nil {
		return "null"
	}
	return n.Kind().String()
}

func isMissing(err error) bool {
	if _, ok := err.(datamodel.ErrNotExists); ok {
		return true
	}
	if _, ok := err.(schema.ErrNoSuchField); ok {
		return true
	}
	if _, ok := err.(schema.ErrInvalidKey); ok {
		return true
	}
	return false
}

type resolutionerr struct {
	msg string
	at  []string
}

func (r resolutionerr) Name() string {
	return "ResolutionError"
}

func (r resolutionerr) Message() string {
	return fmt.Sprintf("can not resolve path: .%s", strings.Join(r.at, "."))
}

func (r resolutionerr) At() []string {
	return r.at
}

func (r resolutionerr) Error() string {
	return r.Message()
}

func newResolutionError(message string, at []string) error {
	return resolutionerr{message, at}
}
