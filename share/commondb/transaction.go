package commondb

import (
	"context"
	"database/sql"
	constants "wellbe-common/share/commonsettings/constants"
	errordef "wellbe-common/share/errordef"
	log "wellbe-common/share/log"
)

var txKey = struct{}{}

func DoInTx(ctx *context.Context, db *sql.DB, f func(tx *sql.Tx) (interface{}, *errordef.LogicError)) (interface{}, *errordef.LogicError) {
	logger := log.GetLogger()
	defer logger.Sync()
	tx, ok := getTx(ctx)
	if !ok {
		txBegin, err := db.Begin()
		if  err != nil {
			logger.Error(err.Error())
			return nil, &errordef.LogicError{Msg: err.Error(), Code: constants.LOGIC_ERROR_CODE_DBERROR}
		}
		tx = txBegin
	}
	*ctx = context.WithValue(*ctx, txKey, tx)
	v, err := f(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return v, nil
}

func Beging(ctx *context.Context, db *sql.DB) *errordef.LogicError {
	logger := log.GetLogger()
	defer logger.Sync()
	tx, ok := getTx(ctx)
	if !ok {
		txBegin, err := db.Begin()
		if  err != nil {
			logger.Error(err.Error())
			return  &errordef.LogicError{Msg: err.Error(), Code: constants.LOGIC_ERROR_CODE_DBERROR}
		}
		tx = txBegin
	}
	*ctx = context.WithValue(*ctx, txKey, tx)
	return nil
}

func Rollback(ctx *context.Context) {
	tx, ok := getTx(ctx)
	if !ok {
		return
	}
	tx.Rollback()
}

func Commit(ctx *context.Context) *errordef.LogicError {
	logger := log.GetLogger()
	defer logger.Sync()
	tx, ok := getTx(ctx)
	if !ok {
		return nil
	}
	if err := tx.Commit(); err != nil {
		logger.Error(err.Error())
		return &errordef.LogicError{Msg: err.Error(), Code: constants.LOGIC_ERROR_CODE_DBERROR}
	}
	return nil
}

func getTx(ctx *context.Context) (*sql.Tx, bool) {
	logger := log.GetLogger()
	defer logger.Sync()
	tx, ok := (*ctx).Value(txKey).(*sql.Tx)
	return tx, ok
}