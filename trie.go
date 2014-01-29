package ethutil

import (
	"fmt"
	"reflect"
)

/*
 * Trie helper functions
 */
// Helper function for printing out the raw contents of a slice
func PrintSlice(slice []string) {
	fmt.Printf("[")
	for i, val := range slice {
		fmt.Printf("%q", val)
		if i != len(slice)-1 {
			fmt.Printf(",")
		}
	}
	fmt.Printf("]\n")
}

func PrintSliceT(slice interface{}) {
	c := Conv(slice)
	for i := 0; i < c.Length(); i++ {
		val := c.Get(i)
		if val.Type() == reflect.Slice {
			PrintSliceT(val.AsRaw())
		} else {
			fmt.Printf("%q", val)
			if i != c.Length()-1 {
				fmt.Printf(",")
			}
		}
	}
}

// RLP Decodes a node in to a [2] or [17] string slice
func DecodeNode(data []byte) []string {
	dec, _ := Decode(data, 0)
	if slice, ok := dec.([]interface{}); ok {
		strSlice := make([]string, len(slice))

		for i, s := range slice {
			if str, ok := s.([]byte); ok {
				strSlice[i] = string(str)
			}
		}

		return strSlice
	} else {
		fmt.Printf("It wasn't a []. It's a %T\n", dec)
	}

	return nil
}

// A (modified) Radix Trie implementation
type Trie struct {
	Root interface{}
	db   Database
}

func NewTrie(db Database, Root interface{}) *Trie {
	return &Trie{db: db, Root: Root}
}

/*
 * Public (query) interface functions
 */
func (t *Trie) Update(key string, value string) {
	k := CompactHexDecode(key)

	t.Root = t.UpdateState(t.Root, k, value)
}

func (t *Trie) Get(key string) string {
	k := CompactHexDecode(key)
	c := Conv(t.GetStateT(t.Root, k))

	return c.AsString()
}

func (t *Trie) GetStateT(node interface{}, key []int) interface{} {
	n := Conv(node)
	// Return the node if key is empty (= found)
	if len(key) == 0 || n.IsNil() {
		return node
	}

	currentNode := t.GetNodeT(node)
	length := currentNode.Length()

	if length == 0 {
		return ""
	} else if length == 2 {
		// Decode the key
		k := CompactDecode(currentNode.Get(0).AsString())
		v := currentNode.Get(1).AsRaw()

		if len(key) >= len(k) && CompareIntSlice(k, key[:len(k)]) {
			return t.GetStateT(v, key[len(k):])
		} else {
			return ""
		}
	} else if length == 17 {
		return t.GetStateT(currentNode.Get(key[0]).AsRaw(), key[1:])
	}

	// It shouldn't come this far
	fmt.Println("GetState unexpected return")
	return ""
}

func (t *Trie) GetNodeT(node interface{}) *RlpValue {
	n := Conv(node)

	//if n.Type() != reflect.String {
	if !n.Get(0).IsNil() {
		return n
	}

	str := n.AsString()
	if len(str) == 0 {
		return n
	} else if len(str) < 32 {
		d, _ := Decode([]byte(str), 0)
		return Conv(d)
	} else {
		// Fetch the encoded node from the db
		o, err := t.db.Get(n.AsBytes())
		if err != nil {
			fmt.Println("Error InsertState", err)
			return Conv("")
		}

		d, _ := Decode(o, 0)
		return Conv(d)
	}

}

func (t *Trie) UpdateT(key string, value string) {
	k := CompactHexDecode(key)

	t.Root = t.UpdateState(t.Root, k, value)
}

func (t *Trie) UpdateState(node interface{}, key []int, value string) interface{} {
	if value != "" {
		return t.InsertState(node, key, value)
	} else {
		// delete it
	}

	return ""
}

func (t *Trie) Put(node interface{}) interface{} {
	enc := Encode(node)

	if len(enc) >= 32 {
		var sha []byte
		sha = Sha3Bin(enc)
		t.db.Put([]byte(sha), enc)

		return sha
	}

	/*
		TODO?
			c := Conv(t.Root)
			fmt.Println(c.Type(), c.Length())
			if c.Type() == reflect.String && c.AsString() == "" {
				return enc
			}
	*/

	return node
}

func EmptyStringSlice(l int) []interface{} {
	slice := make([]interface{}, l)
	for i := 0; i < l; i++ {
		slice[i] = ""
	}
	return slice
}

func (t *Trie) InsertState(node interface{}, key []int, value interface{}) interface{} {
	if len(key) == 0 {
		return value
	}

	// New node
	n := Conv(node)
	if node == nil || (n.Type() == reflect.String && (n.AsString() == "" || n.Get(0).IsNil())) {
		newNode := []interface{}{CompactEncode(key), value}

		return t.Put(newNode)
	}

	currentNode := t.GetNodeT(node)
	// Check for "special" 2 slice type node
	if currentNode.Length() == 2 {
		// Decode the key
		k := CompactDecode(currentNode.Get(0).AsString())
		v := currentNode.Get(1).AsRaw()

		// Matching key pair (ie. there's already an object with this key)
		if CompareIntSlice(k, key) {
			newNode := []interface{}{CompactEncode(key), value}
			return t.Put(newNode)
		}

		var newHash interface{}
		matchingLength := MatchingNibbleLength(key, k)
		if matchingLength == len(k) {
			// Insert the hash, creating a new node
			newHash = t.InsertState(v, key[matchingLength:], value)
		} else {
			// Expand the 2 length slice to a 17 length slice
			oldNode := t.InsertState("", k[matchingLength+1:], v)
			newNode := t.InsertState("", key[matchingLength+1:], value)
			// Create an expanded slice
			scaledSlice := EmptyStringSlice(17)
			// Set the copied and new node
			scaledSlice[k[matchingLength]] = oldNode
			scaledSlice[key[matchingLength]] = newNode

			newHash = t.Put(scaledSlice)
		}

		if matchingLength == 0 {
			// End of the chain, return
			return newHash
		} else {
			newNode := []interface{}{CompactEncode(key[:matchingLength]), newHash}
			return t.Put(newNode)
		}
	} else {
		// Copy the current node over to the new node and replace the first nibble in the key
		newNode := EmptyStringSlice(17)

		for i := 0; i < 17; i++ {
			cpy := currentNode.Get(i).AsRaw()
			if cpy != nil {
				newNode[i] = cpy
			}
		}

		newNode[key[0]] = t.InsertState(currentNode.Get(key[0]).AsRaw(), key[1:], value)

		return t.Put(newNode)
	}

	return ""
}

// Simple compare function which creates a rlp value out of the evaluated objects
func (t *Trie) Cmp(trie *Trie) bool {
	a := NewRlpValue(t.Root)
	b := NewRlpValue(t.Root)

	return a.Cmp(b)
}
