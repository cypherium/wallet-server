package rawdb

import (
	"encoding/json"

	"github.com/cypherium/go-cypherium/common"
	"github.com/cypherium/go-cypherium/log"
	"github.com/cypherium/go-cypherium/params"
	"github.com/cypherium/go-cypherium/rlp"
	"github.com/cypherium/go-cypherium/onet/network"
)

// ReadDatabaseVersion retrieves the version number of the database.
func ReadDatabaseVersion(db DatabaseReader) int {
	var version int

	enc, _ := db.Get(databaseVerisionKey)
	rlp.DecodeBytes(enc, &version)

	return version
}

// WriteDatabaseVersion stores the version number of the database
func WriteDatabaseVersion(db DatabaseWriter, version int) {
	enc, _ := rlp.EncodeToBytes(version)
	if err := db.Put(databaseVerisionKey, enc); err != nil {
		log.Crit("Failed to store the database version", "err", err)
	}
}

// ReadChainConfig retrieves the consensus settings based on the given genesis hash.
func ReadChainConfig(db DatabaseReader, hash common.Hash) *params.ChainConfig {
	data, _ := db.Get(configKey(hash))
	if len(data) == 0 {
		return nil
	}
	var config params.ChainConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Error("Invalid chain config JSON", "hash", hash, "err", err)
		return nil
	}
	return &config
}

// WriteChainConfig writes the chain config settings to the database.
func WriteChainConfig(db DatabaseWriter, hash common.Hash, cfg *params.ChainConfig) {
	if cfg == nil {
		return
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Crit("Failed to JSON encode chain config", "err", err)
	}
	if err := db.Put(configKey(hash), data); err != nil {
		log.Crit("Failed to store chain config", "err", err)
	}
}

// ReadPreimage retrieves a single preimage of the provided hash.
func ReadPreimage(db DatabaseReader, hash common.Hash) []byte {
	data, _ := db.Get(preimageKey(hash))
	return data
}

// WritePreimages writes the provided set of preimages to the database. `number` is the
// current block number, and is used for debug messages only.
func WritePreimages(db DatabaseWriter, number uint64, preimages map[common.Hash][]byte) {
	for hash, preimage := range preimages {
		if err := db.Put(preimageKey(hash), preimage); err != nil {
			log.Crit("Failed to store trie preimage", "err", err)
		}
	}
	preimageCounter.Inc(int64(len(preimages)))
	preimageHitCounter.Inc(int64(len(preimages)))
}

// ReadCommittee retrieves the committee.
func ReadCommittee(db DatabaseReader, keyBlockNumber uint64, hash common.Hash) network.Message {
	if db == nil {
		log.Error("ReadCommittee", "db is nil", "")
		return nil
	}
	data, _ := db.Get(CommitteeKey(keyBlockNumber, hash))
	if len(data) == 0 {
		//log.Error("ReadCommittee", "read data is empty", "")
		return nil
	}

	// For some reason boltdb changes the val before Unmarshal finishes. When
	// copying the value into a buffer, there is no SIGSEGV anymore.
	buf := make([]byte, len(data))
	copy(buf, data)
	_, sbMsg, err := network.Unmarshal(buf, network.EncSuite)
	if err != nil {
		log.Crit("Failed to read committee", "err", err)
	}

	// return sbMsg.(*committee.Committee).Copy()
	return sbMsg
}

// WriteCommittee writes the current committee to the database.
func WriteCommittee(db DatabaseWriter, keyBlockNumber uint64, hash common.Hash, committee network.Message) bool {
	if db == nil {
		log.Error("WriteCommittee", "db is nil", "")
		return false
	}
	if committee == nil {
		log.Warn("WriteCommittee:Try to store nil committee.")
		return false
	}
	val, err := network.Marshal(committee)
	if err != nil {
		log.Crit("Failed to store committee", "err", err)
		return false
	}
	if err := db.Put(CommitteeKey(keyBlockNumber, hash), val); err != nil {
		log.Crit("Failed to store committee", "err", err)
		return false
	}
	return true
}

// DeleteCommittee removes all committee data associated with a keyblock number and leader index.
func DeleteCommittee(db DatabaseDeleter, keyBlockNumber uint64, hash common.Hash) {
	if db == nil {
		log.Error("WriteCommittee", "db is nil", "")
		return
	}
	if err := db.Delete(CommitteeKey(keyBlockNumber, hash)); err != nil {
		log.Crit("Failed to delete committee", "err", err)
	}
}
