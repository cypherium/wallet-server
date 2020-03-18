package bftview

import (
	"bytes"
	"io"

	"github.com/cypherium/go-cypherium/common"
	"github.com/cypherium/go-cypherium/crypto/sha3"
	"github.com/cypherium/go-cypherium/log"
	"github.com/cypherium/go-cypherium/rlp"
)

type View struct {
	TxNumber      uint64
	TxHash        common.Hash
	KeyNumber     uint64
	KeyHash       common.Hash
	CommitteeHash common.Hash
	LeaderIndex   uint
	ReconfigType  uint8
}

func (v *View) EqualAll(other *View) bool {
	return v.TxNumber == other.TxNumber && v.TxHash == other.TxHash && v.KeyNumber == other.KeyNumber && v.KeyHash == other.KeyHash && v.CommitteeHash == other.CommitteeHash && v.LeaderIndex == other.LeaderIndex && v.ReconfigType == other.ReconfigType
}
func (v *View) EqualNoIndex(other *View) bool {
	return v.TxNumber == other.TxNumber && v.TxHash == other.TxHash && v.KeyNumber == other.KeyNumber && v.KeyHash == other.KeyHash
}

func (v *View) Hash() (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, []interface{}{v.TxNumber, v.TxHash, v.KeyNumber, v.KeyHash, v.CommitteeHash, v.LeaderIndex, v.ReconfigType})
	hw.Sum(h[:0])
	return h
}

type ViewExt struct {
	TxNumber      uint64
	TxHash        common.Hash
	KeyNumber     uint64
	KeyHash       common.Hash
	CommitteeHash common.Hash
	LeaderIndex   uint
	ReconfigType  uint8
}

func (v *View) DecodeRLP(s *rlp.Stream) error {
	var eb ViewExt
	if err := s.Decode(&eb); err != nil {
		return err
	}
	v.KeyNumber, v.KeyHash, v.TxNumber, v.TxHash = eb.KeyNumber, eb.KeyHash, eb.TxNumber, eb.TxHash
	v.CommitteeHash, v.LeaderIndex, v.ReconfigType = eb.CommitteeHash, eb.LeaderIndex, eb.ReconfigType

	return nil
}

func (v *View) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, ViewExt{
		TxNumber:      v.TxNumber,
		TxHash:        v.TxHash,
		KeyNumber:     v.KeyNumber,
		KeyHash:       v.KeyHash,
		CommitteeHash: v.CommitteeHash,
		LeaderIndex:   v.LeaderIndex,
		ReconfigType:  v.ReconfigType,
	})
}

func (v *View) EncodeToBytes() []byte {
	m := make([]byte, 0)
	buff := bytes.NewBuffer(m)
	err := v.EncodeRLP(buff)
	if err != nil {
		return nil
	}

	return buff.Bytes()
}

func DecodeToView(data []byte) *View {
	v := &View{}
	buff := bytes.NewBuffer(data)
	c := rlp.NewStream(buff, 0)
	err := v.DecodeRLP(c)
	if err != nil {
		log.Error("DecodeToView", "error", err)
		return nil
	}
	return v
}
