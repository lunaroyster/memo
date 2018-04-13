package db

import (
	"bytes"
	"git.jasonc.me/main/bitcoin/bitcoin/script"
	"git.jasonc.me/main/bitcoin/bitcoin/wallet"
	"github.com/btcsuite/btcutil"
	"github.com/cpacia/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"html"
	"sort"
	"time"
)

type MemoLike struct {
	Id         uint   `gorm:"primary_key"`
	TxHash     []byte `gorm:"unique;size:50"`
	ParentHash []byte
	PkHash     []byte
	PkScript   []byte
	Address    string
	LikeTxHash []byte
	TipAmount  int64
	TipPkHash  []byte
	BlockId    uint
	Block      *Block
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (m MemoLike) Save() error {
	result := save(&m)
	if result.Error != nil {
		return jerr.Get("error saving memo test", result.Error)
	}
	return nil
}

func (m MemoLike) GetTransactionHashString() string {
	hash, err := chainhash.NewHash(m.TxHash)
	if err != nil {
		jerr.Get("error getting chainhash from memo post", err).Print()
		return ""
	}
	return hash.String()
}

func (m MemoLike) GetAddressString() string {
	pkHash, err := btcutil.NewAddressPubKeyHash(m.PkHash, &wallet.MainNetParamsOld)
	if err != nil {
		jerr.Get("error getting pubkeyhash from memo post", err).Print()
		return ""
	}
	return pkHash.EncodeAddress()
}

func (m MemoLike) GetScriptString() string {
	return html.EscapeString(script.GetScriptString(m.PkScript))
}

func (m MemoLike) GetTimeString() string {
	if m.BlockId != 0 {
		return m.Block.Timestamp.Format("2006-01-02 15:04:05")
	}
	return "Unconfirmed"
}

func GetMemoLike(txHash []byte) (*MemoLike, error) {
	var memoLike MemoLike
	err := find(&memoLike, MemoLike{
		TxHash: txHash,
	})
	if err != nil {
		return nil, jerr.Get("error getting memo post", err)
	}
	return &memoLike, nil
}

type memoLikeSortByDate []*MemoLike

func (txns memoLikeSortByDate) Len() int      { return len(txns) }
func (txns memoLikeSortByDate) Swap(i, j int) { txns[i], txns[j] = txns[j], txns[i] }
func (txns memoLikeSortByDate) Less(i, j int) bool {
	if bytes.Equal(txns[i].ParentHash, txns[j].TxHash) {
		return false
	}
	if bytes.Equal(txns[i].TxHash, txns[j].ParentHash) {
		return true
	}
	if txns[i].Block == nil && txns[j].Block == nil {
		return false
	}
	if txns[i].Block == nil {
		return false
	}
	if txns[j].Block == nil {
		return true
	}
	return txns[i].Block.Height < txns[j].Block.Height
}

func GetLikesForTxnHash(txHash []byte) ([]*MemoLike, error) {
	var memoLikes []*MemoLike
	err := findPreloadColumns([]string{
		BlockTable,
	}, &memoLikes, &MemoLike{
		LikeTxHash: txHash,
	})
	if err != nil {
		return nil, jerr.Get("error getting memo likes", err)
	}
	sort.Sort(memoLikeSortByDate(memoLikes))
	return memoLikes, nil
}