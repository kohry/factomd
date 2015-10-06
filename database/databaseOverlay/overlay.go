// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay

import (
	/*"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/FactomProject/factomd/database"*/

	. "github.com/FactomProject/factomd/common/interfaces"
)

const (
	dbVersion     int = 2
	dbMaxTransCnt     = 20000
	dbMaxTransMem     = 64 * 1024 * 1024 // 64 MB
)

// the "table" prefix
const (

	// Directory Block
	TBL_DB uint8 = iota
	TBL_DB_NUM
	TBL_DB_MR
	TBL_DB_INFO

	// Admin Block
	TBL_AB //4
	TBL_AB_NUM

	TBL_SC
	TBL_SC_NUM

	// Entry Credit Block
	TBL_CB //8
	TBL_CB_NUM
	TBL_CB_MR

	// Entry Chain
	TBL_CHAIN_HASH //11

	// The latest Block MR for chains including special chains
	TBL_CHAIN_HEAD

	// Entry Block
	TBL_EB //13
	TBL_EB_CHAIN_NUM
	TBL_EB_MR

	//Entry
	TBL_ENTRY
)

// the process status in db
const (
	STATUS_IN_QUEUE uint8 = iota
	STATUS_PROCESSED
)

var currentChainType uint32 = 1

var isLookupDB bool = true // to be put in property file

type Overlay struct {
	// leveldb pieces
	DB IDatabase

	lastDirBlkShaCached bool
	lastDirBlkSha       IHash
	lastDirBlkHeight    int64
}

func NewOverlay(db IDatabase) *Overlay {
	answer := new(Overlay)
	answer.DB = db

	answer.lastDirBlkHeight = -1

	return answer
}
