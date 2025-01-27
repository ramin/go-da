package test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/rollkit/go-da"
)

// RunDATestSuite runs all tests against given DA
func RunDATestSuite(t *testing.T, d da.DA) {
	t.Run("Basic DA test", func(t *testing.T) {
		BasicDATest(t, d)
	})
	t.Run("Get IDs and all data", func(t *testing.T) {
		GetIDsTest(t, d)
	})
	t.Run("Check Errors", func(t *testing.T) {
		CheckErrors(t, d)
	})
}

// TODO(tzdybal): how to get rid of those aliases?

// Blob is a type alias
type Blob = da.Blob

// ID is a type alias
type ID = da.ID

// BasicDATest tests round trip of messages to DA and back.
func BasicDATest(t *testing.T, da da.DA) {
	msg1 := []byte("message 1")
	msg2 := []byte("message 2")

	id1, proof1, err := da.Submit([]Blob{msg1})
	assert.NoError(t, err)
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, proof1)

	id2, proof2, err := da.Submit([]Blob{msg2})
	assert.NoError(t, err)
	assert.NotEmpty(t, id2)
	assert.NotEmpty(t, proof2)

	id3, proof3, err := da.Submit([]Blob{msg1})
	assert.NoError(t, err)
	assert.NotEmpty(t, id3)
	assert.NotEmpty(t, proof3)

	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id1, id3)

	ret, err := da.Get(id1)
	assert.NoError(t, err)
	assert.Equal(t, []Blob{msg1}, ret)

	commitment1, err := da.Commit([]Blob{msg1})
	assert.NoError(t, err)
	assert.NotEmpty(t, commitment1)

	commitment2, err := da.Commit([]Blob{msg2})
	assert.NoError(t, err)
	assert.NotEmpty(t, commitment2)

	oks, err := da.Validate(id1, proof1)
	assert.NoError(t, err)
	assert.NotEmpty(t, oks)
	for _, ok := range oks {
		assert.True(t, ok)
	}

	oks, err = da.Validate(id2, proof2)
	assert.NoError(t, err)
	assert.NotEmpty(t, oks)
	for _, ok := range oks {
		assert.True(t, ok)
	}

	oks, err = da.Validate(id1, proof2)
	assert.NoError(t, err)
	assert.NotEmpty(t, oks)
	for _, ok := range oks {
		assert.False(t, ok)
	}

	oks, err = da.Validate(id2, proof1)
	assert.NoError(t, err)
	assert.NotEmpty(t, oks)
	for _, ok := range oks {
		assert.False(t, ok)
	}
}

// CheckErrors ensures that errors are handled properly by DA.
func CheckErrors(t *testing.T, da da.DA) {
	blob, err := da.Get([]ID{[]byte("invalid")})
	assert.Error(t, err)
	assert.Empty(t, blob)
}

// GetIDsTest tests iteration over DA
func GetIDsTest(t *testing.T, da da.DA) {
	msgs := [][]byte{[]byte("msg1"), []byte("msg2"), []byte("msg3")}

	ids, proofs, err := da.Submit(msgs)
	assert.NoError(t, err)
	assert.Len(t, ids, len(msgs))
	assert.Len(t, proofs, len(msgs))

	found := false
	end := time.Now().Add(1 * time.Second)

	// To Keep It Simple: we assume working with DA used exclusively for this test (mock, devnet, etc)
	// As we're the only user, we don't need to handle external data (that could be submitted in real world).
	// There is no notion of height, so we need to scan the DA to get test data back.
	for i := uint64(1); !found && !time.Now().After(end); i++ {
		ret, err := da.GetIDs(i)
		if err != nil {
			t.Error("failed to get IDs:", err)
		}
		if len(ret) > 0 {
			blobs, err := da.Get(ret)
			assert.NoError(t, err)

			// Submit ensures atomicity of batch, so it makes sense to compare actual blobs (bodies) only when lengths
			// of slices is the same.
			if len(blobs) == len(msgs) {
				found = true
				for b := 0; b < len(blobs); b++ {
					if !bytes.Equal(blobs[b], msgs[b]) {
						found = false
					}
				}
			}
		}
	}

	assert.True(t, found)
}
