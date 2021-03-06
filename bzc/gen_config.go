// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package bzc

import (
	"math/big"

	"github.com/bazacoin/go-bazacoin/common"
	"github.com/bazacoin/go-bazacoin/common/hexutil"
	"github.com/bazacoin/go-bazacoin/core"
	"github.com/bazacoin/go-bazacoin/bzc/downloader"
	"github.com/bazacoin/go-bazacoin/bzc/gasprice"
)

func (c Config) MarshalTOML() (interface{}, error) {
	type Config struct {
		Genesis                 *core.Genesis `toml:",omitempty"`
		NetworkId               uint64
		SyncMode                downloader.SyncMode
		LightServ               int  `toml:",omitempty"`
		LightPeers              int  `toml:",omitempty"`
		MaxPeers                int  `toml:"-"`
		SkipBcVersionCheck      bool `toml:"-"`
		DatabaseHandles         int  `toml:"-"`
		DatabaseCache           int
		Bazacoinbase               common.Address `toml:",omitempty"`
		MinerThreads            int            `toml:",omitempty"`
		ExtraData               hexutil.Bytes  `toml:",omitempty"`
		GasPrice                *big.Int
		BzhashCacheDir          string
		BzhashCachesInMem       int
		BzhashCachesOnDisk      int
		BzhashDatasetDir        string
		BzhashDatasetsInMem     int
		BzhashDatasetsOnDisk    int
		TxPool                  core.TxPoolConfig
		GPO                     gasprice.Config
		EnablePreimageRecording bool
		DocRoot                 string `toml:"-"`
		PowFake                 bool   `toml:"-"`
		PowTest                 bool   `toml:"-"`
		PowShared               bool   `toml:"-"`
	}
	var enc Config
	enc.Genesis = c.Genesis
	enc.NetworkId = c.NetworkId
	enc.SyncMode = c.SyncMode
	enc.LightServ = c.LightServ
	enc.LightPeers = c.LightPeers
	enc.MaxPeers = c.MaxPeers
	enc.SkipBcVersionCheck = c.SkipBcVersionCheck
	enc.DatabaseHandles = c.DatabaseHandles
	enc.DatabaseCache = c.DatabaseCache
	enc.Bazacoinbase = c.Bazacoinbase
	enc.MinerThreads = c.MinerThreads
	enc.ExtraData = c.ExtraData
	enc.GasPrice = c.GasPrice
	enc.BzhashCacheDir = c.BzhashCacheDir
	enc.BzhashCachesInMem = c.BzhashCachesInMem
	enc.BzhashCachesOnDisk = c.BzhashCachesOnDisk
	enc.BzhashDatasetDir = c.BzhashDatasetDir
	enc.BzhashDatasetsInMem = c.BzhashDatasetsInMem
	enc.BzhashDatasetsOnDisk = c.BzhashDatasetsOnDisk
	enc.TxPool = c.TxPool
	enc.GPO = c.GPO
	enc.EnablePreimageRecording = c.EnablePreimageRecording
	enc.DocRoot = c.DocRoot
	enc.PowFake = c.PowFake
	enc.PowTest = c.PowTest
	enc.PowShared = c.PowShared
	return &enc, nil
}

func (c *Config) UnmarshalTOML(unmarshal func(interface{}) error) error {
	type Config struct {
		Genesis                 *core.Genesis `toml:",omitempty"`
		NetworkId               *uint64
		SyncMode                *downloader.SyncMode
		LightServ               *int  `toml:",omitempty"`
		LightPeers              *int  `toml:",omitempty"`
		MaxPeers                *int  `toml:"-"`
		SkipBcVersionCheck      *bool `toml:"-"`
		DatabaseHandles         *int  `toml:"-"`
		DatabaseCache           *int
		Bazacoinbase               *common.Address `toml:",omitempty"`
		MinerThreads            *int            `toml:",omitempty"`
		ExtraData               hexutil.Bytes   `toml:",omitempty"`
		GasPrice                *big.Int
		BzhashCacheDir          *string
		BzhashCachesInMem       *int
		BzhashCachesOnDisk      *int
		BzhashDatasetDir        *string
		BzhashDatasetsInMem     *int
		BzhashDatasetsOnDisk    *int
		TxPool                  *core.TxPoolConfig
		GPO                     *gasprice.Config
		EnablePreimageRecording *bool
		DocRoot                 *string `toml:"-"`
		PowFake                 *bool   `toml:"-"`
		PowTest                 *bool   `toml:"-"`
		PowShared               *bool   `toml:"-"`
	}
	var dec Config
	if err := unmarshal(&dec); err != nil {
		return err
	}
	if dec.Genesis != nil {
		c.Genesis = dec.Genesis
	}
	if dec.NetworkId != nil {
		c.NetworkId = *dec.NetworkId
	}
	if dec.SyncMode != nil {
		c.SyncMode = *dec.SyncMode
	}
	if dec.LightServ != nil {
		c.LightServ = *dec.LightServ
	}
	if dec.LightPeers != nil {
		c.LightPeers = *dec.LightPeers
	}
	if dec.MaxPeers != nil {
		c.MaxPeers = *dec.MaxPeers
	}
	if dec.SkipBcVersionCheck != nil {
		c.SkipBcVersionCheck = *dec.SkipBcVersionCheck
	}
	if dec.DatabaseHandles != nil {
		c.DatabaseHandles = *dec.DatabaseHandles
	}
	if dec.DatabaseCache != nil {
		c.DatabaseCache = *dec.DatabaseCache
	}
	if dec.Bazacoinbase != nil {
		c.Bazacoinbase = *dec.Bazacoinbase
	}
	if dec.MinerThreads != nil {
		c.MinerThreads = *dec.MinerThreads
	}
	if dec.ExtraData != nil {
		c.ExtraData = dec.ExtraData
	}
	if dec.GasPrice != nil {
		c.GasPrice = dec.GasPrice
	}
	if dec.BzhashCacheDir != nil {
		c.BzhashCacheDir = *dec.BzhashCacheDir
	}
	if dec.BzhashCachesInMem != nil {
		c.BzhashCachesInMem = *dec.BzhashCachesInMem
	}
	if dec.BzhashCachesOnDisk != nil {
		c.BzhashCachesOnDisk = *dec.BzhashCachesOnDisk
	}
	if dec.BzhashDatasetDir != nil {
		c.BzhashDatasetDir = *dec.BzhashDatasetDir
	}
	if dec.BzhashDatasetsInMem != nil {
		c.BzhashDatasetsInMem = *dec.BzhashDatasetsInMem
	}
	if dec.BzhashDatasetsOnDisk != nil {
		c.BzhashDatasetsOnDisk = *dec.BzhashDatasetsOnDisk
	}
	if dec.TxPool != nil {
		c.TxPool = *dec.TxPool
	}
	if dec.GPO != nil {
		c.GPO = *dec.GPO
	}
	if dec.EnablePreimageRecording != nil {
		c.EnablePreimageRecording = *dec.EnablePreimageRecording
	}
	if dec.DocRoot != nil {
		c.DocRoot = *dec.DocRoot
	}
	if dec.PowFake != nil {
		c.PowFake = *dec.PowFake
	}
	if dec.PowTest != nil {
		c.PowTest = *dec.PowTest
	}
	if dec.PowShared != nil {
		c.PowShared = *dec.PowShared
	}
	return nil
}
